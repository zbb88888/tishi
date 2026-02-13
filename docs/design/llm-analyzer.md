# LLM 分析模块设计（LLM Analyzer）

## 概述

LLM Analyzer 是 v1.0 新增模块，负责调用 DeepSeek / Qwen API，对 AI 项目生成中文深度分析报告，写入 `data/projects/*.json` 的 `analysis` 字段。

## LLM 选型

| 提供商 | BaseURL | 模型 | 价格 |
|--------|---------|------|------|
| DeepSeek | `https://api.deepseek.com/v1` | `deepseek-chat` (V3) | ¥1/M input, ¥2/M output |
| Qwen (阿里云) | `https://dashscope.aliyuncs.com/compatible-mode/v1` | `qwen-plus` | ¥0.8/M input, ¥2/M output |

两者均兼容 OpenAI API 格式，使用 `sashabaranov/go-openai` 统一客户端，仅切换 `BaseURL`。

## Prompt 设计

### System Prompt

```
你是一个专业的 AI 技术分析师，专门为中文开发者社区撰写开源项目深度分析报告。
你的报告需要：
1. 使用中文撰写，专业术语保留英文原文
2. 客观准确，基于项目实际功能和代码
3. 面向有一定技术背景的中文开发者
4. 简洁有力，避免空洞的营销话术

请以 JSON 格式输出分析结果。
```

### User Prompt 模板

```
请分析以下 GitHub 开源项目：

项目名称: {{.FullName}}
描述: {{.Description}}
编程语言: {{.Language}}
Star 数: {{.Stars}}
Topics: {{.Topics}}

README 内容（前 3000 字）:
{{.ReadmeContent}}

请输出以下 JSON 格式的分析结果：
{
  "summary": "一句话中文概括（50字以内）",
  "positioning": "项目定位：解决什么问题，面向谁（100字以内）",
  "features": ["核心功能1", "核心功能2", ...],
  "advantages": "相比同类项目的优势（100字以内）",
  "tech_stack": "使用的核心技术栈",
  "use_cases": "适用场景（100字以内）",
  "comparison": [
    {"name": "竞品名", "difference": "差异点"}
  ],
  "ecosystem": "上下游生态（100字以内）"
}
```

## 实现

### LLM 客户端

```go
package llm

import (
    "context"
    openai "github.com/sashabaranov/go-openai"
)

type Client struct {
    client *openai.Client
    model  string
    logger *zap.Logger
}

func NewClient(cfg Config) *Client {
    config := openai.DefaultConfig(cfg.APIKey)
    config.BaseURL = cfg.BaseURL  // DeepSeek 或 Qwen 端点

    return &Client{
        client: openai.NewClientWithConfig(config),
        model:  cfg.Model,
    }
}

func (c *Client) AnalyzeProject(ctx context.Context, project Project) (*Analysis, error) {
    prompt := buildPrompt(project)

    resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model: c.model,
        Messages: []openai.ChatCompletionMessage{
            {Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
            {Role: openai.ChatMessageRoleUser, Content: prompt},
        },
        Temperature:     0.3,  // 低温度保证一致性
        MaxTokens:       2000,
        ResponseFormat:  &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
    })
    if err != nil {
        return nil, fmt.Errorf("LLM API call failed: %w", err)
    }

    var analysis Analysis
    if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &analysis); err != nil {
        return nil, fmt.Errorf("parse LLM response: %w", err)
    }

    analysis.Model = c.model
    analysis.Status = "draft"
    analysis.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
    analysis.TokenUsage = TokenUsage{
        PromptTokens:     resp.Usage.PromptTokens,
        CompletionTokens: resp.Usage.CompletionTokens,
        TotalTokens:      resp.Usage.TotalTokens,
    }

    return &analysis, nil
}
```

### Analysis 数据结构

```go
type Analysis struct {
    Status      string      `json:"status"`       // draft / published / rejected
    Model       string      `json:"model"`        // deepseek-chat / qwen-plus
    Summary     string      `json:"summary"`
    Positioning string      `json:"positioning"`
    Features    []string    `json:"features"`
    Advantages  string      `json:"advantages"`
    TechStack   string      `json:"tech_stack"`
    UseCases    string      `json:"use_cases"`
    Comparison  []CompItem  `json:"comparison"`
    Ecosystem   string      `json:"ecosystem"`
    GeneratedAt string      `json:"generated_at"`
    ReviewedAt  string      `json:"reviewed_at,omitempty"`
    TokenUsage  TokenUsage  `json:"token_usage"`
}

type CompItem struct {
    Name       string `json:"name"`
    Difference string `json:"difference"`
}

type TokenUsage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
```

## 执行流程

```
tishi analyze
  │
  ├── 1. 扫描 data/projects/*.json
  │
  ├── 2. 筛选需要分析的项目：
  │      - analysis 字段不存在
  │      - analysis.status == "draft" 且 generated_at 超过 7 天
  │
  ├── 3. 对每个项目：
  │      ├── 构建 Prompt (description + README 前 3000 字)
  │      ├── 调用 LLM API
  │      ├── 解析 JSON 响应
  │      ├── 写入 project JSON 的 analysis 字段
  │      └── 记录 token 用量
  │
  ├── 4. 输出统计：分析了 N 个项目，消耗 M tokens
  │
  └── 5. 人工审核流程：
         tishi review              # 列出所有 draft 状态的分析
         tishi review --approve=id # 将 draft 改为 published
         tishi review --reject=id  # 将 draft 改为 rejected
```

## 成本估算

| 场景 | 每项目 tokens | 每日项目数 | 日成本 (DeepSeek) |
|------|-------------|-----------|-------------------|
| 新项目首次分析 | ~3000 input + ~800 output | ~20 | ¥0.09 |
| 全量重分析 | ~3000 input + ~800 output | ~100 | ¥0.46 |

**结论**：日常运行成本 < ¥1/天，极低。

## 人工审核工作流

```
                    ┌──────────┐
LLM 生成 ─────────→ │  draft   │
                    └────┬─────┘
                         │
           ┌─────────────┼─────────────┐
           ▼             │             ▼
    ┌──────────┐         │      ┌──────────┐
    │published │         │      │ rejected │
    └──────────┘         │      └──────────┘
           │             │
           ▼             ▼
    前端展示           重新生成 / 人工编辑
```

- `draft`: LLM 自动生成，待审核
- `published`: 审核通过，前端展示
- `rejected`: 质量不达标，需重新生成或人工编辑

## 错误处理

| 错误类型 | 处理方式 |
|----------|----------|
| LLM API 超时 | 重试 2 次，间隔 5s |
| LLM 返回非 JSON | 重试 1 次，仍失败则跳过 |
| LLM 返回内容质量差 | 标记 draft，人工审核时处理 |
| API Key 额度用完 | 切换备用 Provider |
| 单项目分析失败 | 跳过，不阻塞其他项目 |

## CLI 命令

```bash
tishi analyze                    # 分析所有未分析的项目
tishi analyze --id=owner__repo   # 分析指定项目
tishi analyze --force            # 强制重新分析所有项目
tishi analyze --dry-run          # 仅打印 Prompt，不调用 API
tishi analyze --provider=qwen    # 使用 Qwen 而非 DeepSeek

tishi review                     # 列出待审核的分析
tishi review --approve=id        # 审核通过
tishi review --reject=id         # 审核拒绝
```

## 相关文档

- [数据采集](collector.md) — 上游爬取数据
- [评分排名](analyzer.md) — 下游评分
- [数据契约](../../data/schemas/project.schema.json) — analysis 字段 Schema
