# 配置说明

## 概述

tishi v1.0 使用 Viper 管理配置，支持配置文件 + 环境变量 + 命令行参数，优先级：命令行 > 环境变量 > 配置文件 > 默认值。

> **v1.0 变更**：移除了所有 PostgreSQL/Server/Scheduler 相关配置，新增 LLM 和数据目录配置。

## 环境变量

### .env.example

```bash
# =========================
# GitHub Token（Trending 页面不需要，但 API enrichment 需要）
# 多个 Token 逗号分隔，用于轮换
# =========================
TISHI_GITHUB_TOKENS=ghp_token1,ghp_token2

# =========================
# LLM 配置
# =========================
TISHI_LLM_PROVIDER=deepseek        # deepseek 或 qwen
TISHI_LLM_API_KEY=sk-xxx           # LLM API Key（必填）
TISHI_LLM_MODEL=deepseek-chat      # 模型名称（可选，有默认值）

# =========================
# 数据目录
# =========================
TISHI_DATA_DIR=./data              # JSON 数据文件根目录

# =========================
# 日志配置
# =========================
TISHI_LOG_LEVEL=info               # debug / info / warn / error
TISHI_LOG_FORMAT=json              # json / console

# =========================
# 评分权重
# =========================
TISHI_SCORER_WEIGHT_DAILY_STAR=0.35
TISHI_SCORER_WEIGHT_WEEKLY_STAR=0.25
TISHI_SCORER_WEIGHT_FORK_RATIO=0.15
TISHI_SCORER_WEIGHT_ISSUE_ACTIVITY=0.10
TISHI_SCORER_WEIGHT_RECENCY=0.15

# =========================
# Git 同步（可选，用于 push 子命令）
# =========================
TISHI_GIT_REMOTE=origin
TISHI_GIT_BRANCH=main

# =========================
# 时区
# =========================
TZ=UTC
```

## 配置文件

支持 `config.yaml`（放在项目根目录或 `/etc/tishi/`）：

```yaml
# config.yaml

# GitHub Token 列表（Trending 抓取不需要，API enrichment 需要）
github:
  tokens: []               # 生产环境建议通过环境变量传入

# LLM 配置
llm:
  provider: deepseek       # deepseek 或 qwen
  api_key: ""              # 生产环境建议通过环境变量传入
  model: ""                # 留空使用默认值
  max_tokens: 2000         # 最大输出 token
  temperature: 0.3         # 生成温度
  timeout: 60s             # 单次请求超时
  retry_max: 3             # 最大重试次数

# 数据目录
data:
  dir: "./data"            # JSON 数据文件根目录

# Scraper 配置
scraper:
  trending_url: "https://github.com/trending"
  languages:               # 抓取的编程语言页面（空=所有语言）
    - ""                   # 总榜
    - "python"
    - "typescript"
  timeout: 30s             # 单次请求超时
  delay: 2s                # 请求间隔（避免被封）

# Scorer 评分权重
scorer:
  weights:
    daily_star: 0.35
    weekly_star: 0.25
    fork_ratio: 0.15
    issue_activity: 0.10
    recency: 0.15

# Git 同步
git:
  remote: origin
  branch: main
  auto_push: true          # scrape 完成后是否自动 push

# 日志
logging:
  level: info              # debug / info / warn / error
  format: json             # json / console
  output: stdout           # stdout / file
```

## 配置项详解

### LLM 配置

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|----------|--------|------|
| `llm.provider` | `TISHI_LLM_PROVIDER` | `deepseek` | LLM 提供商 |
| `llm.api_key` | `TISHI_LLM_API_KEY` | - | API Key（**必填**） |
| `llm.model` | `TISHI_LLM_MODEL` | 按 provider | 模型名称 |
| `llm.max_tokens` | - | `2000` | 最大输出 token |
| `llm.temperature` | - | `0.3` | 生成温度 |
| `llm.timeout` | - | `60s` | 单次请求超时 |

#### Provider 默认值

| Provider | Base URL | 默认 Model |
|----------|----------|-----------|
| `deepseek` | `https://api.deepseek.com/v1` | `deepseek-chat` |
| `qwen` | `https://dashscope.aliyuncs.com/compatible-mode/v1` | `qwen-plus` |

### GitHub 配置

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|----------|--------|------|
| `github.tokens` | `TISHI_GITHUB_TOKENS` | - | GitHub PAT 列表（逗号分隔） |

> **注意**：抓取 Trending HTML 页面不需要 Token。Token 仅用于 GitHub REST API enrichment（获取 README、topics、详细指标）。单个 Token 配额完全够用。

### 评分权重

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|----------|--------|------|
| `scorer.weights.daily_star` | `TISHI_SCORER_WEIGHT_DAILY_STAR` | `0.35` | 日增 Star 权重 |
| `scorer.weights.weekly_star` | `TISHI_SCORER_WEIGHT_WEEKLY_STAR` | `0.25` | 周增 Star 权重 |
| `scorer.weights.fork_ratio` | `TISHI_SCORER_WEIGHT_FORK_RATIO` | `0.15` | Fork 率权重 |
| `scorer.weights.issue_activity` | `TISHI_SCORER_WEIGHT_ISSUE_ACTIVITY` | `0.10` | Issue 活跃度权重 |
| `scorer.weights.recency` | `TISHI_SCORER_WEIGHT_RECENCY` | `0.15` | 更新活跃度权重 |

> **注意**：所有权重之和必须为 1.0。

### 数据目录

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|----------|--------|------|
| `data.dir` | `TISHI_DATA_DIR` | `./data` | JSON 数据文件根目录 |

目录结构：

```
{data.dir}/
├── projects/       # 项目 JSON 文件
├── snapshots/      # 每日快照 JSONL
├── rankings/       # 每日排行榜
├── posts/          # 博客文章
├── schemas/        # JSON Schema 定义
├── categories.json # 分类定义
└── meta.json       # 元数据
```

## v0.x 已废弃的配置

以下配置项在 v1.0 中已移除，Phase 4 清理阶段会删除相关代码：

| 配置项 | 说明 | 替代方案 |
|--------|------|---------|
| `database.*` | PostgreSQL 配置 | JSON 文件存储 |
| `server.*` | HTTP 服务器配置 | Astro SSG 静态页面 |
| `scheduler.*` | 内置 cron 配置 | 系统 crontab |
| `collector.search_queries` | GitHub Search API 查询 | Trending HTML 抓取 |

## 相关文档

- [开发指南](development.md) — 本地开发环境
- [部署指南](deployment.md) — 生产部署
