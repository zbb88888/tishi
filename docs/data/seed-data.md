# 种子数据 — AI 领域关键词库

## 概述

tishi 通过关键词搜索从 GitHub 海量仓库中筛选 AI 相关项目。本文档定义关键词库和分类映射规则。

## 搜索关键词

### 一级搜索（topic based）

使用 GitHub `topic:` 搜索，精确度最高：

```
topic:llm
topic:large-language-model
topic:gpt
topic:chatbot
topic:ai-agent
topic:langchain
topic:rag
topic:retrieval-augmented-generation
topic:stable-diffusion
topic:diffusion-models
topic:text-to-image
topic:machine-learning
topic:deep-learning
topic:neural-network
topic:pytorch
topic:tensorflow
topic:transformers
topic:mlops
topic:vector-database
topic:embedding
topic:text-to-speech
topic:speech-recognition
topic:computer-vision
topic:reinforcement-learning
topic:rlhf
topic:multimodal
topic:ai-assistant
topic:code-generation
topic:ai-coding
```

### 二级搜索（description/readme based）

对于 topic 覆盖不到的项目，使用描述搜索：

```
"large language model" in:description stars:>200
"retrieval augmented generation" in:description stars:>100
"text to image" in:description stars:>200
"ai agent" in:description stars:>200
"vector database" in:description stars:>200
"model serving" in:description stars:>100
"prompt engineering" in:description stars:>100
```

### Stars 过滤阈值

| 搜索类型 | 最低 Stars | 说明 |
|----------|-----------|------|
| topic 精确搜索 | ≥ 100 | 低门槛发现新项目 |
| description 模糊搜索 | ≥ 200 | 高门槛减少噪音 |
| 最终入榜 | Top 100 by score | 评分排序取顶 |

## 分类映射规则

### 分类 → 关键词映射

| 分类 (slug) | topic 关键词 | description 关键词 |
|-------------|-------------|-------------------|
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

### 匹配优先级

1. **topic 精确匹配** → confidence 1.0
2. **description 关键词匹配** → confidence 0.8
3. **name 包含关键词** → confidence 0.6
4. **未匹配任何分类** → 归入 `other`

### 多分类

一个项目可以属于多个分类。例如：
- `langchain` → `llm` + `agent` + `rag`
- `ComfyUI` → `diffusion` + `tool`
- `vLLM` → `llm` + `mlops`

## 排除规则

以下项目应排除在 AI 相关范围之外：

### 排除关键词

```
awesome-list        # 纯列表项目
tutorial            # 教程项目（除非 Stars 极高）
course              # 课程项目
interview           # 面试题
cheatsheet          # 速查表
```

### 排除条件

- 项目描述中仅有中文/非英文，且无 AI 相关关键词
- Fork 数为 0 且 Star 数 < 50（可能是测试仓库）
- 最后推送时间超过 1 年（不活跃项目不参与排名，但保留历史数据）

## 关键词维护策略

关键词库需要定期更新以跟踪 AI 领域变化：

1. **季度审查** — 每季度审查一次关键词列表，添加新出现的技术方向
2. **自动发现** — 分析新入榜项目的 topics，发现高频但未收录的关键词
3. **手动补充** — 关注 AI 领域重大发布（如新模型、新框架），及时添加

### 示例：2026 年可能新增的关键词

```
world-model         # 世界模型
ai-safety           # AI 安全
ai-alignment        # AI 对齐
test-time-compute   # 测试时计算
reasoning-model     # 推理模型
mcp                 # Model Context Protocol
```

## 相关文档

- [数据采集](../design/collector.md) — 如何使用关键词搜索
- [趋势分析](../design/analyzer.md) — 自动分类逻辑
