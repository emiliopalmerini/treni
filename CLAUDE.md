# Treni

Telegram bot for Italian train tracking. Users message queries like `S9 Desio`
and the bot replies with live departures filtered by line.

## Build and Run

```bash
make build          # Build trenibot binary
make run            # Build + run (requires env vars below)
make test           # Run all tests (includes live-API integration tests)
make test-unit      # Run unit tests only (-short)
make clean          # Clean build artifacts
```

## Environment Variables

| Var | Required | Description |
|---|---|---|
| `TELEGRAM_BOT_TOKEN` | yes | Bot token from @BotFather |
| `TELEGRAM_ALLOWED_CHAT_IDS` | yes | Comma-separated allowed chat IDs |
| `STATE_FILE` | no | Path to JSON state file (default `./state.json`) |

## Project Structure

```
cmd/
  trenibot/       # Bot entry point
internal/
  api/            # ViaggiaTreno client + TrainClient interface
  bot/            # Dispatcher, handlers, Telegram adapter
  config/         # Env-var loader
  domain/         # Core types (Train, Station, Departure, Arrival)
  service/        # Business logic (wraps TrainClient)
docs/adr/         # Architecture Decision Records
```

## Tech Stack

- Go 1.26
- `github.com/go-telegram/bot` for Telegram long-polling
- No database; optional JSON state file for favorites (future ADR)

## Data Sources

- ViaggiaTreno (Trenitalia)
