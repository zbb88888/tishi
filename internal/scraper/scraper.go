// Package scraper fetches AI projects from GitHub Trending and enriches them.
//
// Pipeline: Trending HTML -> Colly parse -> AI keyword filter -> GitHub API enrich -> data/ JSON output
package scraper

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/datastore"
)

// TrendingItem is a raw item parsed from the GitHub Trending HTML page.
type TrendingItem struct {
	FullName    string // owner/repo
	Description string
	Language    string
	Stars       int // total stars
	Forks       int // total forks
	PeriodStars int // stars gained in the period (daily/weekly)
	Rank        int // position on the Trending page (1-based)
}

// Scraper orchestrates the full scrape pipeline.
type Scraper struct {
	store      *datastore.Store
	log        *zap.Logger
	categories []datastore.Category
	tokens     *TokenRotator
	since      string // "daily" or "weekly"
	language   string // optional language filter
	dryRun     bool   // if true, don't write files
}

// Option configures the Scraper.
type Option func(*Scraper)

// WithSince sets the trending period (daily/weekly). Default: daily.
func WithSince(since string) Option {
	return func(s *Scraper) { s.since = since }
}

// WithLanguage sets a language filter for the Trending page.
func WithLanguage(lang string) Option {
	return func(s *Scraper) { s.language = lang }
}

// WithDryRun enables dry-run mode (parse+filter only, no file writes).
func WithDryRun(dry bool) Option {
	return func(s *Scraper) { s.dryRun = dry }
}

// New creates a Scraper instance.
func New(store *datastore.Store, log *zap.Logger, tokens []string, opts ...Option) (*Scraper, error) {
	cats, err := store.LoadCategories()
	if err != nil {
		return nil, fmt.Errorf("loading categories: %w", err)
	}

	sc := &Scraper{
		store:      store,
		log:        log,
		categories: cats,
		tokens:     NewTokenRotator(tokens),
		since:      "daily",
	}
	for _, o := range opts {
		o(sc)
	}
	return sc, nil
}

// Run executes the full scrape pipeline: fetch trending -> filter AI -> enrich -> save.
func (s *Scraper) Run(ctx context.Context) error {
	start := time.Now()

	// 1. Fetch Trending items
	items, err := s.fetchTrending(ctx)
	if err != nil {
		return fmt.Errorf("fetching trending: %w", err)
	}
	s.log.Info("Trending 爬取完成", zap.Int("total", len(items)), zap.String("since", s.since))

	// 2. Filter AI projects
	type filtered struct {
		item       TrendingItem
		categories []datastore.CategoryMatch
	}
	var aiItems []filtered
	for _, item := range items {
		matches := s.matchAIProject(item)
		if len(matches) > 0 {
			aiItems = append(aiItems, filtered{item: item, categories: matches})
		}
	}
	s.log.Info("AI 项目过滤完成", zap.Int("passed", len(aiItems)), zap.Int("total", len(items)))

	if s.dryRun {
		for _, f := range aiItems {
			cats := make([]string, 0, len(f.categories))
			for _, c := range f.categories {
				cats = append(cats, c.Slug)
			}
			s.log.Info("候选 AI 项目",
				zap.String("repo", f.item.FullName),
				zap.Strings("categories", cats),
				zap.Int("period_stars", f.item.PeriodStars),
			)
		}
		return nil
	}

	// 3. Enrich via GitHub API + save projects + append snapshots
	today := time.Now().UTC().Format("2006-01-02")
	var saved, enriched int

	for _, f := range aiItems {
		proj, err := s.enrichAndSave(ctx, f.item, f.categories, today)
		if err != nil {
			s.log.Warn("enrichAndSave 失败",
				zap.String("repo", f.item.FullName),
				zap.Error(err),
			)
			continue
		}
		saved++

		// Append snapshot
		snap := &datastore.Snapshot{
			ProjectID:  proj.ID,
			Date:       today,
			Stars:      proj.Stars,
			Forks:      proj.Forks,
			OpenIssues: proj.OpenIssues,
			DailyStars: f.item.intPtrPeriodStars(s.since),
		}
		if err := s.store.AppendSnapshot(snap); err != nil {
			s.log.Warn("追加快照失败", zap.String("repo", f.item.FullName), zap.Error(err))
			continue
		}
		enriched++
	}

	s.log.Info("采集完成",
		zap.Int("saved", saved),
		zap.Int("snapshots", enriched),
		zap.Duration("elapsed", time.Since(start)),
	)
	return nil
}

// intPtrPeriodStars returns a pointer to PeriodStars if daily, or nil.
func (item TrendingItem) intPtrPeriodStars(since string) *int {
	if since == "daily" && item.PeriodStars > 0 {
		return &item.PeriodStars
	}
	return nil
}

// fetchTrending uses Colly to scrape the GitHub Trending HTML page.
func (s *Scraper) fetchTrending(_ context.Context) ([]TrendingItem, error) {
	url := "https://github.com/trending"
	params := []string{}
	if s.since != "" {
		params = append(params, "since="+s.since)
	}
	if s.language != "" {
		url += "/" + s.language
	}
	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	var items []TrendingItem
	rank := 0

	c := colly.NewCollector(
		colly.AllowedDomains("github.com"),
	)
	c.SetRequestTimeout(30 * time.Second)

	// Limit request rate to avoid 429
	_ = c.Limit(&colly.LimitRule{
		DomainGlob:  "github.com",
		Delay:       1 * time.Second,
		RandomDelay: 500 * time.Millisecond,
	})

	c.OnHTML("article.Box-row", func(e *colly.HTMLElement) {
		rank++
		item := TrendingItem{Rank: rank}

		// full_name: h2 > a href = "/owner/repo"
		rawName := e.ChildAttr("h2 a", "href")
		rawName = strings.TrimPrefix(rawName, "/")
		item.FullName = strings.TrimSpace(rawName)

		item.Description = strings.TrimSpace(e.ChildText("p")) // description

		// language
		item.Language = strings.TrimSpace(e.ChildText("[itemprop=programmingLanguage]"))

		// stars and forks from the .Link--muted elements
		e.ForEach(".Link--muted.d-inline-block.mr-3", func(i int, el *colly.HTMLElement) {
			val := parseNumber(el.Text)
			switch i {
			case 0:
				item.Stars = val
			case 1:
				item.Forks = val
			}
		})

		// period stars (e.g., "523 stars today")
		periodText := strings.TrimSpace(e.ChildText(".float-sm-right"))
		if periodText == "" {
			periodText = strings.TrimSpace(e.ChildText(".d-inline-block.float-sm-right"))
		}
		item.PeriodStars = parseNumber(periodText)

		if item.FullName != "" {
			items = append(items, item)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		s.log.Error("Colly 请求失败",
			zap.String("url", r.Request.URL.String()),
			zap.Int("status", r.StatusCode),
			zap.Error(err),
		)
	})

	s.log.Info("开始爬取 Trending", zap.String("url", url))
	if err := c.Visit(url); err != nil {
		return nil, fmt.Errorf("visiting %s: %w", url, err)
	}

	return items, nil
}

// parseNumber extracts the first integer from text like "523 stars today" or "1,234".
var numRe = regexp.MustCompile(`[\d,]+`)

func parseNumber(text string) int {
	match := numRe.FindString(text)
	if match == "" {
		return 0
	}
	cleaned := strings.ReplaceAll(match, ",", "")
	n, _ := strconv.Atoi(cleaned)
	return n
}
