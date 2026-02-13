// Package datastore provides JSON file-based data storage for tishi.
// All data resides in the data/ directory and syncs via Git between stages.
package datastore

import "time"

// Project represents an AI open-source project (data/projects/{owner}__{repo}.json).
type Project struct {
	ID          string   `json:"id"`                    // owner__repo
	FullName    string   `json:"full_name"`             // owner/repo
	Description *string  `json:"description,omitempty"` // GitHub 原始描述
	Language    *string  `json:"language,omitempty"`
	License     *string  `json:"license,omitempty"` // SPDX ID
	Topics      []string `json:"topics,omitempty"`
	Homepage    *string  `json:"homepage,omitempty"`

	Stars      int  `json:"stars"`
	Forks      int  `json:"forks"`
	OpenIssues int  `json:"open_issues"`
	Watchers   int  `json:"watchers"`
	IsArchived bool `json:"is_archived"`

	PushedAt    *time.Time `json:"pushed_at,omitempty"`
	CreatedAtGH *time.Time `json:"created_at_gh,omitempty"`

	Score    float64 `json:"score"`
	Rank     *int    `json:"rank,omitempty"`
	Category *string `json:"category,omitempty"` // primary category slug

	Trending   *Trending       `json:"trending,omitempty"`
	Analysis   *Analysis       `json:"analysis,omitempty"`
	Categories []CategoryMatch `json:"categories,omitempty"`

	FirstSeenAt time.Time `json:"first_seen_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Trending holds GitHub Trending page data for a project.
type Trending struct {
	DailyStars       *int    `json:"daily_stars,omitempty"`
	WeeklyStars      *int    `json:"weekly_stars,omitempty"`
	RankDaily        *int    `json:"rank_daily,omitempty"`
	LastSeenTrending *string `json:"last_seen_trending,omitempty"` // YYYY-MM-DD
}

// Analysis holds LLM-generated Chinese project analysis.
type Analysis struct {
	Status      string            `json:"status"`                // draft | published | rejected
	Model       string            `json:"model"`                 // e.g. deepseek-chat
	Summary     string            `json:"summary"`               // ≤50 chars
	Positioning string            `json:"positioning,omitempty"` // ~200 chars
	Features    []Feature         `json:"features,omitempty"`
	Advantages  string            `json:"advantages,omitempty"`
	TechStack   string            `json:"tech_stack,omitempty"`
	UseCases    string            `json:"use_cases,omitempty"`
	Comparison  []ComparisonEntry `json:"comparison,omitempty"`
	Ecosystem   string            `json:"ecosystem,omitempty"`
	GeneratedAt time.Time         `json:"generated_at"`
	ReviewedAt  *time.Time        `json:"reviewed_at,omitempty"`
	TokenUsage  *int              `json:"token_usage,omitempty"`
}

// Feature describes a single project feature.
type Feature struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

// ComparisonEntry compares with a competing project.
type ComparisonEntry struct {
	Project string `json:"project"`
	Diff    string `json:"diff"`
}

// CategoryMatch records a matched AI category with confidence score.
type CategoryMatch struct {
	Slug       string  `json:"slug"`
	Confidence float64 `json:"confidence"`
}

// Snapshot is a single-line entry in data/snapshots/{date}.jsonl.
type Snapshot struct {
	ProjectID  string   `json:"project_id"` // owner__repo
	Date       string   `json:"date"`       // YYYY-MM-DD
	Stars      int      `json:"stars"`
	Forks      int      `json:"forks"`
	OpenIssues int      `json:"open_issues"`
	Watchers   int      `json:"watchers,omitempty"`
	Score      *float64 `json:"score,omitempty"`
	Rank       *int     `json:"rank,omitempty"`
	DailyStars *int     `json:"daily_stars,omitempty"`
}

// Ranking is the daily ranking file (data/rankings/{date}.json).
type Ranking struct {
	Date  string        `json:"date"` // YYYY-MM-DD
	Total int           `json:"total"`
	Items []RankingItem `json:"items"`
}

// RankingItem is a single entry in the ranking.
type RankingItem struct {
	Rank        int     `json:"rank"`
	ProjectID   string  `json:"project_id"`
	FullName    string  `json:"full_name"`
	Summary     *string `json:"summary,omitempty"`
	Language    *string `json:"language,omitempty"`
	Category    *string `json:"category,omitempty"`
	Stars       int     `json:"stars"`
	DailyStars  *int    `json:"daily_stars,omitempty"`
	WeeklyStars *int    `json:"weekly_stars,omitempty"`
	Score       float64 `json:"score"`
	RankChange  *int    `json:"rank_change,omitempty"` // positive=up, negative=down, nil=new
}

// Post is a blog post (data/posts/{slug}.json).
type Post struct {
	Slug          string     `json:"slug"`
	Title         string     `json:"title"`
	Content       string     `json:"content"` // Markdown
	PostType      string     `json:"post_type"`
	CoverImageURL *string    `json:"cover_image_url,omitempty"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

// Category represents an AI sub-category from data/categories.json.
type Category struct {
	Slug        string           `json:"slug"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	SortOrder   int              `json:"sort_order"`
	Keywords    CategoryKeywords `json:"keywords"`
	ProjectIDs  []string         `json:"project_ids"`
}

// CategoryKeywords holds keyword lists for matching.
type CategoryKeywords struct {
	Topics      []string `json:"topics"`
	Description []string `json:"description"`
}
