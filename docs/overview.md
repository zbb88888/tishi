# 项目概述

## 愿景

tishi（提示）致力于成为 AI 开源生态的"风向标"——帮助开发者快速了解 GitHub 上 AI 领域最活跃、最有潜力的开源项目，捕捉技术趋势变化。

## 目标

1. **每日自动采集** GitHub 上 AI 相关项目的核心指标（Star、Fork、Issue、Contributor 等）
2. **维护 Top 100 排行榜**，支持按热度评分、Star 增速、分类等多维度排序
3. **趋势可视化**，提供 Star 增长曲线、热度变化等图表
4. **自动生成博客内容**，包括周报、月报、新项目速递
5. **细分领域分类**，覆盖 LLM / Agent / RAG / Diffusion / MLOps / Vector DB 等方向

## 核心功能

### 数据采集

- 定时通过 GitHub API（Search API + GraphQL）采集 AI 相关项目
- 基于关键词库过滤 AI 领域项目（详见 [种子数据](data/seed-data.md)）
- 支持增量更新，避免重复采集

### 趋势排行榜

- Top 100 项目实时排名
- 多维排序：热度评分 / Star 总量 / 日增 Star / 周增 Star
- 按分类筛选：LLM / Agent / RAG / Diffusion / MLOps 等

### 趋势分析

- 每日快照存储，支持任意时间范围的趋势查询
- 热度评分模型：综合 Star 增速、Fork 活跃度、Issue 响应速度等加权计算
- 异常检测：识别突然爆火或持续下滑的项目

### 博客内容

- 自动生成周报/月报 Markdown 文件
- 新项目速递：首次进入 Top 100 的项目专题
- 排名变动分析

### 项目详情

- 项目基本信息：描述、语言、License、创建时间
- 历史趋势图表：Star/Fork/Issue 增长曲线
- 社区活跃度指标

## 目标用户

- 关注 AI 开源生态的 **开发者**
- 需要技术选型参考的 **架构师**
- 追踪 AI 行业动态的 **技术管理者**
- 希望发现新项目的 **开源爱好者**

## 非目标（Explicit Non-Goals）

- 不做 AI 模型训练 / 推理服务
- 不做通用 GitHub 项目追踪（仅聚焦 AI 领域）
- 不做社交功能（评论、点赞等）
- 不做付费功能

## 参考链接

- [GitHub Trending](https://github.com/trending)
- [GitHub Search API](https://docs.github.com/en/rest/search)
- [OSS Insight](https://ossinsight.io/) — 类似项目参考
