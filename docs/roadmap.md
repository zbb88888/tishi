# 版本路线图

> 最后更新: 2025-07-15

## 历史版本

### v0.1 — Full-Stack MVP ✅

跑通 GitHub Search API → PostgreSQL → chi API → Astro SSR 排行榜全链路。

- cobra CLI: server / collect / analyze / generate / migrate / version
- chi HTTP API: 8 个端点 + Middleware + CORS + 分页
- GitHub 数据采集器 (go-github/v67 + Token 轮换)
- 多维加权评分分析器
- 博客自动生成器 (weekly/monthly 模板)
- Cron 调度器 (robfig/cron + taskLock)
- PostgreSQL 迁移 (4 张表)
- Astro SSR 前端 (5 个页面 + TrendChart)
- Docker 多阶段构建 + Compose + CI + 23 篇设计文档

### v0.2 — Quality & Observability ✅

Prometheus metrics、golangci-lint v2、单元测试（覆盖率 30.1%）、404 页面、Pagination 组件。

---

## v1.0 — AI 项目深度分析平台（当前）

### 产品差异点（vs GitHub Trending）

1. **中文深度项目解读** — LLM 自动生成面向中文开发者的完整项目报告（定位/功能/优势/对比/场景），不用翻墙看英文 README
2. **AI 垂直精选 + 分类导航** — 只收录 AI 相关项目，12 个 AI 子方向分类（LLM/Agent/RAG/图像生成/向量数据库/MLOps 等）
3. **趋势追踪 + 历史可回溯** — 每日快照持久化，Star 增长曲线、排名变动追踪，辅助技术选型

### 架构

三阶段完全解耦，JSON + Git 交换数据，无共享数据库：

| Stage | 职责 | 运行环境 | 输入 | 输出 |
|-------|------|---------|------|------|
| 1. 采集+分析 | Trending 爬取 → AI 过滤 → LLM 中文分析 | Machine A | GitHub Trending HTML + README | data/ JSON → Git |
| 2. SSG 构建 | Astro 静态页面生成 | Machine B | data/ JSON (Git pull) | dist/ 静态文件 |
| 3. 发布 | Nginx / CDN 静态服务 | Machine C | dist/ 静态文件 | 用户可访问的网站 |

### 里程碑

- [x] Phase 0: 数据契约设计（JSON Schema + 目录结构）
- [ ] Phase 1: Stage 1 采集 + LLM 分析 CLI
- [ ] Phase 2: Stage 2 Astro SSG 重构
- [ ] Phase 3: Stage 3 静态部署
- [ ] Phase 4: 代码清理（移除 PG/API Server/Scheduler）

---

## 远期展望（v1.x+）

- **Topic 订阅** — 自定义 topic（如 mcp, function-calling），基于 GitHub Search API 抓取
- **邮件订阅** — 周报/月报推送
- **项目对比** — 多项目横向对比
- **RSS Feed** — 博客 RSS 输出
- **Dark Mode** — 暗色主题
