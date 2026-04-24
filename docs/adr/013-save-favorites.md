# ADR-013: Save Favorite Routes

## Status
Accepted

## Context
Users repeat the same `FROM > TO` queries every day (commute, usual
connections). Re-typing the full grammar each time is friction that
the bot can remove cheaply. CLAUDE.md has long earmarked an "optional
JSON state file for favorites (future ADR)"; this is that ADR.

`/start` and `/help` today both send a single `welcomeText` that
still references the pre-ADR-010 grammar (`S9 Desio`). They need to
be updated regardless; this ADR folds that fix in.

## Decision

### Commands

| Command | Behavior |
|---|---|
| `/save <name> <FROM> > <TO>` | Save a route under `<name>` for this chat. Overwrites if `<name>` already exists. |
| `/unsave <name>` | Delete the saved route named `<name>` in this chat. |
| `/favorites` | List this chat's saved routes with Run/Delete buttons. |
| `<name>` (plain text, no `>`) | Run the saved route for `<name>` in this chat. |

### Nickname rules
- Case-insensitive; stored and matched lowercase.
- 1–32 characters.
- No whitespace, no `>`, must not start with `/`.
- Collision with an existing name overwrites and replies
  `Updated '<name>': <FROM> > <TO>.`

### Route parsing
`<FROM> > <TO>` inside `/save` is parsed with the **same rule as
ADR-010**: split on the first `>`, trim both sides, reject if
either is empty. FROM and TO are stored verbatim (no station-code
resolution at save time, so renamed or disambiguated stations keep
working; disambiguation happens at run time via the existing
picker).

### Retrieval paths

**`/favorites` list.**
Text body, one line per favorite (sorted by name):
```
Saved routes (this chat):

home — Desio > Milano
office — Milano Centrale > Brescia
```
Inline keyboard: for each favorite one row with two buttons,
`▶ <name>` (callback `fr:<name>`) and `🗑` (callback `fd:<name>`).
Delete edits the message text and keyboard in place, removing the
row. When the last favorite is deleted, the message is edited to
`No favorites yet. Use /save <name> <FROM> > <TO>.` with no
keyboard.

Empty state on first `/favorites`: same "No favorites yet" body,
no keyboard.

**Nickname as plain text.**
The bot's text handler chain, in order:
1. If the message contains `>` → existing query handler (ADR-010).
2. Else look up `lower(text)` in this chat's favorites.
   - Hit → run as if the user had sent `FROM > TO`. Reuses the
     existing station picker flow (`q:<CODE>:<TO>` callback)
     unchanged.
   - Miss → existing query handler, which returns the format hint.

### Callback data format
- `fr:<name>` — run favorite.
- `fd:<name>` — delete favorite.

`<name>` is ≤32 chars, so both comfortably fit the 64-byte
Telegram callback_data limit. Chat ID is read from the callback's
message, not embedded in the data (same pattern as `q:`).

### Persistence

Uses the existing `STATE_FILE` env var (default `./state.json`).

Schema:
```json
{
  "favorites": {
    "<chat_id>": {
      "<name>": { "from": "Desio", "to": "Milano" }
    }
  }
}
```

- Load on startup. Missing file → empty state. Corrupt JSON → bot
  fails to start with a clear error (don't silently discard user
  data).
- Writes are serialized through an in-memory mutex and performed as
  write-tmp + `os.Rename` for atomicity on the same filesystem.
- Every `/save` and `/unsave` (and each `fd:` delete) flushes the
  whole state file. The state is small (≤10 entries × N chats); no
  incremental-write complexity.

### Limits

- Hard cap: **10 favorites per chat**.
- `/save` at cap replies:
  `Favorite limit reached (10). Delete one with /unsave <name>.`
  Overwrites of an existing name do not count against the cap.

### Start and help copy

`/start` (welcome, short):
```
Ciao. Send me <FROM> > <TO> (e.g. Desio > Milano) to see departures.
/help for commands and favorites.
```

`/help` (full reference):
```
Query:
  <FROM> > <TO>        e.g. Desio > Milano

Favorites (per chat, max 10):
  /save <name> <FROM> > <TO>   save a route
  /unsave <name>               delete a saved route
  /favorites                   list saved routes
  <name>                       run a saved route

Times are the next 60 min from now.
```

Tests that currently send `/start` and `/help` only assert
non-error; they'll keep passing with the rewritten copy.

### Error messages

| Case | Reply |
|---|---|
| `/save` with no args or missing `>` | `Usage: /save <name> <FROM> > <TO>.` |
| `/save` with empty FROM or TO | `Usage: /save <name> <FROM> > <TO>.` |
| `/save` with invalid name (spaces, `>`, `/` prefix, >32 chars) | `Invalid name. Use 1–32 chars, no spaces, no '>'.` |
| `/save` at cap (and name is new) | `Favorite limit reached (10). Delete one with /unsave <name>.` |
| `/unsave` with no arg | `Usage: /unsave <name>.` |
| `/unsave` for missing name | `No favorite named '<name>'.` |
| `fr:<name>` callback where name no longer exists | `That favorite is gone. /favorites to refresh.` |
| `fd:<name>` callback where name no longer exists | Silent: dismiss the spinner, leave the message as-is. |

### Edge cases

| Case | Handling |
|---|---|
| Same `<name>` in different chats | Independent; storage is keyed by `(chat_id, name)`. |
| `/save` with extra whitespace between tokens | Collapsed on parse; FROM/TO trimmed. |
| Name shadows a station name (`milano`) | Favorites lookup wins only when the message has no `>`. `milano > bergamo` still goes to the query parser. |
| State file exists but `favorites` key missing | Treat as empty; upgrade on next write. |
| Two `/save` writes race | Serialized by the store's mutex. |
| State file dir doesn't exist | Bot fails to start (same as any other startup file error). |

## Consequences

- New package `internal/state` with a `FavoritesStore` interface
  and a JSON-file implementation.
- New domain type `domain.Favorite{Name, From, To}`.
- Service layer gains favorites operations (List, Get, Save,
  Delete) so handlers stay thin.
- New bot handlers: `/save`, `/unsave`, `/favorites`, plus callback
  handlers for `fr:` and `fd:`.
- Dispatcher's text path gets a favorites-alias shim in front of
  the query handler.
- `welcomeText` replaced by separate `startText` and `helpText`
  constants; both reflect ADR-010 grammar.
- `cmd/trenibot` wires the store: resolve `STATE_FILE`, load on
  start, inject into the service.
- Tests:
  - unit: nickname validation, `/save` parser, favorites store
    (round-trip, atomic write, cap enforcement, corrupt-file
    startup error).
  - handler: `/save`, `/unsave`, `/favorites` happy paths and each
    error row above; callback handlers for `fr:` and `fd:`.
  - dispatcher: alias shim — message with `>` bypasses favorites,
    message without `>` looks them up, miss falls through.
  - no changes to ADR-010/011/012 query logic or its tests.
