# tishi — GitHub AI 趋势追踪博客

> 自动追踪 GitHub 上 AI 相关 Top 100 热门开源项目的发展趋势，提供排行榜、趋势分析与博客内容。

## 项目简介

tishi（提示）是一个面向 AI 开源生态的趋势追踪平台。通过定时采集 GitHub 数据，分析 AI 领域 Top 100 项目的 Star 增长、活跃度变化等指标，自动生成排行榜和趋势博客文章。

### 核心功能

- **趋势排行榜** — AI 项目 Top 100 实时排名，支持多维度排序
- **Star 增长曲线** — 可视化每个项目的 Star/Fork/Issue 历史趋势
- **热度评分** — 基于多维指标的加权评分模型
- **自动博客** — 周报/月报自动生成，追踪 AI 生态动态
- **项目分类** — LLM / Agent / RAG / Diffusion / MLOps 等细分领域标签

## 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| 后端 | Go | 数据采集、分析引擎、API Server |
| 前端 | Astro | SSG 静态站点生成，内容驱动 |
| 数据库 | PostgreSQL | 项目元数据 + 时序快照存储 |
| 部署 | Docker Compose | 单机一键部署 |

## 快速开始

```bash
# 克隆项目
git clone https://github.com/zbb88888/tishi.git
cd tishi

# 启动服务（需要 Docker + Docker Compose）
docker-compose up -d

# 查看服务状态
docker-compose ps
```

### 本地开发

```bash
# 前置条件: Go 1.22+, Node 20+, PostgreSQL 16+

# 1. 安装后端依赖
go mod download

# 2. 复制配置文件并修改数据库连接
cp config.yaml.example config.yaml
cp .env.example .env

# 3. 运行数据库迁移
make migrate-up

# 4. 启动后端 (带热重载)
make dev

# 5. 安装前端依赖并启动开发服务器
cd web && npm install && npm run dev
```

### 常用命令

```bash
make build          # 编译后端
make test           # 运行单元测试
make lint           # 代码检查 (golangci-lint)
make collect        # 手动采集 GitHub 数据
make analyze        # 手动运行趋势分析
make docker-up      # Docker Compose 一键启动
```

## 项目结构

```
tishi/
├── cmd/tishi/           # CLI 入口
│   └── main.go
├── internal/
│   ├── cmd/             # cobra 子命令 (server/collect/analyze/generate/migrate/version)
│   ├── config/          # viper 配置管理
│   ├── db/              # pgxpool 数据库 + golang-migrate 迁移
│   │   ├── migrations/  # SQL 迁移文件 (4 up + 4 down)
│   │   └── queries/     # sqlc 查询定义
│   ├── collector/       # GitHub API 数据采集 + Token 轮换
│   ├── analyzer/        # 多维加权评分 + 排名计算
│   ├── content/         # 周报/月报自动生成 (Go template)
│   ├── server/          # chi HTTP API + middleware
│   └── scheduler/       # robfig/cron 定时调度
├── web/                 # Astro 4.x 前端
│   ├── src/pages/       # 排行榜 / 分类 / 博客 / 项目详情
│   ├── src/components/  # TrendChart (Chart.js) 等
│   └── src/lib/         # TypeScript API 客户端
├── deploy/nginx/        # Nginx 反向代理配置
├── docs/                # 23 篇设计文档
├── .github/workflows/   # CI (Go lint/test/build + Frontend build + Docker)
├── Dockerfile           # 多阶段构建 (distroless)
├── docker-compose.yml   # Go + PostgreSQL + Nginx
└── Makefile             # 开发/构建/部署任务
```

> 详细开发环境搭建请参考 [开发指南](docs/guides/development.md)

## 文档导航

### 总览

- [项目概述](docs/overview.md) — 愿景、目标、核心功能

### 架构设计

- [系统架构](docs/architecture/system-architecture.md) — 整体架构与模块关系
- [数据流转](docs/architecture/data-flow.md) — 数据全链路流转
- [技术选型](docs/architecture/tech-stack.md) — 技术栈选择与理由
- [部署拓扑](docs/architecture/deployment-topology.md) — Docker Compose 部署架构

### 模块设计

- [数据采集](docs/design/collector.md) — GitHub 数据采集策略
- [数据存储](docs/design/storage.md) — PostgreSQL 数据模型
- [趋势分析](docs/design/analyzer.md) — 热度评分与趋势检测
- [内容生成](docs/design/content-generator.md) — 博客文章自动生成
- [API 服务](docs/design/api-server.md) — RESTful API 设计
- [定时调度](docs/design/scheduler.md) — 任务编排与调度
- [前端展示](docs/design/web-frontend.md) — Astro 页面设计

### API 规范

- [RESTful API](docs/api/restful-api.md) — 接口定义
- [GitHub 集成](docs/api/github-integration.md) — GitHub API 调用策略

### 数据定义

- [数据库 Schema](docs/data/schema.md) — 表结构定义
- [数据字典](docs/data/data-dictionary.md) — 字段说明
- [种子数据](docs/data/seed-data.md) — AI 领域关键词库

### 操作指南

- [开发指南](docs/guides/development.md) — 本地开发环境搭建
- [部署指南](docs/guides/deployment.md) — 生产环境部署
- [配置说明](docs/guides/configuration.md) — 配置项参考
- [贡献指南](docs/guides/contributing.md) — 代码规范与协作流程

### 决策记录

- [ADR 模板](docs/decisions/001-template.md) — 架构决策记录模板

### 规划

- [版本路线图](docs/roadmap.md) — 版本规划与里程碑

## License

MIT