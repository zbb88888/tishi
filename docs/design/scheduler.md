# 定时调度模块设计（Scheduler）

## 概述

Scheduler 基于 `robfig/cron/v3`，内嵌在 tishi 主进程中，负责编排数据采集、趋势分析、内容生成等定时任务。

## 任务清单

| 任务 | Cron 表达式 | UTC 时间 | 说明 |
|------|-------------|----------|------|
| 数据采集 | `0 0 * * *` | 每日 00:00 | 采集 GitHub AI 项目数据 |
| 趋势分析 | `0 1 * * *` | 每日 01:00 | 计算评分、排名、趋势检测 |
| 前端重建 | `0 2 * * *` | 每日 02:00 | 触发 Astro SSG 重新构建 |
| 周报生成 | `0 6 * * 0` | 每周日 06:00 | 生成 AI 开源周报 |
| 月报生成 | `0 6 1 * *` | 每月 1 日 06:00 | 生成 AI 开源月报 |

## 任务编排

```
每日任务链（串行执行）：

00:00 ─── Collector ──── 01:00 ─── Analyzer ──── 02:00 ─── Astro Build
  │         ▲                         ▲                       ▲
  │         │                         │                       │
  │    采集 GitHub 数据          计算评分排名            重建静态页面
  │    写入 DB                   趋势检测
  │
  └── 如采集失败，Analyzer 使用上一次数据运行

周报/月报（独立触发，不阻塞每日任务）：

周日 06:00 ─── Content Generator (weekly)
月初 06:00 ─── Content Generator (monthly)
```

## 实现

```go
package scheduler

import (
    "context"
    "github.com/robfig/cron/v3"
    "go.uber.org/zap"
)

type Scheduler struct {
    cron      *cron.Cron
    collector *collector.Collector
    analyzer  *analyzer.Analyzer
    generator *content.Generator
    logger    *zap.Logger
}

func New(opts ...Option) *Scheduler {
    s := &Scheduler{
        cron: cron.New(cron.WithSeconds(), cron.WithLogger(cronLogger)),
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

func (s *Scheduler) Start(ctx context.Context) error {
    // 每日采集 → 分析 → 构建（串行链）
    s.cron.AddFunc("0 0 0 * * *", s.dailyPipeline)

    // 周报
    s.cron.AddFunc("0 0 6 * * 0", s.weeklyReport)

    // 月报
    s.cron.AddFunc("0 0 6 1 * *", s.monthlyReport)

    s.cron.Start()
    s.logger.Info("scheduler started")

    <-ctx.Done()
    s.cron.Stop()
    return nil
}

func (s *Scheduler) dailyPipeline() {
    ctx := context.Background()

    // 1. 采集
    s.logger.Info("daily pipeline: starting collection")
    if err := s.collector.Run(ctx); err != nil {
        s.logger.Error("collection failed", zap.Error(err))
        // 采集失败不阻塞分析（使用上次数据）
    }

    // 2. 分析
    s.logger.Info("daily pipeline: starting analysis")
    if err := s.analyzer.Run(ctx); err != nil {
        s.logger.Error("analysis failed", zap.Error(err))
        return
    }

    // 3. 触发 Astro 重建
    s.logger.Info("daily pipeline: triggering SSG rebuild")
    if err := s.triggerBuild(ctx); err != nil {
        s.logger.Error("SSG rebuild failed", zap.Error(err))
    }

    s.logger.Info("daily pipeline: completed")
}
```

## 任务锁

防止任务重叠执行（如上一次采集未完成，新的 cron 又触发）：

```go
type TaskLock struct {
    mu      sync.Mutex
    running map[string]bool
}

func (l *TaskLock) TryLock(taskName string) bool {
    l.mu.Lock()
    defer l.mu.Unlock()
    if l.running[taskName] {
        return false // 已在运行
    }
    l.running[taskName] = true
    return true
}

func (l *TaskLock) Unlock(taskName string) {
    l.mu.Lock()
    defer l.mu.Unlock()
    delete(l.running, taskName)
}
```

## 手动触发

除定时执行外，所有任务支持通过 CLI 手动触发：

```bash
tishi collect    # 手动采集
tishi analyze    # 手动分析
tishi generate weekly   # 手动生成周报
tishi generate monthly  # 手动生成月报
```

## 超时控制

| 任务 | 超时 | 说明 |
|------|------|------|
| Collector | 30min | GitHub API 受 Rate Limit 影响 |
| Analyzer | 5min | 纯数据库计算 |
| Content Generator | 5min | 查询 + 渲染 |
| Astro Build | 10min | Node.js 构建 |

## 可观测性

每次任务执行记录：
- 开始时间、结束时间、耗时
- 成功/失败状态
- 错误信息（如有）

```go
s.logger.Info("task completed",
    zap.String("task", "collector"),
    zap.Duration("duration", elapsed),
    zap.Int("projects_collected", count),
    zap.Error(err),
)
```

## 相关文档

- [数据采集](collector.md) — Collector 详设
- [趋势分析](analyzer.md) — Analyzer 详设
- [内容生成](content-generator.md) — Generator 详设
