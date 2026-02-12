# 数据采集模块设计（Collector）

## 概述

Collector 负责定时从 GitHub API 采集 AI 相关项目数据，是 tishi 数据链路的源头。

## 采集策略

### 数据源

| 数据源 | API | 用途 |
|--------|-----|------|
| GitHub Search API | `GET /search/repositories` | 按关键词搜索 AI 项目，获取候选列表 |
| GitHub GraphQL API | `POST /graphql` | 批量获取项目详细信息（减少请求次数） |

### 搜索策略

**关键词搜索**：使用预定义的 AI 领域关键词库（详见 [种子数据](../data/seed-data.md)），组合多个搜索查询：

```
# 搜索查询示例
topic:llm stars:>100
topic:machine-learning stars:>500
topic:deep-learning stars:>500
topic:ai-agent stars:>100
"large language model" in:description stars:>200
"retrieval augmented generation" in:readme stars:>50
```

**排序**：按 `stars` 降序，每个查询取前 100 条。

**去重合并**：所有查询结果按 `repo_id` 去重，最终保留 Star 数最高的 Top 100+（预留 buffer）。

### 采集字段

| 字段 | 来源 | 说明 |
|------|------|------|
| `github_id` | Search API | GitHub 仓库唯一 ID |
| `full_name` | Search API | `owner/repo` 格式 |
| `description` | Search API | 项目描述 |
| `language` | Search API | 主要编程语言 |
| `license` | Search API | 开源协议 |
| `created_at` | Search API | 仓库创建时间 |
| `stargazers_count` | GraphQL | 当前 Star 数 |
| `forks_count` | GraphQL | 当前 Fork 数 |
| `open_issues_count` | GraphQL | 当前 Open Issue 数 |
| `watchers_count` | GraphQL | 关注者数 |
| `pushed_at` | GraphQL | 最后推送时间 |
| `topics` | GraphQL | 项目标签列表 |
| `default_branch` | GraphQL | 默认分支 |
| `homepage` | GraphQL | 项目主页 URL |

## GitHub API Rate Limit 应对

### Rate Limit 概况

| API 类型 | 限制 | 说明 |
|----------|------|------|
| REST Search API | 30 次/分钟（认证） | 搜索专用限制 |
| REST Core API | 5,000 次/小时（认证） | 通用 API |
| GraphQL API | 5,000 点/小时 | 按复杂度计费 |

### 应对策略

1. **Token 轮换** — 配置多个 GitHub Personal Access Token，轮流使用
2. **请求节流** — 实现令牌桶限流器，确保不超过 Rate Limit
3. **GraphQL 批量查询** — 单次 GraphQL 请求获取多个仓库信息，减少请求次数
4. **条件请求** — 使用 `If-None-Match` / `ETag` 头，减少不必要的数据传输
5. **退避重试** — 收到 `403 rate limit exceeded` 时，读取 `X-RateLimit-Reset` 等待后重试

### Token 管理

```go
// TokenRotator 管理多个 GitHub Token，轮换使用
type TokenRotator struct {
    tokens    []string
    current   int
    mu        sync.Mutex
    limiters  map[string]*rate.Limiter  // 每个 token 独立限流
}

func (r *TokenRotator) Next() string {
    r.mu.Lock()
    defer r.mu.Unlock()
    token := r.tokens[r.current]
    r.current = (r.current + 1) % len(r.tokens)
    return token
}
```

## 增量更新策略

- **项目列表**：每次全量搜索，但 Upsert 到 `projects` 表（基于 `github_id`）
- **每日快照**：每日 Insert 一条新记录到 `daily_snapshots`，不更新历史数据
- **新项目检测**：首次入库的项目标记 `first_seen_at`，用于"新项目速递"

## 错误处理

| 错误类型 | 处理方式 |
|----------|----------|
| 网络超时 | 指数退避重试（最多 3 次） |
| Rate Limit (403) | 读取 `X-RateLimit-Reset`，等待后重试 |
| 仓库不存在 (404) | 跳过，标记 `projects.is_archived = true` |
| API 降级 (500/502/503) | 指数退避重试，超过阈值告警 |
| 数据校验失败 | 记录日志，跳过该条数据 |

## 调度

- **频率**：每日 00:00 UTC 执行一次
- **超时**：单次采集最长 30 分钟
- **幂等性**：重复执行同一天的采集不会产生重复快照（`unique(project_id, snapshot_date)`）

## 可观测性

- **日志**：每次采集记录开始时间、结束时间、采集数量、错误数
- **指标**（future）：
  - `collector_projects_total` — 采集到的项目总数
  - `collector_errors_total` — 错误计数
  - `collector_duration_seconds` — 采集耗时
  - `collector_github_rate_limit_remaining` — 剩余配额

## 相关文档

- [种子数据](../data/seed-data.md) — AI 领域关键词库
- [存储设计](storage.md) — 数据如何落库
- [GitHub 集成](../api/github-integration.md) — GitHub API 调用详情
