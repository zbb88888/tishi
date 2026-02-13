package scraper

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v67/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/zbb88888/tishi/internal/datastore"
)

// enrichAndSave calls GitHub API for full project details, merges with existing
// data if present, and writes to data/projects/{owner}__{repo}.json.
func (s *Scraper) enrichAndSave(
	ctx context.Context,
	item TrendingItem,
	preFilterMatches []datastore.CategoryMatch,
	today string,
) (*datastore.Project, error) {
	parts := splitFullName(item.FullName)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid full_name: %s", item.FullName)
	}
	owner, repo := parts[0], parts[1]

	client := s.newGitHubClient(ctx)

	// Fetch full repo metadata
	ghRepo, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("fetching repo %s: %w", item.FullName, err)
	}

	// Fetch topics
	topics, _, err := client.Repositories.ListAllTopics(ctx, owner, repo)
	if err != nil {
		s.log.Warn("获取 topics 失败", zap.String("repo", item.FullName), zap.Error(err))
	}

	// Enrich category matching with topics
	topicMatches := matchAIProjectWithTopics(topics, s.categories)
	allMatches := mergeCategories(preFilterMatches, topicMatches)

	// Try to load existing project (for merge)
	projID := datastore.ProjectIDFromFullName(item.FullName)
	existing, _ := s.store.LoadProject(projID)

	now := time.Now().UTC()

	proj := &datastore.Project{
		ID:         projID,
		FullName:   item.FullName,
		Stars:      ghRepo.GetStargazersCount(),
		Forks:      ghRepo.GetForksCount(),
		OpenIssues: ghRepo.GetOpenIssuesCount(),
		Watchers:   ghRepo.GetWatchersCount(),
		IsArchived: ghRepo.GetArchived(),
		Categories: allMatches,
		Category:   primaryCategory(allMatches),
		Trending: &datastore.Trending{
			RankDaily:        intPtr(item.Rank),
			LastSeenTrending: &today,
		},
		UpdatedAt: now,
	}

	// String pointer fields
	if d := ghRepo.GetDescription(); d != "" {
		proj.Description = &d
	}
	if l := ghRepo.GetLanguage(); l != "" {
		proj.Language = &l
	}
	if lic := ghRepo.GetLicense(); lic != nil {
		spdx := lic.GetSPDXID()
		proj.License = &spdx
	}
	if hp := ghRepo.GetHomepage(); hp != "" {
		proj.Homepage = &hp
	}
	if topics != nil {
		proj.Topics = topics
	}

	// Time fields
	if t := ghRepo.GetPushedAt(); !t.IsZero() {
		pushed := t.Time
		proj.PushedAt = &pushed
	}
	if t := ghRepo.GetCreatedAt(); !t.IsZero() {
		created := t.Time
		proj.CreatedAtGH = &created
	}

	// Period stars from Trending
	if s.since == "daily" && item.PeriodStars > 0 {
		proj.Trending.DailyStars = &item.PeriodStars
	}
	if s.since == "weekly" && item.PeriodStars > 0 {
		proj.Trending.WeeklyStars = &item.PeriodStars
	}

	// Merge with existing project data
	if existing != nil {
		proj.FirstSeenAt = existing.FirstSeenAt
		proj.Score = existing.Score
		proj.Rank = existing.Rank
		// Preserve analysis if already exists
		if existing.Analysis != nil {
			proj.Analysis = existing.Analysis
		}
		// Merge weekly stars from existing if we only have daily now
		if s.since == "daily" && existing.Trending != nil && existing.Trending.WeeklyStars != nil {
			proj.Trending.WeeklyStars = existing.Trending.WeeklyStars
		}
	} else {
		proj.FirstSeenAt = now
	}

	if err := s.store.SaveProject(proj); err != nil {
		return nil, fmt.Errorf("saving project %s: %w", item.FullName, err)
	}

	s.log.Debug("项目已保存",
		zap.String("repo", item.FullName),
		zap.Int("stars", proj.Stars),
		zap.Stringp("category", proj.Category),
	)

	return proj, nil
}

// FetchREADME fetches the README content for a project (used by LLM analyzer).
func FetchREADME(ctx context.Context, client *github.Client, owner, repo string) (string, error) {
	readme, _, err := client.Repositories.GetReadme(ctx, owner, repo, nil)
	if err != nil {
		return "", fmt.Errorf("fetching README: %w", err)
	}

	rawContent, err := readme.GetContent()
	if err != nil {
		return "", fmt.Errorf("decoding README content: %w", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(rawContent)
	if err != nil {
		// Content might not be base64-encoded
		return rawContent, nil
	}
	return string(decoded), nil
}

// newGitHubClient creates a GitHub API client with token rotation.
func (s *Scraper) newGitHubClient(ctx context.Context) *github.Client {
	token := s.tokens.Next()
	if token == "" {
		// Unauthenticated client (60 req/hr)
		return github.NewClient(nil)
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

// NewGitHubClientFromEnv creates a GitHub client using the first available token.
// Exported for use by other packages (e.g., LLM analyzer needs README).
func NewGitHubClientFromEnv(ctx context.Context) *github.Client {
	tokensEnv := os.Getenv("TISHI_GITHUB_TOKENS")
	if tokensEnv == "" {
		return github.NewClient(nil)
	}
	token := tokensEnv
	if idx := len(token); idx > 0 {
		// Take first token if comma-separated
		for i, c := range tokensEnv {
			if c == ',' {
				token = tokensEnv[:i]
				break
			}
		}
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func splitFullName(fullName string) []string {
	parts := make([]string, 0, 2)
	idx := 0
	for i, c := range fullName {
		if c == '/' {
			parts = append(parts, fullName[idx:i])
			idx = i + 1
		}
	}
	parts = append(parts, fullName[idx:])
	return parts
}

func intPtr(v int) *int {
	return &v
}
