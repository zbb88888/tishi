// Package llm provides LLM-based project analysis using OpenAI-compatible APIs.
// Supports DeepSeek and Qwen providers via configurable BaseURL.
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/config"
	"github.com/zbb88888/tishi/internal/datastore"
)

// providerEndpoints maps provider names to their API base URLs.
var providerEndpoints = map[string]string{
	"deepseek": "https://api.deepseek.com/v1",
	"qwen":     "https://dashscope.aliyuncs.com/compatible-mode/v1",
}

// providerDefaultModels maps provider names to default model names.
var providerDefaultModels = map[string]string{
	"deepseek": "deepseek-chat",
	"qwen":     "qwen-plus",
}

// Client wraps an OpenAI-compatible API client for project analysis.
type Client struct {
	client *openai.Client
	model  string
	cfg    config.LLMConfig
	log    *zap.Logger
}

// NewClient creates an LLM client based on provider configuration.
func NewClient(cfg config.LLMConfig, log *zap.Logger) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("llm.api_key is required (set TISHI_LLM_API_KEY)")
	}

	provider := strings.ToLower(cfg.Provider)
	baseURL, ok := providerEndpoints[provider]
	if !ok {
		return nil, fmt.Errorf("unsupported LLM provider %q (use deepseek or qwen)", provider)
	}

	model := cfg.Model
	if model == "" {
		model = providerDefaultModels[provider]
	}

	ocfg := openai.DefaultConfig(cfg.APIKey)
	ocfg.BaseURL = baseURL

	return &Client{
		client: openai.NewClientWithConfig(ocfg),
		model:  model,
		cfg:    cfg,
		log:    log.Named("llm"),
	}, nil
}

// systemPrompt instructs the LLM on its role and output requirements.
var systemPrompt = strings.Join([]string{
	"你是一个专业的 AI 技术分析师，专门为中文开发者社区撰写开源项目深度分析报告。",
	"你的报告需要：",
	"1. 使用中文撰写，专业术语保留英文原文",
	"2. 客观准确，基于项目实际功能和代码",
	"3. 面向有一定技术背景的中文开发者",
	"4. 简洁有力，避免空洞的营销话术",
	"",
	"请严格按照 JSON 格式输出分析结果，不要输出其他内容。",
}, "\n")

// buildUserPrompt constructs the user prompt for a project.
func buildUserPrompt(p *datastore.Project, readme string) string {
	desc := ""
	if p.Description != nil {
		desc = *p.Description
	}
	lang := ""
	if p.Language != nil {
		lang = *p.Language
	}
	topics := strings.Join(p.Topics, ", ")

	return fmt.Sprintf(`请分析以下 GitHub 开源项目：

项目名称: %s
描述: %s
编程语言: %s
Star 数: %d
Topics: %s

README 内容（前 3000 字）:
%s

请输出以下 JSON 格式的分析结果：
{
  "summary": "一句话中文概括（50字以内）",
  "positioning": "项目定位：解决什么问题，面向谁（200字以内）",
  "features": [{"name": "功能名", "desc": "功能描述"}],
  "advantages": "相比同类项目的优势（200字以内）",
  "tech_stack": "使用的核心技术栈",
  "use_cases": "适用场景（200字以内）",
  "comparison": [{"project": "竞品名", "diff": "差异点"}],
  "ecosystem": "上下游生态（200字以内）"
}`, p.FullName, desc, lang, p.Stars, topics, readme)
}

// llmResponse is the expected JSON structure from LLM output.
type llmResponse struct {
	Summary     string              `json:"summary"`
	Positioning string              `json:"positioning"`
	Features    []datastore.Feature `json:"features"`
	Advantages  string              `json:"advantages"`
	TechStack   string              `json:"tech_stack"`
	UseCases    string              `json:"use_cases"`
	Comparison  []llmComparison     `json:"comparison"`
	Ecosystem   string              `json:"ecosystem"`
}

type llmComparison struct {
	Project    string `json:"project"`
	Name       string `json:"name"`       // fallback field
	Diff       string `json:"diff"`
	Difference string `json:"difference"` // fallback field
}

// AnalyzeProject sends project data to the LLM and returns a structured analysis.
func (c *Client) AnalyzeProject(ctx context.Context, p *datastore.Project, readme string) (*datastore.Analysis, error) {
	// Truncate README to ~3000 chars
	if len(readme) > 3000 {
		readme = readme[:3000]
	}

	userPrompt := buildUserPrompt(p, readme)

	c.log.Debug("调用 LLM API",
		zap.String("project", p.FullName),
		zap.String("model", c.model),
		zap.Int("prompt_len", len(userPrompt)),
	)

	req := openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userPrompt},
		},
		Temperature: float32(c.cfg.Temperature),
		MaxTokens:   c.cfg.MaxTokens,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM API: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("LLM returned empty choices")
	}

	content := resp.Choices[0].Message.Content

	var raw llmResponse
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		return nil, fmt.Errorf("parsing LLM JSON: %w (content: %.500s)", err, content)
	}

	// Build Analysis
	now := time.Now().UTC()
	totalTokens := resp.Usage.TotalTokens

	analysis := &datastore.Analysis{
		Status:      "draft",
		Model:       c.model,
		Summary:     raw.Summary,
		Positioning: raw.Positioning,
		Features:    raw.Features,
		Advantages:  raw.Advantages,
		TechStack:   raw.TechStack,
		UseCases:    raw.UseCases,
		Ecosystem:   raw.Ecosystem,
		GeneratedAt: now,
		TokenUsage:  &totalTokens,
	}

	// Convert comparison entries
	for _, comp := range raw.Comparison {
		name := comp.Project
		if name == "" {
			name = comp.Name
		}
		diff := comp.Diff
		if diff == "" {
			diff = comp.Difference
		}
		if name != "" {
			analysis.Comparison = append(analysis.Comparison, datastore.ComparisonEntry{
				Project: name,
				Diff:    diff,
			})
		}
	}

	c.log.Info("LLM 分析完成",
		zap.String("project", p.FullName),
		zap.String("summary", raw.Summary),
		zap.Int("tokens", totalTokens),
	)

	return analysis, nil
}
