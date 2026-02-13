# 数据字典

## Project JSON (data/projects/{owner}__{repo}.json)

完整 JSON Schema 见 `data/schemas/project.schema.json`。

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | string | Y | `owner__repo` 格式，文件名即 ID |
| `full_name` | string | Y | `owner/repo` 格式 |
| `owner` | string | Y | 仓库所有者 |
| `repo` | string | Y | 仓库名称 |
| `description` | string | N | 项目描述 |
| `language` | string | N | 主要编程语言 |
| `license` | string | N | SPDX ID，如 `MIT`、`Apache-2.0` |
| `topics` | string[] | N | GitHub Topics 标签 |
| `homepage` | string | N | 项目主页 URL |
| `stars` | integer | Y | 当前 Star 数 |
| `forks` | integer | Y | 当前 Fork 数 |
| `open_issues` | integer | N | 当前 Open Issue 数 |
| `created_at` | string | N | GitHub 仓库创建时间 (ISO 8601) |
| `pushed_at` | string | N | 最后推送时间 (ISO 8601) |
| `trending` | object | N | Trending 相关数据 |
| `trending.daily_stars` | integer | N | 当日 Star 增长 |
| `trending.weekly_stars` | integer | N | 当周 Star 增长 |
| `trending.rank_daily` | integer | N | 当日 Trending 排名 |
| `trending.last_seen_trending` | string | N | 最近一次出现在 Trending 的日期 |
| `analysis` | object | N | LLM 中文分析结果 |
| `analysis.status` | string | Y* | `draft` / `published` / `rejected` |
| `analysis.model` | string | N | LLM 模型名称 |
| `analysis.summary` | string | N | 一句话中文概括 |
| `analysis.positioning` | string | N | 项目定位 |
| `analysis.features` | string[] | N | 核心功能列表 |
| `analysis.advantages` | string | N | 技术优势 |
| `analysis.tech_stack` | string | N | 核心技术栈 |
| `analysis.use_cases` | string | N | 适用场景 |
| `analysis.comparison` | array | N | 同类对比 [{name, difference}] |
| `analysis.ecosystem` | string | N | 上下游生态 |
| `analysis.generated_at` | string | N | 分析生成时间 (ISO 8601) |
| `analysis.reviewed_at` | string | N | 人工审核时间 (ISO 8601) |
| `analysis.token_usage` | object | N | LLM token 用量 |
| `categories` | string[] | N | AI 分类标签（slug 数组） |
| `score` | number | N | 热度评分 0-100 |
| `rank` | integer | N | 当前排名 |
| `first_seen_at` | string | N | 首次采集时间 (ISO 8601) |
| `updated_at` | string | Y | 最后更新时间 (ISO 8601) |

---

## Snapshot JSONL (data/snapshots/{date}.jsonl)

每行一个 JSON 对象，Schema 见 `data/schemas/snapshot.schema.json`。

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `project_id` | string | Y | 项目 ID (`owner__repo`) |
| `date` | string | Y | 快照日期 (YYYY-MM-DD) |
| `stars` | integer | Y | 当日 Star 数 |
| `forks` | integer | Y | 当日 Fork 数 |
| `open_issues` | integer | N | 当日 Open Issue 数 |
| `watchers` | integer | N | 当日 Watcher 数 |
| `score` | number | N | 当日评分 |
| `rank` | integer | N | 当日排名 |
| `daily_stars` | integer | N | 当日 Star 增长 |

---

## Ranking JSON (data/rankings/{date}.json)

Schema 见 `data/schemas/ranking.schema.json`。

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `date` | string | Y | 排名日期 (YYYY-MM-DD) |
| `total` | integer | Y | 排名项目总数 |
| `items[]` | array | Y | 排名条目列表 |
| `items[].rank` | integer | Y | 排名 |
| `items[].project_id` | string | Y | 项目 ID |
| `items[].full_name` | string | Y | `owner/repo` |
| `items[].summary` | string | N | 一句话中文概括 |
| `items[].language` | string | N | 编程语言 |
| `items[].category` | string | N | 主分类 |
| `items[].stars` | integer | N | 总 Star 数 |
| `items[].daily_stars` | integer | N | 日增 Star |
| `items[].weekly_stars` | integer | N | 周增 Star |
| `items[].score` | number | N | 评分 |
| `items[].rank_change` | integer | N | 排名变化 (正=上升) |

---

## Post JSON (data/posts/{slug}.json)

Schema 见 `data/schemas/post.schema.json`。

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `slug` | string | Y | URL 路径标识 |
| `title` | string | Y | 文章标题 |
| `content` | string | Y | Markdown 格式内容 |
| `post_type` | string | Y | `weekly` / `monthly` / `spotlight` |
| `published_at` | string | Y | 发布时间 (ISO 8601) |
| `projects` | string[] | N | 关联项目 ID 列表 |
| `metadata` | object | N | 文章元数据 |

---

## Categories JSON (data/categories.json)

12 个 AI 领域分类定义，详见 `data/categories.json` 文件。

| 分类 slug | 中文名 | 说明 |
|-----------|--------|------|
| llm | 大语言模型 | GPT, LLaMA, Mistral 等 |
| agent | AI Agent | 自主代理框架 |
| rag | RAG 检索增强 | 检索增强生成 |
| diffusion | 图像生成 | Stable Diffusion, DALL-E 等 |
| mlops | MLOps | 模型训练/部署/监控 |
| vector-db | 向量数据库 | 向量存储和检索 |
| framework | 框架 | PyTorch, TensorFlow 等 |
| tool | AI 工具 | AI 助手、代码生成 |
| multimodal | 多模态 | 视觉语言模型等 |
| speech | 语音 | TTS, ASR 等 |
| rl | 强化学习 | RLHF, PPO 等 |
| other | 其他 | 未分类 |

## 相关文档

- [JSON Schema](../../data/schemas/) — 完整 Schema 定义
- [存储设计](../design/storage.md) — 设计决策
