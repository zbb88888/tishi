# 数据字典

## projects 表

| 字段 | 类型 | 可空 | 默认值 | 说明 |
|------|------|------|--------|------|
| `id` | BIGSERIAL | N | auto | 主键 |
| `github_id` | BIGINT | N | - | GitHub 仓库唯一 ID，用于 Upsert |
| `full_name` | VARCHAR(255) | N | - | `owner/repo` 格式，如 `langchain-ai/langchain` |
| `description` | TEXT | Y | NULL | 项目描述 |
| `language` | VARCHAR(50) | Y | NULL | 主要编程语言，如 `Python`、`TypeScript` |
| `license` | VARCHAR(50) | Y | NULL | 开源协议 SPDX ID，如 `MIT`、`Apache-2.0` |
| `topics` | TEXT[] | N | `{}` | GitHub Topics 标签数组 |
| `homepage` | VARCHAR(500) | Y | NULL | 项目主页 URL |
| `created_at_gh` | TIMESTAMPTZ | Y | NULL | GitHub 仓库创建时间 |
| `pushed_at` | TIMESTAMPTZ | Y | NULL | 最后一次 Push 时间 |
| `metadata` | JSONB | N | `{}` | 扩展字段，存储不常用但可能有用的信息 |
| `stargazers_count` | INT | N | 0 | 当前 Star 数 |
| `forks_count` | INT | N | 0 | 当前 Fork 数 |
| `open_issues_count` | INT | N | 0 | 当前 Open Issue 数 |
| `watchers_count` | INT | N | 0 | 当前 Watcher 数 |
| `score` | NUMERIC(5,2) | N | 0 | 热度评分，0-100 |
| `rank` | INT | Y | NULL | 当前 Top 100 排名，不在榜内为 NULL |
| `first_seen_at` | TIMESTAMPTZ | N | NOW() | tishi 首次采集到该项目的时间 |
| `is_archived` | BOOLEAN | N | FALSE | 项目是否已归档/删除 |
| `created_at` | TIMESTAMPTZ | N | NOW() | 记录创建时间 |
| `updated_at` | TIMESTAMPTZ | N | NOW() | 记录最后更新时间（trigger 自动更新） |

### metadata JSONB 示例

```json
{
    "default_branch": "main",
    "contributor_count": 1234,
    "has_wiki": true,
    "has_discussions": true,
    "size_kb": 52340,
    "github_url": "https://github.com/langchain-ai/langchain"
}
```

---

## daily_snapshots 表

| 字段 | 类型 | 可空 | 默认值 | 说明 |
|------|------|------|--------|------|
| `id` | BIGSERIAL | N | auto | 主键 |
| `project_id` | BIGINT | N | - | 外键 → projects.id，级联删除 |
| `snapshot_date` | DATE | N | - | 快照日期（UTC），与 project_id 联合唯一 |
| `stargazers_count` | INT | N | 0 | 当日 Star 数 |
| `forks_count` | INT | N | 0 | 当日 Fork 数 |
| `open_issues_count` | INT | N | 0 | 当日 Open Issue 数 |
| `watchers_count` | INT | N | 0 | 当日 Watcher 数 |
| `score` | NUMERIC(5,2) | Y | NULL | 当日热度评分 |
| `rank` | INT | Y | NULL | 当日排名 |
| `created_at` | TIMESTAMPTZ | N | NOW() | 记录创建时间 |

**唯一约束**：`UNIQUE(project_id, snapshot_date)` — 每个项目每天只有一条快照。

---

## categories 表

| 字段 | 类型 | 可空 | 默认值 | 说明 |
|------|------|------|--------|------|
| `id` | SERIAL | N | auto | 主键 |
| `name` | VARCHAR(100) | N | - | 分类名称（中文），如 `大语言模型` |
| `slug` | VARCHAR(100) | N | - | URL 友好标识，如 `llm`，唯一 |
| `parent_id` | INT | Y | NULL | 父分类 ID，支持树状分类（目前不用） |
| `description` | TEXT | Y | NULL | 分类描述 |
| `sort_order` | INT | N | 0 | 排序权重，越小越靠前 |
| `created_at` | TIMESTAMPTZ | N | NOW() | 记录创建时间 |

---

## project_categories 表

| 字段 | 类型 | 可空 | 默认值 | 说明 |
|------|------|------|--------|------|
| `project_id` | BIGINT | N | - | 外键 → projects.id |
| `category_id` | INT | N | - | 外键 → categories.id |
| `confidence` | NUMERIC(3,2) | N | 1.0 | 分类置信度：1.0=精确匹配，0.8=关键词匹配，0.6=模糊匹配 |
| `created_at` | TIMESTAMPTZ | N | NOW() | 关联创建时间 |

**主键**：`(project_id, category_id)` — 多对多关联。

---

## blog_posts 表

| 字段 | 类型 | 可空 | 默认值 | 说明 |
|------|------|------|--------|------|
| `id` | BIGSERIAL | N | auto | 主键 |
| `title` | VARCHAR(500) | N | - | 文章标题 |
| `slug` | VARCHAR(500) | N | - | URL 路径标识，唯一 |
| `content` | TEXT | N | - | Markdown 格式文章内容 |
| `post_type` | VARCHAR(20) | N | - | 文章类型枚举 |
| `cover_image_url` | VARCHAR(500) | Y | NULL | 封面图片 URL |
| `published_at` | TIMESTAMPTZ | Y | NULL | 发布时间，NULL 表示草稿 |
| `created_at` | TIMESTAMPTZ | N | NOW() | 记录创建时间 |
| `updated_at` | TIMESTAMPTZ | N | NOW() | 记录最后更新时间 |

### post_type 枚举值

| 值 | 说明 |
|-----|------|
| `weekly` | AI 开源周报 |
| `monthly` | AI 开源月报 |
| `spotlight` | 新项目速递 |

---

## 相关文档

- [数据库 Schema](schema.md) — 完整 DDL
- [存储设计](../design/storage.md) — 设计决策
