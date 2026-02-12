// Package cmd provides the CLI command definitions for tishi.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/config"
)

var (
	cfgFile string
	logger  *zap.Logger
)

// rootCmd is the base command for the tishi CLI.
var rootCmd = &cobra.Command{
	Use:   "tishi",
	Short: "tishi — GitHub AI 趋势追踪",
	Long:  "tishi 自动追踪 GitHub 上 AI 相关 Top 100 热门开源项目的发展趋势，提供排行榜、趋势分析与博客内容。",
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认: ./config.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "日志级别 (debug/info/warn/error)")

	// 添加子命令
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(collectCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(versionCmd)
}

// initConfig reads in config file and ENV variables.
func initConfig() {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logLevel := viper.GetString("logging.level")
	if lvl, _ := rootCmd.Flags().GetString("log-level"); lvl != "info" {
		logLevel = lvl
	}

	logger, err = newLogger(logLevel, cfg.Logging.Format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}
}

// newLogger creates a zap.Logger based on level and format.
func newLogger(level, format string) (*zap.Logger, error) {
	var cfg zap.Config
	if format == "console" {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	if err := cfg.Level.UnmarshalText([]byte(level)); err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", level, err)
	}

	return cfg.Build()
}
