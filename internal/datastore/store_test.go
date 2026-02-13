package datastore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
)

func testLogger() *zap.Logger {
	l, _ := zap.NewDevelopment()
	return l
}

func TestProjectFilename(t *testing.T) {
	tests := []struct {
		fullName string
		want     string
	}{
		{"owner/repo", "owner__repo.json"},
		{"open-ai/gpt-4", "open-ai__gpt-4.json"},
	}
	for _, tt := range tests {
		got := ProjectFilename(tt.fullName)
		if got != tt.want {
			t.Errorf("ProjectFilename(%q) = %q, want %q", tt.fullName, got, tt.want)
		}
	}
}

func TestProjectIDFromFullName(t *testing.T) {
	got := ProjectIDFromFullName("owner/repo")
	if got != "owner__repo" {
		t.Errorf("got %q, want owner__repo", got)
	}
}

func TestSaveAndLoadProject(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir, testLogger())

	desc := "test project"
	lang := "Go"
	now := time.Now().UTC().Truncate(time.Second)

	p := &Project{
		ID:          "owner__repo",
		FullName:    "owner/repo",
		Description: &desc,
		Language:    &lang,
		Stars:       100,
		Forks:       20,
		Topics:      []string{"ai", "ml"},
		FirstSeenAt: now,
		UpdatedAt:   now,
	}

	if err := s.SaveProject(p); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	// Verify file exists
	path := filepath.Join(dir, "projects", "owner__repo.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("project file not found: %v", err)
	}

	loaded, err := s.LoadProject("owner__repo")
	if err != nil {
		t.Fatalf("LoadProject: %v", err)
	}

	if loaded.ID != p.ID {
		t.Errorf("ID = %q, want %q", loaded.ID, p.ID)
	}
	if loaded.FullName != p.FullName {
		t.Errorf("FullName = %q, want %q", loaded.FullName, p.FullName)
	}
	if *loaded.Description != *p.Description {
		t.Errorf("Description = %q, want %q", *loaded.Description, *p.Description)
	}
	if loaded.Stars != p.Stars {
		t.Errorf("Stars = %d, want %d", loaded.Stars, p.Stars)
	}
	if len(loaded.Topics) != 2 || loaded.Topics[0] != "ai" {
		t.Errorf("Topics = %v, want [ai ml]", loaded.Topics)
	}

	// No .tmp file left
	matches, _ := filepath.Glob(filepath.Join(dir, "projects", "*.tmp"))
	if len(matches) > 0 {
		t.Errorf("leftover tmp files: %v", matches)
	}
}

func TestLoadProject_NotFound(t *testing.T) {
	s := NewStore(t.TempDir(), testLogger())
	_, err := s.LoadProject("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent project")
	}
}

func TestListProjects(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir, testLogger())

	projects, err := s.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects empty: %v", err)
	}
	if projects != nil {
		t.Errorf("expected nil for empty dir, got %d", len(projects))
	}

	now := time.Now().UTC()
	for _, name := range []string{"a/b", "c/d"} {
		p := &Project{
			ID: ProjectIDFromFullName(name), FullName: name,
			Stars: 50, FirstSeenAt: now, UpdatedAt: now,
		}
		if err := s.SaveProject(p); err != nil {
			t.Fatalf("SaveProject(%s): %v", name, err)
		}
	}

	projects, err = s.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("got %d projects, want 2", len(projects))
	}
}

func TestAppendAndLoadSnapshots(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir, testLogger())
	date := "2026-02-13"

	snap1 := &Snapshot{ProjectID: "owner__repo", Date: date, Stars: 100, Forks: 20}
	snap2 := &Snapshot{ProjectID: "other__proj", Date: date, Stars: 200, Forks: 40}

	if err := s.AppendSnapshot(snap1); err != nil {
		t.Fatalf("AppendSnapshot 1: %v", err)
	}
	if err := s.AppendSnapshot(snap2); err != nil {
		t.Fatalf("AppendSnapshot 2: %v", err)
	}

	snaps, err := s.LoadSnapshots(date)
	if err != nil {
		t.Fatalf("LoadSnapshots: %v", err)
	}
	if len(snaps) != 2 {
		t.Fatalf("got %d snapshots, want 2", len(snaps))
	}
	if snaps[0].ProjectID != "owner__repo" {
		t.Errorf("snap[0].ProjectID = %q", snaps[0].ProjectID)
	}
	if snaps[1].Stars != 200 {
		t.Errorf("snap[1].Stars = %d, want 200", snaps[1].Stars)
	}
}

func TestLoadSnapshots_NoFile(t *testing.T) {
	s := NewStore(t.TempDir(), testLogger())
	snaps, err := s.LoadSnapshots("2026-01-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snaps != nil {
		t.Errorf("expected nil, got %d", len(snaps))
	}
}

func TestSaveAndLoadRanking(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir, testLogger())

	daily := 50
	r := &Ranking{
		Date:  "2026-02-13",
		Total: 2,
		Items: []RankingItem{
			{Rank: 1, ProjectID: "a__b", FullName: "a/b", Stars: 1000, Score: 95.5, DailyStars: &daily},
			{Rank: 2, ProjectID: "c__d", FullName: "c/d", Stars: 500, Score: 80.2},
		},
	}

	if err := s.SaveRanking(r); err != nil {
		t.Fatalf("SaveRanking: %v", err)
	}

	loaded, err := s.LoadRanking("2026-02-13")
	if err != nil {
		t.Fatalf("LoadRanking: %v", err)
	}
	if loaded.Date != r.Date {
		t.Errorf("Date = %q, want %q", loaded.Date, r.Date)
	}
	if loaded.Total != 2 {
		t.Errorf("Total = %d, want 2", loaded.Total)
	}
	if loaded.Items[0].Score != 95.5 {
		t.Errorf("Items[0].Score = %f, want 95.5", loaded.Items[0].Score)
	}
	if loaded.Items[0].DailyStars == nil || *loaded.Items[0].DailyStars != 50 {
		t.Errorf("DailyStars mismatch")
	}
}

func TestLoadLatestRanking(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir, testLogger())

	r, err := s.LoadLatestRanking()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r != nil {
		t.Errorf("expected nil for empty dir")
	}

	for _, date := range []string{"2026-02-10", "2026-02-13", "2026-02-11"} {
		rk := &Ranking{Date: date, Total: 1, Items: []RankingItem{{Rank: 1, ProjectID: "a__b"}}}
		if err := s.SaveRanking(rk); err != nil {
			t.Fatalf("SaveRanking(%s): %v", date, err)
		}
	}

	latest, err := s.LoadLatestRanking()
	if err != nil {
		t.Fatalf("LoadLatestRanking: %v", err)
	}
	if latest.Date != "2026-02-13" {
		t.Errorf("latest date = %q, want 2026-02-13", latest.Date)
	}
}

func TestSaveAndListPosts(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir, testLogger())

	now := time.Now().UTC().Truncate(time.Second)
	post := &Post{
		Slug: "test-post", Title: "Test Post",
		Content: "# Hello\n\nWorld", PostType: "weekly", CreatedAt: now,
	}

	if err := s.SavePost(post); err != nil {
		t.Fatalf("SavePost: %v", err)
	}

	posts, err := s.ListPosts()
	if err != nil {
		t.Fatalf("ListPosts: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("got %d posts, want 1", len(posts))
	}
	if posts[0].Slug != "test-post" {
		t.Errorf("Slug = %q", posts[0].Slug)
	}
}

func TestLoadCategories(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir, testLogger())

	cats := []Category{
		{Slug: "llm", Name: "LLM", Keywords: CategoryKeywords{Topics: []string{"llm"}}},
		{Slug: "cv", Name: "CV", Keywords: CategoryKeywords{Topics: []string{"vision"}}},
	}
	data, _ := json.MarshalIndent(cats, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "categories.json"), data, 0o644); err != nil {
		t.Fatalf("write categories.json: %v", err)
	}

	loaded, err := s.LoadCategories()
	if err != nil {
		t.Fatalf("LoadCategories: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("got %d, want 2", len(loaded))
	}
	if loaded[0].Slug != "llm" {
		t.Errorf("cats[0].Slug = %q", loaded[0].Slug)
	}
}

func TestDataDir(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir, testLogger())
	if s.DataDir() != dir {
		t.Errorf("DataDir() = %q, want %q", s.DataDir(), dir)
	}
}
