# Treni

Train tracking application with real-time status, station information, and historical analytics.

## Build and Run

```bash
make build          # Build both CLI and server
make build-cli      # Build CLI only
make build-server   # Build server only
make run            # Run web server
make test           # Run tests
make clean          # Clean build artifacts
```

## Code Generation

```bash
make generate       # Run both sqlc and templ generate
make sqlc           # Generate SQL queries
make templ          # Generate templ templates
```

## CLI Usage

```bash
./treni train <train_number>     # Get train status
./treni station <station_code>   # Get station arrivals/departures
./treni history <train_number>   # Get historical data for a train
```

## Project Structure

```
cmd/
  treni/          # CLI application
  trenid/         # Web server daemon
internal/
  api/            # External API clients (ViaggiaTreno, Trenord)
  domain/         # Core types (Train, Station, Delay)
  storage/        # SQLite/Turso repository
  service/        # Business logic
web/
  handlers/       # HTTP handlers
  templates/      # templ templates
  static/         # CSS/JS assets
migrations/       # Database migrations
```

## Tech Stack

- Go with chi router
- SQLite (Turso for production)
- templ for templates
- HTMX for interactivity
- sqlc for type-safe SQL

## Data Sources

- ViaggiaTreno (Trenitalia)
- Trenord API
