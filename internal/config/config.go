// Package config manages application configuration using Viper.
// It supports config file, environment variables, and CLI flags.
package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	DataDir string        `mapstructure:"data_dir"` // path to data/ directory
	GitHub  GitHubConfig  `mapstructure:"github"`
	Scraper ScraperConfig `mapstructure:"scraper"`
	Scorer  ScorerConfig  `mapstructure:"scorer"`
	LLM     LLMConfig     `mapstructure:"llm"`
	Logging LoggingConfig `mapstructure:"logging"`
	Site    SiteConfig    `mapstructure:"site"`
}

// GitHubConfig holds GitHub API settings.
type GitHubConfig struct {
	Tokens []string `mapstructure:"tokens"`
}

// ScraperConfig holds Trending scraper settings.
type ScraperConfig struct {
	Since    string        `mapstructure:"since"`    // daily or weekly
	Language string        `mapstructure:"language"` // optional language filter
	Timeout  time.Duration `mapstructure:"timeout"`
	RetryMax int           `mapstructure:"retry_max"`
}

// ScorerConfig holds scoring weight parameters.
type ScorerConfig struct {
	DailyStars    float64 `mapstructure:"daily_stars"`
	WeeklyStars   float64 `mapstructure:"weekly_stars"`
	ForksRate     float64 `mapstructure:"forks_rate"`
	IssueActivity float64 `mapstructure:"issue_activity"`
	TopN          int     `mapstructure:"top_n"`
}

// LLMConfig holds LLM provider settings for project analysis.
type LLMConfig struct {
	Provider    string  `mapstructure:"provider"` // deepseek or qwen
	APIKey      string  `mapstructure:"api_key"`
	Model       string  `mapstructure:"model"` // empty = provider default
	MaxTokens   int     `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`
	RetryMax    int     `mapstructure:"retry_max"`
}

// LoggingConfig holds logging settings.
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// SiteConfig holds website metadata.
type SiteConfig struct {
	Domain      string `mapstructure:"domain"`
	Title       string `mapstructure:"title"`
	Description string `mapstructure:"description"`
}

// global holds the singleton config instance.
var global *Config

// Load reads configuration from file and environment variables.
func Load(cfgFile string) (*Config, error) {
	setDefaults()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/tishi/")
	}

	// Environment variables: TISHI_DATA_DIR, TISHI_GITHUB_TOKENS, etc.
	viper.SetEnvPrefix("TISHI")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Support non-prefixed env vars for common ones
	_ = viper.BindEnv("github.tokens", "TISHI_GITHUB_TOKENS", "GITHUB_TOKENS")
	_ = viper.BindEnv("data_dir", "TISHI_DATA_DIR")
	_ = viper.BindEnv("logging.level", "TISHI_LOG_LEVEL", "LOG_LEVEL")
	_ = viper.BindEnv("logging.format", "TISHI_LOG_FORMAT", "LOG_FORMAT")

	// LLM bindings
	_ = viper.BindEnv("llm.provider", "TISHI_LLM_PROVIDER")
	_ = viper.BindEnv("llm.api_key", "TISHI_LLM_API_KEY")
	_ = viper.BindEnv("llm.model", "TISHI_LLM_MODEL")

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
		// Config file not found is acceptable - use defaults + env vars
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	// Handle comma-separated GITHUB_TOKENS env var
	if len(cfg.GitHub.Tokens) == 1 && strings.Contains(cfg.GitHub.Tokens[0], ",") {
		cfg.GitHub.Tokens = strings.Split(cfg.GitHub.Tokens[0], ",")
	}

	global = &cfg
	return &cfg, nil
}

// Get returns the global config. Must call Load first.
func Get() *Config {
	if global == nil {
		panic("config.Load() must be called before config.Get()")
	}
	return global
}

// setDefaults configures default values for all settings.
func setDefaults() {
	// v1.0 defaults
	viper.SetDefault("data_dir", "./data")

	// Scraper
	viper.SetDefault("scraper.since", "daily")
	viper.SetDefault("scraper.language", "")
	viper.SetDefault("scraper.timeout", "5m")
	viper.SetDefault("scraper.retry_max", 3)

	// Scorer weights (sum = 1.0 excluding recency)
	viper.SetDefault("scorer.daily_stars", 0.35)
	viper.SetDefault("scorer.weekly_stars", 0.25)
	viper.SetDefault("scorer.forks_rate", 0.15)
	viper.SetDefault("scorer.issue_activity", 0.10)
	viper.SetDefault("scorer.top_n", 100)

	// Logging
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")

	// Site
	// LLM defaults
	viper.SetDefault("llm.provider", "deepseek")
	viper.SetDefault("llm.api_key", "")
	viper.SetDefault("llm.model", "")
	viper.SetDefault("llm.max_tokens", 2000)
	viper.SetDefault("llm.temperature", 0.3)
	viper.SetDefault("llm.retry_max", 3)

	viper.SetDefault("site.domain", "localhost")
	viper.SetDefault("site.title", "tishi — AI 开源项目深度分析")
	viper.SetDefault("site.description", "追踪 GitHub AI 热门开源项目趋势，提供中文深度分析报告")

}
