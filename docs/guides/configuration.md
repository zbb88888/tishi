# 配置说明

## 概述

tishi 使用 Viper 管理配置，支持配置文件 + 环境变量 + 命令行参数三种方式，优先级：命令行 > 环境变量 > 配置文件 > 默认值。

## 环境变量

### .env.example

```bash
# =========================
# 数据库配置
# =========================
DB_HOST=postgres
DB_PORT=5432
DB_NAME=tishi_db
DB_USER=tishi
DB_PASSWORD=change_me_to_a_secure_password
DB_SSLMODE=disable

# 或使用完整 DSN
# DATABASE_URL=postgres://tishi:password@postgres:5432/tishi_db?sslmode=disable

# =========================
# GitHub API 配置
# =========================
# 多个 Token 用逗号分隔
GITHUB_TOKENS=ghp_token1,ghp_token2

# =========================
# 服务器配置
# =========================
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=30s

# =========================
# 日志配置
# =========================
LOG_LEVEL=info       # debug / info / warn / error
LOG_FORMAT=json      # json / console

# =========================
# 采集配置
# =========================
COLLECTOR_TOP_N=100          # 追踪项目数量
COLLECTOR_MIN_STARS=100      # 最低 Star 数
COLLECTOR_TIMEOUT=30m        # 单次采集超时

# =========================
# 分析配置
# =========================
ANALYZER_WEIGHT_DAILY_STAR=0.30
ANALYZER_WEIGHT_WEEKLY_STAR=0.25
ANALYZER_WEIGHT_FORK_RATIO=0.15
ANALYZER_WEIGHT_ISSUE_ACTIVITY=0.15
ANALYZER_WEIGHT_RECENCY=0.15

# =========================
# 站点配置
# =========================
SITE_DOMAIN=localhost
SITE_TITLE=tishi — AI 开源趋势追踪
SITE_DESCRIPTION=追踪 GitHub AI Top 100 热门项目

# =========================
# 时区
# =========================
TZ=UTC
```

## 配置文件

支持 `config.yaml`（放在项目根目录或 `/etc/tishi/`）：

```yaml
# config.yaml

database:
  host: postgres
  port: 5432
  name: tishi_db
  user: tishi
  password: ""         # 生产环境建议通过环境变量传入
  sslmode: disable
  max_open_conns: 10
  max_idle_conns: 5
  conn_max_lifetime: 1h

server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 10s
  write_timeout: 30s
  idle_timeout: 60s

github:
  tokens: []           # 生产环境建议通过环境变量传入
  search_queries:      # 可自定义搜索查询
    - "topic:llm stars:>100"
    - "topic:ai-agent stars:>100"
    - "topic:machine-learning stars:>500"
    # ... 更多查询见种子数据文档

collector:
  top_n: 100           # 追踪项目数
  min_stars: 100       # 最低 Star 数
  timeout: 30m
  retry_max: 3
  retry_backoff: 1s

analyzer:
  weights:
    daily_star: 0.30
    weekly_star: 0.25
    fork_ratio: 0.15
    issue_activity: 0.15
    recency: 0.15

scheduler:
  collect_cron: "0 0 * * *"      # 每日 00:00 UTC
  analyze_cron: "0 1 * * *"      # 每日 01:00 UTC
  build_cron: "0 2 * * *"        # 每日 02:00 UTC
  weekly_cron: "0 6 * * 0"       # 每周日 06:00 UTC
  monthly_cron: "0 6 1 * *"      # 每月 1 日 06:00 UTC

logging:
  level: info          # debug / info / warn / error
  format: json         # json / console
  output: stdout       # stdout / file
  file_path: ""        # 当 output=file 时的日志路径

site:
  domain: localhost
  title: "tishi — AI 开源趋势追踪"
  description: "追踪 GitHub AI Top 100 热门项目"
```

## 配置项详解

### 数据库配置

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|----------|--------|------|
| `database.host` | `DB_HOST` | `localhost` | 数据库地址 |
| `database.port` | `DB_PORT` | `5432` | 数据库端口 |
| `database.name` | `DB_NAME` | `tishi_db` | 数据库名称 |
| `database.user` | `DB_USER` | `tishi` | 数据库用户 |
| `database.password` | `DB_PASSWORD` | - | 数据库密码（**必填**） |
| `database.sslmode` | `DB_SSLMODE` | `disable` | SSL 模式 |
| `database.max_open_conns` | - | `10` | 最大连接数 |
| `database.max_idle_conns` | - | `5` | 最大空闲连接数 |
| `database.conn_max_lifetime` | - | `1h` | 连接最大生命周期 |

### GitHub 配置

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|----------|--------|------|
| `github.tokens` | `GITHUB_TOKENS` | - | GitHub PAT 列表（**必填**） |

### 评分权重

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|----------|--------|------|
| `analyzer.weights.daily_star` | `ANALYZER_WEIGHT_DAILY_STAR` | `0.30` | 日增 Star 权重 |
| `analyzer.weights.weekly_star` | `ANALYZER_WEIGHT_WEEKLY_STAR` | `0.25` | 周增 Star 权重 |
| `analyzer.weights.fork_ratio` | `ANALYZER_WEIGHT_FORK_RATIO` | `0.15` | Fork 率权重 |
| `analyzer.weights.issue_activity` | `ANALYZER_WEIGHT_ISSUE_ACTIVITY` | `0.15` | Issue 活跃度权重 |
| `analyzer.weights.recency` | `ANALYZER_WEIGHT_RECENCY` | `0.15` | 更新活跃度权重 |

> **注意**：所有权重之和必须为 1.0。

## 相关文档

- [开发指南](development.md) — 本地开发环境
- [部署指南](deployment.md) — 生产部署
