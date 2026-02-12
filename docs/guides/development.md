# 开发指南

## 环境要求

| 工具 | 版本 | 用途 |
|------|------|------|
| Go | ≥ 1.22 | 后端开发 |
| Node.js | ≥ 20 LTS | Astro 前端构建 |
| pnpm | ≥ 9.x | Node 包管理器 |
| Docker | ≥ 24.x | 容器化运行 |
| Docker Compose | v2 | 本地服务编排 |
| PostgreSQL | ≥ 16 | 数据库（可用 Docker 启动） |
| sqlc | ≥ 1.25 | SQL → Go 代码生成 |
| golang-migrate | ≥ 4.x | 数据库迁移 |

## 项目结构

```
tishi/
├── cmd/
│   └── tishi/
│       └── main.go              # 入口，cobra CLI
├── internal/
│   ├── collector/               # 数据采集模块
│   │   ├── collector.go
│   │   ├── github_client.go
│   │   └── token_rotator.go
│   ├── analyzer/                # 趋势分析模块
│   │   ├── analyzer.go
│   │   └── scorer.go
│   ├── content/                 # 内容生成模块
│   │   ├── generator.go
│   │   └── templates.go
│   ├── server/                  # API Server
│   │   ├── server.go
│   │   ├── router.go
│   │   ├── handlers/
│   │   └── middleware/
│   ├── scheduler/               # 定时调度
│   │   └── scheduler.go
│   ├── db/                      # 数据库访问层
│   │   ├── sqlc/                # sqlc 生成的代码
│   │   ├── queries/             # SQL 查询文件
│   │   └── migrations/          # 数据库迁移文件
│   └── config/                  # 配置管理
│       └── config.go
├── web/                         # Astro 前端
│   ├── astro.config.mjs
│   ├── package.json
│   ├── src/
│   └── dist/
├── templates/                   # Go 模板文件
│   ├── weekly.md.tmpl
│   ├── monthly.md.tmpl
│   └── spotlight.md.tmpl
├── deploy/                      # 部署配置
│   ├── docker-compose.yml
│   ├── Dockerfile
│   └── nginx/
│       └── nginx.conf
├── docs/                        # 项目文档（当前目录）
├── scripts/                     # 辅助脚本
│   ├── setup.sh
│   └── seed.sh
├── Makefile
├── go.mod
├── go.sum
├── .env.example
├── .gitignore
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
cp .env.example .env
# 编辑 .env，配置 GitHub Token 和数据库密码
```

### 3. 启动数据库

```bash
# 使用 Docker 启动 PostgreSQL
docker-compose up -d postgres

# 等待数据库就绪
docker-compose exec postgres pg_isready -U tishi
```

### 4. 执行数据库迁移

```bash
make migrate-up
```

### 5. 启动后端服务

```bash
# 开发模式（热重载，需要 air）
make dev

# 或直接运行
go run ./cmd/tishi server
```

### 6. 启动前端开发服务器

```bash
cd web
pnpm install
pnpm dev
```

### 7. 手动测试采集

```bash
# 手动触发一次数据采集
go run ./cmd/tishi collect

# 手动触发一次趋势分析
go run ./cmd/tishi analyze
```

## Makefile 命令

```makefile
# 构建
make build          # 构建 Go 二进制
make build-web      # 构建 Astro 前端

# 开发
make dev            # 启动开发模式（air 热重载）
make dev-web        # 启动前端开发服务器

# 数据库
make migrate-up     # 执行迁移
make migrate-down   # 回滚迁移
make migrate-create # 创建新迁移文件
make sqlc           # 重新生成 sqlc 代码

# 测试
make test           # 运行单元测试
make test-e2e       # 运行端到端测试
make test-cover     # 生成覆盖率报告

# 代码质量
make lint           # 运行 golangci-lint
make fmt            # 格式化代码

# Docker
make docker-build   # 构建 Docker 镜像
make docker-up      # 启动所有服务
make docker-down    # 停止所有服务
make docker-logs    # 查看日志

# 手动操作
make collect        # 手动采集
make analyze        # 手动分析
make generate-weekly  # 手动生成周报
```

## 开发工具推荐

| 工具 | 用途 |
|------|------|
| [air](https://github.com/cosmtrek/air) | Go 热重载 |
| [golangci-lint](https://golangci-lint.run/) | Go 代码检查 |
| [sqlc](https://sqlc.dev/) | SQL → Go 代码生成 |
| [pgAdmin](https://www.pgadmin.org/) | PostgreSQL GUI |
| [httpie](https://httpie.io/) | API 测试 |

## 相关文档

- [配置说明](configuration.md) — 所有配置项
- [贡献指南](contributing.md) — 代码规范与 PR 流程
- [系统架构](../architecture/system-architecture.md) — 架构总览
