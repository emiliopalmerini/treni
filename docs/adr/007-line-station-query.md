# ADR-007: Line-at-Station Query

## Status
Accepted

## Context
ADR-006 shipped an empty bot skeleton. This ADR adds the primary use-case:
a user types a short query like `S9 Desio` and the bot replies with the
next hour of departures for that line at that station.

This ADR is the **minimum viable query path**. Station disambiguation and
direction disambiguation are follow-up ADRs (008 and 009).

## Decision

### Grammar (this ADR)
```
QUERY   := LINE WS STATION
LINE    := [A-Za-z]+[0-9]*          ; e.g. S9, RV, R, IC, FR
STATION := .+                        ; one or more words; free text
```

Normalized at parse time: `LINE` is upper-cased; `STATION` is trimmed.

A message that does not match this shape and does not start with `/` is
rejected with `"Query format: <LINE> <STATION>. Example: S9 Desio. /help for more."`

A message starting with `/` retains ADR-006 command semantics.

### Flow
1. Parse message into `(line, stationQuery)`.
2. Call `service.SearchStations(ctx, stationQuery)`.
3. If `len(results) == 0` → reply `"No station found for 'X'."`
4. Else take `results[0]` (first match). (ADR-008 adds the picker for the
   ambiguous case.)
5. Call `service.QueryDepartures(ctx, line, stationCode, window)`.
6. Format results as a single text message.

### New service method
```go
func (s *Service) QueryDepartures(
    ctx context.Context, line, stationCode string, window time.Duration,
) ([]domain.Departure, error)
```

- Calls `api.GetStation` (or a narrower `GetDepartures`; see below) to fetch
  the departure board.
- Filters by `TrainCategory` matching `line` case-insensitively (exact equal).
- Filters by `ScheduledTime` within `[now, now+window]`.
- Returns results sorted by `ScheduledTime` ascending.

For the filter to have the right granularity, expose `GetDepartures` on
`api.TrainClient` (the concrete `*viaggiatreno.Client` already has it;
we just need to add it to the interface).

### Line-matching rules
- Case-insensitive exact string equality against `TrainCategory`.
- If a departure has empty `TrainCategory`, it does not match any line.
- No fuzzy matching in this ADR; real-world quirks (S-line trains, RV vs
  Regionale Veloce) are deferred to a follow-up once we observe live data.

### Window
- Fixed at 60 minutes in this ADR. Configurable window is ADR-011.

### Reply format
One message, monospace block (` ``` ` fenced) so times align:

```
S9 @ MILANO BOVISA FNS · next 60min

14:32  +3'  bin 2   → Milano Porta Garibaldi
14:47  —    bin 2   → Milano Porta Garibaldi
15:02  canc bin ?   → Milano Porta Garibaldi
15:17  +12' bin 1   → Saronno
```

- `+N'` for positive delay in minutes; `—` for on-time; `canc` for cancelled.
- `bin ?` when platform unknown.
- Destination is the train's terminus.

Telegram parse mode: `MarkdownV2`. Formatter must escape MarkdownV2 special
characters in station/destination names.

### Empty results
If `QueryDepartures` returns an empty slice:
`"No <LINE> departures from <StationName> in the next 60 min."`

### Dispatcher change
`Dispatcher.Handle` currently routes any non-command to `fallbackHandler`.
Change: route non-slash messages to a new `queryHandler`; slash messages
with no matching command keep the existing `fallback` ("Unknown command").

### Inputs / outputs
- **Input:** Telegram text messages from whitelisted chats.
- **Output:** one `SendMessage` call with the formatted board, or an error
  message.
- **Side effects:** one external HTTP call per query (ViaggiaTreno).

### Edge cases
| Case | Handling |
|---|---|
| Empty message text | Ignore (no reply). |
| Text without a space | Treat as parse failure → format-hint reply. |
| LINE token with only digits (e.g. `2419`) | Parse failure in this ADR (bare train number is ADR-010). |
| Station query with punctuation / unicode | Passed through to search as-is. |
| `SearchStations` network error | Reply `"Couldn't reach ViaggiaTreno. Try again in a sec."` |
| `GetDepartures` network error | Same reply. |
| Zero departures at station (API returns `[]`) | "No departures …" message. |

### Error conditions
All errors from the service are logged and translated to the generic
"Couldn't reach ViaggiaTreno" reply. Details never leak to the user.

## Consequences
- `api.TrainClient` gains `GetDepartures(ctx, stationCode)`.
- `service` gains `QueryDepartures`.
- `internal/bot` gains `parser.go`, `query_handler.go`, `format.go`.
- Dispatcher wires plain text → queryHandler and keeps `/cmd` routing.
- Service needs constructor injection of a `time.Now` clock for testable
  windowing — passed as a `func() time.Time` field, defaulting to
  `time.Now`.
