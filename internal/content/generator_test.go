package content

import (
	"strings"
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	g := &Generator{}

	data := WeeklyData{
		WeekNumber:  42,
		Year:        2026,
		DateRange:   "2026-01-01 ~ 2026-01-07",
		PublishedAt: "2026-01-07T06:00:00Z",
		TopGainers: []TopGainer{
			{Rank: 1, FullName: "org/repo-a", WeeklyGain: 500, Stars: 10000},
			{Rank: 2, FullName: "org/repo-b", WeeklyGain: 300, Stars: 8000},
		},
	}

	result, err := g.renderTemplate(weeklyTemplate, data)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	checks := []string{
		"#42",
		"2026-01-01 ~ 2026-01-07",
		"org/repo-a",
		"+500",
		"10000",
		"org/repo-b",
		"+300",
		"tishi",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("expected %q in rendered output", check)
		}
	}
}

func TestRenderTemplate_EmptyGainers(t *testing.T) {
	g := &Generator{}

	data := WeeklyData{
		WeekNumber:  1,
		Year:        2026,
		DateRange:   "2026-01-01 ~ 2026-01-07",
		PublishedAt: "2026-01-07T00:00:00Z",
		TopGainers:  nil,
	}

	result, err := g.renderTemplate(weeklyTemplate, data)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(result, "#1") {
		t.Error("expected week number in output")
	}
}
