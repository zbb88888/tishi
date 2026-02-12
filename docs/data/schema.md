# 数据库 Schema

## 概述

完整的 PostgreSQL DDL 定义。使用 `golang-migrate/migrate` 管理版本。

## Migration 001: projects

```sql
-- 000001_create_projects.up.sql

CREATE TABLE projects (
    id              BIGSERIAL PRIMARY KEY,
    github_id       BIGINT NOT NULL UNIQUE,
    full_name       VARCHAR(255) NOT NULL UNIQUE,
    description     TEXT,
    language        VARCHAR(50),
    license         VARCHAR(50),
    topics          TEXT[] DEFAULT '{}',
    homepage        VARCHAR(500),
    created_at_gh   TIMESTAMPTZ,
    pushed_at       TIMESTAMPTZ,
    metadata        JSONB DEFAULT '{}',

    -- 当前指标（每日更新）
    stargazers_count  INT NOT NULL DEFAULT 0,
    forks_count       INT NOT NULL DEFAULT 0,
    open_issues_count INT NOT NULL DEFAULT 0,
    watchers_count    INT NOT NULL DEFAULT 0,

    -- 分析结果
    score           NUMERIC(5,2) DEFAULT 0,
    rank            INT,

    -- 系统字段
    first_seen_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_archived     BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_projects_github_id ON projects(github_id);
CREATE INDEX idx_projects_full_name ON projects(full_name);
CREATE INDEX idx_projects_score ON projects(score DESC);
CREATE INDEX idx_projects_rank ON projects(rank ASC) WHERE rank IS NOT NULL;
CREATE INDEX idx_projects_language ON projects(language);
CREATE INDEX idx_projects_topics ON projects USING GIN(topics);
CREATE INDEX idx_projects_metadata ON projects USING GIN(metadata);
CREATE INDEX idx_projects_first_seen ON projects(first_seen_at DESC);

-- 自动更新 updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_projects_updated_at
    BEFORE UPDATE ON projects
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

```sql
-- 000001_create_projects.down.sql

DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS projects;
```

## Migration 002: daily_snapshots

```sql
-- 000002_create_daily_snapshots.up.sql

CREATE TABLE daily_snapshots (
    id                BIGSERIAL PRIMARY KEY,
    project_id        BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    snapshot_date     DATE NOT NULL,

    stargazers_count  INT NOT NULL DEFAULT 0,
    forks_count       INT NOT NULL DEFAULT 0,
    open_issues_count INT NOT NULL DEFAULT 0,
    watchers_count    INT NOT NULL DEFAULT 0,

    score             NUMERIC(5,2),
    rank              INT,

    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(project_id, snapshot_date)
);

CREATE INDEX idx_snapshots_date ON daily_snapshots(snapshot_date DESC);
CREATE INDEX idx_snapshots_project_date ON daily_snapshots(project_id, snapshot_date DESC);
CREATE INDEX idx_snapshots_rank ON daily_snapshots(snapshot_date, rank ASC) WHERE rank IS NOT NULL;
```

```sql
-- 000002_create_daily_snapshots.down.sql

DROP TABLE IF EXISTS daily_snapshots;
```

## Migration 003: categories

```sql
-- 000003_create_categories.up.sql

CREATE TABLE categories (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    parent_id   INT REFERENCES categories(id) ON DELETE SET NULL,
    description TEXT,
    sort_order  INT DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE project_categories (
    project_id  BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    category_id INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    confidence  NUMERIC(3,2) DEFAULT 1.0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, category_id)
);

CREATE INDEX idx_project_categories_category ON project_categories(category_id);

-- 初始分类数据
INSERT INTO categories (name, slug, description, sort_order) VALUES
    ('大语言模型', 'llm', 'LLM、ChatBot、文本生成相关项目', 1),
    ('AI Agent', 'agent', 'AI Agent 框架、自主代理、Agentic 工具', 2),
    ('RAG', 'rag', '检索增强生成、知识库、文档问答', 3),
    ('图像生成', 'diffusion', 'Diffusion 模型、文生图、图像编辑', 4),
    ('MLOps', 'mlops', 'ML 工程化、模型训练/部署/监控', 5),
    ('向量数据库', 'vector-db', '向量存储、相似度搜索、Embedding', 6),
    ('AI 框架', 'framework', '深度学习框架、训练工具', 7),
    ('AI 工具', 'tool', 'AI 辅助开发、代码生成、AI 助手', 8),
    ('多模态', 'multimodal', '多模态模型、视觉语言模型', 9),
    ('语音', 'speech', 'TTS、ASR、语音克隆', 10),
    ('强化学习', 'rl', 'RLHF、强化学习框架', 11),
    ('其他', 'other', '未归类的 AI 相关项目', 99);
```

```sql
-- 000003_create_categories.down.sql

DROP TABLE IF EXISTS project_categories;
DROP TABLE IF EXISTS categories;
```

## Migration 004: blog_posts

```sql
-- 000004_create_blog_posts.up.sql

CREATE TABLE blog_posts (
    id              BIGSERIAL PRIMARY KEY,
    title           VARCHAR(500) NOT NULL,
    slug            VARCHAR(500) NOT NULL UNIQUE,
    content         TEXT NOT NULL,
    post_type       VARCHAR(20) NOT NULL CHECK (post_type IN ('weekly', 'monthly', 'spotlight')),
    cover_image_url VARCHAR(500),
    published_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_posts_type ON blog_posts(post_type);
CREATE INDEX idx_posts_published ON blog_posts(published_at DESC);
CREATE INDEX idx_posts_slug ON blog_posts(slug);

CREATE TRIGGER update_blog_posts_updated_at
    BEFORE UPDATE ON blog_posts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

```sql
-- 000004_create_blog_posts.down.sql

DROP TRIGGER IF EXISTS update_blog_posts_updated_at ON blog_posts;
DROP TABLE IF EXISTS blog_posts;
```

## 相关文档

- [数据字典](data-dictionary.md) — 字段详细说明
- [存储设计](../design/storage.md) — 数据模型设计理由
