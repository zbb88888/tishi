# GitHub 数据获取策略

## 概述

v1.0 采用**两阶段数据获取**：

1. **Trending HTML 抓取** — Colly 解析 `github.com/trending` 页面，无需 Token
2. **REST API Enrichment** — 对已过滤的 AI 项目调用 GitHub API 获取详细信息，需要 Token

> **v0.x 使用 GitHub Search API 主动搜索仓库，已废弃。** v1.0 仅从 Trending 页面被动获取。

## 阶段 1：Trending HTML 抓取

### 目标 URL

```
https://github.com/trending                    # 总榜
https://github.com/trending/python             # Python
https://github.com/trending/typescript         # TypeScript
https://github.com/trending?since=daily        # 每日（默认）
https://github.com/trending?since=weekly       # 每周
```

### CSS 选择器

| 选择器 | 提取内容 |
|--------|---------|
| `article.Box-row` | 每个 Trending 项目行 |
| `article.Box-row h2 a` | 仓库全名 `owner/repo` |
| `article.Box-row p` | 项目描述 |
| `article.Box-row span[itemprop="programmingLanguage"]` | 编程语言 |
| `article.Box-row .Link--muted:last-of-type` | 当日 Star 增长 |

### 提取数据

```go
type TrendingItem struct {
    FullName    string // "owner/repo"
    Description string // 项目描述
    Language    string // 编程语言
    Stars       int    // 总 Star 数
    Forks       int    // 总 Fork 数
    PeriodStars int    // 当日/周 Star 增长
}
```

### 抓取限制

- **无需 Token** — HTML 页面公开访问
- **请求频率** — 建议间隔 2-3 秒，避免被 GitHub 限流
- **数据量** — 每页 25 个项目，3 个语言页面 ≈ 75 个项目/次
- **User-Agent** — 设置合理的 UA，避免被识别为爬虫

```go
c := colly.NewCollector(
    colly.UserAgent("Mozilla/5.0 (compatible; tishi/1.0)"),
)
c.Limit(&colly.LimitRule{
    DomainGlob: "*github.com*",
    Delay:      2 * time.Second,
})
```

## 阶段 2：REST API Enrichment

对阶段 1 过滤后的 AI 项目，调用 GitHub REST API 获取完整信息。

### 使用的 API 端点

#### 获取仓库详情

```
GET https://api.github.com/repos/{owner}/{repo}
```

返回：stars, forks, open_issues, watchers, license, topics, homepage, created_at, pushed_at 等完整信息。

#### 获取 README

```
GET https://api.github.com/repos/{owner}/{repo}/readme
Accept: application/vnd.github.raw+json
```

返回 README 原始内容，用作 LLM 分析输入。

#### 获取 Topics

```
GET https://api.github.com/repos/{owner}/{repo}/topics
Accept: application/vnd.github.mercy-preview+json
```

返回 topics 数组，用于 AI 分类匹配。

## Rate Limit 策略

### 限额概况

| API | 限额 | 重置周期 |
|-----|------|----------|
| REST Core（无 Token） | 60 次/小时 | 每小时 |
| REST Core（Token 认证） | 5,000 次/小时 | 每小时 |
| Trending HTML | 无官方限制 | 建议 2s 间隔 |

### 多 Token 轮换

```bash
# .env
TISHI_GITHUB_TOKENS=ghp_token1,ghp_token2
```

N 个 Token 可用配额：N × 5,000 次/小时。

### Token 权限要求

Fine-grained Personal Access Token，**无需任何仓库写权限**：

- Public Repositories: Read-only（默认）

### 使用量估算

| 步骤 | 请求数/次 | 说明 |
|------|----------|------|
| Trending 页面抓取 | 3-6 次 | HTML，无需 Token |
| Repo 详情 | ~30 次 | 过滤后的 AI 项目数 |
| README 获取 | ~30 次 | 每个 AI 项目获取一次 README |
| **合计** | ~65 次 | 远低于单 Token 5000 配额 |

**结论**：单个 Token 完全够用。多 Token 仅作备份和容错。

### 限流器实现

```go
type TokenRotator struct {
    tokens  []string
    current int
    mu      sync.Mutex
    limiter *rate.Limiter // 1 req/sec 安全速率
}

func (r *TokenRotator) Next() string {
    r.mu.Lock()
    defer r.mu.Unlock()
    token := r.tokens[r.current]
    r.current = (r.current + 1) % len(r.tokens)
    return token
}
```

## 错误处理

| HTTP Status | 处理方式 |
|-------------|----------|
| 200 | 正常处理 |
| 304 | Not Modified（ETag 缓存命中） |
| 403 Rate Limit | 切换 Token 或等待 Reset |
| 404 | 仓库不存在/已删除 — 跳过 |
| 500/502/503 | 指数退避重试（最多 3 次） |

### ETag 缓存

对 API enrichment 使用 ETag 条件请求，304 响应不消耗配额：

```go
req.Header.Set("If-None-Match", cachedETag)
// HTTP 304 → 使用缓存数据，不消耗 Rate Limit
```

## v0.x 已废弃

| v0.x 内容 | v1.0 替代 |
|-----------|----------|
| GitHub Search API (`/search/repositories`) | Trending HTML 抓取 |
| GraphQL API 批量查询 | REST API 逐个 enrichment |
| `github.search_queries` 配置 | `scraper.languages` 配置 |

## 相关文档

- [Scraper 设计](../design/collector.md) — Colly 抓取实现
- [配置说明](../guides/configuration.md) — Token 配置方法
