# 种子数据 — AI 领域关键词库

## 概述

tishi 通过抓取 GitHub Trending 页面获取热门项目，再使用关键词库过滤 AI 相关项目。

v1.0 变更：

- **数据来源**：GitHub Trending HTML 页面（Colly 抓取），不再使用 GitHub Search API
- **关键词用途**：过滤 Trending 列表中的 AI 项目，而非主动搜索全量仓库
- **分类定义**：存储在 `data/categories.json`，不再使用数据库 INSERT

## 关键词定义

所有关键词定义在 `data/categories.json` 中，包含 12 个 AI 分类，每个分类有两组关键词：

- `topics[]` — 匹配 GitHub Topics 标签（精确匹配）
- `description[]` — 匹配项目描述和 README 文本（模糊匹配）

### 分类关键词一览

| 分类 slug | Topics 关键词 | Description 关键词 |
|-----------|--------------|-------------------|
| `llm` | llm, large-language-model, gpt, chatgpt, llama, mistral, chatbot, text-generation, language-model | large language model, LLM, chatbot, text generation |
| `agent` | ai-agent, autonomous-agent, agent-framework, agentic, langchain, autogpt, crew-ai | AI agent, autonomous agent, agentic, agent framework |
| `rag` | rag, retrieval-augmented-generation, vector-search, document-qa, knowledge-base | retrieval augmented, RAG, document Q&A, knowledge base |
| `diffusion` | stable-diffusion, diffusion-models, text-to-image, image-generation, comfyui, sdxl | stable diffusion, text to image, image generation, diffusion model |
| `mlops` | mlops, ml-pipeline, model-serving, feature-store, ml-platform, model-training | MLOps, model serving, ML pipeline, feature store |
| `vector-db` | vector-database, vector-store, similarity-search, embedding-database | vector database, vector store, similarity search, embedding |
| `framework` | pytorch, tensorflow, deep-learning, machine-learning, neural-network, jax, transformers | deep learning framework, ML framework, neural network |
| `tool` | ai-tool, ai-assistant, copilot, code-generation, ai-coding, ai-powered | AI tool, AI assistant, code generation, AI powered |
| `multimodal` | multimodal, vision-language-model, vlm, visual-question-answering | multimodal, vision language, VLM |
| `speech` | text-to-speech, tts, asr, speech-recognition, voice-cloning, speech-synthesis | text to speech, TTS, speech recognition, voice clone |
| `rl` | reinforcement-learning, rlhf, reward-model, ppo | reinforcement learning, RLHF, reward model |
| `other` | — | — |

## 过滤流程

v1.0 的 AI 项目过滤发生在 Scraper 抓取 Trending 页面之后：

```
GitHub Trending HTML
  → Colly 解析 article.Box-row
  → 提取 repo name, description, language, stars
  → 加载 data/categories.json 关键词
  → 匹配：topics ∪ description ∪ repo name
  → 命中任意关键词 → 标记为 AI 项目，写入 data/projects/
  → 未命中 → 跳过
```

### 匹配优先级

1. **GitHub Topics 精确匹配** — 项目的 topics 标签包含关键词 → 高置信度
2. **Description 关键词匹配** — 项目描述包含关键词 → 中置信度
3. **Repo 名称包含关键词** — 仓库名包含 `llm`、`agent` 等 → 低置信度
4. **全未匹配** → 归入 `other` 或跳过

### 多分类

一个项目可以属于多个分类，存储在 `project.categories[]` 数组中：

- `langchain` → `["llm", "agent", "rag"]`
- `ComfyUI` → `["diffusion", "tool"]`
- `vLLM` → `["llm", "mlops"]`

## 排除规则

以下项目在过滤阶段排除：

### 排除关键词（出现在 repo name 或 description 中）

```
awesome-list        # 纯列表项目
tutorial            # 教程项目（除非 Stars 极高）
course              # 课程项目
interview           # 面试题
cheatsheet          # 速查表
```

### 排除条件

- Trending 列表中的 Fork 项目（GitHub 已标注）
- 最后推送时间超过 1 年（从 GitHub API enrichment 获取）

## 关键词维护策略

关键词库通过编辑 `data/categories.json` 文件维护：

1. **季度审查** — 每季度审查一次关键词列表，添加新出现的技术方向
2. **自动发现** — 分析新入榜项目的 topics，发现高频但未收录的关键词（Phase 3）
3. **手动补充** — 关注 AI 领域重大发布（新模型、新框架），及时添加到 categories.json

### 新增关键词示例

```json
// 在 data/categories.json 对应分类的 topics 或 description 数组中添加
{
  "slug": "rl",
  "topics": ["reinforcement-learning", "rlhf", "reward-model", "ppo", "world-model"],
  "description": ["reinforcement learning", "RLHF", "reward model", "world model"]
}
```

### 可能新增的技术方向

| 技术 | 可能的关键词 | 归属分类 |
|------|------------|---------|
| 世界模型 | world-model | rl |
| AI 安全 | ai-safety, ai-alignment | 新分类或 other |
| MCP | model-context-protocol, mcp | tool |
| 推理模型 | reasoning-model, test-time-compute | llm |

## 相关文档

- [Scraper 设计](../design/collector.md) — Trending 页面抓取和过滤
- [分类定义](../../data/categories.json) — 完整关键词 JSON
