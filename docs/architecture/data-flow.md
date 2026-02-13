# 数据流转

## 概述

tishi v1.0 的数据从 GitHub Trending 爬取，经 AI 过滤、LLM 分析、评分排名，最终以 JSON 文件输出到 `data/` 目录。Stage 2 读取 JSON 生成静态页面。全链路无数据库。

## 全链路数据流

```
           Stage 1: 采集+分析 (Machine A)
           ═══════════════════════════════

  ┌────────────────────┐
  │ GitHub Trending    │  HTTP GET https://github.com/trending
  │ HTML 页面          │
  └────────┬───────────┘
           │ Colly 爬取
           ▼
  ┌────────────────────┐
  │ 候选项目列表        │  article.Box-row 解析
  │ (repo, desc, lang, │  ~25 项/页 × 3 语种
  │  stars, period_stars)│
  └────────┬───────────┘
           │ AI 关键词过滤 (categories.json)
           ▼
  ┌────────────────────┐
  │ AI 项目列表         │  过滤掉非 AI 项目
  └────────┬───────────┘
           │ GitHub API 补充 (README + 详细指标)
           ▼
  ┌────────────────────┐
  │ 完整项目数据        │  owner, description, stars, forks,
  │                    │  open_issues, readme_content, topics[]
  └────────┬───────────┘
           │ LLM 中文分析 (DeepSeek / Qwen)
           ▼
  ┌────────────────────┐
  │ 项目 + analysis    │  summary, positioning, features[],
  │                    │  advantages, tech_stack, use_cases,
  │                    │  comparison[], ecosystem
  └────────┬───────────┘
           │ 评分 + 排名 + 快照
           ▼
  ┌────────────────────┐
  │ data/ 目录输出      │
  │  projects/*.json   │  每个 AI 项目一个文件
  │  snapshots/*.jsonl │  当日快照 (追加)
  │  rankings/*.json   │  当日排行
  │  posts/*.json      │  博客文章
  └────────┬───────────┘
           │ git push
           ▼
           Stage 2: SSG 构建 (Machine B)
           ═══════════════════════════════

  ┌────────────────────┐
  │ git pull data/     │
  └────────┬───────────┘
           │ Astro SSG build
           ▼
  ┌────────────────────┐
  │ dist/ 静态 HTML    │  首页排行榜 / 项目详情 / 分类页 / 博客
  └────────┬───────────┘
           │ rsync / scp
           ▼
           Stage 3: 发布 (Machine C)
           ═══════════════════════════════

  ┌────────────────────┐
  │ Nginx / CDN        │  serve dist/
  │ 用户浏览器访问       │
  └────────────────────┘
```

## 阶段详解

### 阶段 1a：Trending 爬取 (Scraper)

**触发**：`tishi scrape` CLI 命令（外部 cron 每日调度）

**输入**：GitHub Trending 页面 URL

**处理流程**：

1. Colly 请求 `https://github.com/trending?since=daily` (可选 weekly)
2. CSS 选择器 `article.Box-row` 提取：repo full_name, description, language, stars_total, period_stars
3. 读取 `data/categories.json` 关键词映射，按 12 类 AI 关键词过滤
4. 对通过过滤的项目，调用 GitHub REST API 补充：README 内容、topics、forks、issues 等
5. Merge 到 `data/projects/{owner}__{repo}.json`（存在则更新 trending 字段，不存在则新建）
6. Append 到 `data/snapshots/{YYYY-MM-DD}.jsonl`

**输出**：projects/*.json + snapshots/*.jsonl

### 阶段 1b：LLM 分析 (Analyzer)

**触发**：`tishi analyze` CLI 命令

**输入**：`data/projects/*.json` 中 `analysis.status == "draft"` 或未分析的项目

**处理流程**：

1. 遍历 projects/ 目录，筛选需要分析的项目
2. 构建 Prompt：项目名 + description + README 前 3000 字 + topics
3. 调用 LLM API (OpenAI-compatible)，模型 deepseek-chat 或 qwen-plus
4. 解析返回的 JSON：summary, positioning, features[], advantages, tech_stack, use_cases, comparison[], ecosystem
5. 写入项目 JSON 的 `analysis` 字段，status 设为 `"draft"`（需人工 review 后改为 `"published"`）
6. 记录 token_usage 用于成本监控

**输出**：更新 projects/*.json 的 analysis 字段

### 阶段 1c：评分排名 (Scorer)

**触发**：`tishi score` CLI 命令

**输入**：projects/*.json + snapshots/*.jsonl

**处理流程**：

1. 读取所有项目 JSON + 最近 7 日快照
2. 多维加权评分：daily_stars × 0.4 + weekly_stars × 0.3 + forks_rate × 0.15 + issues_activity × 0.15
3. 按评分排序，生成 Top N 排行榜
4. 与前一日排名对比，计算 rank_change

**输出**：`data/rankings/{YYYY-MM-DD}.json`

### 阶段 2：Astro SSG 构建

**触发**：`npm run build`（Machine B 的 cron 或 CI）

**输入**：`data/` 目录全部 JSON 文件

**处理流程**：

1. Astro 构建脚本读取 data/projects/*.json、data/rankings/latest.json、data/posts/*.json
2. 生成静态 HTML：首页排行榜、项目详情页、分类浏览页、博客列表页、博客详情页
3. 输出到 dist/

**输出**：dist/ 目录下的静态 HTML/CSS/JS

## 数据时效性

| 数据类型 | 更新频率 | 来源 |
|----------|----------|------|
| Trending 项目数据 | 每日 | GitHub Trending HTML |
| LLM 分析报告 | 新项目/定期更新 | DeepSeek / Qwen API |
| 排行榜 | 每日 | 评分计算 |
| 静态页面 | 每日 | Astro SSG rebuild |

## 数据量估算

| 数据 | 单文件大小 | 日增 | 年累计 |
|------|-----------|------|--------|
| project JSON | ~5KB | ~5-20 个新项目 | ~2000 个文件 |
| snapshot JSONL | ~10KB/日 | 1 个文件 | 365 个文件 |
| ranking JSON | ~20KB | 1 个文件 | 365 个文件 |
| post JSON | ~5KB | ~0.3 个 | ~100 个文件 |

**结论**：全部数据 < 100MB/年，Git 仓库轻松承载。
