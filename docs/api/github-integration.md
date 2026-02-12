# GitHub API 集成策略

## 概述

tishi 依赖 GitHub API 作为唯一数据源，需要合理管理 API 调用以应对 Rate Limit 和保证数据质量。

## API 选择

### REST vs GraphQL

| 场景 | 选择 | 理由 |
|------|------|------|
| 搜索项目 | REST Search API | GraphQL 不支持 search |
| 批量获取项目详情 | GraphQL | 单次请求获取多个仓库信息，节省配额 |
| 获取单个仓库信息 | GraphQL | 字段可选，减少数据传输 |

### 使用的 API 端点

#### REST Search API

```
GET https://api.github.com/search/repositories
    ?q=topic:llm+stars:>100
    &sort=stars
    &order=desc
    &per_page=100
```

**限制**：每次搜索最多返回 1,000 条结果（前 10 页）。

#### GraphQL API

```graphql
query ($ids: [ID!]!) {
  nodes(ids: $ids) {
    ... on Repository {
      id
      nameWithOwner
      description
      primaryLanguage { name }
      licenseInfo { spdxId }
      stargazerCount
      forkCount
      issues(states: OPEN) { totalCount }
      watchers { totalCount }
      pushedAt
      repositoryTopics(first: 20) {
        nodes { topic { name } }
      }
      defaultBranchRef { name }
      homepageUrl
    }
  }
}
```

**优势**：单次请求可查询多达 100 个仓库，大幅减少请求次数。

## Rate Limit 策略

### 限额概况

| API | 方式 | 限额 | 重置周期 |
|-----|------|------|----------|
| REST Search | Token 认证 | 30 次/分钟 | 每分钟 |
| REST Core | Token 认证 | 5,000 次/小时 | 每小时 |
| GraphQL | Token 认证 | 5,000 点/小时 | 每小时 |

### 多 Token 轮换

配置多个 GitHub Personal Access Token（Fine-grained PAT），轮换使用以提高总配额：

```
# .env
GITHUB_TOKENS=ghp_token1,ghp_token2,ghp_token3
```

N 个 Token 可用配额：
- Search：N × 30 次/分钟
- Core：N × 5,000 次/小时
- GraphQL：N × 5,000 点/小时

### Token 权限要求

Fine-grained Personal Access Token，**无需任何仓库权限**：
- Metadata: Read-only（默认，搜索和查看公开仓库信息）

### 限流器实现

```go
type RateLimiter struct {
    search  *rate.Limiter  // 30 req/min → 0.5 req/sec
    core    *rate.Limiter  // 5000 req/hour → ~1.4 req/sec
    graphql *rate.Limiter  // 5000 points/hour
}

func NewRateLimiter() *RateLimiter {
    return &RateLimiter{
        search:  rate.NewLimiter(rate.Every(2*time.Second), 1),      // 0.5/sec
        core:    rate.NewLimiter(rate.Every(720*time.Millisecond), 1), // ~1.4/sec
        graphql: rate.NewLimiter(rate.Every(720*time.Millisecond), 1),
    }
}
```

### 响应头监控

每次 API 调用后检查 Rate Limit 响应头：

```
X-RateLimit-Limit: 5000
X-RateLimit-Remaining: 4950
X-RateLimit-Reset: 1707696000
X-RateLimit-Used: 50
X-RateLimit-Resource: core
```

当 `Remaining < 100` 时，切换到下一个 Token 或等待 Reset。

## 错误处理 & 重试

### 重试策略

```go
type RetryConfig struct {
    MaxRetries     int           // 最多重试 3 次
    InitialBackoff time.Duration // 初始退避：1s
    MaxBackoff     time.Duration // 最大退避：60s
    BackoffFactor  float64       // 退避因子：2.0（指数退避）
}
```

### 错误分类

| HTTP Status | 处理方式 |
|-------------|----------|
| 200 | 正常处理 |
| 304 | Not Modified - 使用缓存数据 |
| 401 | Token 无效 - 移除该 Token，切换下一个 |
| 403 Rate Limit | 读取 Reset 时间，等待后重试 |
| 403 Abuse | 等待 Retry-After 头指定的时间 |
| 404 | 仓库不存在/已删除 - 标记 is_archived |
| 422 | 搜索语法错误 - 记录错误，跳过该查询 |
| 500/502/503 | GitHub 服务端错误 - 指数退避重试 |

### 条件请求（ETag）

对重复查询使用 ETag 缓存，减少配额消耗：

```go
type CachedRequest struct {
    ETag         string
    LastModified string
    Data         []byte
}

// 请求时带上 ETag
req.Header.Set("If-None-Match", cached.ETag)

// 304 响应不消耗 Rate Limit 配额
```

## 数据采集量估算

### 单次采集（每日）

| 步骤 | API | 请求数 | 消耗 |
|------|-----|--------|------|
| 搜索 AI 项目 | REST Search | ~10 次（10 个关键词组合） | 10 search quota |
| 获取项目详情 | GraphQL | ~2 次（每次 50 个仓库） | ~200 points |
| **合计** | | ~12 次 | 单 Token 配额内 |

**结论**：单个 Token 完全够用，多 Token 仅作备份。

## Token 管理最佳实践

1. **使用 Fine-grained PAT** — 权限最小化，仅需 public repo metadata
2. **设置过期时间** — 建议 90 天，定期轮换
3. **环境变量存储** — 不提交到代码仓库
4. **监控配额消耗** — 日志记录每次请求的 Remaining 值

## 相关文档

- [数据采集](../design/collector.md) — Collector 详细设计
- [配置说明](../guides/configuration.md) — Token 配置方法
