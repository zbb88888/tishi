# API 服务模块设计（API Server）

## 概述

API Server 是 tishi 后端的 HTTP 服务，基于 Go + chi 路由，为 Astro 前端构建时提供数据。

## 架构

```
HTTP Request
     │
     ▼
┌─────────────────────────┐
│     Middleware Chain     │
│  Logger → Recovery →    │
│  CORS → RateLimit       │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│       chi Router        │
│                         │
│  /api/v1/rankings       │
│  /api/v1/projects       │
│  /api/v1/projects/:id   │
│  /api/v1/posts          │
│  /api/v1/categories     │
│  /healthz               │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│       Handlers          │
│  (请求校验 + 业务逻辑)    │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│    Repository Layer      │
│  (sqlc generated code)   │
└────────────┬────────────┘
             │
             ▼
        PostgreSQL
```

## 路由定义

详细接口规范见 [RESTful API](../api/restful-api.md)。

```go
func NewRouter(h *Handlers) chi.Router {
    r := chi.NewRouter()

    // 全局中间件
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.RealIP)
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins: []string{"*"},
        AllowedMethods: []string{"GET"},
    }))

    // 健康检查
    r.Get("/healthz", h.Healthz)

    // API v1
    r.Route("/api/v1", func(r chi.Router) {
        r.Use(middleware.Throttle(100)) // 并发限制

        r.Get("/rankings", h.GetRankings)
        r.Get("/projects", h.ListProjects)
        r.Get("/projects/{id}", h.GetProject)
        r.Get("/projects/{id}/trends", h.GetProjectTrends)
        r.Get("/posts", h.ListPosts)
        r.Get("/posts/{slug}", h.GetPost)
        r.Get("/categories", h.ListCategories)
    })

    return r
}
```

## 中间件

| 中间件 | 功能 |
|--------|------|
| Logger | 请求日志（zap structured logging） |
| Recoverer | panic 恢复，返回 500 |
| RealIP | 从 X-Forwarded-For 获取真实 IP |
| CORS | 跨域支持（Astro 开发模式需要） |
| Throttle | 并发请求限制 |

## 响应格式

统一 JSON 响应格式：

```json
// 成功响应
{
    "data": { ... },
    "meta": {
        "total": 100,
        "page": 1,
        "per_page": 20
    }
}

// 错误响应
{
    "error": {
        "code": "NOT_FOUND",
        "message": "Project not found"
    }
}
```

## 分页

使用 offset-based 分页（数据量小，无需 cursor-based）：

```
GET /api/v1/projects?page=1&per_page=20
```

默认 `per_page=20`，最大 `per_page=100`。

## 缓存策略

| 端点 | Cache-Control | 理由 |
|------|---------------|------|
| `/api/v1/rankings` | `max-age=3600` | 每日更新一次，1小时缓存足够 |
| `/api/v1/projects/:id` | `max-age=3600` | 同上 |
| `/api/v1/projects/:id/trends` | `max-age=3600` | 同上 |
| `/api/v1/posts` | `max-age=86400` | 文章更新频率更低 |
| `/healthz` | `no-cache` | 实时状态 |

> 主要缓存依赖 Nginx 层的 proxy_cache，Go 应用层设置 HTTP 缓存头。

## 错误处理

```go
type AppError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Status  int    `json:"-"`
}

var (
    ErrNotFound     = &AppError{Code: "NOT_FOUND", Message: "Resource not found", Status: 404}
    ErrBadRequest   = &AppError{Code: "BAD_REQUEST", Message: "Invalid request", Status: 400}
    ErrInternal     = &AppError{Code: "INTERNAL", Message: "Internal server error", Status: 500}
)
```

## 配置

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 10s
  write_timeout: 30s
  idle_timeout: 60s
  max_request_body: 1MB
```

## 相关文档

- [RESTful API](../api/restful-api.md) — 接口详细定义
- [部署拓扑](../architecture/deployment-topology.md) — Nginx 反向代理配置
