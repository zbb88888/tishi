# tishi - AI Trends Top 100 Tracker

.PHONY: build test lint clean tidy scrape score analyze review push pipeline docker-build help

# Build vars
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS  = -s -w \
	-X github.com/zbb88888/tishi/internal/cmd.Version=$(VERSION) \
	-X github.com/zbb88888/tishi/internal/cmd.GitCommit=$(COMMIT) \
	-X github.com/zbb88888/tishi/internal/cmd.BuildDate=$(DATE)

BINARY = bin/tishi

## build: Compile the binary
build:
	@echo "==> Building..."
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/tishi

## ──────────────── v1.0 Pipeline ────────────────

## scrape: 抓取 GitHub Trending 数据
scrape: build
	./$(BINARY) scrape

## score: 计算项目评分与排行
score: build
	./$(BINARY) score

## analyze: LLM 中文分析新入榜项目
analyze: build
	./$(BINARY) analyze

## review: 列出待审核分析 (--approve/--reject)
review: build
	./$(BINARY) review

## push: 提交并推送 data/ 到 Git
push: build
	./$(BINARY) push

## pipeline: 完整日常流水线 (scrape → score → analyze → push)
pipeline: build
	./$(BINARY) scrape
	./$(BINARY) score
	./$(BINARY) analyze
	./$(BINARY) push

## ──────────────── Dev Tools ────────────────

## test: Run unit tests
test:
	go test -race -coverprofile=coverage.out ./...

## lint: Run golangci-lint
lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed"; exit 1; }
	golangci-lint run ./...

## tidy: Tidy Go modules
tidy:
	go mod tidy

## docker-build: Build Docker image
docker-build:
	docker build -t tishi:$(VERSION) .

## clean: Remove build artifacts
clean:
	rm -rf bin/ coverage.out

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'
