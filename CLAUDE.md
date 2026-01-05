# Treni

Train tracking web application built with Go, HTMX, and templ.

## Build and Run

```bash
make build          # Build binary
make run            # Build and run
make test           # Run tests
make clean          # Clean build artifacts
```

## Code Generation

```bash
make generate       # Run both sqlc and templ generate
make sqlc           # Generate SQL queries
make templ          # Generate templ templates
```

## Database

Migrations are applied automatically when the app starts.

## Code Quality

```bash
make fmt            # Format code
make vet            # Run go vet (includes fmt)
```

## Tech Stack

- Go with chi router
- SQLite with sqlc
- templ for templates
- HTMX for interactivity
