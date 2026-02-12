package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/analyzer"
	"github.com/zbb88888/tishi/internal/config"
	"github.com/zbb88888/tishi/internal/db"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "手动触发一次趋势分析",
	Long:  "对已采集的数据执行热度评分、排名和趋势检测。",
	RunE:  runAnalyze,
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	log := logger.Named("analyze")

	pool, err := db.Connect(cmd.Context(), cfg.Database)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	defer pool.Close()

	a := analyzer.New(pool, log, cfg)

	log.Info("开始趋势分析")
	if err := a.Run(cmd.Context()); err != nil {
		log.Error("趋势分析失败", zap.Error(err))
		return err
	}

	log.Info("趋势分析完成")
	return nil
}
