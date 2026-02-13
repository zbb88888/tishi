package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/config"
	"github.com/zbb88888/tishi/internal/datastore"
	"github.com/zbb88888/tishi/internal/generator"
)

var generateCmd = &cobra.Command{
	Use:   "generate [weekly|spotlight]",
	Short: "从 data/ 数据生成博客文章",
	Long:  "基于排行榜和项目分析数据，自动生成周报或项目深度解读文章。",
	Args:  cobra.ExactArgs(1),
	RunE:  runGenerate,
}

var (
	generateID  string
	generateDry bool
)

func init() {
	generateCmd.Flags().StringVar(&generateID, "id", "", "项目 ID（spotlight 类型必填）")
	generateCmd.Flags().BoolVar(&generateDry, "dry-run", false, "仅打印内容，不写文件")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	postType := args[0]
	cfg := config.Get()
	log := logger.Named("generate")

	store := datastore.NewStore(cfg.DataDir, log)

	g := generator.New(store, log)

	opts := generator.RunOptions{
		ProjectID: generateID,
		DryRun:    generateDry,
	}

	if err := g.Run(postType, opts); err != nil {
		log.Error("内容生成失败", zap.Error(err))
		return err
	}

	return nil
}
