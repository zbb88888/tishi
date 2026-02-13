package scraper

import (
	"strings"

	"github.com/zbb88888/tishi/internal/datastore"
)

// matchAIProject checks if a TrendingItem matches any AI category keywords.
// Returns matched categories with confidence scores.
func (s *Scraper) matchAIProject(item TrendingItem) []datastore.CategoryMatch {
	var matches []datastore.CategoryMatch
	seen := make(map[string]bool)

	descLower := strings.ToLower(item.Description)
	nameLower := strings.ToLower(item.FullName)

	for _, cat := range s.categories {
		if cat.Slug == "other" {
			continue // "other" only used as fallback
		}

		var bestConfidence float64

		// Check description keywords (confidence 0.8)
		for _, kw := range cat.Keywords.Description {
			if strings.Contains(descLower, strings.ToLower(kw)) {
				if bestConfidence < 0.8 {
					bestConfidence = 0.8
				}
				break
			}
		}

		// Check repo name keywords (confidence 0.6)
		for _, kw := range cat.Keywords.Topics {
			kwLower := strings.ToLower(kw)
			if strings.Contains(nameLower, kwLower) {
				if bestConfidence < 0.6 {
					bestConfidence = 0.6
				}
				break
			}
		}

		if bestConfidence > 0 && !seen[cat.Slug] {
			seen[cat.Slug] = true
			matches = append(matches, datastore.CategoryMatch{
				Slug:       cat.Slug,
				Confidence: bestConfidence,
			})
		}
	}

	return matches
}

// matchAIProjectWithTopics checks against both Trending data and GitHub API topics.
// Called after enrichment when topics[] is available, providing higher confidence.
func matchAIProjectWithTopics(topics []string, categories []datastore.Category) []datastore.CategoryMatch {
	var matches []datastore.CategoryMatch
	seen := make(map[string]bool)

	topicSet := make(map[string]bool, len(topics))
	for _, t := range topics {
		topicSet[strings.ToLower(t)] = true
	}

	for _, cat := range categories {
		if cat.Slug == "other" {
			continue
		}

		// Topics exact match → confidence 1.0
		for _, kw := range cat.Keywords.Topics {
			if topicSet[strings.ToLower(kw)] {
				if !seen[cat.Slug] {
					seen[cat.Slug] = true
					matches = append(matches, datastore.CategoryMatch{
						Slug:       cat.Slug,
						Confidence: 1.0,
					})
				}
				break
			}
		}
	}

	return matches
}

// mergeCategories combines pre-filter matches with topic-based matches,
// keeping the highest confidence for each category.
func mergeCategories(a, b []datastore.CategoryMatch) []datastore.CategoryMatch {
	best := make(map[string]float64)
	for _, m := range a {
		if m.Confidence > best[m.Slug] {
			best[m.Slug] = m.Confidence
		}
	}
	for _, m := range b {
		if m.Confidence > best[m.Slug] {
			best[m.Slug] = m.Confidence
		}
	}

	result := make([]datastore.CategoryMatch, 0, len(best))
	for slug, conf := range best {
		result = append(result, datastore.CategoryMatch{
			Slug:       slug,
			Confidence: conf,
		})
	}
	return result
}

// primaryCategory returns the slug of the highest-confidence category.
func primaryCategory(matches []datastore.CategoryMatch) *string {
	if len(matches) == 0 {
		return nil
	}
	best := matches[0]
	for _, m := range matches[1:] {
		if m.Confidence > best.Confidence {
			best = m
		}
	}
	return &best.Slug
}
