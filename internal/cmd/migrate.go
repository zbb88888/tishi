package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/config"
	"github.com/zbb88888/tishi/internal/db"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "执行数据库迁移",
	Long:  "执行所有未应用的数据库迁移脚本。",
	RunE:  runMigrate,
}

func init() {
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "执行所有待迁移",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMigration("up")
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "回滚最后一次迁移",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMigration("down")
	},
}

func runMigrate(cmd *cobra.Command, args []string) error {
	return runMigration("up")
}

func runMigration(direction string) error {
	cfg := config.Get()
	log := logger.Named("migrate")

	dsn := db.BuildDSN(cfg.Database)
	log.Info("执行数据库迁移",
		zap.String("direction", direction),
		zap.String("host", cfg.Database.Host),
		zap.String("database", cfg.Database.Name),
	)

	if err := db.RunMigrations(dsn, direction); err != nil {
		log.Error("数据库迁移失败", zap.Error(err))
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	log.Info("数据库迁移完成")
	return nil
}
