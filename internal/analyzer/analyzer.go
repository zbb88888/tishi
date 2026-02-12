// Package analyzer computes project scores, rankings, and trend detection.
package analyzer

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/config"
)

// Analyzer computes project scores and rankings.
type Analyzer struct {
	pool *pgxpool.Pool
	log  *zap.Logger
	cfg  *config.Config
}

// New creates a new Analyzer.
func New(pool *pgxpool.Pool, log *zap.Logger, cfg *config.Config) *Analyzer {
	return &Analyzer{
		pool: pool,
		log:  log,
		cfg:  cfg,
	}
}

type projectData struct {
	ID               int64
	FullName         string
	DailyStarGain    int
	WeeklyStarGain   int
	Stars            int
	Forks            int
	OpenIssues       int
	DaysSinceCreated int
	DaysSincePush    int
	Score            float64
	Rank             int
}

// Run executes the full analysis pipeline.
func (a *Analyzer) Run(ctx context.Context) error {
	start := time.Now()

	projects, err := a.loadProjectMetrics(ctx)
	if err != nil {
		return fmt.Errorf("loading project metrics: %w", err)
	}

	if len(projects) == 0 {
		a.log.Warn("没有项目可分析")
		return nil
	}

	a.log.Info("加载项目指标", zap.Int("count", len(projects)))
	a.computeScores(projects)
	a.generateRankings(projects)

	if err := a.updateDatabase(ctx, projects); err != nil {
		return fmt.Errorf("updating database: %w", err)
	}

	a.log.Info("分析完成",
		zap.Int("projects_scored", len(projects)),
		zap.Duration("elapsed", time.Since(start)),
	)

	return nil
}

func (a *Analyzer) loadProjectMetrics(ctx context.Context) ([]*projectData, error) {
	rows, err := a.pool.Query(ctx, `
		SELECT
			p.id,
			p.full_name,
			p.stargazers_count,
			p.forks_count,
			p.open_issues_count,
			EXTRACT(DAY FROM NOW() - p.created_at_gh)::INT AS days_since_created,
			EXTRACT(DAY FROM NOW() - p.pushed_at)::INT AS days_since_push,
			COALESCE(today.stargazers_count - yesterday.stargazers_count, 0) AS daily_gain,
			COALESCE(today.stargazers_count - week_ago.stargazers_count, 0) AS weekly_gain
		FROM projects p
		LEFT JOIN daily_snapshots today
			ON p.id = today.project_id AND today.snapshot_date = CURRENT_DATE
		LEFT JOIN daily_snapshots yesterday
			ON p.id = yesterday.project_id AND yesterday.snapshot_date = CURRENT_DATE - 1
		LEFT JOIN daily_snapshots week_ago
			ON p.id = week_ago.project_id AND week_ago.snapshot_date = CURRENT_DATE - 7
		WHERE p.is_archived = FALSE
		ORDER BY p.stargazers_count DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*projectData
	for rows.Next() {
		p := &projectData{}
		if err := rows.Scan(
			&p.ID, &p.FullName, &p.Stars, &p.Forks, &p.OpenIssues,
			&p.DaysSinceCreated, &p.DaysSincePush,
			&p.DailyStarGain, &p.WeeklyStarGain,
		); err != nil {
			return nil, fmt.Errorf("scanning project row: %w", err)
		}
		projects = append(projects, p)
	}

	return projects, rows.Err()
}

func (a *Analyzer) computeScores(projects []*projectData) {
	w := a.cfg.Analyzer.Weights

	dailyGains := make([]float64, len(projects))
	weeklyGains := make([]float64, len(projects))
	forkRatios := make([]float64, len(projects))
	issueActivities := make([]float64, len(projects))
	recencies := make([]float64, len(projects))

	for i, p := range projects {
		dailyGains[i] = float64(p.DailyStarGain)
		weeklyGains[i] = float64(p.WeeklyStarGain)

		if p.Stars > 0 {
			forkRatios[i] = float64(p.Forks) / float64(p.Stars)
		}
		if p.DaysSinceCreated > 0 {
			issueActivities[i] = float64(p.OpenIssues) / float64(p.DaysSinceCreated+1)
		}
		if p.DaysSincePush >= 0 {
			recencies[i] = 1.0 / float64(p.DaysSincePush+1)
		}
	}

	normDaily := minMaxNormalize(dailyGains)
	normWeekly := minMaxNormalize(weeklyGains)
	normFork := minMaxNormalize(forkRatios)
	normIssue := minMaxNormalize(issueActivities)
	normRecency := minMaxNormalize(recencies)

	for i, p := range projects {
		p.Score = w.DailyStar*normDaily[i] +
			w.WeeklyStar*normWeekly[i] +
			w.ForkRatio*normFork[i] +
			w.IssueActivity*normIssue[i] +
			w.Recency*normRecency[i]

		p.Score = math.Max(0, math.Min(100, p.Score))
		p.Score = math.Round(p.Score*100) / 100
	}
}

func (a *Analyzer) generateRankings(projects []*projectData) {
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Score > projects[j].Score
	})

	for i, p := range projects {
		p.Rank = i + 1
	}
}

func (a *Analyzer) updateDatabase(ctx context.Context, projects []*projectData) error {
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	today := time.Now().UTC().Truncate(24 * time.Hour)

	for _, p := range projects {
		if _, err := tx.Exec(ctx,
			`UPDATE projects SET score = $2, rank = $3 WHERE id = $1`,
			p.ID, p.Score, p.Rank,
		); err != nil {
			return fmt.Errorf("updating project %d score: %w", p.ID, err)
		}

		if _, err := tx.Exec(ctx,
			`UPDATE daily_snapshots SET score = $3, rank = $4 WHERE project_id = $1 AND snapshot_date = $2`,
			p.ID, today, p.Score, p.Rank,
		); err != nil {
			a.log.Debug("更新快照评分失败",
				zap.Int64("project_id", p.ID),
				zap.Error(err),
			)
		}
	}

	return tx.Commit(ctx)
}

func minMaxNormalize(values []float64) []float64 {
	if len(values) == 0 {
		return nil
	}

	minVal, maxVal := values[0], values[0]
	for _, v := range values[1:] {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	result := make([]float64, len(values))
	rang := maxVal - minVal
	if rang == 0 {
		for i := range result {
			result[i] = 50
		}
		return result
	}

	for i, v := range values {
		result[i] = ((v - minVal) / rang) * 100
	}
	return result
}
