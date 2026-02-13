package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v67/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/zbb88888/tishi/internal/config"
	"github.com/zbb88888/tishi/internal/datastore"
)

// Analyzer orchestrates LLM analysis for projects that need it.
type Analyzer struct {
	store  *datastore.Store
	client *Client
	gh     *github.Client
	log    *zap.Logger
	cfg    config.LLMConfig
}

// NewAnalyzer creates an Analyzer with LLM client and GitHub client for README fetching.
func NewAnalyzer(store *datastore.Store, llmCfg config.LLMConfig, ghToken string, log *zap.Logger) (*Analyzer, error) {
	client, err := NewClient(llmCfg, log)
	if err != nil {
		return nil, fmt.Errorf("creating LLM client: %w", err)
	}

	var gh *github.Client
	if ghToken != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ghToken})
		gh = github.NewClient(oauth2.NewClient(context.Background(), ts))
	} else {
		gh = github.NewClient(nil)
	}

	return &Analyzer{
		store:  store,
		client: client,
		gh:     gh,
		log:    log.Named("analyzer"),
		cfg:    llmCfg,
	}, nil
}

// RunOptions configures an analysis run.
type RunOptions struct {
	ProjectID string // empty = all eligible projects
	Force     bool   // force re-analyze even if analysis exists
	DryRun    bool   // print prompts only, don't call LLM
}

// Run executes the analysis pipeline.
func (a *Analyzer) Run(ctx context.Context, opts RunOptions) error {
	start := time.Now()

	var projects []*datastore.Project

	if opts.ProjectID != "" {
		// Single project
		p, err := a.store.LoadProject(opts.ProjectID)
		if err != nil {
			return fmt.Errorf("loading project %s: %w", opts.ProjectID, err)
		}
		projects = []*datastore.Project{p}
	} else {
		// All projects
		all, err := a.store.ListProjects()
		if err != nil {
			return fmt.Errorf("listing projects: %w", err)
		}
		projects = all
	}

	// Filter projects that need analysis
	var candidates []*datastore.Project
	for _, p := range projects {
		if opts.Force || needsAnalysis(p) {
			candidates = append(candidates, p)
		}
	}

	a.log.Info("待分析项目",
		zap.Int("candidates", len(candidates)),
		zap.Int("total", len(projects)),
		zap.Bool("force", opts.Force),
	)

	if len(candidates) == 0 {
		a.log.Info("没有需要分析的项目")
		return nil
	}

	var analyzed, failed, totalTokens int

	for _, p := range candidates {
		select {
		case <-ctx.Done():
			a.log.Warn("分析被取消", zap.Error(ctx.Err()))
			return ctx.Err()
		default:
		}

		if opts.DryRun {
			a.log.Info("dry-run: 将分析项目",
				zap.String("project", p.FullName),
				zap.Bool("has_analysis", p.Analysis != nil),
			)
			continue
		}

		if err := a.analyzeOne(ctx, p); err != nil {
			a.log.Warn("分析失败，跳过",
				zap.String("project", p.FullName),
				zap.Error(err),
			)
			failed++
			continue
		}

		if p.Analysis != nil && p.Analysis.TokenUsage != nil {
			totalTokens += *p.Analysis.TokenUsage
		}
		analyzed++

		// Rate limiting: pause between API calls
		time.Sleep(2 * time.Second)
	}

	a.log.Info("LLM 分析完成",
		zap.Int("analyzed", analyzed),
		zap.Int("failed", failed),
		zap.Int("total_tokens", totalTokens),
		zap.Duration("elapsed", time.Since(start)),
	)
	return nil
}

// analyzeOne fetches README + calls LLM + saves result for a single project.
func (a *Analyzer) analyzeOne(ctx context.Context, p *datastore.Project) error {
	parts := strings.SplitN(p.FullName, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid full_name: %s", p.FullName)
	}
	owner, repo := parts[0], parts[1]

	// Fetch README
	readme, err := fetchREADME(ctx, a.gh, owner, repo)
	if err != nil {
		a.log.Warn("README 获取失败，继续不含 README",
			zap.String("project", p.FullName),
			zap.Error(err),
		)
		readme = "(README 不可用)"
	}

	// Call LLM
	analysis, err := a.client.AnalyzeProject(ctx, p, readme)
	if err != nil {
		return fmt.Errorf("LLM analyze: %w", err)
	}

	// Save
	p.Analysis = analysis
	p.UpdatedAt = time.Now().UTC()

	if err := a.store.SaveProject(p); err != nil {
		return fmt.Errorf("saving project: %w", err)
	}

	return nil
}

// needsAnalysis returns true if a project should be (re-)analyzed.
func needsAnalysis(p *datastore.Project) bool {
	if p.Analysis == nil {
		return true
	}
	// Re-analyze if draft and older than 7 days
	if p.Analysis.Status == "draft" {
		return time.Since(p.Analysis.GeneratedAt) > 7*24*time.Hour
	}
	return false
}

// fetchREADME fetches the README content from GitHub.
func fetchREADME(ctx context.Context, gh *github.Client, owner, repo string) (string, error) {
	readme, _, err := gh.Repositories.GetReadme(ctx, owner, repo, nil)
	if err != nil {
		return "", fmt.Errorf("GitHub API: %w", err)
	}
	content, err := readme.GetContent()
	if err != nil {
		return "", fmt.Errorf("decode content: %w", err)
	}
	return content, nil
}
