// Package scorer computes project scores and rankings from JSON data.
package scorer

import (
	"fmt"
	"math"
	"sort"
	"time"

	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/config"
	"github.com/zbb88888/tishi/internal/datastore"
)

// Scorer computes scores and generates daily rankings.
type Scorer struct {
	store *datastore.Store
	log   *zap.Logger
	cfg   config.ScorerConfig
}

// New creates a Scorer instance.
func New(store *datastore.Store, log *zap.Logger, cfg config.ScorerConfig) *Scorer {
	return &Scorer{
		store: store,
		log:   log,
		cfg:   cfg,
	}
}

// Run executes the scoring + ranking pipeline.
func (s *Scorer) Run() error {
	start := time.Now()

	projects, err := s.store.ListProjects()
	if err != nil {
		return fmt.Errorf("listing projects: %w", err)
	}
	if len(projects) == 0 {
		s.log.Warn("没有项目可评分")
		return nil
	}

	s.log.Info("加载项目数据", zap.Int("count", len(projects)))

	// Compute scores
	s.computeScores(projects)

	// Sort by score descending
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Score > projects[j].Score
	})

	// Assign ranks (top N)
	topN := s.cfg.TopN
	if topN <= 0 || topN > len(projects) {
		topN = len(projects)
	}
	for i := 0; i < topN; i++ {
		rank := i + 1
		projects[i].Rank = &rank
	}

	// Load previous ranking for rank_change computation
	prevRanking, _ := s.store.LoadLatestRanking()
	prevRankMap := make(map[string]int)
	if prevRanking != nil {
		for _, item := range prevRanking.Items {
			prevRankMap[item.ProjectID] = item.Rank
		}
	}

	// Build today's ranking
	today := time.Now().UTC().Format("2006-01-02")
	ranking := &datastore.Ranking{
		Date:  today,
		Total: topN,
	}

	for i := 0; i < topN; i++ {
		p := projects[i]
		item := datastore.RankingItem{
			Rank:      i + 1,
			ProjectID: p.ID,
			FullName:  p.FullName,
			Language:  p.Language,
			Category:  p.Category,
			Stars:     p.Stars,
			Score:     math.Round(p.Score*100) / 100,
		}

		// summary from analysis
		if p.Analysis != nil && p.Analysis.Summary != "" {
			item.Summary = &p.Analysis.Summary
		}

		// trending stars
		if p.Trending != nil {
			item.DailyStars = p.Trending.DailyStars
			item.WeeklyStars = p.Trending.WeeklyStars
		}

		// rank change
		if prevRank, ok := prevRankMap[p.ID]; ok {
			change := prevRank - (i + 1) // positive = moved up
			item.RankChange = &change
		}
		// nil RankChange means "new entry"

		ranking.Items = append(ranking.Items, item)
	}

	// Save ranking file
	if err := s.store.SaveRanking(ranking); err != nil {
		return fmt.Errorf("saving ranking: %w", err)
	}

	// Update project files with new score and rank
	for i := 0; i < topN; i++ {
		p := projects[i]
		p.UpdatedAt = time.Now().UTC()
		if err := s.store.SaveProject(p); err != nil {
			s.log.Warn("更新项目评分失败", zap.String("id", p.ID), zap.Error(err))
		}
	}

	s.log.Info("评分排名完成",
		zap.Int("ranked", topN),
		zap.String("date", today),
		zap.Duration("elapsed", time.Since(start)),
	)
	return nil
}

// computeScores calculates a weighted score for each project.
// Formula: daily_stars * w1 + weekly_stars * w2 + forks_rate * w3 + issue_activity * w4
func (s *Scorer) computeScores(projects []*datastore.Project) {
	// Find max values for normalization
	var maxDailyStars, maxWeeklyStars, maxForks, maxIssues float64

	for _, p := range projects {
		if p.Trending != nil {
			if p.Trending.DailyStars != nil && float64(*p.Trending.DailyStars) > maxDailyStars {
				maxDailyStars = float64(*p.Trending.DailyStars)
			}
			if p.Trending.WeeklyStars != nil && float64(*p.Trending.WeeklyStars) > maxWeeklyStars {
				maxWeeklyStars = float64(*p.Trending.WeeklyStars)
			}
		}
		if float64(p.Forks) > maxForks {
			maxForks = float64(p.Forks)
		}
		if float64(p.OpenIssues) > maxIssues {
			maxIssues = float64(p.OpenIssues)
		}
	}

	// Avoid division by zero
	if maxDailyStars == 0 {
		maxDailyStars = 1
	}
	if maxWeeklyStars == 0 {
		maxWeeklyStars = 1
	}
	if maxForks == 0 {
		maxForks = 1
	}
	if maxIssues == 0 {
		maxIssues = 1
	}

	for _, p := range projects {
		var dailyNorm, weeklyNorm, forkNorm, issueNorm float64

		if p.Trending != nil {
			if p.Trending.DailyStars != nil {
				dailyNorm = float64(*p.Trending.DailyStars) / maxDailyStars
			}
			if p.Trending.WeeklyStars != nil {
				weeklyNorm = float64(*p.Trending.WeeklyStars) / maxWeeklyStars
			}
		}
		if p.Stars > 0 {
			forkNorm = float64(p.Forks) / maxForks
		}
		issueNorm = float64(p.OpenIssues) / maxIssues

		score := dailyNorm*s.cfg.DailyStars +
			weeklyNorm*s.cfg.WeeklyStars +
			forkNorm*s.cfg.ForksRate +
			issueNorm*s.cfg.IssueActivity

		// Scale to 0-100
		p.Score = math.Round(score*10000) / 100
	}
}
