# 趋势分析模块设计（Analyzer）

## 概述

Analyzer 在每日数据采集完成后运行，负责计算热度评分、生成排名、检测趋势变化、自动分类打标。

## 热度评分模型

### 评分公式

$$
\text{Score} = w_1 \cdot S_{daily} + w_2 \cdot S_{weekly} + w_3 \cdot F_{ratio} + w_4 \cdot I_{activity} + w_5 \cdot R_{recency}
$$

各分量标准化到 0-100 后加权求和，最终评分范围 0-100。

### 指标定义与权重

| 指标 | 符号 | 计算方式 | 权重 | 说明 |
|------|------|----------|------|------|
| 日增 Star | $S_{daily}$ | 今日 Star - 昨日 Star | 0.30 | 最能反映当前热度 |
| 周增 Star | $S_{weekly}$ | 本周 Star - 上周 Star | 0.25 | 平滑短期波动 |
| Fork 活跃率 | $F_{ratio}$ | forks / stars | 0.15 | 反映项目实用性 |
| Issue 活跃度 | $I_{activity}$ | open_issues / (age_days + 1) | 0.15 | 反映社区互动 |
| 更新活跃度 | $R_{recency}$ | 1 / (now - pushed_at).days | 0.15 | 反映维护活跃度 |

### 标准化方法

每个原始指标通过 **Min-Max 标准化** 映射到 [0, 100]：

$$
X_{norm} = \frac{X - X_{min}}{X_{max} - X_{min}} \times 100
$$

> 对于 $S_{daily}$ 和 $S_{weekly}$，使用当日所有项目中的 min/max 值。

### 权重调优

权重初始值基于经验设定，后续可通过以下方式调优：
- 观察排行榜与 GitHub Trending 的吻合度
- 分析用户反馈（如果采集的话）
- A/B 比较不同权重配置

## 排名生成

```sql
-- 基于评分生成排名
WITH scored AS (
    SELECT id,
           score,
           ROW_NUMBER() OVER (ORDER BY score DESC) AS new_rank
    FROM projects
    WHERE is_archived = FALSE
)
UPDATE projects p
SET rank = s.new_rank,
    updated_at = NOW()
FROM scored s
WHERE p.id = s.id;
```

## 趋势变动检测

### 排名变动

```go
type RankChange struct {
    ProjectID  int64
    OldRank    int
    NewRank    int
    Direction  string  // "up" / "down" / "new" / "out"
    Delta      int     // 排名变化绝对值
}
```

检测规则：
- **新入榜**：昨日不在 Top 100，今日进入
- **跌出榜**：昨日在 Top 100，今日未进入
- **大幅上升**：排名上升 ≥ 10 位
- **大幅下降**：排名下降 ≥ 10 位

### 异常检测

**Star 爆发检测**：日增 Star 超过过去 7 天日均的 3 倍标准差。

```go
// detectStarBurst 检测 Star 异常增长
func detectStarBurst(projectID int64, dailyGain int, avgGain float64, stdDev float64) bool {
    threshold := avgGain + 3*stdDev
    return float64(dailyGain) > threshold
}
```

## 自动分类

### 分类策略

1. **基于 Topics 标签**：GitHub 项目的 topics 字段直接映射到预定义分类
2. **基于关键词匹配**：项目 description 和 full_name 中的关键词匹配
3. **多分类支持**：一个项目可属于多个分类（如既是 LLM 又是 Agent）

### 分类映射规则

```go
var categoryRules = map[string][]string{
    "llm":        {"llm", "large-language-model", "gpt", "llama", "mistral", "chatbot"},
    "agent":      {"ai-agent", "autonomous-agent", "agent-framework", "agentic"},
    "rag":        {"rag", "retrieval-augmented", "vector-search", "embedding"},
    "diffusion":  {"diffusion", "stable-diffusion", "text-to-image", "image-generation"},
    "mlops":      {"mlops", "ml-pipeline", "model-serving", "feature-store"},
    "vector-db":  {"vector-database", "vector-store", "similarity-search"},
    "framework":  {"deep-learning", "machine-learning", "neural-network", "pytorch", "tensorflow"},
    "tool":       {"ai-tool", "ai-assistant", "copilot", "code-generation"},
}
```

### 置信度

- Topics 精确匹配：confidence = 1.0
- Description 关键词匹配：confidence = 0.8
- full_name 包含关键词：confidence = 0.6

## 执行流程

```
Analyzer.Run()
  │
  ├── 1. 计算每个项目的各项指标
  │      ├── 从 daily_snapshots 计算日增/周增 Star
  │      ├── 查询 projects 获取 Fork/Issue/Push 数据
  │      └── 标准化所有指标
  │
  ├── 2. 计算加权评分
  │      └── 写入 projects.score + daily_snapshots.score
  │
  ├── 3. 生成排名
  │      └── 写入 projects.rank + daily_snapshots.rank
  │
  ├── 4. 检测趋势变动
  │      ├── 排名变动检测
  │      └── Star 异常检测
  │
  ├── 5. 自动分类打标
  │      └── 写入 project_categories
  │
  └── 6. 记录分析日志
```

## 相关文档

- [数据采集](collector.md) — 上游数据来源
- [内容生成](content-generator.md) — 下游消费分析结果
- [存储设计](storage.md) — 分析结果存储
