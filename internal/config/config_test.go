package config

import (
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	for _, k := range []string{"GITHUB_TOKENS", "TISHI_LLM_API_KEY"} {
		t.Setenv(k, "")
	}

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("expected default log level=info, got %q", cfg.Logging.Level)
	}
	if cfg.DataDir != "./data" {
		t.Errorf("expected default data_dir=./data, got %q", cfg.DataDir)
	}
	if cfg.Scorer.TopN != 100 {
		t.Errorf("expected default scorer.top_n=100, got %d", cfg.Scorer.TopN)
	}
	if cfg.LLM.Provider != "deepseek" {
		t.Errorf("expected default llm.provider=deepseek, got %q", cfg.LLM.Provider)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("GITHUB_TOKENS", "tok1,tok2")
	t.Setenv("TISHI_LLM_API_KEY", "test-key-123")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("expected log level=debug from env, got %q", cfg.Logging.Level)
	}
	if len(cfg.GitHub.Tokens) != 2 {
		t.Errorf("expected 2 tokens, got %d", len(cfg.GitHub.Tokens))
	}
	if cfg.LLM.APIKey != "test-key-123" {
		t.Errorf("expected LLM API key from env, got %q", cfg.LLM.APIKey)
	}
}
