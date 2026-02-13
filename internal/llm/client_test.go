package llm

import (
	"testing"

	"github.com/zbb88888/tishi/internal/config"
	"github.com/zbb88888/tishi/internal/datastore"
	"go.uber.org/zap"
)

func testLogger() *zap.Logger {
	l, _ := zap.NewDevelopment()
	return l
}

func TestNewClient_MissingAPIKey(t *testing.T) {
	cfg := config.LLMConfig{Provider: "deepseek"}
	_, err := NewClient(cfg, testLogger())
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
}

func TestNewClient_UnsupportedProvider(t *testing.T) {
	cfg := config.LLMConfig{Provider: "unknown", APIKey: "test-key"}
	_, err := NewClient(cfg, testLogger())
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
}

func TestNewClient_DeepSeek(t *testing.T) {
	cfg := config.LLMConfig{Provider: "deepseek", APIKey: "test-key"}
	c, err := NewClient(cfg, testLogger())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.model != "deepseek-chat" {
		t.Errorf("model = %q, want deepseek-chat", c.model)
	}
}

func TestNewClient_Qwen(t *testing.T) {
	cfg := config.LLMConfig{Provider: "qwen", APIKey: "test-key"}
	c, err := NewClient(cfg, testLogger())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.model != "qwen-plus" {
		t.Errorf("model = %q, want qwen-plus", c.model)
	}
}

func TestNewClient_CustomModel(t *testing.T) {
	cfg := config.LLMConfig{Provider: "deepseek", APIKey: "key", Model: "deepseek-coder"}
	c, err := NewClient(cfg, testLogger())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.model != "deepseek-coder" {
		t.Errorf("model = %q, want deepseek-coder", c.model)
	}
}

func TestBuildUserPrompt(t *testing.T) {
	desc := "An awesome AI tool"
	lang := "Python"
	p := &datastore.Project{
		FullName:    "owner/repo",
		Description: &desc,
		Language:    &lang,
		Stars:       5000,
		Topics:      []string{"ai", "ml"},
	}

	prompt := buildUserPrompt(p, "# README content")

	// Should contain project name
	if !containsStr(prompt, "owner/repo") {
		t.Error("prompt missing project name")
	}
	if !containsStr(prompt, "An awesome AI tool") {
		t.Error("prompt missing description")
	}
	if !containsStr(prompt, "Python") {
		t.Error("prompt missing language")
	}
	if !containsStr(prompt, "5000") {
		t.Error("prompt missing star count")
	}
	if !containsStr(prompt, "ai, ml") {
		t.Error("prompt missing topics")
	}
	if !containsStr(prompt, "README content") {
		t.Error("prompt missing README")
	}
}

func TestBuildUserPrompt_NilFields(t *testing.T) {
	p := &datastore.Project{
		FullName: "owner/repo",
		Stars:    100,
	}
	prompt := buildUserPrompt(p, "")
	if !containsStr(prompt, "owner/repo") {
		t.Error("prompt missing project name")
	}
}

func TestNeedsAnalysis(t *testing.T) {
	tests := []struct {
		name string
		p    *datastore.Project
		want bool
	}{
		{"no analysis", &datastore.Project{}, true},
		{"published", &datastore.Project{Analysis: &datastore.Analysis{Status: "published"}}, false},
		{"rejected", &datastore.Project{Analysis: &datastore.Analysis{Status: "rejected"}}, false},
		{"fresh draft", &datastore.Project{
			Analysis: &datastore.Analysis{Status: "draft", GeneratedAt: timeNow()},
		}, false},
		// Stale draft older than 7 days would need re-analysis but we can't easily
		// test time-dependent behavior without mocking
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := needsAnalysis(tt.p)
			if got != tt.want {
				t.Errorf("needsAnalysis = %v, want %v", got, tt.want)
			}
		})
	}
}

func containsStr(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && contains(s, substr)
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
