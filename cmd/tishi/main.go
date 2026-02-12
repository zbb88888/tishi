// Package main is the entry point for the tishi CLI application.
// tishi tracks the top 100 AI-related trending projects on GitHub.
package main

import (
	"os"

	"github.com/zbb88888/tishi/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
