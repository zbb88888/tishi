package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/config"
	"github.com/zbb88888/tishi/internal/datastore"
	"github.com/zbb88888/tishi/internal/scraper"
)

var scrapeCmd = &cobra.Command{
	Use:   "scrape",
	Short: "爬取 GitHub Trending AI 项目",
	Long:  "从 GitHub Trending 页面爬取项目列表，过滤 AI 项目，调用 GitHub API 补充数据，输出到 data/ 目录。",
	RunE:  runScrape,
}

var (
	scrapeSince    string
	scrapeLanguage string
	scrapeDryRun   bool
)

func init() {
	scrapeCmd.Flags().StringVar(&scrapeSince, "since", "", "Trending 周期: daily (默认) 或 weekly")
	scrapeCmd.Flags().StringVar(&scrapeLanguage, "language", "", "按编程语言过滤 (如 python)")
	scrapeCmd.Flags().BoolVar(&scrapeDryRun, "dry-run", false, "仅打印候选列表，不写文件")
}

func runScrape(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	log := logger.Named("scrape")

	store := datastore.NewStore(cfg.DataDir, log)

	var opts []scraper.Option
	since := cfg.Scraper.Since
	if scrapeSince != "" {
		since = scrapeSince
	}
	opts = append(opts, scraper.WithSince(since))

	lang := cfg.Scraper.Language
	if scrapeLanguage != "" {
		lang = scrapeLanguage
	}
	if lang != "" {
		opts = append(opts, scraper.WithLanguage(lang))
	}

	if scrapeDryRun {
		opts = append(opts, scraper.WithDryRun(true))
	}

	sc, err := scraper.New(store, log, cfg.GitHub.Tokens, opts...)
	if err != nil {
		return err
	}

	log.Info("开始爬取 Trending",
		zap.String("since", since),
		zap.String("language", lang),
		zap.Bool("dry_run", scrapeDryRun),
	)

	if err := sc.Run(cmd.Context()); err != nil {
		log.Error("爬取失败", zap.Error(err))
		return err
	}

	log.Info("爬取完成")
	return nil
}
