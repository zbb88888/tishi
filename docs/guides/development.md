# 开发指南

## 环境要求

| 工具 | 版本 | 用途 |
|------|------|------|
| Go | ≥ 1.23 | 后端 CLI 开发 |
| Node.js | ≥ 20 LTS | Astro 前端构建 |
| pnpm | ≥ 9.x | Node 包管理器 |
| Git | ≥ 2.x | 版本控制 + 数据同步 |

> **v1.0 不再需要**：PostgreSQL、Docker（开发阶段）、sqlc、golang-migrate。

## 项目结构

```
tishi/
├── cmd/
│   └── tishi/
│       └── main.go              # 入口，cobra CLI
├── internal/
│   ├── cmd/                     # cobra 子命令
│   │   ├── root.go
│   │   ├── scrape.go            # v1.0 新增
│   │   ├── analyze.go           # v1.0 新增（LLM 分析）
│   │   ├── score.go             # v1.0 新增
│   │   ├── generate.go
│   │   ├── push.go              # v1.0 新增（git push）
│   │   ├── review.go            # v1.0 新增（人工审核）
│   │   └── version.go
│   ├── scraper/                 # Trending 页面抓取（v1.0 新增）
│   │   ├── scraper.go
│   │   ├── filter.go
│   │   └── enricher.go
│   ├── llm/                     # LLM 中文分析（v1.0 新增）
│   │   ├── client.go
│   │   ├── prompt.go
│   │   └── analyzer.go
│   ├── scorer/                  # 热度评分（重构自 analyzer）
│   │   └── scorer.go
│   ├── content/                 # 内容生成（周报/月报）
│   │   ├── generator.go
│   │   └── templates.go
│   ├── datastore/               # JSON 文件存储（v1.0 新增）
│   │   └── store.go
│   └── config/                  # 配置管理
│       └── config.go
├── data/                        # JSON 数据文件（Git 管理）
│   ├── projects/
│   ├── snapshots/
│   ├── rankings/
│   ├── posts/
│   ├── schemas/
│   ├── categories.json
│   └── meta.json
├── web/                         # Astro 前端（纯 SSG）
│   ├── astro.config.mjs
│   ├── package.json
│   ├── src/
│   └── dist/                    # 构建产物
├── docs/                        # 项目文档
├── deploy/
│   └── nginx/
│       └── nginx.conf
├── Makefile
├── go.mod
├── config.yaml.example
└── README.md
```

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/zbb88888/tishi.git
cd tishi
```

### 2. 配置环境变量

```bash
cp config.yaml.example config.yaml
# 编辑 config.yaml，配置 LLM API Key 和 GitHub Token
```

最小配置：

```yaml
llm:
  provider: deepseek
  api_key: "sk-xxx"     # 必填

github:
  tokens:
    - "ghp_xxx"         # API enrichment 用
```

### 3. 检查数据目录

```bash
# data/ 目录已包含在仓库中
ls data/
# categories.json  meta.json  posts/  projects/  rankings/  schemas/  snapshots/
```

### 4. 构建后端

```bash
make build
# 或
go build -o bin/tishi ./cmd/tishi
```

### 5. 手动测试 Pipeline

```bash
# 1. 抓取 Trending + 过滤 AI 项目 + API enrichment
./bin/tishi scrape

# 2. LLM 中文分析（对新项目生成分析报告）
./bin/tishi analyze

# 3. 评分排名
./bin/tishi score

# 4. 生成周报
./bin/tishi generate --type weekly

# 5. 查看数据
ls data/projects/
cat data/rankings/$(date +%Y-%m-%d).json | jq .
```

### 6. 启动前端开发服务器

```bash
cd web
pnpm install
pnpm dev
# 访问 http://localhost:4321
```

### 7. 构建前端静态站点

```bash
cd web
pnpm build
# 产物在 web/dist/
```

## Makefile 命令

```makefile
# 构建
make build          # 构建 Go CLI 二进制
make build-web      # 构建 Astro SSG 前端

# 开发
make dev-web        # 启动前端开发服务器

# Pipeline 命令
make scrape         # 抓取 Trending
make analyze        # LLM 分析
make score          # 评分排名
make generate       # 生成内容
make push           # Git push 数据
make pipeline       # 运行完整 pipeline

# 测试
make test           # 运行单元测试
make test-e2e       # 运行端到端测试
make test-cover     # 生成覆盖率报告

# 代码质量
make lint           # 运行 golangci-lint
make fmt            # 格式化代码

# Docker（仅 Nginx 部署用）
make docker-build   # 构建 Nginx 镜像
```

## 开发工具推荐

| 工具 | 用途 |
|------|------|
| [golangci-lint](https://golangci-lint.run/) | Go 代码检查 |
| [jq](https://jqlang.github.io/jq/) | JSON 数据查看/调试 |
| [httpie](https://httpie.io/) | GitHub API 测试 |

## 相关文档

- [配置说明](configuration.md) — 所有配置项
- [贡献指南](contributing.md) — 代码规范与 PR 流程
- [系统架构](../architecture/system-architecture.md) — 架构总览
