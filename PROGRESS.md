# tishi 开发进度记录

> 最后更新: 2026-02-13

## 当前状态

- **分支**: main (`5d429de`)
- **Go 测试**: 全部通过 (5 个包, race detector)
- **golangci-lint**: 0 issues (v2.7.1 配置)
- **Astro 构建**: SSR 模式构建通过
- **测试覆盖率**: 30.1% (仅 internal/ 包)

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

### 高优先级
- [ ] **提升测试覆盖率** — 目标 60%+, 重点: collector (mock GitHub API), server handlers (httptest), db/migrate
- [ ] **E2E / 集成测试** — httptest 测试完整 API 链路
- [ ] **sqlc codegen** — 运行 `sqlc generate`, 替换 handler 中的 raw SQL

### 中优先级
- [ ] **Redis 缓存层** — 排行榜 API 缓存 (TTL 5min), 减轻 DB 压力
- [ ] **搜索 / 过滤** — 排行榜支持按语言、分类筛选; 项目名模糊搜索
- [ ] **RSS Feed** — /feed.xml 输出博客文章 RSS
- [ ] **SEO** — sitemap.xml, Open Graph meta tags, 结构化数据

### 低优先级
- [ ] **Grafana Dashboard** — 对接 Prometheus metrics 的预置 dashboard JSON
- [ ] **Rate Limit** — API 接口限流 (per-IP)
- [ ] **Webhook 通知** — Star 大幅增长时发送通知 (Slack/Telegram)
- [ ] **Dark Mode** — Tailwind dark 模式支持

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
