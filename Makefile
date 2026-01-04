.PHONY: all fmt vet templ build run test clean

BINARY_NAME=treni
CMD_PATH=./cmd

all: build

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

templ:
	templ generate

build: vet templ
	go build -o $(BINARY_NAME) $(CMD_PATH)

run: build
	./$(BINARY_NAME)

test: vet
	go test -v ./...

clean:
	go clean
	rm -f $(BINARY_NAME)
