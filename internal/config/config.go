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
	Database  DatabaseConfig  `mapstructure:"database"`
	Server    ServerConfig    `mapstructure:"server"`
	GitHub    GitHubConfig    `mapstructure:"github"`
	Collector CollectorConfig `mapstructure:"collector"`
	Analyzer  AnalyzerConfig  `mapstructure:"analyzer"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Site      SiteConfig      `mapstructure:"site"`
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Name            string        `mapstructure:"name"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// GitHubConfig holds GitHub API settings.
type GitHubConfig struct {
	Tokens        []string `mapstructure:"tokens"`
	SearchQueries []string `mapstructure:"search_queries"`
}

// CollectorConfig holds data collection settings.
type CollectorConfig struct {
	TopN     int           `mapstructure:"top_n"`
	MinStars int           `mapstructure:"min_stars"`
	Timeout  time.Duration `mapstructure:"timeout"`
	RetryMax int           `mapstructure:"retry_max"`
}

// AnalyzerConfig holds trend analysis settings.
type AnalyzerConfig struct {
	Weights WeightsConfig `mapstructure:"weights"`
}

// WeightsConfig holds scoring weight parameters.
type WeightsConfig struct {
	DailyStar     float64 `mapstructure:"daily_star"`
	WeeklyStar    float64 `mapstructure:"weekly_star"`
	ForkRatio     float64 `mapstructure:"fork_ratio"`
	IssueActivity float64 `mapstructure:"issue_activity"`
	Recency       float64 `mapstructure:"recency"`
}

// SchedulerConfig holds cron schedule expressions.
type SchedulerConfig struct {
	CollectCron string `mapstructure:"collect_cron"`
	AnalyzeCron string `mapstructure:"analyze_cron"`
	BuildCron   string `mapstructure:"build_cron"`
	WeeklyCron  string `mapstructure:"weekly_cron"`
	MonthlyCron string `mapstructure:"monthly_cron"`
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

	// Environment variables: TISHI_DATABASE_HOST, TISHI_SERVER_PORT, etc.
	viper.SetEnvPrefix("TISHI")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Also support non-prefixed env vars for common ones
	_ = viper.BindEnv("database.host", "DB_HOST")
	_ = viper.BindEnv("database.port", "DB_PORT")
	_ = viper.BindEnv("database.name", "DB_NAME")
	_ = viper.BindEnv("database.user", "DB_USER")
	_ = viper.BindEnv("database.password", "DB_PASSWORD")
	_ = viper.BindEnv("database.sslmode", "DB_SSLMODE")
	_ = viper.BindEnv("github.tokens", "GITHUB_TOKENS")
	_ = viper.BindEnv("logging.level", "LOG_LEVEL")
	_ = viper.BindEnv("logging.format", "LOG_FORMAT")
	_ = viper.BindEnv("server.host", "SERVER_HOST")
	_ = viper.BindEnv("server.port", "SERVER_PORT")

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
	// Database
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.name", "tishi_db")
	viper.SetDefault("database.user", "tishi")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_open_conns", 10)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "1h")

	// Server
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "10s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "60s")

	// Collector
	viper.SetDefault("collector.top_n", 100)
	viper.SetDefault("collector.min_stars", 100)
	viper.SetDefault("collector.timeout", "30m")
	viper.SetDefault("collector.retry_max", 3)

	// Analyzer weights
	viper.SetDefault("analyzer.weights.daily_star", 0.30)
	viper.SetDefault("analyzer.weights.weekly_star", 0.25)
	viper.SetDefault("analyzer.weights.fork_ratio", 0.15)
	viper.SetDefault("analyzer.weights.issue_activity", 0.15)
	viper.SetDefault("analyzer.weights.recency", 0.15)

	// Scheduler cron expressions
	viper.SetDefault("scheduler.collect_cron", "0 0 * * *")
	viper.SetDefault("scheduler.analyze_cron", "0 1 * * *")
	viper.SetDefault("scheduler.build_cron", "0 2 * * *")
	viper.SetDefault("scheduler.weekly_cron", "0 6 * * 0")
	viper.SetDefault("scheduler.monthly_cron", "0 6 1 * *")

	// Logging
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")

	// Site
	viper.SetDefault("site.domain", "localhost")
	viper.SetDefault("site.title", "tishi — AI 开源趋势追踪")
	viper.SetDefault("site.description", "追踪 GitHub AI Top 100 热门项目")

	// Default search queries
	viper.SetDefault("github.search_queries", []string{
		"topic:llm stars:>100",
		"topic:large-language-model stars:>100",
		"topic:ai-agent stars:>100",
		"topic:machine-learning stars:>500",
		"topic:deep-learning stars:>500",
		"topic:stable-diffusion stars:>100",
		"topic:rag stars:>100",
		"topic:transformers stars:>200",
		"topic:vector-database stars:>100",
		"topic:text-to-speech stars:>100",
		"topic:generative-ai stars:>100",
	})
}
