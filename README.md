# tishi — AI 开源项目深度分析平台

> 面向中文开发者，自动追踪 GitHub Trending 上的 AI 热门项目，通过 LLM 生成深度中文分析报告。

## 项目简介

tishi（提示）是一个面向中文开发者的 AI 开源项目深度分析平台。每日抓取 GitHub Trending 页面，过滤 AI 相关项目，通过 DeepSeek/Qwen 大模型自动生成**中文深度分析报告**（定位、功能、优势、技术栈、适用场景、同类对比），并以静态站点形式展示排行榜、趋势曲线和博客文章。

### 核心特性

- **GitHub Trending 追踪** — 每日抓取 Trending 页面，聚焦 AI 领域项目
- **LLM 深度分析** — DeepSeek/Qwen 自动生成 7 维度中文分析报告
- **热度排行榜** — 多维加权评分（日增 Star、周增 Star、Fork 率、Issue 活跃度、更新频率）
- **趋势可视化** — Star/Fork/Issue 历史趋势曲线
- **自动博客** — 周报/月报/新项目速递，自动生成
- **12 个 AI 分类** — LLM / Agent / RAG / 图像生成 / MLOps / 向量数据库 / 框架 / 工具 / 多模态 / 语音 / 强化学习

### 与 GitHub Trending 的区别

| 特性 | GitHub Trending | tishi |
|------|----------------|-------|
| 语言 | 英文 | **中文** |
| 范围 | 全部项目 | **AI 专精** |
| 分析深度 | 仅名称+描述 | **LLM 7 维度深度分析** |
| 历史数据 | 无 | **完整趋势曲线** |
| 分类标签 | 无 | **12 个 AI 细分领域** |
| 内容形式 | 列表 | **排行榜 + 博客 + 报告** |

## 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| 后端 CLI | Go 1.23 | Cobra CLI，Pipeline 式数据处理 |
| 前端 | Astro 4.x | 纯 SSG 静态站点，直接读取 JSON |
| LLM | DeepSeek / Qwen | OpenAI 兼容 API，中文分析 |
| 数据抓取 | Colly | GitHub Trending HTML 解析 |
| 数据存储 | JSON + Git | 文件存储 + Git 版本管理和跨机同步 |
| 部署 | Nginx | 静态文件托管 |

## 架构概览

三阶段独立 Pipeline，通过 Git 仓库交换数据：

```
┌─────────────────────────────────────────────────────────┐
│  Stage 1: 数据采集 + 分析 (Machine A, Go CLI)            │
│  tishi scrape → analyze → score → generate → push       │
│  输出: data/ 目录 (JSON 文件)                            │
└──────────────────────┬──────────────────────────────────┘
                       │ git push / git pull
┌──────────────────────▼──────────────────────────────────┐
│  Stage 2: 静态站点构建 (Machine B, Node.js)              │
│  git pull → pnpm build (Astro SSG)                      │
│  输出: web/dist/ (HTML/CSS/JS)                           │
└──────────────────────┬──────────────────────────────────┘
                       │ rsync / deploy
┌──────────────────────▼──────────────────────────────────┐
│  Stage 3: 发布 (Machine C, Nginx)                        │
│  Nginx serve web/dist/                                   │
└─────────────────────────────────────────────────────────┘
```

## 快速开始

```bash
# 克隆项目
git clone https://github.com/zbb88888/tishi.git
cd tishi

# 配置
cp config.yaml.example config.yaml
# 编辑 config.yaml，填入 LLM API Key 和 GitHub Token

# 构建
make build

# 运行 Pipeline
./bin/tishi scrape    # 抓取 Trending + 过滤 AI 项目
./bin/tishi analyze   # LLM 深度分析
./bin/tishi score     # 评分排名
./bin/tishi generate  # 生成周报

# 构建前端
cd web && pnpm install && pnpm build
# 用 Nginx 托管 web/dist/
```

### 环境要求

| 工具 | 版本 | 用途 |
|------|------|------|
| Go | ≥ 1.23 | 后端 CLI |
| Node.js | ≥ 20 LTS | Astro 构建 |
| pnpm | ≥ 9.x | 包管理 |

> 详见 [开发指南](docs/guides/development.md)

### 常用命令

```bash
make build          # 编译 Go CLI
make build-web      # 构建 Astro 前端
make test           # 单元测试
make lint           # 代码检查
make pipeline       # 运行完整 Pipeline
```

## 项目结构

```
tishi/
├── cmd/tishi/           # CLI 入口
├── internal/
│   ├── cmd/             # cobra 子命令 (scrape/analyze/score/generate/push/review/version)
│   ├── config/          # viper 配置管理
│   ├── scraper/         # Trending HTML 抓取 + AI 过滤 + API enrichment
│   ├── llm/             # DeepSeek/Qwen 中文分析
│   ├── scorer/          # 多维加权评分 + 排名
│   ├── content/         # 周报/月报生成 (Go template)
│   └── datastore/       # JSON 文件存储
├── data/                # JSON 数据 (Git 管理)
│   ├── projects/        # 项目档案 (*.json)
│   ├── snapshots/       # 每日快照 (*.jsonl)
│   ├── rankings/        # 每日排行榜 (*.json)
│   ├── posts/           # 博客文章 (*.json)
│   ├── schemas/         # JSON Schema 定义
│   └── categories.json  # 12 个 AI 分类
├── web/                 # Astro 4.x 前端 (纯 SSG)
│   └── src/pages/       # 排行榜 / 分类 / 博客 / 项目分析报告
├── docs/                # 项目文档
├── deploy/nginx/        # Nginx 配置
├── Makefile
└── config.yaml.example
```

## 文档导航

### 总览

- [项目概述](docs/overview.md) — 愿景、目标、对比 GitHub Trending
- [版本路线图](docs/roadmap.md) — 版本规划与里程碑

### 架构设计

- [系统架构](docs/architecture/system-architecture.md) — 三阶段 Pipeline 架构
- [数据流转](docs/architecture/data-flow.md) — 数据全链路流转
- [技术选型](docs/architecture/tech-stack.md) — 技术栈选择与理由
- [部署拓扑](docs/architecture/deployment-topology.md) — 三机/单机部署方案

### 模块设计

- [Scraper](docs/design/collector.md) — Trending 页面抓取 + AI 过滤
- [LLM 分析器](docs/design/llm-analyzer.md) — DeepSeek/Qwen 深度分析
- [Scorer](docs/design/analyzer.md) — 热度评分与排名
- [内容生成](docs/design/content-generator.md) — 博客文章自动生成
- [存储](docs/design/storage.md) — JSON + Git 数据存储
- [前端](docs/design/web-frontend.md) — Astro SSG 页面设计
- [~~API Server~~](docs/design/api-server.md) — v1.0 已废弃
- [~~Scheduler~~](docs/design/scheduler.md) — v1.0 已废弃，改用系统 crontab

### API / 集成

- [~~RESTful API~~](docs/api/restful-api.md) — v1.0 已废弃
- [GitHub 数据获取](docs/api/github-integration.md) — Trending 抓取 + REST API

### 数据定义

- [数据 Schema](docs/data/schema.md) — JSON Schema 定义
- [数据字典](docs/data/data-dictionary.md) — 字段说明
- [种子数据](docs/data/seed-data.md) — AI 领域关键词库

### 操作指南

- [配置说明](docs/guides/configuration.md) — 配置项参考
- [开发指南](docs/guides/development.md) — 本地开发环境搭建
- [部署指南](docs/guides/deployment.md) — 三机/单机部署
- [贡献指南](docs/guides/contributing.md) — 代码规范与协作流程

## License

MIT
