package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/config"
	"github.com/zbb88888/tishi/internal/content"
	"github.com/zbb88888/tishi/internal/db"
)

var generateCmd = &cobra.Command{
	Use:   "generate [weekly|monthly]",
	Short: "手动生成博客内容",
	Long:  "手动触发生成周报或月报博客文章。",
	Args:  cobra.ExactArgs(1),
	RunE:  runGenerate,
}

func runGenerate(cmd *cobra.Command, args []string) error {
	postType := args[0]
	if postType != "weekly" && postType != "monthly" {
		return fmt.Errorf("不支持的文章类型 %q，可选: weekly, monthly", postType)
	}

	cfg := config.Get()
	log := logger.Named("generate")

	pool, err := db.Connect(cmd.Context(), cfg.Database)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	defer pool.Close()

	g := content.NewGenerator(pool, log, cfg)

	log.Info("开始生成内容", zap.String("type", postType))
	if err := g.Run(cmd.Context(), postType); err != nil {
		log.Error("内容生成失败", zap.Error(err), zap.String("type", postType))
		return err
	}

	log.Info("内容生成完成", zap.String("type", postType))
	return nil
}
