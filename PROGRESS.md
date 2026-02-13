# tishi 开发进度记录

> 最后更新: 2026-02-13

## 当前状态

- **版本**: v1.0-dev
- **分支**: main
- **阶段**: Phase 0-4 全部完成，v1.0 核心功能就绪

## 已完成

### v1.0 — AI 项目深度分析平台

#### Phase 0: 数据契约 ✅ (`0a77d5d`)

- [x] 设计 `data/` 目录结构（projects/, snapshots/, rankings/, posts/, schemas/）
- [x] 定义 JSON Schema（project/snapshot/ranking/post 四个 schema）
- [x] 创建 categories.json 初始分类数据（12 个 AI 子分类 + 关键词映射）
- [x] 创建 meta.json 元信息
- [x] 重写全部 20+ 设计文档，对齐 v1.0 架构

#### Phase 1: Stage 1 — 采集 + 分析 CLI ✅ (`741a2ec` + `05bad8c`)

**Go CLI Pipeline (6 个命令)**

- [x] `tishi scrape` — Colly 爬取 GitHub Trending HTML + AI 关键词过滤 + GitHub API 补充详情
- [x] `tishi score` — 多维加权评分（daily_stars 0.35, weekly_stars 0.25, forks_rate 0.15, issue_activity 0.10）+ 排行榜生成
- [x] `tishi analyze` — DeepSeek/Qwen LLM 中文项目分析（OpenAI 兼容 API）
- [x] `tishi generate` — 博客文章生成（周报 + Spotlight 深度分析）
- [x] `tishi review` — LLM 分析草稿审核（--approve / --reject）
- [x] `tishi push` — Git add/commit/push data/ 变更

**核心包**

- [x] `internal/datastore` — JSON 文件 CRUD + 原子写入（12 tests）
- [x] `internal/scraper` — Colly 爬虫 + AI 过滤 + Token 轮换（11 tests）
- [x] `internal/scorer` — 评分 + 排名（6 tests）
- [x] `internal/llm` — LLM 客户端 + 分析编排（8 tests）
- [x] `internal/generator` — 博客生成（7 tests）
- [x] `internal/config` — Viper 配置（4 tests来自 clean 后的 test）

**测试**: 44 个单元测试全绿，go vet clean

#### Phase 2: Stage 2 — Astro SSG 前端 ✅ (`9c3eb29`)

- [x] `output: 'static'`，移除 `@astrojs/node` adapter
- [x] `web/src/lib/data.ts` — 构建时读取 data/*.json（替代 HTTP API 客户端）
- [x] 首页（`/`）— 排行榜表格 + 分类导航 + 最新博客
- [x] 项目详情页（`/projects/{id}`）— 完整 LLM 分析报告 + 趋势图
- [x] 分类总览（`/categories/`）— 12 个 AI 子分类卡片
- [x] 分类详情（`/categories/{slug}`）— 分类下项目列表
- [x] 博客列表（`/blog/`）— 已发布文章
- [x] 博客详情（`/blog/{slug}`）— Markdown 渲染（marked）
- [x] TrendChart 组件 — 构建时注入快照数据 + CDN Chart.js
- [x] 验证: 18 页 682ms 构建通过

#### Phase 4: Legacy 清理 ✅ (`98c6d0f`)

- [x] 删除 6 个 legacy 包: analyzer, collector, content, scheduler, server, db
- [x] 删除 legacy CLI 命令: server, collect, migrate
- [x] 清理 config.go: 移除所有 legacy 结构体/字段/默认值/绑定
- [x] go mod tidy: 移除 pgx, chi, cors, prometheus, migrate, cron
- [x] 重写 Makefile, Dockerfile, docker-compose.yml, nginx.conf, config.yaml.example
- [x] -3147 行代码，build + vet + 44 tests pass

#### Phase 3: Stage 3 — 部署 ✅

- [x] `deploy/nginx/nginx.conf` — 纯静态站点配置（Astro SSG 产物）
- [x] `deploy/deploy.sh` — 自动部署脚本（git pull + npm build + rsync）
- [x] GitHub Actions CI 更新（Go 1.24 + Astro SSG build）

### 历史版本

#### v0.1 — Full-Stack MVP (`a0e9fcc`)

- cobra CLI + chi HTTP API + PostgreSQL + Astro SSR + Docker + CI + 23 篇设计文档

#### v0.2 — Quality & Observability (`5d429de`)

- Prometheus metrics + golangci-lint v2 + 单元测试 + 404 页面 + Pagination 组件

---

## 技术栈

| 层级 | 技术 | 用途 |
|------|------|------|
| Go CLI | Go 1.24, Cobra, Viper, Zap | Stage 1 数据流水线 |
| 爬虫 | Colly v2 | GitHub Trending HTML 解析 |
| LLM | DeepSeek / Qwen (go-openai) | 中文项目分析 |
| 前端 | Astro 4 SSG + Tailwind + Chart.js | Stage 2 静态站点 |
| 部署 | Nginx + Git | Stage 3 静态托管 |

## 文件结构

```
cmd/tishi/main.go                    # CLI 入口
internal/
├── cmd/                             # 8 个子命令 (scrape/score/analyze/generate/review/push/version/root)
├── config/config.go (+test)         # Viper 配置
├── datastore/                       # JSON 文件 CRUD + 原子写入
│   ├── models.go                    # 数据模型 (Project/Snapshot/Ranking/Post/Category)
│   └── store.go (+test)
├── scraper/                         # GitHub Trending 爬虫
│   ├── scraper.go                   # Colly 爬虫主流程
│   ├── filter.go (+test)            # AI 关键词过滤
│   ├── enricher.go                  # GitHub API 补充
│   └── token_rotator.go (+test)     # Token 轮换
├── scorer/scorer.go (+test)         # 评分 + 排名
├── llm/                             # LLM 客户端
│   ├── client.go                    # OpenAI 兼容 (DeepSeek/Qwen)
│   └── analyzer.go (+test)          # 分析编排
└── generator/generator.go (+test)   # 博客生成 (weekly + spotlight)
web/
├── astro.config.mjs                 # SSG 配置
├── src/pages/                       # 7 个页面 (index/projects/categories/blog/404)
├── src/components/TrendChart.astro  # Chart.js 趋势图
├── src/layouts/BaseLayout.astro     # 公共布局
├── src/lib/data.ts                  # 数据加载层 (读 JSON 文件)
└── src/styles/global.css            # Tailwind + 组件样式
data/                                # JSON 数据目录 (Git 同步)
├── projects/                        # 项目 JSON
├── snapshots/                       # 每日快照 JSONL
├── rankings/                        # 每日排行榜
├── posts/                           # 博客文章
├── categories.json                  # 12 个 AI 分类
└── meta.json                        # 元信息
```
