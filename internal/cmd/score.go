package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/config"
	"github.com/zbb88888/tishi/internal/datastore"
	"github.com/zbb88888/tishi/internal/scorer"
)

var scoreCmd = &cobra.Command{
	Use:   "score",
	Short: "计算项目评分并生成排行榜",
	Long:  "基于 Trending 数据和项目指标进行多维加权评分，生成 data/rankings/{date}.json。",
	RunE:  runScore,
}

func runScore(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	log := logger.Named("score")

	store := datastore.NewStore(cfg.DataDir, log)
	sc := scorer.New(store, log, cfg.Scorer)

	log.Info("开始评分排名")
	if err := sc.Run(); err != nil {
		log.Error("评分排名失败", zap.Error(err))
		return err
	}

	log.Info("评分排名完成")
	return nil
}
