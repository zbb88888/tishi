package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/config"
	"github.com/zbb88888/tishi/internal/datastore"
	"github.com/zbb88888/tishi/internal/llm"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "使用 LLM 对 AI 项目生成中文分析报告",
	Long:  "扫描 data/projects/，对未分析或过期的项目调用 DeepSeek/Qwen API 生成结构化中文分析。",
	RunE:  runAnalyze,
}

var (
	analyzeID    string
	analyzeForce bool
	analyzeDry   bool
)

func init() {
	analyzeCmd.Flags().StringVar(&analyzeID, "id", "", "指定项目 ID (owner__repo)")
	analyzeCmd.Flags().BoolVar(&analyzeForce, "force", false, "强制重新分析所有项目")
	analyzeCmd.Flags().BoolVar(&analyzeDry, "dry-run", false, "仅打印待分析项目，不调用 LLM")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	log := logger.Named("analyze")

	store := datastore.NewStore(cfg.DataDir, log)

	// Pick the first GitHub token for README fetching
	ghToken := ""
	if len(cfg.GitHub.Tokens) > 0 {
		ghToken = cfg.GitHub.Tokens[0]
	}

	analyzer, err := llm.NewAnalyzer(store, cfg.LLM, ghToken, log)
	if err != nil {
		return err
	}

	opts := llm.RunOptions{
		ProjectID: analyzeID,
		Force:     analyzeForce,
		DryRun:    analyzeDry,
	}

	if err := analyzer.Run(cmd.Context(), opts); err != nil {
		log.Error("LLM 分析失败", zap.Error(err))
		return err
	}

	return nil
}
