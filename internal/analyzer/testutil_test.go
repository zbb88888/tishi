package analyzer

import "github.com/zbb88888/tishi/internal/config"

// configStub provides a minimal config for testing.
type configStub struct{}

func (c *configStub) get() *config.Config {
	return &config.Config{
		Analyzer: config.AnalyzerConfig{
			Weights: config.WeightsConfig{
				DailyStar:     0.30,
				WeeklyStar:    0.25,
				ForkRatio:     0.15,
				IssueActivity: 0.15,
				Recency:       0.15,
			},
		},
	}
}
