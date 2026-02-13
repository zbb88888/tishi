# tishi 开发进度记录

> 最后更新: 2026-02-13

## 当前状态

- **版本**: v1.0-dev（架构重构中）
- **分支**: main
- **阶段**: Phase 0 完成，Phase 1 准备开始

## 已完成

### v0.1 — Full-Stack MVP (`a0e9fcc`)

**Go 后端**

- [x] cobra CLI: server / collect / analyze / generate / migrate / version
- [x] chi HTTP API: /healthz, /api/v1/{rankings,projects,projects/{id},projects/{id}/trends,posts,posts/{slug},categories}
- [x] GitHub 数据采集器 (go-github/v67 + Token 轮换)
- [x] 多维加权评分分析器 (daily_star 0.30, weekly_star 0.25, fork_ratio 0.15, issue_activity 0.15, recency 0.15)
- [x] 博客自动生成器 (weekly/monthly 两种模板)
- [x] Cron 调度器 (robfig/cron + taskLock 防重叠)
- [x] PostgreSQL 迁移 (4 张表: projects, daily_snapshots, categories, blog_posts)
- [x] Viper 配置 (config file + env vars + defaults)

**Astro 前端**

- [x] SSR 模式 (@astrojs/node adapter)
- [x] 5 个页面: 排行榜 / 分类 / 博客列表 / 博客详情[slug] / 项目详情[id]
- [x] TrendChart 组件 (Chart.js CDN, 7/30/90 天切换)
- [x] Tailwind CSS + Typography 插件
- [x] TypeScript API 客户端

**基础设施**

- [x] Docker 多阶段构建 (distroless)
- [x] Docker Compose (Go + PostgreSQL + Nginx)
- [x] GitHub Actions CI (5 jobs: go-lint → go-test → go-build → frontend-build → docker-build)
- [x] Makefile (build/test/lint/dev/docker 等 target)
- [x] 23 篇设计文档

### v0.2 — Quality & Observability (`5d429de`)

- [x] Prometheus /metrics endpoint (requests_total, duration_seconds, requests_active)
- [x] golangci-lint v2 配置 + 修复 12 个 lint issue
- [x] config_test.go (Load defaults, env overrides)
- [x] metrics_test.go (normalisePath, middleware, statusWriter)
- [x] 404 自定义错误页
- [x] Pagination 分页组件 (博客列表已集成)

## 明天继续 (v0.3 建议方向)

> ⚠️ 以下 v0.3 方向已作废，被 v1.0 新计划取代。保留供参考。

### 高优先级

- [ ] ~~**提升测试覆盖率** — 目标 60%+, 重点: collector (mock GitHub API), server handlers (httptest), db/migrate~~
- [ ] ~~**E2E / 集成测试** — httptest 测试完整 API 链路~~
- [ ] ~~**sqlc codegen** — 运行 `sqlc generate`, 替换 handler 中的 raw SQL~~

### 中优先级

- [ ] ~~**Redis 缓存层** — 排行榜 API 缓存 (TTL 5min), 减轻 DB 压力~~
- [ ] ~~**搜索 / 过滤** — 排行榜支持按语言、分类筛选; 项目名模糊搜索~~
- [ ] ~~**RSS Feed** — /feed.xml 输出博客文章 RSS~~
- [ ] ~~**SEO** — sitemap.xml, Open Graph meta tags, 结构化数据~~

### 低优先级

- [ ] ~~**Grafana Dashboard** — 对接 Prometheus metrics 的预置 dashboard JSON~~
- [ ] ~~**Rate Limit** — API 接口限流 (per-IP)~~
- [ ] ~~**Webhook 通知** — Star 大幅增长时发送通知 (Slack/Telegram)~~
- [ ] ~~**Dark Mode** — Tailwind dark 模式支持~~

---

## v1.0 新计划 — AI 项目深度分析平台

> 决策日期: 2026-02-13
> 架构方向: 三阶段完全解耦 + SSG + LLM 中文分析

### 产品定位

对标 <https://github.com/trending> ，做 AI 领域垂直增强版，三个核心差异点：

1. **中文深度项目解读** — LLM 自动生成面向中文开发者的完整项目报告（定位/功能/优势/对比/场景），不用翻墙看英文 README
2. **AI 垂直精选 + 分类导航** — 只收录 AI 相关项目，12 个 AI 子方向分类（LLM/Agent/RAG/图像生成/向量数据库/MLOps 等）
3. **趋势追踪 + 历史可回溯** — 每日快照持久化，Star 增长曲线、排名变动追踪，辅助技术选型

### 架构决策

三个 Stage 完全独立，可在不同机器运行，通过 JSON 文件 + Git 仓库交换数据：

```
Stage 1 (Machine A)          Stage 2 (Machine B)         Stage 3 (Machine C)
采集 + LLM 翻译/分析    →     Astro SSG 静态构建     →     Nginx / CDN 发布
输出 data/ JSON + git push    git pull + npm build         git pull / rsync
```

- 去掉 PostgreSQL，数据层改为 JSON 文件 + Git 版本控制
- 去掉 API Server + 内置 Scheduler，Go 二进制变为纯 CLI 工具
- Astro 从 SSR 切换为 SSG，输出纯静态 HTML

### 实施步骤

#### Phase 0: 数据契约 ✅

- [x] 设计 `data/` 目录结构（projects/, snapshots/, rankings/, posts/, schemas/）
- [x] 定义 JSON Schema（project/snapshot/ranking/post 四个 schema 文件）
- [x] 创建 categories.json 初始分类数据（12 个 AI 子分类 + 关键词映射）
- [x] 创建 meta.json 元信息

#### Phase 1: Stage 1 — 采集 + 分析

- [ ] `internal/collector/trending.go` — colly 爬取 GitHub Trending HTML
- [ ] `internal/collector/ai_filter.go` — AI 相关性多层过滤（topic + description + 排除规则）
- [ ] GitHub API 补充详情 + README 获取
- [ ] `internal/llm/` package — DeepSeek/Qwen 客户端（OpenAI-compatible, BaseURL 可切换）
- [ ] LLM Prompt 设计 — 输出结构化中文项目分析 JSON
- [ ] 草稿审核流程（status: draft → published, `tishi review` CLI）
- [ ] Snapshot 追加 + 评分排名（读写 JSON 文件，去除 PG 依赖）
- [ ] 自动分类打标（topic + description 关键词匹配）
- [ ] `tishi push` — git add/commit/push 数据变更

#### Phase 2: Stage 2 — SSG 构建

- [ ] Astro 切换 `output: 'static'`，移除 @astrojs/node
- [ ] 数据读取层重写（fetch API → 读本地 JSON 文件）
- [ ] 所有动态路由添加 `getStaticPaths()`
- [ ] 排行榜页面：中文简介 + 今日新增 Star
- [ ] 项目详情页重构：完整分析报告（定位/功能/优势/技术架构/对比/场景/生态）
- [ ] 分类页面激活（12 个 AI 子分类）

#### Phase 3: Stage 3 — 发布

- [ ] Nginx 简化为纯静态服务
- [ ] 部署脚本（git pull dist/ 或 rsync）

#### Phase 4: 清理

- [ ] 移除 internal/server/, internal/scheduler/, internal/db/
- [ ] go.mod 移除 pgx, chi, prometheus, golang-migrate
- [ ] go.mod 新增 colly, go-openai
- [ ] CLI 子命令调整：scrape / analyze / llm-analyze / review / generate / push
- [ ] 更新 Makefile, Dockerfile, docker-compose.yml, CI

### TODO（后续迭代）

- [ ] **Topic 订阅** — 支持用户自定义 topic（如 mcp, function-calling, code-generation），基于 topic 从 GitHub Search API 抓取内容，与 Trending 数据合并

## 文件清单

```
cmd/tishi/main.go                    # CLI 入口
internal/
├── cmd/                             # 7 个子命令
├── config/config.go (+test)         # Viper 配置
├── db/db.go, migrate.go             # pgxpool + 迁移
│   ├── migrations/ (8 SQL files)
│   └── queries/ (4 sqlc files)
├── collector/collector.go (+test)   # GitHub 采集
│   └── token_rotator.go (+test)
├── analyzer/analyzer.go (+test)     # 评分排名
├── content/generator.go (+test)     # 博客生成
├── server/server.go (+test)         # HTTP API
│   ├── middleware.go                # zap logger
│   └── metrics.go (+test)          # Prometheus
└── scheduler/scheduler.go           # Cron 调度
web/
├── src/pages/ (6 pages)
├── src/components/ (TrendChart, Pagination)
├── src/layouts/BaseLayout.astro
├── src/lib/api.ts
└── src/styles/global.css
```
