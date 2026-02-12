// Package content generates blog posts (weekly/monthly reports) from analysis data.
package content

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/config"
)

// Generator produces Markdown blog posts from project data.
type Generator struct {
	pool *pgxpool.Pool
	log  *zap.Logger
	cfg  *config.Config
}

// NewGenerator creates a new content Generator.
func NewGenerator(pool *pgxpool.Pool, log *zap.Logger, cfg *config.Config) *Generator {
	return &Generator{
		pool: pool,
		log:  log,
		cfg:  cfg,
	}
}

// TopGainer represents a project with highest star gains.
type TopGainer struct {
	Rank       int
	FullName   string
	WeeklyGain int
	Stars      int
	Category   string
}

// WeeklyData holds data for weekly report generation.
type WeeklyData struct {
	WeekNumber  int
	DateRange   string
	Year        int
	TopGainers  []TopGainer
	NewEntries  int
	PublishedAt string
}

// Run generates a blog post of the specified type.
func (g *Generator) Run(ctx context.Context, postType string) error {
	switch postType {
	case "weekly":
		return g.generateWeekly(ctx)
	case "monthly":
		return g.generateMonthly(ctx)
	default:
		return fmt.Errorf("unsupported post type: %s", postType)
	}
}

func (g *Generator) generateWeekly(ctx context.Context) error {
	now := time.Now().UTC()
	_, week := now.ISOWeek()

	topGainers, err := g.queryTopGainers(ctx, 10)
	if err != nil {
		return fmt.Errorf("querying top gainers: %w", err)
	}

	data := WeeklyData{
		WeekNumber:  week,
		Year:        now.Year(),
		DateRange:   fmt.Sprintf("%s ~ %s", now.AddDate(0, 0, -7).Format("2006-01-02"), now.Format("2006-01-02")),
		TopGainers:  topGainers,
		PublishedAt: now.Format(time.RFC3339),
	}

	postContent, err := g.renderTemplate(weeklyTemplate, data)
	if err != nil {
		return fmt.Errorf("rendering weekly template: %w", err)
	}

	slug := fmt.Sprintf("ai-weekly-%d-w%02d", now.Year(), week)
	title := fmt.Sprintf("AI 开源周报 #%d | %s", week, data.DateRange)

	_, err = g.pool.Exec(ctx, `
		INSERT INTO blog_posts (title, slug, content, post_type, published_at)
		VALUES ($1, $2, $3, 'weekly', $4)
		ON CONFLICT (slug) DO UPDATE SET
			title = EXCLUDED.title,
			content = EXCLUDED.content,
			published_at = EXCLUDED.published_at`,
		title, slug, postContent, now,
	)
	if err != nil {
		return fmt.Errorf("upserting blog post: %w", err)
	}

	g.log.Info("周报生成完成", zap.String("slug", slug))
	return nil
}

func (g *Generator) generateMonthly(ctx context.Context) error {
	now := time.Now().UTC()

	slug := fmt.Sprintf("ai-monthly-%d-%02d", now.Year(), now.Month())
	title := fmt.Sprintf("AI 开源月报 | %d年%d月", now.Year(), now.Month())
	postContent := fmt.Sprintf("# %s\n\n> 月报生成中...\n", title)

	_, err := g.pool.Exec(ctx, `
		INSERT INTO blog_posts (title, slug, content, post_type, published_at)
		VALUES ($1, $2, $3, 'monthly', $4)
		ON CONFLICT (slug) DO UPDATE SET
			title = EXCLUDED.title,
			content = EXCLUDED.content,
			published_at = EXCLUDED.published_at`,
		title, slug, postContent, now,
	)
	if err != nil {
		return fmt.Errorf("upserting monthly post: %w", err)
	}

	g.log.Info("月报生成完成", zap.String("slug", slug))
	return nil
}

func (g *Generator) queryTopGainers(ctx context.Context, limit int) ([]TopGainer, error) {
	rows, err := g.pool.Query(ctx, `
		SELECT
			p.full_name,
			p.stargazers_count,
			COALESCE(today.stargazers_count - week_ago.stargazers_count, 0) AS weekly_gain
		FROM projects p
		LEFT JOIN daily_snapshots today
			ON p.id = today.project_id AND today.snapshot_date = CURRENT_DATE
		LEFT JOIN daily_snapshots week_ago
			ON p.id = week_ago.project_id AND week_ago.snapshot_date = CURRENT_DATE - 7
		WHERE p.is_archived = FALSE
		ORDER BY weekly_gain DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gainers []TopGainer
	rank := 1
	for rows.Next() {
		var tg TopGainer
		if err := rows.Scan(&tg.FullName, &tg.Stars, &tg.WeeklyGain); err != nil {
			return nil, err
		}
		tg.Rank = rank
		rank++
		gainers = append(gainers, tg)
	}
	return gainers, rows.Err()
}

func (g *Generator) renderTemplate(tmpl string, data interface{}) (string, error) {
	t, err := template.New("post").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}
	return buf.String(), nil
}

var weeklyTemplate = "# AI 开源周报 #{{.WeekNumber}} | {{.DateRange}}\n\n" +
	"> 发布时间: {{.PublishedAt}}\n\n" +
	"## 本周 Star 增长 Top 10\n\n" +
	"| 排名 | 项目 | 周增 Star | 总 Star |\n" +
	"|------|------|-----------|---------|" +
	"{{range .TopGainers}}\n" +
	"| {{.Rank}} | {{.FullName}} | +{{.WeeklyGain}} | {{.Stars}} |" +
	"{{end}}\n\n" +
	"---\n\n" +
	"*本报告由 tishi 自动生成*\n"
