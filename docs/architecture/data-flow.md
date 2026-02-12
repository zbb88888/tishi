# 数据流转

## 概述

tishi 的数据从 GitHub API 采集，经过清洗、存储、分析、生成，最终以静态页面形式展示给用户。本文档描述完整的数据流转链路。

## 全链路数据流

```
                    ┌─── 每日 00:00 UTC ───┐
                    │                       │
                    ▼                       │
┌──────────┐   ┌──────────┐   ┌─────────────────┐
│ GitHub   │──▶│Collector │──▶│   PostgreSQL    │
│ Search   │   │          │   │                 │
│ API      │   │ 1.搜索    │   │ projects 表     │
│          │   │ 2.过滤    │   │ daily_snapshots │
│ GraphQL  │   │ 3.清洗    │   │                 │
│ API      │   │ 4.写入    │   └────────┬────────┘
└──────────┘   └──────────┘            │
                                        │ 采集完成后触发
                                        ▼
                               ┌──────────────┐
                               │   Analyzer   │
                               │              │
                               │ 1.计算热度评分 │
                               │ 2.排名变动检测 │
                               │ 3.趋势异常检测 │
                               │ 4.分类打标     │
                               │ 5.写回数据库   │
                               └──────┬───────┘
                                      │
                          ┌───────────┴───────────┐
                          │ 周日/月末触发           │ 实时可用
                          ▼                       ▼
                 ┌────────────────┐     ┌──────────────┐
                 │Content Generator│    │  API Server   │
                 │                │     │              │
                 │ 1.查询本周数据   │     │ JSON API     │
                 │ 2.渲染模板      │     │ 供 Astro     │
                 │ 3.生成 Markdown │     │ 构建时消费    │
                 │ 4.存入 DB/文件  │     └──────┬───────┘
                 └────────────────┘            │
                                               │ Astro build
                                               ▼
                                     ┌──────────────────┐
                                     │  Astro SSG 构建   │
                                     │                  │
                                     │ 1.调用 API 获取数据│
                                     │ 2.生成静态 HTML   │
                                     │ 3.输出 dist/      │
                                     └────────┬─────────┘
                                              │
                                              ▼
                                     ┌──────────────────┐
                                     │  Nginx / CDN     │
                                     │  静态文件托管      │
                                     │  用户浏览器访问    │
                                     └──────────────────┘
```

## 阶段详解

### 阶段 1：数据采集（Collector）

**触发**：Scheduler 每日 00:00 UTC

**输入**：AI 领域关键词库（见 [种子数据](../data/seed-data.md)）

**处理流程**：

1. 使用 GitHub Search API 按关键词搜索项目
2. 按 Star 数排序，取 Top N（N > 100 以备去重后仍够 100）
3. 对每个项目调用 GraphQL API 获取详细信息：
   - 基本信息（name, description, language, license, created_at）
   - 指标数据（stargazers_count, forks_count, open_issues_count, watchers_count）
   - 最近活跃度（last_push, last_commit, contributor_count）
4. 数据清洗：去重、字段校验、异常值过滤
5. Upsert 到 `projects` 表，Insert 到 `daily_snapshots` 表

**输出**：PostgreSQL 中更新的项目数据 + 当日快照

### 阶段 2：趋势分析（Analyzer）

**触发**：Collector 完成后由 Scheduler 串行触发

**输入**：`projects` + `daily_snapshots` 表数据

**处理流程**：

1. **热度评分计算**：
   - 日增 Star × 权重 + 周增 Star × 权重 + Fork 率 × 权重 + Issue 响应速度 × 权重
   - 输出标准化评分 0-100
2. **排名计算**：按评分排序，生成 Top 100 排行榜
3. **变动检测**：与前一日排名对比，标记 ↑↓ 变动幅度
4. **异常检测**：识别 Star 单日暴涨（>200%均值）的项目
5. **分类打标**：基于关键词 + README 内容，映射到预定义分类树

**输出**：更新 `projects.score`, `projects.rank`, `project_categories` 关联

### 阶段 3：内容生成（Content Generator）

**触发**：周日 06:00 UTC（周报）、每月 1 日 06:00 UTC（月报）

**输入**：分析后的项目数据 + 历史快照

**处理流程**：

1. 查询时间范围内的排名变动、新入榜项目、Star 增长最快项目
2. 套用 Markdown 模板生成文章
3. 插入图表占位符（前端渲染时替换）
4. 存入 `blog_posts` 表

**输出**：Markdown 格式的博客文章记录

### 阶段 4：API 服务 + 前端构建

**API Server** 提供以下核心数据端点（详见 [RESTful API](../api/restful-api.md)）：

- `GET /api/v1/rankings` — 当前 Top 100 排行榜
- `GET /api/v1/projects/:id` — 项目详情
- `GET /api/v1/projects/:id/trends` — 项目趋势数据（时序）
- `GET /api/v1/posts` — 博客文章列表

**Astro SSG** 在构建时调用上述 API，生成静态 HTML 页面。

## 数据时效性

| 数据类型 | 更新频率 | 延迟 |
|----------|----------|------|
| 项目基本信息 | 每日 | < 1小时（采集耗时） |
| 排行榜 | 每日 | 采集 + 分析完成后 |
| 趋势图表 | 每日 | 同上 |
| 周报博客 | 每周日 | 固定时间生成 |
| 静态页面 | 每日 | 需触发 Astro rebuild |

## 数据量估算

| 数据 | 日增 | 年增 | 5年累计 |
|------|------|------|---------|
| projects 记录 | ~0（缓慢增长） | ~200 | ~1,000 |
| daily_snapshots | 100-200 | 36,500-73,000 | ~365,000 |
| blog_posts | ~0.3（周报+月报） | ~64 | ~320 |

**结论**：数据量极小，单 PostgreSQL 实例轻松承载，无需分库分表或时序数据库。
