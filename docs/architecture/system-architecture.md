# 系统架构

## 架构总览

tishi 采用经典的 **采集-存储-分析-展示** 四层架构，所有组件通过 PostgreSQL 解耦。

```
┌─────────────────────────────────────────────────────────────────────┐
│                          tishi 系统架构                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   ┌──────────┐     ┌──────────┐     ┌───────────────┐              │
│   │ GitHub   │────▶│Collector │────▶│  PostgreSQL   │              │
│   │ API      │     │ (Go)     │     │               │              │
│   └──────────┘     └──────────┘     │  - projects   │              │
│                                      │  - snapshots  │              │
│   ┌──────────┐                      │  - categories │              │
│   │Scheduler │──── 触发 ──────┐     │  - blog_posts │              │
│   │ (Go Cron)│               │     └───────┬───────┘              │
│   └──────────┘               │             │                       │
│        │                     │             │                       │
│        ├── 触发 ─▶ Collector  │             ▼                       │
│        ├── 触发 ─▶ Analyzer ◀┘     ┌──────────────┐               │
│        └── 触发 ─▶ Content         │  Analyzer    │               │
│                    Generator       │  (Go)        │               │
│                         │          │  - 热度评分    │               │
│                         │          │  - 趋势检测    │               │
│                         │          │  - 分类打标    │               │
│                         │          └──────────────┘               │
│                         ▼                                          │
│                  ┌──────────────┐                                   │
│                  │ Content      │                                   │
│                  │ Generator    │                                   │
│                  │ (Go)         │                                   │
│                  │ - 周报/月报   │                                   │
│                  │ - Markdown   │                                   │
│                  └──────┬───────┘                                   │
│                         │                                          │
│                         ▼                                          │
│   ┌──────────────────────────────────────────┐                     │
│   │              API Server (Go)              │                    │
│   │  GET /api/v1/projects                     │                    │
│   │  GET /api/v1/projects/:id/trends          │                    │
│   │  GET /api/v1/rankings                     │                    │
│   │  GET /api/v1/posts                        │                    │
│   └──────────────────┬───────────────────────┘                     │
│                      │                                             │
│                      ▼                                             │
│   ┌──────────────────────────────────────────┐                     │
│   │           Astro SSG (Frontend)            │                    │
│   │  - 首页排行榜                               │                    │
│   │  - 项目详情 + 趋势图表                       │                    │
│   │  - 博客文章列表                              │                    │
│   │  - RSS Feed                               │                    │
│   └──────────────────────────────────────────┘                     │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

## 模块职责

| 模块 | 进程 | 职责 | 触发方式 |
|------|------|------|----------|
| **Collector** | Go binary | 调用 GitHub API 采集项目数据，写入 PostgreSQL | Scheduler 定时触发 |
| **Analyzer** | Go binary | 计算热度评分，检测趋势变化，分类打标 | Scheduler 在采集完成后触发 |
| **Content Generator** | Go binary | 基于分析结果生成 Markdown 博客文章 | Scheduler 周/月定时触发 |
| **API Server** | Go HTTP server | 提供 RESTful API，供前端消费 | 常驻运行 |
| **Scheduler** | Go cron | 编排定时任务（采集→分析→生成） | 常驻运行 |
| **Web Frontend** | Astro SSG | 静态站点，展示排行榜/图表/博客 | 构建时生成，Nginx 托管 |
| **PostgreSQL** | 数据库 | 持久化所有数据 | 常驻运行 |

## 模块间通信

所有模块通过 **PostgreSQL** 作为数据总线进行间接通信，不使用消息队列：

```
Collector ──写入──▶ PostgreSQL ◀──读取── Analyzer
                        ▲                    │
                        │                    │写入评分/标签
                        │                    ▼
                        ├◀─────── Content Generator（读取分析结果，写入博客）
                        │
                        └◀─────── API Server（只读查询）
```

**设计理由**：
- 个人项目，数据量小（Top 100 × 365天/年 ≈ 36,500 条快照/年），无需消息队列
- PostgreSQL 事务保证数据一致性
- 模块间通过数据库解耦，可独立运行和测试

## 进程模型

生产环境下，所有 Go 模块编译为**单一二进制文件**，通过子命令启动不同模式：

```bash
tishi server     # 启动 API Server + Scheduler（常驻）
tishi collect    # 手动触发一次数据采集
tishi analyze    # 手动触发一次趋势分析
tishi generate   # 手动触发一次内容生成
tishi migrate    # 执行数据库迁移
```

## 关键设计决策

1. **单二进制部署** — 降低运维复杂度，适合个人项目
2. **SSG 而非 SSR** — 数据更新频率低（每日一次），SSG 足够，性能更好
3. **PostgreSQL 单库** — 数据量小，无需分库分表，JSONB 灵活存储元数据
4. **Scheduler 内置** — 使用 Go cron 库，不依赖外部 cron/Airflow

## 相关文档

- [数据流转](data-flow.md) — 数据在各模块间的流转细节
- [技术选型](tech-stack.md) — 各技术组件的选择理由
- [部署拓扑](deployment-topology.md) — Docker Compose 部署方案
