# tishi - AI Trends Top 100 Tracker

.PHONY: build run dev test lint clean migrate-up migrate-down docker-build docker-up docker-down tidy generate

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

## run: Build and run the server
run: build
	./$(BINARY) server

## dev: Run with live reload (requires air)
dev:
	@command -v air >/dev/null 2>&1 || { echo "air not installed: go install github.com/air-verse/air@latest"; exit 1; }
	air

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

## generate: Run code generators (sqlc)
generate:
	@command -v sqlc >/dev/null 2>&1 || { echo "sqlc not installed: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"; exit 1; }
	sqlc generate

## migrate-up: Run database migrations up
migrate-up: build
	./$(BINARY) migrate up

## migrate-down: Run database migrations down
migrate-down: build
	./$(BINARY) migrate down

## collect: Trigger a manual data collection
collect: build
	./$(BINARY) collect

## analyze: Trigger a manual analysis
analyze: build
	./$(BINARY) analyze

## docker-build: Build Docker image
docker-build:
	docker build -t tishi:$(VERSION) .

## docker-up: Start all services with Docker Compose
docker-up:
	docker compose up -d

## docker-down: Stop all Docker Compose services
docker-down:
	docker compose down

## clean: Remove build artifacts
clean:
	rm -rf bin/ coverage.out

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'
