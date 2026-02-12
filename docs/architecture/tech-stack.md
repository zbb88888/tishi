# 技术选型

## 总览

| 层级 | 技术 | 版本要求 | 选型理由 |
|------|------|----------|----------|
| 后端语言 | Go | ≥ 1.22 | 高性能、单二进制部署、强并发支持 |
| 前端框架 | Astro | ≥ 4.x | 内容驱动型 SSG、性能极佳、Islands Architecture |
| 数据库 | PostgreSQL | ≥ 16 | JSONB 灵活存储、社区成熟、扩展性好 |
| 部署 | Docker Compose | v2 | 个人项目一键部署、运维简单 |
| Web Server | Nginx | ≥ 1.25 | 托管 Astro 构建产物、反向代理 API |

## 后端：Go

### 选择理由

1. **单二进制部署** — 编译产物无依赖，Docker 镜像可用 scratch/distroless，极小
2. **并发模型** — goroutine 适合并发调用 GitHub API（Rate Limit 下多 Token 轮换）
3. **标准库丰富** — `net/http` + `encoding/json` + `database/sql` 覆盖核心需求
4. **生态成熟** — ORM（sqlc/GORM）、HTTP 路由（chi/gin）、定时任务（robfig/cron）

### 核心库选择

| 功能 | 库 | 理由 |
|------|-----|------|
| HTTP 路由 | `go-chi/chi` | 轻量、标准 `net/http` 兼容、中间件生态好 |
| 数据库访问 | `sqlc` | SQL-first，类型安全，编译时生成 Go 代码 |
| 数据库迁移 | `golang-migrate/migrate` | 成熟、支持 PostgreSQL、可嵌入二进制 |
| 定时任务 | `robfig/cron/v3` | 标准 cron 表达式、goroutine 安全 |
| GitHub 客户端 | `google/go-github` | 官方维护、覆盖 REST + GraphQL |
| 日志 | `uber-go/zap` | 结构化日志、高性能 |
| 配置 | `spf13/viper` | 支持文件 + 环境变量 + 命令行参数 |
| CLI | `spf13/cobra` | 子命令支持（server/collect/analyze/generate） |

### 不选的替代方案

| 替代 | 不选理由 |
|------|----------|
| Python | 部署复杂（需 runtime）、性能差、不适合长运行服务 |
| Rust | 开发效率低、生态在 Web 领域不如 Go 成熟 |
| Node.js | 运行时依赖重、类型安全弱（TS 也仅编译时） |

## 前端：Astro

### 选择理由

1. **SSG 优先** — 数据每日更新一次，SSG 完全够用，无需 SSR
2. **性能极佳** — 默认零 JS 发送到客户端（Islands Architecture）
3. **内容驱动** — 原生支持 Markdown/MDX，适合博客内容
4. **灵活集成** — 可在 Island 中使用 React/Vue/Svelte 组件（图表用）
5. **SEO 友好** — 纯静态 HTML，搜索引擎收录优秀

### 图表库

| 功能 | 库 | 理由 |
|------|-----|------|
| 趋势图表 | Chart.js 或 ECharts | 轻量、交互性好、在 Astro Island 中按需加载 |

### 不选的替代方案

| 替代 | 不选理由 |
|------|----------|
| Next.js | SSR 能力过剩，部署需 Node runtime，运维复杂 |
| Hugo | 模板语言受限，无法嵌入交互式图表组件 |
| Vue/Nuxt | SPA/SSR 模式对博客过重 |

## 数据库：PostgreSQL

### 选择理由

1. **JSONB 支持** — GitHub API 返回的项目元数据字段多变，JSONB 灵活存储
2. **窗口函数** — 排名计算、环比同比分析用 SQL 窗口函数高效实现
3. **扩展性** — 未来如需全文搜索可用 `pg_trgm`，时序可加 `timescaledb`
4. **社区成熟** — 文档丰富、工具链完善、运维经验多

### 不选的替代方案

| 替代 | 不选理由 |
|------|----------|
| SQLite | 单文件不支持并发写入（API Server + Collector 同时运行时） |
| MySQL | JSONB 支持弱、窗口函数生态不如 PG |
| MongoDB | 对这个场景过重，关系查询能力弱 |

## 部署：Docker Compose

### 选择理由

1. **一键启停** — `docker-compose up -d` 搞定所有服务
2. **环境一致** — 开发/生产环境一致，避免 "works on my machine"
3. **适合个人项目** — 无需 K8s 的复杂度

### 容器编排

```yaml
services:
  tishi:        # Go 二进制（API Server + Scheduler）
  postgres:     # PostgreSQL 数据库
  nginx:        # 静态文件托管 + 反向代理
```

### 不选的替代方案

| 替代 | 不选理由 |
|------|----------|
| Kubernetes | 个人项目不需要，运维成本高 |
| 裸机部署 | 环境不可复现，运维麻烦 |
| Serverless | 定时任务 + 数据库连接管理复杂 |

## 相关文档

- [系统架构](system-architecture.md)
- [部署拓扑](deployment-topology.md)
