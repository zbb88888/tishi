package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/config"
	"github.com/zbb88888/tishi/internal/datastore"
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "审核 LLM 生成的项目分析",
	Long:  "列出所有 draft 状态的分析，或批准/拒绝指定项目的分析结果。",
	RunE:  runReview,
}

var (
	reviewApprove string
	reviewReject  string
)

func init() {
	reviewCmd.Flags().StringVar(&reviewApprove, "approve", "", "批准指定项目 ID 的分析 (owner__repo)")
	reviewCmd.Flags().StringVar(&reviewReject, "reject", "", "拒绝指定项目 ID 的分析 (owner__repo)")
}

func runReview(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	log := logger.Named("review")

	store := datastore.NewStore(cfg.DataDir, log)

	// Handle approve
	if reviewApprove != "" {
		return setAnalysisStatus(store, log, reviewApprove, "published")
	}

	// Handle reject
	if reviewReject != "" {
		return setAnalysisStatus(store, log, reviewReject, "rejected")
	}

	// Default: list all draft analyses
	projects, err := store.ListProjects()
	if err != nil {
		return fmt.Errorf("listing projects: %w", err)
	}

	var drafts int
	for _, p := range projects {
		if p.Analysis != nil && p.Analysis.Status == "draft" {
			drafts++
			fmt.Printf("[draft] %-40s  %s\n", p.FullName, p.Analysis.Summary)
		}
	}

	if drafts == 0 {
		fmt.Println("没有待审核的分析。")
	} else {
		fmt.Printf("\n共 %d 个待审核分析。使用 --approve=ID 或 --reject=ID 审核。\n", drafts)
	}

	return nil
}

func setAnalysisStatus(store *datastore.Store, log *zap.Logger, projectID, status string) error {
	p, err := store.LoadProject(projectID)
	if err != nil {
		return fmt.Errorf("loading project %s: %w", projectID, err)
	}

	if p.Analysis == nil {
		return fmt.Errorf("项目 %s 没有分析结果", p.FullName)
	}

	oldStatus := p.Analysis.Status
	p.Analysis.Status = status
	now := time.Now().UTC()
	p.Analysis.ReviewedAt = &now
	p.UpdatedAt = now

	if err := store.SaveProject(p); err != nil {
		return fmt.Errorf("saving project: %w", err)
	}

	log.Info("分析状态已更新",
		zap.String("project", p.FullName),
		zap.String("from", oldStatus),
		zap.String("to", status),
	)
	fmt.Printf("✓ %s: %s → %s\n", p.FullName, oldStatus, status)
	return nil
}
