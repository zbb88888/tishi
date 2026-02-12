# RESTful API 定义

## 基础信息

| 属性 | 值 |
|------|-----|
| Base URL | `http://localhost:8080/api/v1` |
| 协议 | HTTP/HTTPS |
| 格式 | JSON |
| 编码 | UTF-8 |
| 认证 | 无（公开只读 API） |

## 通用规范

### 响应格式

```json
// 成功 - 列表
{
    "data": [...],
    "meta": {
        "total": 100,
        "page": 1,
        "per_page": 20,
        "total_pages": 5
    }
}

// 成功 - 单条
{
    "data": { ... }
}

// 错误
{
    "error": {
        "code": "NOT_FOUND",
        "message": "Resource not found"
    }
}
```

### 分页参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `page` | int | 1 | 页码，从 1 开始 |
| `per_page` | int | 20 | 每页条数，最大 100 |

### HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 429 | 请求频率超限 |
| 500 | 服务器内部错误 |

---

## 接口列表

### 1. 获取排行榜

**GET** `/api/v1/rankings`

获取当前 AI 项目 Top 100 排行榜。

**Query Parameters**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `page` | int | 否 | 页码 |
| `per_page` | int | 否 | 每页条数 |
| `category` | string | 否 | 按分类筛选，如 `llm`、`agent` |
| `sort` | string | 否 | 排序字段：`score`(默认)、`stars`、`daily_gain`、`weekly_gain` |

**Response**

```json
{
    "data": [
        {
            "rank": 1,
            "project": {
                "id": 12345,
                "github_id": 567890,
                "full_name": "langchain-ai/langchain",
                "description": "Build context-aware reasoning applications",
                "language": "Python",
                "license": "MIT",
                "stars": 95234,
                "forks": 15432,
                "open_issues": 2341,
                "score": 92.5,
                "categories": ["llm", "agent", "framework"],
                "daily_star_gain": 523,
                "weekly_star_gain": 3245,
                "pushed_at": "2026-02-11T18:30:00Z"
            },
            "rank_change": 2,
            "rank_direction": "up"
        }
    ],
    "meta": {
        "total": 100,
        "page": 1,
        "per_page": 20,
        "total_pages": 5,
        "updated_at": "2026-02-12T01:00:00Z"
    }
}
```

---

### 2. 获取项目列表

**GET** `/api/v1/projects`

搜索和浏览所有已追踪的项目。

**Query Parameters**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `page` | int | 否 | 页码 |
| `per_page` | int | 否 | 每页条数 |
| `q` | string | 否 | 搜索关键词（匹配 full_name 和 description） |
| `language` | string | 否 | 按编程语言筛选 |
| `category` | string | 否 | 按分类筛选 |
| `sort` | string | 否 | 排序：`score`(默认)、`stars`、`name`、`created` |

**Response**

```json
{
    "data": [
        {
            "id": 12345,
            "github_id": 567890,
            "full_name": "langchain-ai/langchain",
            "description": "Build context-aware reasoning applications",
            "language": "Python",
            "license": "MIT",
            "stars": 95234,
            "forks": 15432,
            "open_issues": 2341,
            "score": 92.5,
            "rank": 1,
            "categories": ["llm", "agent", "framework"],
            "homepage": "https://langchain.com",
            "pushed_at": "2026-02-11T18:30:00Z",
            "created_at_gh": "2022-10-17T00:00:00Z",
            "first_seen_at": "2026-01-15T00:00:00Z"
        }
    ],
    "meta": { "total": 150, "page": 1, "per_page": 20, "total_pages": 8 }
}
```

---

### 3. 获取项目详情

**GET** `/api/v1/projects/{id}`

**Path Parameters**

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | int | 项目 ID |

**Response**

```json
{
    "data": {
        "id": 12345,
        "github_id": 567890,
        "full_name": "langchain-ai/langchain",
        "description": "Build context-aware reasoning applications",
        "language": "Python",
        "license": "MIT",
        "topics": ["llm", "ai", "langchain", "agents"],
        "homepage": "https://langchain.com",
        "stars": 95234,
        "forks": 15432,
        "open_issues": 2341,
        "watchers": 892,
        "score": 92.5,
        "rank": 1,
        "categories": [
            {"slug": "llm", "name": "大语言模型"},
            {"slug": "agent", "name": "AI Agent"},
            {"slug": "framework", "name": "框架"}
        ],
        "daily_star_gain": 523,
        "weekly_star_gain": 3245,
        "monthly_star_gain": 12456,
        "pushed_at": "2026-02-11T18:30:00Z",
        "created_at_gh": "2022-10-17T00:00:00Z",
        "first_seen_at": "2026-01-15T00:00:00Z",
        "github_url": "https://github.com/langchain-ai/langchain"
    }
}
```

---

### 4. 获取项目趋势

**GET** `/api/v1/projects/{id}/trends`

获取项目的历史趋势数据（时序快照）。

**Path Parameters**

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | int | 项目 ID |

**Query Parameters**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `days` | int | 否 | 查询天数，默认 30，最大 365 |

**Response**

```json
{
    "data": {
        "project_id": 12345,
        "full_name": "langchain-ai/langchain",
        "trends": [
            {
                "date": "2026-02-12",
                "stars": 95234,
                "forks": 15432,
                "open_issues": 2341,
                "score": 92.5,
                "rank": 1
            },
            {
                "date": "2026-02-11",
                "stars": 94711,
                "forks": 15398,
                "open_issues": 2330,
                "score": 91.8,
                "rank": 3
            }
        ]
    }
}
```

---

### 5. 获取博客文章列表

**GET** `/api/v1/posts`

**Query Parameters**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `page` | int | 否 | 页码 |
| `per_page` | int | 否 | 每页条数 |
| `type` | string | 否 | 文章类型：`weekly`、`monthly`、`spotlight` |

**Response**

```json
{
    "data": [
        {
            "id": 1,
            "title": "AI 开源周报 #07 | 2026-02-02 ~ 2026-02-08",
            "slug": "ai-weekly-2026-w07",
            "type": "weekly",
            "published_at": "2026-02-08T06:00:00Z",
            "excerpt": "本周 Top 100 共有 5 个新入榜项目..."
        }
    ],
    "meta": { "total": 20, "page": 1, "per_page": 20, "total_pages": 1 }
}
```

---

### 6. 获取博客文章详情

**GET** `/api/v1/posts/{slug}`

**Response**

```json
{
    "data": {
        "id": 1,
        "title": "AI 开源周报 #07 | 2026-02-02 ~ 2026-02-08",
        "slug": "ai-weekly-2026-w07",
        "type": "weekly",
        "content": "## 本周概览\n\n本周 Top 100 共有 5 个新入榜项目...",
        "published_at": "2026-02-08T06:00:00Z",
        "created_at": "2026-02-08T06:00:00Z"
    }
}
```

---

### 7. 获取分类列表

**GET** `/api/v1/categories`

**Response**

```json
{
    "data": [
        {
            "id": 1,
            "name": "大语言模型",
            "slug": "llm",
            "description": "LLM 相关项目",
            "project_count": 35,
            "children": []
        },
        {
            "id": 2,
            "name": "AI Agent",
            "slug": "agent",
            "description": "AI Agent 框架和工具",
            "project_count": 18,
            "children": []
        }
    ]
}
```

---

### 8. 健康检查

**GET** `/healthz`

**Response**

```json
{
    "status": "ok",
    "version": "0.1.0",
    "uptime": "24h30m",
    "database": "connected"
}
```

## 相关文档

- [API 服务设计](../design/api-server.md) — 实现细节
- [GitHub 集成](github-integration.md) — 上游 API
