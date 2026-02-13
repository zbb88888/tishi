# 定时调度设计（Scheduler）

> **⚠️ v1.0 架构变更** — v1.0 不再使用内置的 `robfig/cron` 调度器。改为外部 cron（系统 crontab）触发 CLI 命令。本文档描述新的调度方案。

## v1.0 调度方案

Stage 1（采集+分析）通过系统 crontab 触发 CLI 命令链：

```cron
# Machine A: 每日 UTC 00:00 运行 Stage 1 完整 pipeline
0 0 * * * cd /opt/tishi && ./run-pipeline.sh >> /var/log/tishi/pipeline.log 2>&1
```

### Pipeline 脚本

```bash
#!/bin/bash
# run-pipeline.sh — Stage 1 每日 pipeline
set -euo pipefail

DATE=$(date -u +%Y-%m-%d)
LOG_PREFIX="[${DATE}]"

echo "${LOG_PREFIX} Starting daily pipeline..."

# 1. 爬取 Trending + AI 过滤 + GitHub API 补充
echo "${LOG_PREFIX} Step 1: Scraping..."
./tishi scrape --since=daily 2>&1

# 2. LLM 分析未分析的新项目
echo "${LOG_PREFIX} Step 2: LLM Analysis..."
./tishi analyze 2>&1

# 3. 评分排名
echo "${LOG_PREFIX} Step 3: Scoring..."
./tishi score 2>&1

# 4. 生成博客文章（如果是周日/月初）
DOW=$(date -u +%u)
DOM=$(date -u +%d)
if [ "$DOW" = "7" ]; then
    echo "${LOG_PREFIX} Step 4a: Generating weekly report..."
    ./tishi generate --type=weekly 2>&1
fi
if [ "$DOM" = "01" ]; then
    echo "${LOG_PREFIX} Step 4b: Generating monthly report..."
    ./tishi generate --type=monthly 2>&1
fi

# 5. Git push
echo "${LOG_PREFIX} Step 5: Pushing data..."
./tishi push 2>&1

echo "${LOG_PREFIX} Pipeline completed."
```

### Stage 2 调度

```cron
# Machine B: 每日 UTC 01:00 (给 Stage 1 一小时完成)
0 1 * * * cd /opt/tishi && git pull && cd web && npm run build && rsync -avz dist/ machineC:/var/www/tishi/
```

## 与 v0.x 对比

| 维度 | v0.x | v1.0 |
|------|------|------|
| 调度方式 | `robfig/cron/v3` 内置 | 系统 crontab 外部 |
| 进程模型 | Go 常驻进程 + 内置 cron | CLI 按需执行 |
| 任务编排 | Go 代码串行调用 | Shell 脚本串行 |
| 任务锁 | Go sync.Mutex | flock(1) 文件锁 |
| 依赖 | robfig/cron 库 | 系统 crontab (零依赖) |

## 任务锁（防重复执行）

使用 `flock` 文件锁防止任务重叠：

```bash
# crontab 中使用 flock
0 0 * * * flock -n /tmp/tishi-pipeline.lock /opt/tishi/run-pipeline.sh
```

## 超时控制

| 步骤 | 超时 | 说明 |
|------|------|------|
| tishi scrape | 15min | Colly 爬取 + GitHub API |
| tishi analyze | 30min | LLM API 调用（每项目 ~10s） |
| tishi score | 1min | 本地 JSON 计算 |
| tishi generate | 1min | 模板渲染 |
| tishi push | 2min | Git push |

可通过 `timeout` 命令控制：

```bash
timeout 15m ./tishi scrape
timeout 30m ./tishi analyze
```

## Phase 4 清理计划

将移除：

- `internal/scheduler/` — 整个调度器包
- `robfig/cron/v3` 依赖
- `internal/cmd/server.go` 中的 scheduler 启动逻辑

## 相关文档

- [数据采集](collector.md) — Scraper 详设
- [LLM 分析](llm-analyzer.md) — LLM 分析详设
- [评分排名](analyzer.md) — Scorer 详设
- [部署拓扑](../architecture/deployment-topology.md) — cron 配置
