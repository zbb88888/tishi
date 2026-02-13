# RESTful API 定义

> ⚠️ **v1.0 已废弃** — v1.0 移除了 HTTP API Server，前端改为 Astro SSG 直接读取 JSON 文件。

## v1.0 替代方案

v1.0 不再提供运行时 HTTP API。数据通过以下方式访问：

### 前端数据加载（Astro SSG 构建时）

```typescript
// web/src/lib/data.ts
import fs from 'node:fs';
import path from 'node:path';

const DATA_DIR = path.resolve('../data');

export function loadRanking(date: string) {
  const file = path.join(DATA_DIR, 'rankings', `${date}.json`);
  return JSON.parse(fs.readFileSync(file, 'utf-8'));
}

export function loadProject(id: string) {
  const file = path.join(DATA_DIR, 'projects', `${id}.json`);
  return JSON.parse(fs.readFileSync(file, 'utf-8'));
}

export function loadCategories() {
  const file = path.join(DATA_DIR, 'categories.json');
  return JSON.parse(fs.readFileSync(file, 'utf-8'));
}
```

### 数据文件即 API

| v0.x API 路由 | v1.0 数据文件 |
|---------------|-------------|
| `GET /api/v1/rankings` | `data/rankings/{date}.json` |
| `GET /api/v1/projects` | `data/projects/*.json` (列举) |
| `GET /api/v1/projects/{id}` | `data/projects/{owner}__{repo}.json` |
| `GET /api/v1/projects/{id}/trends` | `data/snapshots/{date}.jsonl` |
| `GET /api/v1/posts` | `data/posts/*.json` (列举) |
| `GET /api/v1/posts/{slug}` | `data/posts/{slug}.json` |
| `GET /api/v1/categories` | `data/categories.json` |
| `GET /healthz` | 不再需要（无运行时服务） |

## v0.x API 参考（已废弃）

以下 API 路由保留作为参考，将在 Phase 4 清理阶段删除相关代码：

| 方法 | 路由 | 说明 |
|------|------|------|
| GET | `/api/v1/rankings` | 排行榜（分页、分类筛选） |
| GET | `/api/v1/projects` | 项目列表（搜索、筛选） |
| GET | `/api/v1/projects/{id}` | 项目详情 |
| GET | `/api/v1/projects/{id}/trends` | 项目趋势数据 |
| GET | `/api/v1/posts` | 文章列表 |
| GET | `/api/v1/posts/{slug}` | 文章详情 |
| GET | `/api/v1/categories` | 分类列表 |
| GET | `/healthz` | 健康检查 |

## Phase 4 清理计划

移除以下代码和依赖：

- `internal/server/` — HTTP Server 代码
- `internal/cmd/server.go` — cobra server 子命令
- `go-chi/chi/v5` — HTTP 路由器
- `prometheus/client_golang` — Metrics 中间件

## 相关文档

- [API Server 设计](../design/api-server.md) — 设计决策和废弃说明
- [Web 前端设计](../design/web-frontend.md) — SSG 数据加载方式
