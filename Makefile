.PHONY: all build test test-unit test-integration clean fmt lint run tidy check help

BINARY := trenibot

all: build

# === Build ===

build:
	go build -o $(BINARY) ./cmd/trenibot

build-release:
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/trenibot

install:
	go install ./cmd/trenibot

# === Testing ===

test:
	go test -v ./...

test-unit:
	go test -v -short ./...

test-integration:
	go test -v -run Integration ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# === Code Quality ===

fmt:
	go fmt ./...
	gofmt -s -w .

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

check: fmt lint test-unit

# === Run ===

run: build
	./$(BINARY)

# === Cleanup ===

clean:
	rm -f $(BINARY)
	rm -f coverage.out coverage.html

# === Help ===

help:
	@echo "trenibot - Italian train tracking Telegram bot"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build:   build | build-release | install | clean"
	@echo "Test:    test | test-unit | test-integration | test-coverage"
	@echo "Quality: fmt | lint | tidy | check"
	@echo "Run:     run"
	@echo ""
	@echo "Required env to run: TELEGRAM_BOT_TOKEN, TELEGRAM_ALLOWED_CHAT_IDS"
