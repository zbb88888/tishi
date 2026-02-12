# 贡献指南

## 代码规范

### Go 代码

- 遵循 [Effective Go](https://go.dev/doc/effective_go) 和 [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- 使用 `golangci-lint` 进行代码检查
- 所有导出函数/类型必须有注释
- 使用 `context.Context` 传递请求级别数据
- 使用 `uber-go/zap` 进行结构化日志记录
- 错误处理使用 `fmt.Errorf("xxx: %w", err)` 包装

### 目录规范

```
internal/           # 私有代码，不暴露给外部
cmd/                # 可执行程序入口
pkg/                # 可被外部引用的公共包（如果有）
```

### 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 包名 | 小写，单词 | `collector`, `analyzer` |
| 文件名 | 小写，下划线分隔 | `token_rotator.go` |
| 结构体 | PascalCase | `ProjectCollector` |
| 接口 | PascalCase，-er 后缀 | `Scorer`, `Generator` |
| 方法 | PascalCase | `CalculateScore` |
| 常量 | PascalCase 或 ALL_CAPS | `MaxRetries`, `DEFAULT_TIMEOUT` |

### Astro/TypeScript 代码

- TypeScript strict mode
- 组件文件 PascalCase：`RankingTable.astro`
- 工具函数 camelCase：`fetchRankings.ts`

## Git 工作流

### 分支策略

```
main                 # 生产分支，始终可部署
├── feat/xxx         # 功能分支
├── fix/xxx          # 修复分支
├── docs/xxx         # 文档分支
└── refactor/xxx     # 重构分支
```

### 分支命名

```
feat/add-collector
feat/ranking-page
fix/rate-limit-handling
docs/update-api-spec
refactor/scorer-algorithm
```

### Commit Message 规范

使用 [Conventional Commits](https://www.conventionalcommits.org/)：

```
<type>(<scope>): <description>

[body]

[footer]
```

**Type**:

| Type | 说明 |
|------|------|
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `docs` | 文档变更 |
| `style` | 格式调整（不影响逻辑） |
| `refactor` | 重构（不改变功能） |
| `test` | 测试 |
| `chore` | 构建/工具变更 |
| `perf` | 性能优化 |

**Scope**：`collector`, `analyzer`, `server`, `web`, `db`, `scheduler`, `config`

**示例**:

```
feat(collector): add GitHub GraphQL batch query support

Implement batched GraphQL queries to fetch up to 50 repositories
in a single request, reducing API quota consumption.

Closes #12
```

## PR 流程

1. 从 `main` 创建功能分支
2. 开发并提交 commits
3. 运行测试：`make test && make lint`
4. 创建 Pull Request，填写描述：
   - 变更说明
   - 关联的 Issue
   - 测试方式
5. Code Review
6. Squash Merge 到 `main`

## 测试

### 单元测试

```bash
# 运行所有单元测试
make test

# 运行特定包的测试
go test ./internal/collector/... -v

# 生成覆盖率报告
make test-cover
```

- 测试文件与源文件同目录，命名 `xxx_test.go`
- 使用 table-driven tests
- Mock 外部依赖（GitHub API、数据库）

### 端到端测试

```bash
make test-e2e
```

- 需要运行中的 PostgreSQL 实例
- 测试完整的数据链路：采集 → 分析 → API 响应

## 相关文档

- [开发指南](development.md) — 环境搭建
- [系统架构](../architecture/system-architecture.md) — 了解项目结构
