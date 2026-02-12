# 内容生成模块设计（Content Generator）

## 概述

Content Generator 基于 Analyzer 的分析结果，自动生成 Markdown 格式的博客文章，包括周报、月报和新项目速递。

## 文章类型

| 类型 | 触发频率 | 内容 |
|------|----------|------|
| **weekly** | 每周日 06:00 UTC | AI 开源周报：本周 Top 10 变动、新入榜项目、Star 增长最快 |
| **monthly** | 每月 1 日 06:00 UTC | AI 开源月报：月度排名变化、分类趋势、年度对比 |
| **spotlight** | 新项目首次入榜时 | 新项目速递：项目介绍、核心特点、快速上手 |

## 文章模板

### 周报模板

```markdown
---
title: "AI 开源周报 #{{.WeekNumber}} | {{.DateRange}}"
date: {{.PublishedAt}}
type: weekly
---

## 本周概览

本周 Top 100 共有 {{.NewEntries}} 个新入榜项目，{{.BigMovers}} 个项目排名大幅变动。

## 🔥 本周 Star 增长 Top 10

| 排名 | 项目 | 周增 Star | 总 Star | 分类 |
|------|------|-----------|---------|------|
{{range .TopGainers}}
| {{.Rank}} | [{{.FullName}}]({{.URL}}) | +{{.WeeklyGain}} | {{.Stars}} | {{.Category}} |
{{end}}

## 🆕 新入榜项目

{{range .NewProjects}}
### {{.FullName}}

> {{.Description}}

- ⭐ Star: {{.Stars}} | 🍴 Fork: {{.Forks}} | 📝 语言: {{.Language}}
- 🏷️ 分类: {{.Categories}}
- 📅 首次入榜: {{.FirstSeenAt}}
{{end}}

## 📊 排名变动

{{range .BigChanges}}
- {{.Direction}} **{{.FullName}}**: {{.OldRank}} → {{.NewRank}} ({{.Delta}})
{{end}}

## 📈 分类趋势

| 分类 | 项目数 | 本周新增 | 平均评分 |
|------|--------|----------|----------|
{{range .CategoryStats}}
| {{.Name}} | {{.Count}} | {{.NewCount}} | {{.AvgScore}} |
{{end}}
```

### 月报模板

在周报基础上增加：
- 月度 Star 增长曲线图表数据
- Top 100 月度稳定性分析（留存率）
- 各分类占比环形图数据

### 新项目速递模板

```markdown
---
title: "新项目速递 | {{.FullName}}"
date: {{.PublishedAt}}
type: spotlight
---

## {{.FullName}}

> {{.Description}}

### 基本信息

| 属性 | 值 |
|------|-----|
| GitHub | [{{.FullName}}]({{.URL}}) |
| 语言 | {{.Language}} |
| License | {{.License}} |
| Star | {{.Stars}} |
| 创建时间 | {{.CreatedAt}} |

### 项目亮点

{{.Highlights}}

### 快速上手

{{.QuickStart}}
```

## 模板引擎

使用 Go 标准库 `text/template`，模板文件存放在 `templates/` 目录：

```
templates/
├── weekly.md.tmpl
├── monthly.md.tmpl
└── spotlight.md.tmpl
```

## 数据查询

Content Generator 运行时需要查询以下数据：

```sql
-- 本周 Star 增长 Top 10
SELECT p.full_name, p.description, p.language,
       s_today.stargazers_count - s_week_ago.stargazers_count AS weekly_gain
FROM projects p
JOIN daily_snapshots s_today ON p.id = s_today.project_id AND s_today.snapshot_date = CURRENT_DATE
JOIN daily_snapshots s_week_ago ON p.id = s_week_ago.project_id AND s_week_ago.snapshot_date = CURRENT_DATE - 7
ORDER BY weekly_gain DESC
LIMIT 10;

-- 本周新入榜项目
SELECT p.*
FROM projects p
WHERE p.first_seen_at >= CURRENT_DATE - INTERVAL '7 days'
  AND p.rank IS NOT NULL AND p.rank <= 100
ORDER BY p.rank ASC;
```

## 生成流程

```
ContentGenerator.Run(postType)
  │
  ├── 1. 根据 postType 选择模板
  │
  ├── 2. 查询所需数据
  │      ├── 排行榜变动
  │      ├── Star 增长 Top N
  │      ├── 新入榜项目
  │      └── 分类统计
  │
  ├── 3. 构造模板数据结构
  │
  ├── 4. 渲染模板 → Markdown 字符串
  │
  ├── 5. 生成 slug（日期 + 类型）
  │
  └── 6. Upsert 到 blog_posts 表
```

## Slug 生成规则

```
weekly:    ai-weekly-2026-w07
monthly:   ai-monthly-2026-02
spotlight: new-project-owner-repo-name
```

## 相关文档

- [趋势分析](analyzer.md) — 上游分析结果
- [前端展示](web-frontend.md) — 博客文章展示
