package scorer

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/config"
	"github.com/zbb88888/tishi/internal/datastore"
)

func testLogger() *zap.Logger {
	l, _ := zap.NewDevelopment()
	return l
}

func defaultScorerCfg() config.ScorerConfig {
	return config.ScorerConfig{
		DailyStars:    0.35,
		WeeklyStars:   0.25,
		ForksRate:     0.15,
		IssueActivity: 0.10,
		TopN:          100,
	}
}

func setupTestStore(t *testing.T) *datastore.Store {
	t.Helper()
	dir := t.TempDir()
	s := datastore.NewStore(dir, testLogger())

	d10 := 10
	d50 := 50
	w100 := 100
	w500 := 500

	projects := []*datastore.Project{
		{
			ID: "top__project", FullName: "top/project", Stars: 5000, Forks: 500, OpenIssues: 100,
			Trending:    &datastore.Trending{DailyStars: &d50, WeeklyStars: &w500},
			FirstSeenAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		},
		{
			ID: "mid__project", FullName: "mid/project", Stars: 2000, Forks: 200, OpenIssues: 50,
			Trending:    &datastore.Trending{DailyStars: &d10, WeeklyStars: &w100},
			FirstSeenAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		},
		{
			ID: "low__project", FullName: "low/project", Stars: 100, Forks: 10, OpenIssues: 5,
			FirstSeenAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		},
	}

	for _, p := range projects {
		if err := s.SaveProject(p); err != nil {
			t.Fatalf("SaveProject: %v", err)
		}
	}
	return s
}

func TestScorer_Run(t *testing.T) {
	store := setupTestStore(t)
	cfg := defaultScorerCfg()
	cfg.TopN = 10

	sc := New(store, testLogger(), cfg)
	if err := sc.Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verify ranking file created
	r, err := store.LoadLatestRanking()
	if err != nil {
		t.Fatalf("LoadLatestRanking: %v", err)
	}
	if r == nil {
		t.Fatal("no ranking created")
	}
	if r.Total != 3 {
		t.Errorf("Total = %d, want 3", r.Total)
	}

	// First item should be top/project (highest stars)
	if r.Items[0].ProjectID != "top__project" {
		t.Errorf("rank 1 = %q, want top__project", r.Items[0].ProjectID)
	}
	if r.Items[0].Score <= r.Items[1].Score {
		t.Errorf("rank 1 score (%f) should be > rank 2 (%f)", r.Items[0].Score, r.Items[1].Score)
	}

	// Verify project files updated with rank
	p, _ := store.LoadProject("top__project")
	if p.Rank == nil || *p.Rank != 1 {
		t.Errorf("top project rank = %v, want 1", p.Rank)
	}
}

func TestScorer_Run_Empty(t *testing.T) {
	dir := t.TempDir()
	store := datastore.NewStore(dir, testLogger())
	// Create empty projects dir
	os.MkdirAll(filepath.Join(dir, "projects"), 0o755)

	sc := New(store, testLogger(), defaultScorerCfg())
	if err := sc.Run(); err != nil {
		t.Fatalf("Run on empty: %v", err)
	}
}

func TestScorer_RankChange(t *testing.T) {
	dir := t.TempDir()
	store := datastore.NewStore(dir, testLogger())

	// Create a previous ranking
	prevRanking := &datastore.Ranking{
		Date:  "2026-02-12",
		Total: 2,
		Items: []datastore.RankingItem{
			{Rank: 1, ProjectID: "mid__project"},
			{Rank: 2, ProjectID: "top__project"},
		},
	}
	if err := store.SaveRanking(prevRanking); err != nil {
		t.Fatalf("SaveRanking: %v", err)
	}

	// Create projects
	d50 := 50
	w500 := 500
	d10 := 10
	w100 := 100
	now := time.Now().UTC()

	for _, p := range []*datastore.Project{
		{ID: "top__project", FullName: "top/project", Stars: 5000, Forks: 500, OpenIssues: 100,
			Trending:    &datastore.Trending{DailyStars: &d50, WeeklyStars: &w500},
			FirstSeenAt: now, UpdatedAt: now},
		{ID: "mid__project", FullName: "mid/project", Stars: 2000, Forks: 200, OpenIssues: 50,
			Trending:    &datastore.Trending{DailyStars: &d10, WeeklyStars: &w100},
			FirstSeenAt: now, UpdatedAt: now},
	} {
		if err := store.SaveProject(p); err != nil {
			t.Fatalf("SaveProject: %v", err)
		}
	}

	sc := New(store, testLogger(), defaultScorerCfg())
	if err := sc.Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}

	r, _ := store.LoadLatestRanking()
	// top__project should be rank 1 now (was rank 2), so rank_change = +1
	for _, item := range r.Items {
		if item.ProjectID == "top__project" && item.Rank == 1 {
			if item.RankChange == nil {
				t.Error("expected rank_change for top__project, got nil")
			} else if *item.RankChange != 1 {
				t.Errorf("rank_change = %d, want 1", *item.RankChange)
			}
		}
	}
}

func TestComputeScores_Normalization(t *testing.T) {
	dir := t.TempDir()
	store := datastore.NewStore(dir, testLogger())
	cfg := defaultScorerCfg()
	sc := New(store, testLogger(), cfg)

	d100 := 100
	w1000 := 1000
	projects := []*datastore.Project{
		{ID: "max", Stars: 10000, Trending: &datastore.Trending{DailyStars: &d100, WeeklyStars: &w1000}, Forks: 1000, OpenIssues: 500},
	}

	sc.computeScores(projects)

	// Single project with all max values should get max score
	// (0.35 + 0.25 + 0.15 + 0.10) * 100 = 85.0
	expected := (cfg.DailyStars + cfg.WeeklyStars + cfg.ForksRate + cfg.IssueActivity) * 100
	if math.Abs(projects[0].Score-expected) > 0.01 {
		t.Errorf("score = %f, want %f", projects[0].Score, expected)
	}
}

func TestComputeScores_ZeroValues(t *testing.T) {
	dir := t.TempDir()
	store := datastore.NewStore(dir, testLogger())
	sc := New(store, testLogger(), defaultScorerCfg())

	// Project without trending data
	projects := []*datastore.Project{
		{ID: "zero", Stars: 0, Forks: 0, OpenIssues: 0},
	}
	sc.computeScores(projects)

	if projects[0].Score != 0 {
		t.Errorf("score = %f, want 0", projects[0].Score)
	}
}

// Ensure ranking JSON is valid and can be round-tripped
func TestRanking_JSON_RoundTrip(t *testing.T) {
	daily := 42
	change := 3
	lang := "Python"
	r := &datastore.Ranking{
		Date:  "2026-02-13",
		Total: 1,
		Items: []datastore.RankingItem{
			{Rank: 1, ProjectID: "a__b", FullName: "a/b", Stars: 1000,
				DailyStars: &daily, RankChange: &change, Language: &lang, Score: 95.5},
		},
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var loaded datastore.Ranking
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if loaded.Items[0].Score != 95.5 {
		t.Errorf("score = %f", loaded.Items[0].Score)
	}
	if *loaded.Items[0].RankChange != 3 {
		t.Errorf("rank_change = %d", *loaded.Items[0].RankChange)
	}
}
