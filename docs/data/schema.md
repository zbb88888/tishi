# 数据 Schema

## 概述

v1.0 使用 JSON 文件存储，不再使用 PostgreSQL。所有 Schema 定义为 JSON Schema (Draft 2020-12) 格式，存放在 `data/schemas/` 目录。

> **v0.x 使用 PostgreSQL + golang-migrate，已完全废弃。** SQL Migration 文件将在 Phase 4 清理阶段移除。

## Schema 文件列表

| 文件 | 说明 | 对应数据目录 |
|------|------|-------------|
| `data/schemas/project.schema.json` | 项目档案 | `data/projects/{owner}__{repo}.json` |
| `data/schemas/snapshot.schema.json` | 每日快照（JSONL 行） | `data/snapshots/{date}.jsonl` |
| `data/schemas/ranking.schema.json` | 每日排行榜 | `data/rankings/{date}.json` |
| `data/schemas/post.schema.json` | 博客文章 | `data/posts/{slug}.json` |

## Project Schema 结构

```
data/projects/{owner}__{repo}.json
```

顶层字段：

```json
{
  "id": "owner__repo",
  "full_name": "owner/repo",
  "owner": "string",
  "repo": "string",
  "description": "string",
  "language": "string",
  "license": "string",
  "topics": ["string"],
  "homepage": "string",
  "stars": 0,
  "forks": 0,
  "open_issues": 0,
  "created_at": "ISO 8601",
  "pushed_at": "ISO 8601",
  "trending": {
    "daily_stars": 0,
    "weekly_stars": 0,
    "rank_daily": 0,
    "last_seen_trending": "YYYY-MM-DD"
  },
  "analysis": {
    "status": "draft|published|rejected",
    "model": "deepseek-chat",
    "summary": "一句话中文概括",
    "positioning": "项目定位",
    "features": ["功能点"],
    "advantages": "技术优势",
    "tech_stack": "核心技术栈",
    "use_cases": "适用场景",
    "comparison": [{"name": "同类项目", "difference": "差异"}],
    "ecosystem": "上下游生态",
    "generated_at": "ISO 8601",
    "reviewed_at": "ISO 8601",
    "token_usage": {"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0}
  },
  "categories": ["llm", "agent"],
  "score": 85.5,
  "rank": 1,
  "first_seen_at": "ISO 8601",
  "updated_at": "ISO 8601"
}
```

## Snapshot Schema 结构

```
data/snapshots/{date}.jsonl       # 每行一个 JSON 对象
```

```json
{"project_id": "owner__repo", "date": "2025-01-15", "stars": 15000, "forks": 2000, "open_issues": 100, "watchers": 500, "score": 85.5, "rank": 1, "daily_stars": 120}
```

## Ranking Schema 结构

```
data/rankings/{date}.json
```

```json
{
  "date": "2025-01-15",
  "total": 50,
  "items": [
    {
      "rank": 1,
      "project_id": "owner__repo",
      "full_name": "owner/repo",
      "summary": "一句话中文概括",
      "language": "Python",
      "category": "llm",
      "stars": 15000,
      "daily_stars": 120,
      "weekly_stars": 500,
      "score": 85.5,
      "rank_change": 2
    }
  ]
}
```

## Post Schema 结构

```
data/posts/{slug}.json
```

```json
{
  "slug": "weekly-2025-03",
  "title": "AI 开源周报 2025 第 3 期",
  "content": "# 标题\n\n正文 Markdown...",
  "post_type": "weekly",
  "published_at": "2025-01-20T00:00:00Z",
  "projects": ["owner__repo1", "owner__repo2"],
  "metadata": {}
}
```

## Categories 定义

```
data/categories.json
```

12 个 AI 分类在 `data/categories.json` 中静态定义，不使用数据库表。每个分类包含 `topics` 和 `description` 两组关键词用于自动匹配。

## 文件命名规则

| 数据类型 | 文件名格式 | 说明 |
|---------|-----------|------|
| 项目 | `{owner}__{repo}.json` | 双下划线分隔 owner 和 repo |
| 快照 | `{date}.jsonl` | 日期格式 `YYYY-MM-DD` |
| 排行榜 | `{date}.json` | 日期格式 `YYYY-MM-DD` |
| 文章 | `{slug}.json` | URL 友好标识 |

## v0.x 遗留

以下 PostgreSQL 相关文件将在 Phase 4 清理阶段移除：

- `internal/db/migrations/` — SQL Migration 文件
- `internal/db/` — 数据库连接和迁移代码
- `sqlc.yaml` — SQLC 配置

## 相关文档

- [数据字典](data-dictionary.md) — 字段详细说明
- [存储设计](../design/storage.md) — 设计决策
- [JSON Schema 文件](../../data/schemas/) — 完整 Schema 定义
