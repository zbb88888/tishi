package cmd

import (
"fmt"

"github.com/spf13/cobra"
"go.uber.org/zap"

"github.com/zbb88888/tishi/internal/collector"
"github.com/zbb88888/tishi/internal/config"
"github.com/zbb88888/tishi/internal/db"
)

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "手动触发一次数据采集",
	Long:  "立即执行一次 GitHub AI 项目数据采集，将结果写入数据库。",
	RunE:  runCollect,
}

func runCollect(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	log := logger.Named("collect")

	pool, err := db.Connect(cmd.Context(), cfg.Database)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	defer pool.Close()

	c := collector.New(pool, log, cfg)

	log.Info("开始数据采集")
	if err := c.Run(cmd.Context()); err != nil {
		log.Error("数据采集失败", zap.Error(err))
		return err
	}

	log.Info("数据采集完成")
	return nil
}
