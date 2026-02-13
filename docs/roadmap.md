# 版本路线图

> 最后更新: 2026-02-13

## 当前版本

### v1.0 — AI 项目深度分析平台 ✅

三阶段完全解耦架构，JSON + Git 交换数据：

```
Stage 1 (Machine A)          Stage 2 (Machine B)         Stage 3 (Machine C)
Go CLI: 采集+分析      →     Astro SSG 构建         →     Nginx / CDN 发布
scrape→score→analyze→push    git pull → npm build         静态文件服务
```

#### 里程碑

- [x] Phase 0: 数据契约设计 (`0a77d5d`)
- [x] Phase 1: Go CLI Pipeline — scrape/score/analyze/generate/review/push (`741a2ec` + `05bad8c`)
- [x] Phase 2: Astro SSG 前端重写 (`9c3eb29`)
- [x] Phase 3: 静态部署配置 — Nginx + deploy.sh + CI
- [x] Phase 4: Legacy 清理 — 移除 PG/API Server/Scheduler (`98c6d0f`)

---

## 历史版本

### v0.1 — Full-Stack MVP ✅ (`a0e9fcc`)

跑通 GitHub Search API → PostgreSQL → chi API → Astro SSR 排行榜全链路。

### v0.2 — Quality & Observability ✅ (`5d429de`)

Prometheus metrics、golangci-lint v2、单元测试（覆盖率 30.1%）、404 页面、Pagination 组件。

---

## 远期展望（v1.x+）

- [ ] **Topic 订阅** — 自定义 topic（如 mcp, function-calling），基于 GitHub Search API 抓取
- [ ] **邮件订阅** — 周报/月报推送
- [ ] **项目对比** — 多项目横向对比
- [ ] **RSS Feed** — 博客 RSS 输出
- [ ] **Dark Mode** — 暗色主题
- [ ] **SEO 增强** — sitemap.xml, Open Graph meta, 结构化数据
