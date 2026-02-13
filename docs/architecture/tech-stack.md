# 技术选型

## 总览

| 层级 | 技术 | 版本 | 用途 |
|------|------|------|------|
| 后端语言 | Go | ≥ 1.23 | Stage 1 CLI 工具 |
| 前端框架 | Astro | ≥ 4.x | Stage 2 SSG 静态站点生成 |
| LLM | DeepSeek / Qwen | - | 中文项目深度分析 |
| 爬虫 | Colly | v2 | GitHub Trending HTML 解析 |
| 数据存储 | JSON 文件 + Git | - | 全链路数据交换 |
| Web Server | Nginx | ≥ 1.25 | Stage 3 静态文件托管 |

## 后端：Go

### 选择理由

1. **单二进制部署** — 编译产物无依赖，适合跨机器运行
2. **并发模型** — goroutine 适合批量调用 GitHub API + LLM API
3. **标准库丰富** — `net/http` + `encoding/json` + `os` 覆盖全部 IO 需求

### 核心库

| 功能 | 库 | 理由 |
|------|-----|------|
| CLI | `spf13/cobra` | 子命令支持 (scrape/analyze/score/generate/push) |
| 配置 | `spf13/viper` | 支持 YAML + 环境变量 |
| 日志 | `uber-go/zap` | 结构化日志、高性能 |
| 爬虫 | `gocolly/colly/v2` | Go 原生 HTML 爬虫，CSS 选择器+速率限制 |
| LLM 客户端 | `sashabaranov/go-openai` | OpenAI 兼容 API，支持 BaseURL 切换 (DeepSeek/Qwen) |
| GitHub API | `google/go-github/v67` | README 获取、项目详细信息补充 |

### v1.0 移除的库

| 库 | 原用途 | 移除理由 |
|-----|--------|---------|
| `jackc/pgx/v5` | PostgreSQL 驱动 | 不再使用数据库 |
| `go-chi/chi/v5` | HTTP 路由 | 不再需要 API Server |
| `prometheus/client_golang` | 指标暴露 | CLI 工具无需 metrics 端点 |
| `golang-migrate/migrate` | DB 迁移 | 无数据库 |
| `robfig/cron/v3` | 内置调度器 | 改用外部 cron |

## LLM：DeepSeek / Qwen

### 选择理由

1. **中文质量** — DeepSeek-V3 和 Qwen-Plus 在中文理解和生成上表现优异
2. **OpenAI 兼容 API** — 统一使用 `sashabaranov/go-openai`，仅切换 BaseURL
3. **成本极低** — DeepSeek: ¥1/M input tokens, ¥2/M output tokens; 100 个项目/天 < ¥1

### API 端点

| 提供商 | BaseURL | 模型 |
|--------|---------|------|
| DeepSeek | `https://api.deepseek.com/v1` | `deepseek-chat` |
| Qwen (阿里云) | `https://dashscope.aliyuncs.com/compatible-mode/v1` | `qwen-plus` |

## 爬虫：Colly

### 选择理由

1. **Go 原生** — 无需 headless browser，性能好
2. **CSS 选择器** — `article.Box-row` 直接定位 Trending 条目
3. **内置限速** — 自带 Rate Limit，避免被 GitHub 封 IP
4. **回调模式** — `OnHTML` 回调适合流式处理

### 不选 Chromedp/Playwright

GitHub Trending 是服务端渲染的静态 HTML，不需要 JavaScript 执行，Colly 够用。

## 前端：Astro

### 选择理由

1. **SSG 优先** — 数据每日更新，纯静态生成即可，无需 SSR
2. **零 JS** — Islands Architecture，默认不发送 JS 到客户端
3. **内容驱动** — 原生读取 JSON 数据文件，适合 data/ 目录结构
4. **构建时数据加载** — Astro 在 build 时直接 `import` JSON，无需运行时 API

### v1.0 变更

- `output: 'static'`（从 SSR 切换为纯 SSG）
- 移除 `@astrojs/node` adapter
- 数据源从 HTTP API 改为直接读取 `data/*.json`

## 数据存储：JSON + Git

### 选择理由（替换 PostgreSQL）

1. **三阶段解耦** — JSON 文件可通过 Git 在不同机器间同步，无需共享数据库
2. **人类可读** — 开发调试时直接查看/编辑 JSON 文件
3. **版本追溯** — Git 历史天然提供数据变更记录
4. **零运维** — 无需维护数据库进程、备份、连接池

### 不选 SQLite

- 虽然是本地文件，但 Git 无法 diff 二进制 `.db` 文件
- 跨机器传输不如 JSON 直观

## 相关文档

- [系统架构](system-architecture.md)
- [部署拓扑](deployment-topology.md)
- [数据字典](../data/data-dictionary.md)
