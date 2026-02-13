# 数据存储设计（Storage）

## 概述

tishi v1.0 使用 **JSON 文件 + Git** 作为唯一数据存储，替换原有的 PostgreSQL。所有数据以 JSON 格式存储在 `data/` 目录，通过 Git 仓库在三个 Stage 之间同步。

## 目录结构

```
data/
├── projects/              # 每个 AI 项目一个 JSON 文件
│   ├── langchain-ai__langchain.json
│   ├── ollama__ollama.json
│   └── ...
├── snapshots/             # 每日快照 JSONL（追加式）
│   ├── 2025-07-14.jsonl
│   ├── 2025-07-15.jsonl
│   └── ...
├── rankings/              # 每日排行榜 JSON
│   ├── 2025-07-14.json
│   ├── 2025-07-15.json
│   └── ...
├── posts/                 # 博客文章 JSON
│   ├── ai-weekly-2025-w29.json
│   ├── spotlight-langchain-ai-langchain.json
│   └── ...
├── schemas/               # JSON Schema 定义
│   ├── project.schema.json
│   ├── snapshot.schema.json
│   ├── ranking.schema.json
│   └── post.schema.json
├── categories.json        # 12 个 AI 分类 + 关键词映射
└── meta.json              # 版本和元信息
```

## 数据模型

### Project JSON (data/projects/{owner}__{repo}.json)

详细 Schema 见 `data/schemas/project.schema.json`。核心字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | `owner__repo` 格式 |
| full_name | string | `owner/repo` 格式 |
| description | string | 项目描述 |
| language | string | 主要编程语言 |
| stars / forks | integer | 基本指标 |
| topics | string[] | GitHub topics |
| trending | object | 每日/每周 Star 增长、Trending 排名 |
| analysis | object | LLM 分析结果（status/summary/features/...） |
| categories | string[] | AI 分类标签 |
| score / rank | number | 评分和排名 |

### Snapshot JSONL (data/snapshots/{date}.jsonl)

每行一个项目的当日快照：

```jsonl
{"project_id":"langchain-ai__langchain","date":"2025-07-15","stars":95234,"forks":15432,"open_issues":1234,"score":92.5,"rank":1,"daily_stars":523}
{"project_id":"ollama__ollama","date":"2025-07-15","stars":88100,"forks":6200,"open_issues":890,"score":88.2,"rank":2,"daily_stars":412}
```

### Ranking JSON (data/rankings/{date}.json)

```json
{
  "date": "2025-07-15",
  "total": 50,
  "items": [
    {"rank": 1, "project_id": "langchain-ai__langchain", "score": 92.5, "daily_stars": 523, "rank_change": 0},
    {"rank": 2, "project_id": "ollama__ollama", "score": 88.2, "daily_stars": 412, "rank_change": 1}
  ]
}
```

### Post JSON (data/posts/{slug}.json)

```json
{
  "slug": "ai-weekly-2025-w29",
  "title": "AI 开源周报 #29",
  "content": "## 本周概览\n\n...",
  "post_type": "weekly",
  "published_at": "2025-07-20T06:00:00Z"
}
```

## 文件命名规则

| 类型 | 命名 | 示例 |
|------|------|------|
| Project | `{owner}__{repo}.json` | `langchain-ai__langchain.json` |
| Snapshot | `{YYYY-MM-DD}.jsonl` | `2025-07-15.jsonl` |
| Ranking | `{YYYY-MM-DD}.json` | `2025-07-15.json` |
| Post | `{slug}.json` | `ai-weekly-2025-w29.json` |

使用 `__` (双下划线) 分隔 owner 和 repo，因为 `/` 不能用于文件名。

## 数据操作接口

```go
package datastore

// Store 提供对 data/ 目录的读写操作
type Store struct {
    dataDir string
    logger  *zap.Logger
}

func NewStore(dataDir string) *Store { ... }

// Projects
func (s *Store) LoadProject(id string) (*Project, error)
func (s *Store) SaveProject(p *Project) error
func (s *Store) ListProjects() ([]*Project, error)

// Snapshots
func (s *Store) AppendSnapshot(date string, snap *Snapshot) error
func (s *Store) LoadSnapshots(date string) ([]*Snapshot, error)
func (s *Store) LoadSnapshotRange(from, to string) ([]*Snapshot, error)

// Rankings
func (s *Store) SaveRanking(r *Ranking) error
func (s *Store) LoadRanking(date string) (*Ranking, error)
func (s *Store) LoadLatestRanking() (*Ranking, error)

// Posts
func (s *Store) SavePost(p *Post) error
func (s *Store) LoadPost(slug string) (*Post, error)
func (s *Store) ListPosts() ([]*Post, error)
```

## Git 同步

Stage 1 完成所有数据写入后，执行 `tishi push`：

```bash
cd data/
git add -A
git commit -m "daily update $(date +%Y-%m-%d)"
git push origin main
```

Stage 2 在构建前拉取最新数据：

```bash
cd data/
git pull origin main
```

## 与 v0.x 对比

| 维度 | v0.x (PostgreSQL) | v1.0 (JSON + Git) |
|------|-------------------|-------------------|
| 存储方式 | 4 张关系表 | JSON 文件目录 |
| 查询方式 | SQL | Go file I/O + JSON unmarshal |
| 同步方式 | 共享数据库连接 | Git push/pull |
| 版本追溯 | 无（需手动备份） | Git 历史天然提供 |
| 运维 | 需维护 PG 进程/备份/连接池 | 零运维 |
| 可读性 | 需 SQL 客户端 | 直接查看 JSON 文件 |

## 数据量估算

| 数据 | 单文件大小 | 年累计文件数 | 年累计体积 |
|------|-----------|-------------|-----------|
| projects/ | ~5-10KB | ~2000 | ~20MB |
| snapshots/ | ~10-30KB | 365 | ~10MB |
| rankings/ | ~20KB | 365 | ~7MB |
| posts/ | ~5KB | ~100 | ~0.5MB |
| **合计** | | | **< 40MB/年** |

Git 仓库完全承载无压力。

## 相关文档

- [数据契约 Schema](../../data/schemas/) — JSON Schema 定义
- [数据字典](../data/data-dictionary.md) — 字段详细说明
- [数据流转](../architecture/data-flow.md) — 数据如何流入存储
