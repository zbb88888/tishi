# 数据采集模块设计（Scraper）

## 概述

Scraper 负责从 GitHub Trending 页面爬取项目列表，过滤出 AI 相关项目，调用 GitHub API 补充详细数据，输出到 `data/projects/*.json` 和 `data/snapshots/*.jsonl`。

## 数据源

| 数据源 | 方式 | 用途 |
|--------|------|------|
| GitHub Trending HTML | `gocolly/colly/v2` 爬取 | 获取当日/当周 Trending 项目列表 |
| GitHub REST API | `google/go-github/v67` | 补充 README、topics、详细指标 |

### Trending 页面爬取

**URL 模式**：

```
https://github.com/trending?since=daily
https://github.com/trending?since=weekly
https://github.com/trending/python?since=daily    # 按语言
```

**CSS 选择器**（`article.Box-row`）：

```go
c.OnHTML("article.Box-row", func(e *colly.HTMLElement) {
    project := TrendingItem{
        FullName:    strings.TrimSpace(e.ChildText("h2 a")),        // owner/repo
        Description: strings.TrimSpace(e.ChildText("p")),           // 项目描述
        Language:    strings.TrimSpace(e.ChildText("[itemprop=programmingLanguage]")),
        // Stars 和 period_stars 需解析 span 文本
    }
})
```

**爬取字段**：

| 字段 | CSS 选择器 | 说明 |
|------|-----------|------|
| full_name | `h2 a` | `owner/repo` |
| description | `p` | 项目描述 |
| language | `[itemprop=programmingLanguage]` | 编程语言 |
| stars_total | `.Link--muted` (第1个) | 总 Star 数 |
| forks_total | `.Link--muted` (第2个) | 总 Fork 数 |
| period_stars | `.d-inline-block.float-sm-right` | 期间增长 Star |

## AI 项目过滤

读取 `data/categories.json` 中 12 个 AI 分类的关键词映射，对每个 Trending 项目执行匹配：

```go
func isAIProject(item TrendingItem, categories []Category) (bool, []string) {
    matchedCategories := []string{}
    // 1. 检查 description 关键词
    // 2. 检查 full_name 关键词
    // 3. 检查 language (Python 项目加权)
    // 4. 后续补充: 检查 topics (需 GitHub API)
    return len(matchedCategories) > 0, matchedCategories
}
```

匹配逻辑按优先级：

- topics 精确匹配 → confidence 1.0
- description 关键词 → confidence 0.8
- full_name 关键词 → confidence 0.6

## GitHub API 数据补充

对通过过滤的 AI 项目，调用 GitHub API 获取：

| 数据 | API | 说明 |
|------|-----|------|
| README 内容 | `GET /repos/{owner}/{repo}/readme` | 用于 LLM 分析输入 |
| 项目 topics | `GET /repos/{owner}/{repo}/topics` | 精确分类匹配 |
| 详细指标 | `GET /repos/{owner}/{repo}` | forks, issues, watchers, license, created_at |

## Token 轮换

支持配置多个 GitHub Token 轮换，应对 Rate Limit：

```go
type TokenRotator struct {
    tokens  []string
    current int
    mu      sync.Mutex
}

func (r *TokenRotator) Next() string {
    r.mu.Lock()
    defer r.mu.Unlock()
    token := r.tokens[r.current]
    r.current = (r.current + 1) % len(r.tokens)
    return token
}
```

## 输出

### data/projects/{owner}__{repo}.json

每个 AI 项目一个 JSON 文件（符合 `data/schemas/project.schema.json`）。Scraper 负责写入基础字段和 `trending` 字段：

```json
{
  "id": "langchain-ai__langchain",
  "full_name": "langchain-ai/langchain",
  "owner": "langchain-ai",
  "repo": "langchain",
  "description": "Build context-aware reasoning applications",
  "language": "Python",
  "topics": ["llm", "agent", "rag"],
  "stars": 95234,
  "forks": 15432,
  "trending": {
    "daily_stars": 523,
    "weekly_stars": 3245,
    "rank_daily": 1,
    "last_seen_trending": "2025-07-15"
  },
  "categories": ["llm", "agent", "framework"],
  "updated_at": "2025-07-15T00:30:00Z"
}
```

### data/snapshots/{YYYY-MM-DD}.jsonl

追加式 JSONL，每个项目一行（符合 `data/schemas/snapshot.schema.json`）。

## 错误处理

| 错误类型 | 处理方式 |
|----------|----------|
| Trending 页面 403/429 | 等待 30s 重试，最多 3 次 |
| GitHub API Rate Limit | Token 轮换 + 指数退避 |
| 项目 404 | 跳过，日志记录 |
| JSON 写入失败 | 报错终止（数据完整性优先） |

## CLI 命令

```bash
tishi scrape                    # 爬取 daily Trending
tishi scrape --since=weekly     # 爬取 weekly Trending
tishi scrape --language=python  # 仅爬取 Python 项目
tishi scrape --dry-run          # 仅打印候选列表，不写文件
```

## 相关文档

- [LLM 分析](llm-analyzer.md) — 下游 LLM 分析
- [评分排名](analyzer.md) — 下游评分
- [数据契约](../../data/schemas/) — JSON Schema

## 相关文档

- [种子数据](../data/seed-data.md) — AI 领域关键词库
- [存储设计](storage.md) — 数据如何落库
- [GitHub 集成](../api/github-integration.md) — GitHub API 调用详情
