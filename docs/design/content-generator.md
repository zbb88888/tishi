# 内容生成模块设计（Content Generator）

## 概述

Content Generator 基于 data/projects/ 和 data/rankings/ 中的数据，自动生成博客文章 JSON，输出到 `data/posts/*.json`。

## 文章类型

| 类型 | 频率 | 内容 |
|------|------|------|
| **weekly** | 每周 | AI 开源周报：本周 Star 增长 Top 10、新入榜项目、排名变动 |
| **monthly** | 每月 | AI 开源月报：月度趋势、分类统计、年度对比 |
| **spotlight** | 不定期 | 项目深度解读：基于 LLM 分析的完整项目报告 |

## 输出格式

博客文章以 JSON 文件存储在 `data/posts/`，符合 `data/schemas/post.schema.json`：

```json
{
  "slug": "ai-weekly-2025-w29",
  "title": "AI 开源周报 #29 | 2025-07-14 ~ 2025-07-20",
  "content": "## 本周概览\n\n本周 Trending 共收录 35 个 AI 项目...",
  "post_type": "weekly",
  "published_at": "2025-07-20T06:00:00Z",
  "projects": ["langchain-ai__langchain", "ollama__ollama"],
  "metadata": {
    "new_entries": 5,
    "top_gainer": "example__project",
    "total_projects": 35
  }
}
```

## 模板系统

使用 Go `text/template` 渲染 Markdown 内容：

### 周报模板

```markdown
## 本周概览

本周 AI Trending 共收录 {{.TotalProjects}} 个项目，{{.NewEntries}} 个新入榜。

## 🔥 Star 增长 Top 10

| 排名 | 项目 | 语言 | 周增 Star | 总 Star | 分类 |
|------|------|------|-----------|---------|------|
{{range .TopGainers}}| {{.Rank}} | {{.FullName}} | {{.Language}} | +{{.WeeklyStars}} | {{.Stars}} | {{.Category}} |
{{end}}

## 🆕 新入榜项目
{{range .NewProjects}}
### {{.FullName}}

> {{.Summary}}

⭐ {{.Stars}} | 🍴 {{.Forks}} | 📝 {{.Language}} | 🏷️ {{.Categories}}
{{end}}
```

## 数据查询

Content Generator 运行时读取本地 JSON 文件（非数据库）：

```go
func (g *Generator) loadWeeklyData(date time.Time) (*WeeklyData, error) {
    // 1. 读取本周 7 天的 rankings
    rankings := g.loadRankings(date.AddDate(0, 0, -7), date)

    // 2. 读取所有 projects (已有 analysis.status == "published" 的)
    projects := g.loadPublishedProjects()

    // 3. 计算周增 Star = 最新快照 - 7天前快照
    // 4. 找出新入榜项目
    // 5. 排名变动统计
    return &WeeklyData{...}, nil
}
```

## Slug 生成规则

```
weekly:    ai-weekly-2025-w29
monthly:   ai-monthly-2025-07
spotlight: spotlight-langchain-ai-langchain
```

## CLI 命令

```bash
tishi generate                     # 生成所有到期的文章
tishi generate --type=weekly       # 仅生成周报
tishi generate --type=spotlight --id=owner__repo  # 为指定项目生成 Spotlight
tishi generate --dry-run           # 仅打印内容，不写文件
```

## 相关文档

- [评分排名](analyzer.md) — 排行榜数据来源
- [LLM 分析](llm-analyzer.md) — Spotlight 文章的项目分析来源
- [前端展示](web-frontend.md) — 博客页面展示
