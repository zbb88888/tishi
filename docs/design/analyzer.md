# 评分排名模块设计（Scorer / Analyzer）

## 概述

Analyzer 读取 `data/projects/*.json` 和 `data/snapshots/*.jsonl`，计算多维加权评分，生成每日排行榜 `data/rankings/{date}.json`。

## 热度评分模型

### 评分公式

$$
\text{Score} = w_1 \cdot S_{daily} + w_2 \cdot S_{weekly} + w_3 \cdot F_{ratio} + w_4 \cdot I_{activity} + w_5 \cdot R_{recency}
$$

各分量标准化到 0-100 后加权求和，最终评分范围 0-100。

### 指标定义与权重

| 指标 | 符号 | 计算方式 | 权重 | 说明 |
|------|------|----------|------|------|
| 日增 Star | $S_{daily}$ | trending.daily_stars | 0.35 | 最能反映当前热度 |
| 周增 Star | $S_{weekly}$ | trending.weekly_stars | 0.25 | 平滑短期波动 |
| Fork 活跃率 | $F_{ratio}$ | forks / stars | 0.15 | 反映项目实用性 |
| Issue 活跃度 | $I_{activity}$ | open_issues / (age_days + 1) | 0.10 | 反映社区互动 |
| 更新活跃度 | $R_{recency}$ | 1 / (now - pushed_at).days | 0.15 | 反映维护活跃度 |

### 标准化方法

Min-Max 标准化映射到 [0, 100]：

$$
X_{norm} = \frac{X - X_{min}}{X_{max} - X_{min}} \times 100
$$

## 排名生成

```go
// GenerateRanking 读取所有项目，计算评分，输出排行榜 JSON
func (a *Analyzer) GenerateRanking(ctx context.Context, date string) error {
    projects := a.loadAllProjects()         // data/projects/*.json
    snapshots := a.loadSnapshots(date)       // data/snapshots/{date}.jsonl

    scored := make([]ScoredProject, 0, len(projects))
    for _, p := range projects {
        score := a.calculateScore(p, snapshots)
        scored = append(scored, ScoredProject{Project: p, Score: score})
    }

    sort.Slice(scored, func(i, j int) bool {
        return scored[i].Score > scored[j].Score
    })

    // 生成 ranking JSON
    ranking := Ranking{
        Date:  date,
        Total: len(scored),
        Items: make([]RankingItem, 0),
    }
    for rank, sp := range scored {
        ranking.Items = append(ranking.Items, RankingItem{
            Rank:       rank + 1,
            ProjectID:  sp.Project.ID,
            FullName:   sp.Project.FullName,
            Score:      sp.Score,
            DailyStars: sp.Project.Trending.DailyStars,
            // ... 其他字段
        })
    }

    return a.writeRanking(date, ranking)  // data/rankings/{date}.json
}
```

## 排名变动检测

比较今日 ranking 和昨日 ranking，计算 `rank_change`：

```go
type RankChange struct {
    ProjectID  string
    OldRank    int
    NewRank    int
    Direction  string  // "up" / "down" / "new" / "out"
    Delta      int
}
```

## 执行流程

```
tishi score
  │
  ├── 1. 读取 data/projects/*.json
  │
  ├── 2. 读取 data/snapshots/ (最近 7 天)
  │
  ├── 3. 计算每个项目的标准化指标
  │
  ├── 4. 加权评分
  │
  ├── 5. 排序 → 生成排名
  │
  ├── 6. 与昨日排名对比 → rank_change
  │
  ├── 7. 更新 projects/*.json 中的 score/rank 字段
  │
  └── 8. 写入 data/rankings/{date}.json
```

## CLI 命令

```bash
tishi score                    # 计算今日评分和排名
tishi score --date=2025-07-14  # 计算指定日期
tishi score --dry-run          # 仅打印评分结果，不写文件
```

## 相关文档

- [数据采集](collector.md) — 上游数据来源
- [LLM 分析](llm-analyzer.md) — 并行的 LLM 分析
- [内容生成](content-generator.md) — 下游博客生成
