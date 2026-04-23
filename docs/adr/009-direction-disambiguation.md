# ADR-009: Direction Disambiguation via Inline Keyboard

## Status
Superseded by ADR-010

## Context
Lines like S9 run in two directions (Saronno ↔ Albairate via Milano).
From a mid-line station like Desio, the user almost always wants *one*
direction. Today the bot returns all upcoming trains regardless of
terminus; the user has to visually filter.

This ADR adds a direction picker that appears after the line+station is
resolved (either directly from a single search hit, or after an ADR-008
station pick callback).

## Decision

### Trigger
After `QueryDepartures` returns, inspect the distinct `Destination`
values among the returned departures:

| distinct termini | Behavior |
|---|---|
| 0 | Unchanged empty-state reply. |
| 1 | Unchanged: list the departures. |
| 2+ | Send a direction keyboard with one button per terminus plus a "Tutti" (all) button. |

The "all" button is equivalent to the current behavior (list every
departure).

### Keyboard layout
- One button per row: `→ <Terminus>`.
- Final row: `Tutti` (show all).

### Callback data format
```
d:<LINE>:<STATION_CODE>:<TERMINUS_OR_*>
```

- `TERMINUS_OR_*` is either the terminus name truncated to 40 bytes, or
  the literal `*` for "show all".
- Total length well under Telegram's 64-byte limit for realistic Italian
  station names.

Filtering on callback: a departure matches if `strings.HasPrefix(
departure.Destination, terminus)` (case-insensitive). Prefix match keeps
us safe if truncation chopped a trailing word off a long terminus name.

### Shared "post-station" path
The flow diverges at the station level (search vs station-pick callback)
but converges after: both paths need the same "fetch + maybe show
direction picker + maybe show list" logic.

Introduce an internal `messageTarget` abstraction:
```go
type messageTarget struct {
    chatID    int64
    messageID int  // 0 = send new; non-zero = edit existing
}

func (t messageTarget) renderText(ctx, s, text) error
func (t messageTarget) renderWithButtons(ctx, s, text, buttons) error
```

And `routeAfterStation(ctx, s, svc, target, line, station, window) error`
becomes the shared entry point from:
- the text-entry query handler (when there's a single station match), and
- the station-pick callback handler.

Both ADR-007 and ADR-008 `replyWithDepartures` call sites are replaced.

### Sender interface extension
```go
type Sender interface {
    // existing methods …
    EditMessageWithButtons(ctx context.Context, chatID int64, messageID int, text string, buttons []Button) error
}
```

### New callback route
Main registers two callback prefixes:
- `q:` → station pick (ADR-008)
- `d:` → direction pick (this ADR)

### Edge cases
| Case | Handling |
|---|---|
| Terminus containing `:` | Shouldn't occur for Italian station names; if it does, `SplitN` on `:` with limit 3 preserves it in the last field. |
| Truncated terminus that no longer uniquely identifies a train | Prefix match keeps behavior sane; user still sees trains heading that way. |
| Direction-picker callback returns zero departures (e.g., train left between picker and tap) | Edit with the standard empty-state reply. |
| Callback for a station that no longer has trains in window | Same as above. |

### Inputs / outputs
- Input surface unchanged at text-entry; adds a new callback prefix `d:`.
- Output: one extra inline-keyboard render per multi-terminus query;
  an edit per direction pick.

## Consequences
- `Sender` grows one method.
- `query_handler.go` is split: station picker stays; post-station logic
  moves into a shared file (`route.go` or inline).
- `callback_handler.go` grows a direction-callback path.
- No persistent state required; all context travels in the callback data.
- Existing ADR-007/008 tests must be updated to reflect that multi-terminus
  results now produce a keyboard first, not a list directly.
