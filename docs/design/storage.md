# 数据存储模块设计（Storage）

## 概述

tishi 使用 PostgreSQL 作为唯一数据存储，本文档定义数据模型设计与存储策略。

## ER 关系图

```
┌──────────────┐       ┌───────────────────┐
│  projects    │       │ daily_snapshots   │
│──────────────│       │───────────────────│
│ id (PK)      │◀──┐   │ id (PK)           │
│ github_id    │   └──│ project_id (FK)   │
│ full_name    │       │ snapshot_date     │
│ description  │       │ stargazers_count  │
│ language     │       │ forks_count       │
│ license      │       │ open_issues_count │
│ topics       │       │ watchers_count    │
│ homepage     │       │ score             │
│ score        │       │ rank              │
│ rank         │       │ created_at        │
│ metadata     │       └───────────────────┘
│ first_seen_at│
│ is_archived  │       ┌───────────────────┐
│ created_at   │       │ categories        │
│ updated_at   │       │───────────────────│
└──────┬───────┘       │ id (PK)           │
       │               │ name              │
       │               │ slug              │
       │               │ parent_id (FK)    │
       │               │ description       │
       │               └────────┬──────────┘
       │                        │
       ▼                        ▼
┌──────────────────────────────────┐
│   project_categories (M:N)      │
│──────────────────────────────────│
│ project_id (FK)                 │
│ category_id (FK)                │
└──────────────────────────────────┘

┌──────────────────┐
│   blog_posts     │
│──────────────────│
│ id (PK)          │
│ title            │
│ slug             │
│ content          │
│ post_type        │  -- weekly / monthly / spotlight
│ cover_image_url  │
│ published_at     │
│ created_at       │
│ updated_at       │
└──────────────────┘
```

## 表设计

### projects — 项目主表

```sql
CREATE TABLE projects (
    id              BIGSERIAL PRIMARY KEY,
    github_id       BIGINT NOT NULL UNIQUE,
    full_name       VARCHAR(255) NOT NULL UNIQUE,  -- owner/repo
    description     TEXT,
    language        VARCHAR(50),
    license         VARCHAR(50),
    topics          TEXT[],                         -- PostgreSQL 数组类型
    homepage        VARCHAR(500),
    created_at_gh   TIMESTAMPTZ,                   -- GitHub 仓库创建时间
    pushed_at       TIMESTAMPTZ,                   -- 最后推送时间
    metadata        JSONB DEFAULT '{}',            -- 扩展字段（灵活存储）

    -- 分析结果（由 Analyzer 更新）
    score           NUMERIC(5,2) DEFAULT 0,        -- 热度评分 0-100
    rank            INT,                           -- 当前排名

    -- 系统字段
    first_seen_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_archived     BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_projects_score ON projects(score DESC);
CREATE INDEX idx_projects_rank ON projects(rank ASC) WHERE rank IS NOT NULL;
CREATE INDEX idx_projects_language ON projects(language);
CREATE INDEX idx_projects_topics ON projects USING GIN(topics);
CREATE INDEX idx_projects_metadata ON projects USING GIN(metadata);
```

### daily_snapshots — 每日快照

```sql
CREATE TABLE daily_snapshots (
    id                BIGSERIAL PRIMARY KEY,
    project_id        BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    snapshot_date     DATE NOT NULL,

    -- 指标快照
    stargazers_count  INT NOT NULL DEFAULT 0,
    forks_count       INT NOT NULL DEFAULT 0,
    open_issues_count INT NOT NULL DEFAULT 0,
    watchers_count    INT NOT NULL DEFAULT 0,

    -- 分析结果
    score             NUMERIC(5,2),
    rank              INT,

    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(project_id, snapshot_date)
);

-- 索引
CREATE INDEX idx_snapshots_date ON daily_snapshots(snapshot_date DESC);
CREATE INDEX idx_snapshots_project_date ON daily_snapshots(project_id, snapshot_date DESC);
```

### categories — 分类表

```sql
CREATE TABLE categories (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    parent_id   INT REFERENCES categories(id),
    description TEXT,
    sort_order  INT DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 初始分类（见种子数据文档）
```

### project_categories — 项目-分类关联（M:N）

```sql
CREATE TABLE project_categories (
    project_id  BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    category_id INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    confidence  NUMERIC(3,2) DEFAULT 1.0,  -- 分类置信度（自动分类时使用）
    PRIMARY KEY (project_id, category_id)
);
```

### blog_posts — 博客文章

```sql
CREATE TABLE blog_posts (
    id              BIGSERIAL PRIMARY KEY,
    title           VARCHAR(500) NOT NULL,
    slug            VARCHAR(500) NOT NULL UNIQUE,
    content         TEXT NOT NULL,                   -- Markdown 格式
    post_type       VARCHAR(20) NOT NULL,            -- weekly / monthly / spotlight
    cover_image_url VARCHAR(500),
    published_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_posts_type ON blog_posts(post_type);
CREATE INDEX idx_posts_published ON blog_posts(published_at DESC);
```

## 数据库迁移

使用 `golang-migrate/migrate` 管理 Schema 变更：

```
migrations/
├── 000001_create_projects.up.sql
├── 000001_create_projects.down.sql
├── 000002_create_daily_snapshots.up.sql
├── 000002_create_daily_snapshots.down.sql
├── 000003_create_categories.up.sql
├── 000003_create_categories.down.sql
├── 000004_create_blog_posts.up.sql
├── 000004_create_blog_posts.down.sql
└── ...
```

## 查询模式

### 常用查询

```sql
-- Top 100 排行榜
SELECT p.*, array_agg(c.name) AS categories
FROM projects p
LEFT JOIN project_categories pc ON p.id = pc.project_id
LEFT JOIN categories c ON pc.category_id = c.id
WHERE p.rank IS NOT NULL AND p.rank <= 100
GROUP BY p.id
ORDER BY p.rank ASC;

-- 项目趋势（最近 30 天）
SELECT snapshot_date, stargazers_count, forks_count, score
FROM daily_snapshots
WHERE project_id = $1
  AND snapshot_date >= CURRENT_DATE - INTERVAL '30 days'
ORDER BY snapshot_date ASC;

-- Star 日增量（窗口函数）
SELECT snapshot_date,
       stargazers_count,
       stargazers_count - LAG(stargazers_count) OVER (ORDER BY snapshot_date) AS daily_star_gain
FROM daily_snapshots
WHERE project_id = $1
ORDER BY snapshot_date DESC
LIMIT 30;
```

## 数据保留策略

| 数据类型 | 保留策略 | 理由 |
|----------|----------|------|
| projects | 永久 | 主数据，量小 |
| daily_snapshots | 永久（前期）| 年增 ~5 万条，PG 轻松承载 |
| blog_posts | 永久 | 博客内容，量极小 |

> 如未来数据量增长，可考虑将超过 1 年的快照按周聚合。

## 相关文档

- [数据库 Schema](../data/schema.md) — 完整 DDL
- [数据字典](../data/data-dictionary.md) — 字段详细说明
- [数据流转](../architecture/data-flow.md) — 数据如何流入存储
