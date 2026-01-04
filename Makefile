.PHONY: all fmt vet templ sqlc generate build run test clean migrate-up migrate-down migrate-create

BINARY_NAME=treni
CMD_PATH=./cmd
DB_PATH=./treni.db
MIGRATIONS_PATH=./internal/database/migrations

all: build

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

sqlc:
	sqlc generate

templ:
	templ generate

generate: sqlc templ

build: vet generate
	go build -o $(BINARY_NAME) $(CMD_PATH)

run: build
	./$(BINARY_NAME)

test: vet
	go test -v ./...

clean:
	go clean
	rm -f $(BINARY_NAME)

migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "sqlite3://$(DB_PATH)" up

migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "sqlite3://$(DB_PATH)" down 1

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $$name
