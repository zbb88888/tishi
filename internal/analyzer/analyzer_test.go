package analyzer

import (
	"math"
	"testing"
)

func TestMinMaxNormalize_Empty(t *testing.T) {
	result := minMaxNormalize(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestMinMaxNormalize_AllEqual(t *testing.T) {
	values := []float64{5, 5, 5}
	result := minMaxNormalize(values)

	for i, v := range result {
		if v != 50 {
			t.Errorf("index %d: expected 50, got %f", i, v)
		}
	}
}

func TestMinMaxNormalize_Range(t *testing.T) {
	values := []float64{0, 50, 100}
	result := minMaxNormalize(values)

	expected := []float64{0, 50, 100}
	for i, v := range result {
		if math.Abs(v-expected[i]) > 0.001 {
			t.Errorf("index %d: expected %f, got %f", i, expected[i], v)
		}
	}
}

func TestMinMaxNormalize_NegativeValues(t *testing.T) {
	values := []float64{-10, 0, 10}
	result := minMaxNormalize(values)

	expected := []float64{0, 50, 100}
	for i, v := range result {
		if math.Abs(v-expected[i]) > 0.001 {
			t.Errorf("index %d: expected %f, got %f", i, expected[i], v)
		}
	}
}

func TestMinMaxNormalize_SingleValue(t *testing.T) {
	values := []float64{42}
	result := minMaxNormalize(values)

	if result[0] != 50 {
		t.Errorf("expected 50, got %f", result[0])
	}
}

func TestComputeScores_Basic(t *testing.T) {
	stub := (&configStub{}).get()
	a := &Analyzer{
		cfg: stub,
	}

	projects := []*projectData{
		{ID: 1, DailyStarGain: 100, WeeklyStarGain: 500, Stars: 1000, Forks: 200, OpenIssues: 50, DaysSinceCreated: 100, DaysSincePush: 1},
		{ID: 2, DailyStarGain: 10, WeeklyStarGain: 50, Stars: 500, Forks: 50, OpenIssues: 10, DaysSinceCreated: 200, DaysSincePush: 30},
	}

	a.computeScores(projects)

	// Project 1 应该得分更高（日增/周增 star 都更大）
	if projects[0].Score <= projects[1].Score {
		t.Errorf("project 1 score (%f) should be > project 2 score (%f)", projects[0].Score, projects[1].Score)
	}

	// 分数应在 [0, 100]
	for _, p := range projects {
		if p.Score < 0 || p.Score > 100 {
			t.Errorf("project %d: score %f out of [0,100] range", p.ID, p.Score)
		}
	}
}

func TestGenerateRankings(t *testing.T) {
	a := &Analyzer{}

	projects := []*projectData{
		{ID: 1, Score: 30},
		{ID: 2, Score: 80},
		{ID: 3, Score: 50},
	}

	a.generateRankings(projects)

	if projects[0].Rank != 1 || projects[0].ID != 2 {
		t.Errorf("rank 1 should be project 2, got project %d", projects[0].ID)
	}
	if projects[1].Rank != 2 || projects[1].ID != 3 {
		t.Errorf("rank 2 should be project 3, got project %d", projects[1].ID)
	}
	if projects[2].Rank != 3 || projects[2].ID != 1 {
		t.Errorf("rank 3 should be project 1, got project %d", projects[2].ID)
	}
}
