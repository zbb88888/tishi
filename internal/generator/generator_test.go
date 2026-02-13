package generator

import (
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/datastore"
)

func testLogger() *zap.Logger {
	l, _ := zap.NewDevelopment()
	return l
}

func TestGenerateWeekly(t *testing.T) {
	dir := t.TempDir()
	store := datastore.NewStore(dir, testLogger())

	// Create a ranking
	daily := 50
	weekly := 500
	lang := "Python"
	cat := "llm"
	summary := "A great LLM tool"

	ranking := &datastore.Ranking{
		Date:  time.Now().UTC().Format("2006-01-02"),
		Total: 2,
		Items: []datastore.RankingItem{
			{Rank: 1, ProjectID: "a__b", FullName: "a/b", Stars: 5000,
				DailyStars: &daily, WeeklyStars: &weekly, Language: &lang,
				Category: &cat, Summary: &summary, Score: 95.5},
			{Rank: 2, ProjectID: "c__d", FullName: "c/d", Stars: 2000,
				Score: 80.0},
		},
	}

	if err := store.SaveRanking(ranking); err != nil {
		t.Fatalf("SaveRanking: %v", err)
	}

	g := New(store, testLogger())
	if err := g.Run("weekly", RunOptions{DryRun: true}); err != nil {
		t.Fatalf("Run weekly: %v", err)
	}
}

func TestGenerateWeekly_NoRanking(t *testing.T) {
	dir := t.TempDir()
	store := datastore.NewStore(dir, testLogger())

	g := New(store, testLogger())
	err := g.Run("weekly", RunOptions{})
	if err == nil {
		t.Fatal("expected error when no ranking exists")
	}
}

func TestGenerateSpotlight_NoID(t *testing.T) {
	dir := t.TempDir()
	store := datastore.NewStore(dir, testLogger())

	g := New(store, testLogger())
	err := g.Run("spotlight", RunOptions{})
	if err == nil {
		t.Fatal("expected error when no --id for spotlight")
	}
}

func TestGenerateSpotlight_NeedsPublished(t *testing.T) {
	dir := t.TempDir()
	store := datastore.NewStore(dir, testLogger())

	now := time.Now().UTC()
	p := &datastore.Project{
		ID: "owner__repo", FullName: "owner/repo", Stars: 1000,
		Analysis:    &datastore.Analysis{Status: "draft", GeneratedAt: now},
		FirstSeenAt: now, UpdatedAt: now,
	}
	if err := store.SaveProject(p); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	g := New(store, testLogger())
	err := g.Run("spotlight", RunOptions{ProjectID: "owner__repo"})
	if err == nil {
		t.Fatal("expected error for draft analysis")
	}
}

func TestGenerateSpotlight_Success(t *testing.T) {
	dir := t.TempDir()
	store := datastore.NewStore(dir, testLogger())

	now := time.Now().UTC()
	p := &datastore.Project{
		ID: "owner__repo", FullName: "owner/repo", Stars: 1000,
		Analysis: &datastore.Analysis{
			Status:      "published",
			Model:       "deepseek-chat",
			Summary:     "A great tool",
			Positioning: "Solving X for Y",
			Features:    []datastore.Feature{{Name: "f1", Desc: "fast"}},
			Advantages:  "Very fast",
			TechStack:   "Go + gRPC",
			UseCases:    "Backend dev",
			Comparison:  []datastore.ComparisonEntry{{Project: "other/tool", Diff: "Faster"}},
			Ecosystem:   "Cloud native",
			GeneratedAt: now,
		},
		FirstSeenAt: now, UpdatedAt: now,
	}
	if err := store.SaveProject(p); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	g := New(store, testLogger())
	// DryRun to avoid file write
	if err := g.Run("spotlight", RunOptions{ProjectID: "owner__repo", DryRun: true}); err != nil {
		t.Fatalf("Run spotlight: %v", err)
	}
}

func TestGenerateSpotlight_SavePost(t *testing.T) {
	dir := t.TempDir()
	store := datastore.NewStore(dir, testLogger())

	now := time.Now().UTC()
	p := &datastore.Project{
		ID: "owner__repo", FullName: "owner/repo", Stars: 1000,
		Analysis: &datastore.Analysis{
			Status: "published", Model: "deepseek-chat", Summary: "Tool",
			Positioning: "X", Advantages: "Y", TechStack: "Z",
			UseCases: "W", Ecosystem: "E", GeneratedAt: now,
		},
		FirstSeenAt: now, UpdatedAt: now,
	}
	store.SaveProject(p)

	g := New(store, testLogger())
	if err := g.Run("spotlight", RunOptions{ProjectID: "owner__repo"}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	posts, err := store.ListPosts()
	if err != nil {
		t.Fatalf("ListPosts: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("got %d posts, want 1", len(posts))
	}
	if posts[0].PostType != "spotlight" {
		t.Errorf("PostType = %q", posts[0].PostType)
	}
}

func TestRun_InvalidType(t *testing.T) {
	g := New(datastore.NewStore(t.TempDir(), testLogger()), testLogger())
	if err := g.Run("invalid", RunOptions{}); err == nil {
		t.Fatal("expected error for invalid post type")
	}
}
