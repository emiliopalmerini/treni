# ADR-006: Pivot from Web App to Telegram Bot

## Status
Accepted

## Context
The web app requires opening a browser and navigating tabs to check train status. For the "get to my train fast" use-case, Telegram is a better surface: it is already open on the user's phone, supports conversational follow-ups (inline keyboards), and does not need a frontend toolchain.

This ADR is the minimal pivot: delete the web app, stand up a bot skeleton with whitelist enforcement and two no-op commands (`/start`, `/help`). Real train queries land in ADR-007+.

## Decision

### Deletions
- `cmd/trenid/` (HTTP server entry point)
- `web/` (templates, static assets, handlers, cookies)
- `treni.db` (stale artifact from the removed DB layer)
- `templ` and `chi` from `go.mod`
- Makefile targets related to templ/serve

### Additions
- `cmd/trenibot/main.go` — bot entry point.
- `internal/config/` — env var loader.
- `internal/bot/` — update dispatcher, whitelist guard, command handlers.
- `github.com/go-telegram/bot` dependency.

### Preserved
- `internal/api/` (ViaggiaTreno client) — reused by ADR-007+.
- `internal/domain/` — unchanged.
- `internal/service/` — `GetTrain`, `GetStation`, `SearchStations` remain; they will be called by bot handlers in later ADRs.

### Configuration (environment variables)
| Var | Required | Description |
|---|---|---|
| `TELEGRAM_BOT_TOKEN` | yes | Bot token from @BotFather. Startup fails if missing. |
| `TELEGRAM_ALLOWED_CHAT_IDS` | yes | Comma-separated list of integer chat IDs. Startup fails if empty. |
| `STATE_FILE` | no | Path to JSON state file (used by ADR-009). Default `./state.json`. Not used in this ADR. |

### Commands
| Command | Behavior |
|---|---|
| `/start` | Reply: `"Ciao. Send me `<LINE> <STATION>` (e.g. `S9 Desio`) or a train number. /help for more."` |
| `/help` | Reply with the same text as `/start` for now (will be expanded in later ADRs). |
| Anything else | Reply: `"Unknown command. /help for usage."` |

### Whitelist enforcement
- Every incoming update's `Message.Chat.ID` is checked against the allowed-IDs set.
- Non-allowed chats: **silently drop**. No reply, no log at INFO (DEBUG log is fine).
- Allowed chats: dispatch to command handler.

### Dispatcher shape
```go
type Handler func(ctx context.Context, b Sender, msg *models.Message) error

type Dispatcher struct {
    allowed map[int64]struct{}
    cmds    map[string]Handler
    fallback Handler
}

func (d *Dispatcher) Handle(ctx context.Context, b Sender, update *models.Update) error
```

`Sender` is a narrow interface (`SendMessage(...)`) so handlers are testable without a real Telegram client.

### Inputs / outputs (ADR-006 scope)
- **Input:** Telegram `Update` objects via long-polling (`bot.Bot.Start`).
- **Output:** `sendMessage` API calls.
- **Side effects:** none. No state file reads/writes in this ADR.

### Edge cases
- Update without `Message` (edited message, callback query, etc.): ignore.
- Message without `Chat`: drop (can't check whitelist).
- Message text empty or without `/` prefix: route to fallback handler (`"Unknown command. /help for usage."`) — but only for whitelisted users.
- `TELEGRAM_ALLOWED_CHAT_IDS` parse error (non-integer token): startup fails with a clear error.
- Network / Telegram API errors: logged, bot keeps polling (library handles reconnect).
- `SIGINT`/`SIGTERM`: cancel the bot context, `bot.Start` returns, process exits cleanly.

### Error conditions
| Condition | Handling |
|---|---|
| Missing `TELEGRAM_BOT_TOKEN` | `log.Fatal` at startup with `"TELEGRAM_BOT_TOKEN is required"` |
| Empty/malformed `TELEGRAM_ALLOWED_CHAT_IDS` | `log.Fatal` with parse error |
| Bot API init error | `log.Fatal` |
| Handler returns error | Log it; do not reply to the user (avoid leaking internals) |

## Consequences
- The binary name changes from `trenid` → `trenibot`.
- No more HTTP port; deployment needs to provide the two env vars instead.
- The Go module picks up one new dep (`github.com/go-telegram/bot`), drops two (`templ`, `chi`).
- `CLAUDE.md` project section needs to be rewritten; the "Build and Run" table, "CLI Usage" block, and "Tech Stack" block are all stale.
- `flake.nix` build target is renamed.
- Future ADRs (007–010) plug handlers into the dispatcher without needing to revisit the skeleton.
