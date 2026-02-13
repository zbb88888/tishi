package cmd

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/config"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "提交并推送 data/ 变更到 Git",
	Long:  "在 data/ 目录执行 git add -A && git commit && git push，将数据同步到远端仓库。",
	RunE:  runPush,
}

var pushMessage string

func init() {
	pushCmd.Flags().StringVarP(&pushMessage, "message", "m", "", "自定义 commit message")
}

func runPush(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	log := logger.Named("push")

	dataDir := cfg.DataDir
	today := time.Now().UTC().Format("2006-01-02")

	msg := pushMessage
	if msg == "" {
		msg = fmt.Sprintf("daily update %s", today)
	}

	// git add -A
	log.Info("git add", zap.String("dir", dataDir))
	if err := runGit(dataDir, "add", "-A"); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	// Check if there are changes to commit
	out, err := runGitOutput(dataDir, "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	if strings.TrimSpace(out) == "" {
		log.Info("没有数据变更，跳过 commit")
		return nil
	}

	// git commit
	log.Info("git commit", zap.String("message", msg))
	if err := runGit(dataDir, "commit", "-m", msg); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	// git push
	log.Info("git push")
	if err := runGit(dataDir, "push"); err != nil {
		return fmt.Errorf("git push: %w", err)
	}

	log.Info("数据推送完成")
	return nil
}

func runGit(dir string, args ...string) error {
	c := exec.Command("git", args...)
	c.Dir = dir
	output, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(output))
	}
	return nil
}

func runGitOutput(dir string, args ...string) (string, error) {
	c := exec.Command("git", args...)
	c.Dir = dir
	output, err := c.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, string(output))
	}
	return string(output), nil
}
