# 系统架构

## 架构总览

tishi v1.0 采用 **三阶段完全解耦** 架构，通过 JSON 文件 + Git 仓库交换数据，无共享数据库。

```
Stage 1: 采集+分析 (Machine A)         Stage 2: SSG 构建 (Machine B)         Stage 3: 发布 (Machine C)
┌──────────────────────────┐           ┌──────────────────────┐           ┌─────────────────┐
│  GitHub Trending HTML    │           │  git pull data/      │           │  Nginx / CDN    │
│         ↓                │           │       ↓              │           │  serve dist/    │
│  Colly 爬取 Trending     │           │  读取 JSON 数据       │           │       ↓         │
│         ↓                │           │       ↓              │           │  用户浏览器      │
│  AI 项目过滤 (keywords)  │  git push │  Astro SSG build     │  rsync/  │                 │
│         ↓                │ ────────→ │       ↓              │ ───────→ │                 │
│  GitHub API 补充数据      │  data/    │  dist/ 静态 HTML     │  dist/   │                 │
│         ↓                │           └──────────────────────┘           └─────────────────┘
│  LLM 中文深度分析         │
│  (DeepSeek / Qwen)       │
│         ↓                │
│  data/ JSON 输出          │
└──────────────────────────┘
```

## 模块职责

| 模块 | 阶段 | 职责 | 输入 | 输出 |
|------|------|------|------|------|
| **Scraper** | Stage 1 | 爬取 GitHub Trending HTML，提取项目列表 | Trending 页面 | 候选项目列表 |
| **Filter** | Stage 1 | 基于 12 类 AI 关键词过滤非 AI 项目 | 候选列表 + categories.json | AI 项目列表 |
| **Enricher** | Stage 1 | 调用 GitHub API 补充 README/详细指标 | AI 项目列表 | 完整项目数据 |
| **LLM Analyzer** | Stage 1 | 调用 DeepSeek/Qwen 生成中文项目报告 | README + 项目元数据 | analysis 字段 (JSON) |
| **Scorer** | Stage 1 | 多维加权评分 + 排名计算 | 项目数据 + 快照历史 | ranking JSON |
| **Astro SSG** | Stage 2 | 读取 data/ JSON，生成静态 HTML 页面 | data/*.json | dist/ |
| **Nginx/CDN** | Stage 3 | 托管 dist/ 静态文件 | dist/ | HTTP 响应 |

## 数据交换

模块间通过 **JSON 文件 + Git** 交换数据，不使用数据库或消息队列：

```
Scraper → Filter → Enricher → LLM Analyzer → Scorer
    ↓         ↓         ↓            ↓            ↓
    └─────────┴─────────┴────────────┴────────────┘
                         ↓
                   data/ 目录 (JSON)
                         ↓
                      git push
                         ↓
                   Stage 2 git pull
                         ↓
                   Astro SSG build
```

**数据目录结构**：

```
data/
├── projects/          # 每个项目一个 JSON 文件 (owner__repo.json)
├── snapshots/         # 每日快照 JSONL (YYYY-MM-DD.jsonl)
├── rankings/          # 每日排行 JSON (YYYY-MM-DD.json)
├── posts/             # 博客文章 (slug.json)
├── schemas/           # JSON Schema 定义
├── categories.json    # 12 个 AI 分类 + 关键词映射
└── meta.json          # 版本和元信息
```

**设计理由**：

- 三阶段可运行在不同机器上，通过 Git 同步，无需网络直连
- JSON 文件人类可读、版本可追溯、易于调试
- 无状态构建：任何阶段失败可独立重跑，天然幂等

## 进程模型

单一 Go 二进制，通过子命令运行 Stage 1 各步骤：

```bash
tishi scrape          # 爬取 Trending + AI 过滤 + GitHub API 补充
tishi analyze         # LLM 深度分析（可选 --dry-run 跳过 API 调用）
tishi score           # 评分排名 + 生成 ranking JSON
tishi generate        # 生成博客文章 JSON
tishi push            # git add + commit + push data/
tishi version         # 打印版本号
```

Stage 2 和 Stage 3 不需要 Go 二进制，直接使用 npm/Nginx。

## 关键设计决策

1. **单二进制部署** — 降低运维复杂度，适合个人项目
2. **SSG 而非 SSR** — 数据更新频率低（每日一次），SSG 足够，性能更好
3. **PostgreSQL 单库** — 数据量小，无需分库分表，JSONB 灵活存储元数据
4. **Scheduler 内置** — 使用 Go cron 库，不依赖外部 cron/Airflow

## 相关文档

- [数据流转](data-flow.md) — 数据在各模块间的流转细节
- [技术选型](tech-stack.md) — 各技术组件的选择理由
- [部署拓扑](deployment-topology.md) — Docker Compose 部署方案
