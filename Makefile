.PHONY: all build build-cli build-server test test-unit test-integration clean sqlc templ fmt lint run serve migrate reset install help

# Variables
CLI_BINARY := treni
SERVER_BINARY := trenid
BUILD_DIR := .
GO_FILES := $(shell find . -name '*.go' -not -path './sqlc/generated/*')
TEMPL_FILES := $(shell find . -name '*.templ')
SQL_FILES := $(shell find . -path '*/queries/*.sql' 2>/dev/null)

# Default target
all: build

# === Code Generation ===

# Generate sqlc code (depends on SQL files)
sqlc: $(SQL_FILES)
	sqlc generate

# Generate templ templates (depends on templ files)
templ: $(TEMPL_FILES)
	templ generate

# Generate all code
generate: sqlc templ

# === Build ===

# Build both binaries (depends on generated code)
build: generate build-cli build-server

# Build CLI only
build-cli:
	go build -o $(CLI_BINARY) ./cmd/treni

# Build server only
build-server:
	go build -o $(SERVER_BINARY) ./cmd/trenid

# Build with version info
build-release: generate
	go build -ldflags="-s -w" -o $(CLI_BINARY) ./cmd/treni
	go build -ldflags="-s -w" -o $(SERVER_BINARY) ./cmd/trenid

# Install to GOPATH/bin
install: generate
	go install ./cmd/treni
	go install ./cmd/trenid

# === Testing ===

# Run all tests (depends on generated code)
test: generate
	go test -v ./...

# Run unit tests only (skip integration tests)
test-unit: generate
	go test -v -short ./...

# Run integration tests
test-integration: generate
	go test -v -run Integration ./...

# Run tests with coverage
test-coverage: generate
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# === Code Quality ===

# Format code
fmt:
	go fmt ./...
	gofmt -s -w .

# Lint code (depends on generated code)
lint: generate
	golangci-lint run ./...

# Tidy dependencies
tidy:
	go mod tidy

# Check everything (format, lint, test)
check: fmt lint test

# === Database ===

# Run migrations
migrate: build-server
	./$(SERVER_BINARY) migrate

# Reset database (drop all tables)
reset: build-server
	./$(SERVER_BINARY) migrate 0

# === Run ===

# Run the CLI
run: build-cli
	./$(CLI_BINARY)

# Start web server (for development)
serve: build-server
	./$(SERVER_BINARY)

# Development: watch and rebuild
dev:
	@echo "Watching for changes..."
	@while true; do \
		$(MAKE) build; \
		fswatch -1 $(GO_FILES) $(TEMPL_FILES) > /dev/null 2>&1 || inotifywait -q -e modify $(GO_FILES) $(TEMPL_FILES) 2>/dev/null || sleep 2; \
	done

# === Cleanup ===

# Clean build artifacts
clean:
	rm -f $(CLI_BINARY) $(SERVER_BINARY)
	rm -f coverage.out coverage.html
	rm -f *.db *.db-*

# Clean generated code too
clean-all: clean
	rm -f internal/storage/generated/*.go
	rm -f web/templates/*_templ.go

# === Help ===

help:
	@echo "treni - Train tracking application"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build targets:"
	@echo "  all             Build everything (default)"
	@echo "  build           Generate code + build both binaries"
	@echo "  build-cli       Build CLI only"
	@echo "  build-server    Build web server only"
	@echo "  build-release   Generate code + build optimized binaries"
	@echo "  install         Generate code + install to GOPATH/bin"
	@echo "  clean           Remove build artifacts"
	@echo "  clean-all       Remove build + generated code"
	@echo ""
	@echo "Test targets:"
	@echo "  test            Generate + run all tests"
	@echo "  test-unit       Generate + run unit tests only"
	@echo "  test-integration Generate + run integration tests"
	@echo "  test-coverage   Generate + run tests with coverage report"
	@echo ""
	@echo "Code generation:"
	@echo "  sqlc            Generate sqlc code from SQL"
	@echo "  templ           Generate Go code from templ templates"
	@echo "  generate        Generate all code (sqlc + templ)"
	@echo ""
	@echo "Code quality:"
	@echo "  fmt             Format code"
	@echo "  lint            Generate + run linter"
	@echo "  tidy            Tidy go modules"
	@echo "  check           Format + lint + test"
	@echo ""
	@echo "Database:"
	@echo "  migrate         Build + run database migrations"
	@echo "  reset           Build + reset database to version 0"
	@echo ""
	@echo "Run:"
	@echo "  run             Build + run CLI"
	@echo "  serve           Build + start web server"
	@echo "  dev             Watch files and rebuild on changes"
