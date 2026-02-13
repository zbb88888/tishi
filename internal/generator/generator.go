// Package generator creates blog posts from JSON data files.
// Reads rankings, projects, and snapshots to produce weekly reports and spotlight articles.
package generator

import (
	"fmt"
	"sort"
	"strings"
	"text/template"
	"time"

	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/datastore"
)

// Generator produces blog post JSON files from data/ contents.
type Generator struct {
	store *datastore.Store
	log   *zap.Logger
}

// New creates a Generator instance.
func New(store *datastore.Store, log *zap.Logger) *Generator {
	return &Generator{store: store, log: log.Named("generator")}
}

// RunOptions configures a generation run.
type RunOptions struct {
	ProjectID string // for spotlight only
	DryRun    bool
}

// Run generates posts of the given type. Supported: weekly, spotlight.
func (g *Generator) Run(postType string, opts RunOptions) error {
	switch postType {
	case "weekly":
		return g.generateWeekly(opts)
	case "spotlight":
		return g.generateSpotlight(opts)
	default:
		return fmt.Errorf("unsupported post type: %s (use weekly or spotlight)", postType)
	}
}

// ── Weekly Report ──────────────────────────────────────────────

type weeklyData struct {
	Year          int
	WeekNum       int
	StartDate     string
	EndDate       string
	TotalProjects int
	NewEntries    int
	TopGainers    []weeklyProject
	NewProjects   []weeklyProject
}

type weeklyProject struct {
	Rank        int
	FullName    string
	Language    string
	Stars       int
	WeeklyStars int
	Category    string
	Summary     string
	Forks       int
}

var weeklyTpl = template.Must(template.New("weekly").Parse(
	"## 本周概览\n\n" +
		"本周 AI Trending 共追踪 {{.TotalProjects}} 个项目，{{.NewEntries}} 个新入榜。\n\n" +
		"## Star 增长 Top 10\n\n" +
		"| 排名 | 项目 | 语言 | 周增 Star | 总 Star | 分类 |\n" +
		"|------|------|------|-----------|---------|------|\n" +
		"{{- range .TopGainers}}\n" +
		"| {{.Rank}} | {{.FullName}} | {{.Language}} | +{{.WeeklyStars}} | {{.Stars}} | {{.Category}} |\n" +
		"{{- end}}\n" +
		"{{if .NewProjects}}\n" +
		"## 新入榜项目\n" +
		"{{range .NewProjects}}\n" +
		"### {{.FullName}}\n\n" +
		"> {{.Summary}}\n\n" +
		"Stars: {{.Stars}} | Forks: {{.Forks}} | Language: {{.Language}} | Category: {{.Category}}\n" +
		"{{end}}{{end}}",
))

func (g *Generator) generateWeekly(opts RunOptions) error {
	now := time.Now().UTC()
	year, week := now.ISOWeek()
	endDate := now.Format("2006-01-02")
	startDate := now.AddDate(0, 0, -7).Format("2006-01-02")

	slug := fmt.Sprintf("ai-weekly-%d-w%02d", year, week)

	ranking, err := g.store.LoadLatestRanking()
	if err != nil {
		return fmt.Errorf("loading ranking: %w", err)
	}
	if ranking == nil {
		return fmt.Errorf("没有可用的排行数据，请先运行 tishi score")
	}

	data := weeklyData{
		Year:          year,
		WeekNum:       week,
		StartDate:     startDate,
		EndDate:       endDate,
		TotalProjects: ranking.Total,
	}

	// Sort by weekly stars for top gainers
	type ri = datastore.RankingItem
	sortedItems := make([]ri, len(ranking.Items))
	copy(sortedItems, ranking.Items)

	sort.Slice(sortedItems, func(i, j int) bool {
		a, b := 0, 0
		if sortedItems[i].WeeklyStars != nil {
			a = *sortedItems[i].WeeklyStars
		}
		if sortedItems[j].WeeklyStars != nil {
			b = *sortedItems[j].WeeklyStars
		}
		return a > b
	})

	topN := 10
	if topN > len(sortedItems) {
		topN = len(sortedItems)
	}
	for i := 0; i < topN; i++ {
		it := sortedItems[i]
		wp := weeklyProject{
			Rank:     i + 1,
			FullName: it.FullName,
			Stars:    it.Stars,
		}
		if it.Language != nil {
			wp.Language = *it.Language
		}
		if it.Category != nil {
			wp.Category = *it.Category
		}
		if it.WeeklyStars != nil {
			wp.WeeklyStars = *it.WeeklyStars
		}
		if it.Summary != nil {
			wp.Summary = *it.Summary
		}
		data.TopGainers = append(data.TopGainers, wp)
	}

	// New entries (rank_change == nil)
	for _, it := range ranking.Items {
		if it.RankChange == nil {
			wp := weeklyProject{FullName: it.FullName, Stars: it.Stars}
			if it.Language != nil {
				wp.Language = *it.Language
			}
			if it.Category != nil {
				wp.Category = *it.Category
			}
			if it.Summary != nil {
				wp.Summary = *it.Summary
			}
			if p, _ := g.store.LoadProject(it.ProjectID); p != nil {
				wp.Forks = p.Forks
			}
			data.NewProjects = append(data.NewProjects, wp)
			data.NewEntries++
		}
	}

	var buf strings.Builder
	if err := weeklyTpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("rendering weekly: %w", err)
	}

	title := fmt.Sprintf("AI 开源周报 #%d | %s ~ %s", week, startDate, endDate)

	if opts.DryRun {
		g.log.Info("dry-run", zap.String("slug", slug), zap.String("title", title))
		fmt.Println(buf.String())
		return nil
	}

	post := &datastore.Post{
		Slug:     slug,
		Title:    title,
		Content:  buf.String(),
		PostType: "weekly",
	}
	if err := g.store.SavePost(post); err != nil {
		return fmt.Errorf("saving post: %w", err)
	}

	g.log.Info("周报生成完成", zap.String("slug", slug))
	return nil
}

// ── Spotlight ──────────────────────────────────────────────────

type spotlightData struct {
	FullName    string
	Summary     string
	Positioning string
	Features    []datastore.Feature
	Advantages  string
	TechStack   string
	UseCases    string
	Comparison  []datastore.ComparisonEntry
	Ecosystem   string
}

var spotlightTpl = template.Must(template.New("spotlight").Parse(
	"## 概述\n\n> {{.Summary}}\n\n{{.Positioning}}\n\n" +
		"## 核心功能\n{{range .Features}}\n- **{{.Name}}**: {{.Desc}}\n{{- end}}\n\n" +
		"## 技术亮点\n\n{{.Advantages}}\n\n" +
		"## 技术栈\n\n{{.TechStack}}\n\n" +
		"## 适用场景\n\n{{.UseCases}}\n" +
		"{{if .Comparison}}\n## 竞品对比\n\n" +
		"| 项目 | 差异 |\n|------|------|\n" +
		"{{- range .Comparison}}\n| {{.Project}} | {{.Diff}} |\n{{- end}}\n{{end}}\n" +
		"## 生态定位\n\n{{.Ecosystem}}\n",
))

func (g *Generator) generateSpotlight(opts RunOptions) error {
	if opts.ProjectID == "" {
		return fmt.Errorf("--id is required for spotlight posts")
	}

	p, err := g.store.LoadProject(opts.ProjectID)
	if err != nil {
		return fmt.Errorf("loading project: %w", err)
	}

	if p.Analysis == nil || p.Analysis.Status != "published" {
		status := "无"
		if p.Analysis != nil {
			status = p.Analysis.Status
		}
		return fmt.Errorf("项目 %s 没有已发布的分析（当前: %s），请先 tishi analyze + tishi review --approve", p.FullName, status)
	}

	slug := "spotlight-" + strings.ReplaceAll(p.FullName, "/", "-")

	data := spotlightData{
		FullName:    p.FullName,
		Summary:     p.Analysis.Summary,
		Positioning: p.Analysis.Positioning,
		Features:    p.Analysis.Features,
		Advantages:  p.Analysis.Advantages,
		TechStack:   p.Analysis.TechStack,
		UseCases:    p.Analysis.UseCases,
		Comparison:  p.Analysis.Comparison,
		Ecosystem:   p.Analysis.Ecosystem,
	}

	var buf strings.Builder
	if err := spotlightTpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("rendering spotlight: %w", err)
	}

	title := fmt.Sprintf("项目深度解读：%s", p.FullName)

	if opts.DryRun {
		g.log.Info("dry-run", zap.String("slug", slug))
		fmt.Println(buf.String())
		return nil
	}

	post := &datastore.Post{
		Slug:     slug,
		Title:    title,
		Content:  buf.String(),
		PostType: "spotlight",
	}
	if err := g.store.SavePost(post); err != nil {
		return fmt.Errorf("saving post: %w", err)
	}

	g.log.Info("spotlight 生成完成", zap.String("slug", slug), zap.String("project", p.FullName))
	return nil
}
