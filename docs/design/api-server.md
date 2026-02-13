# API 服务模块设计（API Server）

> **⚠️ v1.0 已废弃** — API Server 在 v1.0 架构中不再需要。Astro SSG 构建时直接读取 `data/*.json` 文件，无需 HTTP API。本文档保留作为历史参考。

## v0.x 架构（已废弃）

API Server 基于 Go + chi 路由，为 Astro 前端构建时提供 JSON 数据。

### 路由

```
GET /api/v1/rankings        — Top 100 排行榜
GET /api/v1/projects        — 项目列表
GET /api/v1/projects/:id    — 项目详情
GET /api/v1/posts           — 博客文章列表
GET /api/v1/categories      — 分类列表
GET /healthz                — 健康检查
```

## v1.0 替代方案

前端数据获取方式从 HTTP API 改为直接文件读取：

| v0.x | v1.0 |
|------|------|
| `fetch('/api/v1/rankings')` | `import rankings from 'data/rankings/latest.json'` |
| `fetch('/api/v1/projects/:id')` | `import project from 'data/projects/{id}.json'` |
| `fetch('/api/v1/posts')` | Glob `data/posts/*.json` |

### Astro 数据加载示例

```typescript
// web/src/lib/data.ts
import fs from 'fs';
import path from 'path';

const DATA_DIR = path.resolve('../data');

export function getLatestRanking() {
  const files = fs.readdirSync(path.join(DATA_DIR, 'rankings'))
    .filter(f => f.endsWith('.json'))
    .sort()
    .reverse();
  const content = fs.readFileSync(path.join(DATA_DIR, 'rankings', files[0]), 'utf-8');
  return JSON.parse(content);
}

export function getProject(id: string) {
  const content = fs.readFileSync(path.join(DATA_DIR, 'projects', `${id}.json`), 'utf-8');
  return JSON.parse(content);
}
```

## Phase 4 清理计划

在 Phase 4 代码清理阶段，将移除以下文件：

- `internal/server/` — 整个 API Server 包
- `internal/cmd/server.go` — server 子命令
- `go-chi/chi` 依赖
- `prometheus/client_golang` 依赖
