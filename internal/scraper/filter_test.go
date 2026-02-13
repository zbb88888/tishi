package scraper

import (
	"testing"

	"github.com/zbb88888/tishi/internal/datastore"
)

func TestMatchAIProjectWithTopics(t *testing.T) {
	categories := []datastore.Category{
		{
			Slug: "llm",
			Keywords: datastore.CategoryKeywords{
				Topics:      []string{"llm", "large-language-model"},
				Description: []string{"language model"},
			},
		},
		{
			Slug: "cv",
			Keywords: datastore.CategoryKeywords{
				Topics:      []string{"computer-vision", "image-recognition"},
				Description: []string{"object detection"},
			},
		},
		{
			Slug: "other",
			Keywords: datastore.CategoryKeywords{
				Topics: []string{"ai"},
			},
		},
	}

	tests := []struct {
		name       string
		topics     []string
		wantSlugs  []string
		wantCount  int
	}{
		{
			name:      "exact topic match",
			topics:    []string{"llm", "python"},
			wantSlugs: []string{"llm"},
			wantCount: 1,
		},
		{
			name:      "multiple matches",
			topics:    []string{"llm", "computer-vision"},
			wantSlugs: []string{"llm", "cv"},
			wantCount: 2,
		},
		{
			name:      "no match",
			topics:    []string{"web", "frontend"},
			wantCount: 0,
		},
		{
			name:      "other category skipped",
			topics:    []string{"ai"},
			wantCount: 0,
		},
		{
			name:      "case insensitive",
			topics:    []string{"LLM"},
			wantSlugs: []string{"llm"},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := matchAIProjectWithTopics(tt.topics, categories)
			if len(matches) != tt.wantCount {
				t.Errorf("got %d matches, want %d: %+v", len(matches), tt.wantCount, matches)
			}
			for _, want := range tt.wantSlugs {
				found := false
				for _, m := range matches {
					if m.Slug == want {
						found = true
						if m.Confidence != 1.0 {
							t.Errorf("topic match confidence for %q = %f, want 1.0", want, m.Confidence)
						}
					}
				}
				if !found {
					t.Errorf("expected slug %q in matches", want)
				}
			}
		})
	}
}

func TestMergeCategories(t *testing.T) {
	a := []datastore.CategoryMatch{
		{Slug: "llm", Confidence: 0.8},
		{Slug: "cv", Confidence: 0.6},
	}
	b := []datastore.CategoryMatch{
		{Slug: "llm", Confidence: 1.0},
		{Slug: "nlp", Confidence: 0.9},
	}

	merged := mergeCategories(a, b)

	// Should have 3 unique slugs
	if len(merged) != 3 {
		t.Fatalf("got %d merged, want 3", len(merged))
	}

	bySlug := make(map[string]float64)
	for _, m := range merged {
		bySlug[m.Slug] = m.Confidence
	}

	// llm should take higher confidence (1.0 from b)
	if bySlug["llm"] != 1.0 {
		t.Errorf("llm confidence = %f, want 1.0", bySlug["llm"])
	}
	if bySlug["cv"] != 0.6 {
		t.Errorf("cv confidence = %f, want 0.6", bySlug["cv"])
	}
	if bySlug["nlp"] != 0.9 {
		t.Errorf("nlp confidence = %f, want 0.9", bySlug["nlp"])
	}
}

func TestPrimaryCategory(t *testing.T) {
	// Empty
	if p := primaryCategory(nil); p != nil {
		t.Errorf("expected nil for empty, got %v", p)
	}

	matches := []datastore.CategoryMatch{
		{Slug: "cv", Confidence: 0.6},
		{Slug: "llm", Confidence: 1.0},
		{Slug: "nlp", Confidence: 0.8},
	}
	p := primaryCategory(matches)
	if p == nil || *p != "llm" {
		t.Errorf("primary = %v, want llm", p)
	}
}
