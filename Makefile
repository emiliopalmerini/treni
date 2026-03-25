.PHONY: all build test test-unit test-integration clean templ fmt lint run serve install help

# Variables
BINARY := trenid
GO_FILES := $(shell find . -name '*.go')
TEMPL_FILES := $(shell find . -name '*.templ')

# Default target
all: build

# === Code Generation ===

# Generate templ templates (depends on templ files)
templ: $(TEMPL_FILES)
	templ generate

# Alias
generate: templ

# === Build ===

# Build server (depends on generated code)
build: generate
	go build -o $(BINARY) ./cmd/trenid

# Build with version info
build-release: generate
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/trenid

# Install to GOPATH/bin
install: generate
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

# === Run ===

# Start web server
run: build
	./$(BINARY)

# Alias
serve: run

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
	rm -f $(BINARY)
	rm -f coverage.out coverage.html

# Clean generated code too
clean-all: clean
	rm -f web/templates/*_templ.go

# === Help ===

help:
	@echo "treni - Train tracking web application"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build targets:"
	@echo "  all             Build everything (default)"
	@echo "  build           Generate code + build server"
	@echo "  build-release   Generate code + build optimized binary"
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
	@echo "  templ           Generate Go code from templ templates"
	@echo "  generate        Alias for templ"
	@echo ""
	@echo "Code quality:"
	@echo "  fmt             Format code"
	@echo "  lint            Generate + run linter"
	@echo "  tidy            Tidy go modules"
	@echo "  check           Format + lint + test"
	@echo ""
	@echo "Run:"
	@echo "  run             Build + start web server"
	@echo "  serve           Alias for run"
	@echo "  dev             Watch files and rebuild on changes"
