# 项目概述

## 愿景

**tishi（提示）** 是一个面向中文开发者的 AI 开源项目深度分析平台。数据来源于 [GitHub Trending](https://github.com/trending)，通过 LLM 自动生成中文项目报告，帮助开发者快速理解 AI 领域最活跃的开源项目——定位、功能、优势、适用场景一目了然。

## 核心差异点（vs GitHub Trending）

| 维度 | GitHub Trending | tishi |
|------|----------------|-------|
| 语言 | 英文 | **中文深度解读**（LLM 翻译+分析） |
| 范围 | 全领域 | **AI 垂直精选**，12 个 AI 子方向分类 |
| 深度 | 仅 repo name + description | **完整项目报告**（定位/功能/优势/对比/场景） |
| 历史 | 无持久化 | **趋势追踪**，每日快照，Star 增长曲线 |

## 12 个 AI 分类

LLM · Agent · RAG · Diffusion · MLOps · Vector-DB · Framework · Tool · Multimodal · Speech · RL · Other

## 架构概要

三阶段完全解耦，JSON 文件 + Git 仓库交换数据，无共享数据库：

```
Stage 1 (Machine A)                Stage 2 (Machine B)         Stage 3 (Machine C)
┌─────────────────────┐            ┌──────────────────┐        ┌────────────────┐
│ GitHub Trending HTML │            │  Git pull data/  │        │  Nginx / CDN   │
│        ↓             │            │       ↓          │        │  serve dist/   │
│   Colly 爬取+过滤    │   git push │  Astro SSG build │        │       ↓        │
│        ↓             │ ─────────→ │       ↓          │ ────→  │  用户浏览器     │
│   LLM 中文分析       │   data/    │   dist/ 静态文件  │  dist/ │                │
│        ↓             │            └──────────────────┘        └────────────────┘
│   data/ JSON 输出    │
└─────────────────────┘
```

## LLM 分析内容

每个 AI 项目自动生成的中文报告包含：

- **项目定位** — 这个项目是做什么的，解决什么问题
- **核心功能** — 主要功能点清单
- **技术优势** — 相比同类项目的差异化优势
- **技术栈** — 使用的语言/框架/底层技术
- **适用场景** — 什么情况下适合用
- **同类对比** — 与竞品的横向对比
- **生态系统** — 上下游依赖和集成

## 目标用户

- 关注 AI 开源动态的 **中文开发者**
- 需要 AI 技术选型参考的 **架构师**
- 追踪 AI 领域趋势的 **技术管理者**

## 非目标

- 不做通用 GitHub 项目追踪（仅 AI 领域）
- 不做 AI 模型训练 / 推理服务
- 不做社交功能（评论、点赞）
- 不做付费功能

## 参考

- [GitHub Trending](https://github.com/trending) — 数据来源
- [数据契约 Schema](../data/schemas/) — JSON Schema 定义
