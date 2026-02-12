// Package collector handles GitHub API data collection for AI-related projects.
package collector

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v67/github"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/time/rate"

	"github.com/zbb88888/tishi/internal/config"
)

// Collector fetches AI-related project data from GitHub.
type Collector struct {
	pool   *pgxpool.Pool
	log    *zap.Logger
	cfg    *config.Config
	tokens *TokenRotator
}

// New creates a new Collector instance.
func New(pool *pgxpool.Pool, log *zap.Logger, cfg *config.Config) *Collector {
	return &Collector{
		pool:   pool,
		log:    log,
		cfg:    cfg,
		tokens: NewTokenRotator(cfg.GitHub.Tokens),
	}
}

// Run executes a full collection cycle.
func (c *Collector) Run(ctx context.Context) error {
	start := time.Now()

	client := c.newGitHubClient(ctx)

	candidates, err := c.searchProjects(ctx, client)
	if err != nil {
		return fmt.Errorf("searching projects: %w", err)
	}

	c.log.Info("搜索完成",
		zap.Int("candidates", len(candidates)),
		zap.Duration("elapsed", time.Since(start)),
	)

	unique := c.deduplicateProjects(candidates)
	c.log.Info("去重完成", zap.Int("unique", len(unique)))

	today := time.Now().UTC().Truncate(24 * time.Hour)
	var upserted, snapshotted int

	for _, repo := range unique {
		projectID, err := c.upsertProject(ctx, repo)
		if err != nil {
			c.log.Warn("upsert 项目失败",
				zap.String("repo", repo.GetFullName()),
				zap.Error(err),
			)
			continue
		}
		upserted++

		if err := c.insertSnapshot(ctx, projectID, repo, today); err != nil {
			c.log.Warn("插入快照失败",
				zap.String("repo", repo.GetFullName()),
				zap.Error(err),
			)
			continue
		}
		snapshotted++
	}

	c.log.Info("采集完成",
		zap.Int("upserted", upserted),
		zap.Int("snapshotted", snapshotted),
		zap.Duration("total_elapsed", time.Since(start)),
	)

	return nil
}

func (c *Collector) searchProjects(ctx context.Context, client *github.Client) ([]*github.Repository, error) {
	var (
		all []*github.Repository
		mu  sync.Mutex
	)

	limiter := rate.NewLimiter(rate.Every(2*time.Second), 1)

	for _, query := range c.cfg.GitHub.SearchQueries {
		if err := limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter wait: %w", err)
		}

		c.log.Debug("执行搜索", zap.String("query", query))

		opts := &github.SearchOptions{
			Sort:  "stars",
			Order: "desc",
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		}

		result, resp, err := client.Search.Repositories(ctx, query, opts)
		if err != nil {
			if resp != nil && resp.StatusCode == 403 {
				c.log.Warn("rate limit, 等待重试", zap.String("query", query))
				c.waitForRateLimit(resp)
				result, _, err = client.Search.Repositories(ctx, query, opts)
				if err != nil {
					c.log.Error("重试后仍失败", zap.String("query", query), zap.Error(err))
					continue
				}
			} else {
				c.log.Error("搜索失败", zap.String("query", query), zap.Error(err))
				continue
			}
		}

		if result != nil {
			mu.Lock()
			all = append(all, result.Repositories...)
			mu.Unlock()
			c.log.Debug("搜索结果",
				zap.String("query", query),
				zap.Int("count", len(result.Repositories)),
				zap.Int("total", result.GetTotal()),
			)
		}
	}

	return all, nil
}

func (c *Collector) deduplicateProjects(repos []*github.Repository) []*github.Repository {
	seen := make(map[int64]*github.Repository)
	for _, repo := range repos {
		id := repo.GetID()
		if existing, ok := seen[id]; !ok || repo.GetStargazersCount() > existing.GetStargazersCount() {
			seen[id] = repo
		}
	}

	unique := make([]*github.Repository, 0, len(seen))
	for _, repo := range seen {
		unique = append(unique, repo)
	}
	return unique
}

func (c *Collector) upsertProject(ctx context.Context, repo *github.Repository) (int64, error) {
	topics := repo.Topics
	if topics == nil {
		topics = []string{}
	}

	var projectID int64
	err := c.pool.QueryRow(ctx, `
		INSERT INTO projects (
			github_id, full_name, description, language, license,
			topics, homepage, created_at_gh, pushed_at, metadata,
			stargazers_count, forks_count, open_issues_count, watchers_count
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb, $11, $12, $13, $14
		)
		ON CONFLICT (github_id) DO UPDATE SET
			full_name = EXCLUDED.full_name,
			description = EXCLUDED.description,
			language = EXCLUDED.language,
			license = EXCLUDED.license,
			topics = EXCLUDED.topics,
			homepage = EXCLUDED.homepage,
			pushed_at = EXCLUDED.pushed_at,
			metadata = EXCLUDED.metadata,
			stargazers_count = EXCLUDED.stargazers_count,
			forks_count = EXCLUDED.forks_count,
			open_issues_count = EXCLUDED.open_issues_count,
			watchers_count = EXCLUDED.watchers_count
		RETURNING id`,
		repo.GetID(),
		repo.GetFullName(),
		repo.GetDescription(),
		repo.GetLanguage(),
		getLicense(repo),
		topics,
		repo.GetHomepage(),
		repo.GetCreatedAt().Time,
		repo.GetPushedAt().Time,
		`{}`,
		repo.GetStargazersCount(),
		repo.GetForksCount(),
		repo.GetOpenIssuesCount(),
		repo.GetWatchersCount(),
	).Scan(&projectID)

	return projectID, err
}

func (c *Collector) insertSnapshot(ctx context.Context, projectID int64, repo *github.Repository, date time.Time) error {
	_, err := c.pool.Exec(ctx, `
		INSERT INTO daily_snapshots (
			project_id, snapshot_date,
			stargazers_count, forks_count, open_issues_count, watchers_count
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (project_id, snapshot_date) DO UPDATE SET
			stargazers_count = EXCLUDED.stargazers_count,
			forks_count = EXCLUDED.forks_count,
			open_issues_count = EXCLUDED.open_issues_count,
			watchers_count = EXCLUDED.watchers_count`,
		projectID,
		date,
		repo.GetStargazersCount(),
		repo.GetForksCount(),
		repo.GetOpenIssuesCount(),
		repo.GetWatchersCount(),
	)
	return err
}

func (c *Collector) newGitHubClient(ctx context.Context) *github.Client {
	token := c.tokens.Next()
	if token == "" {
		return github.NewClient(nil)
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func (c *Collector) waitForRateLimit(resp *github.Response) {
	if resp.Rate.Reset.Time.After(time.Now()) {
		wait := time.Until(resp.Rate.Reset.Time) + time.Second
		c.log.Info("等待 rate limit 重置", zap.Duration("wait", wait))
		time.Sleep(wait)
	}
}

func getLicense(repo *github.Repository) string {
	if repo.License != nil {
		return repo.License.GetSPDXID()
	}
	return ""
}

// TopicMatches checks if any of the repo topics match the given keywords.
func TopicMatches(topics []string, keywords []string) bool {
	for _, t := range topics {
		lower := strings.ToLower(t)
		for _, kw := range keywords {
			if lower == kw {
				return true
			}
		}
	}
	return false
}
