# ADR-008: Station Disambiguation via Inline Keyboard

## Status
Accepted

## Context
ADR-007 silently takes the first `SearchStations` result. When multiple
stations match a short name (`Milano` → several stations; `San Giovanni`
→ many), the user has no way to steer the lookup. This ADR adds an
inline-keyboard picker for the ambiguous case.

## Decision

### Trigger
In the query handler, after `SearchStations(q)`:

| `len(results)` | Behavior |
|---|---|
| 0 | Unchanged: `"No station found for 'X'."` |
| 1 | Unchanged: proceed to `QueryDepartures` with `results[0]`. |
| 2–`maxChoices` | Reply with inline keyboard (one button per station). |
| > `maxChoices` | Take the first `maxChoices` and show them; list is truncated silently (rationale: a better-specified query is the right fix). |

`maxChoices = 5`. Telegram supports more, but five fits on a phone without
scrolling the keyboard.

### Keyboard layout
One button per row (vertical list; station names wrap badly side-by-side).
Each button's label is the station's `Name` (truncated to 40 chars if needed).

### Callback data format
```
q:<LINE>:<STATION_CODE>
```

Station codes are ~6–7 chars (`S01700`); lines are ≤4 chars (`S9`, `RV`, `FR`).
Total length comfortably under Telegram's 64-byte `callback_data` limit.

### Callback flow
1. User taps a button. Telegram sends a `CallbackQuery` update.
2. Bot parses the callback data, extracts `line` and `stationCode`.
3. Bot calls `service.QueryDepartures(line, stationCode, window)`.
4. Bot **edits** the original message to replace the picker with the
   formatted departures board.
5. Bot answers the callback query (`AnswerCallbackQuery`) so Telegram
   dismisses the button's loading spinner.

Editing (instead of sending a new message) keeps the conversation clean.
The picker message and the result occupy the same message slot.

### Sender interface extension
```go
type Sender interface {
    SendMessage(ctx context.Context, chatID int64, text string) error

    // ADR-008 additions:
    SendMessageWithButtons(ctx context.Context, chatID int64, text string, buttons []Button) error
    EditMessageText(ctx context.Context, chatID int64, messageID int, text string) error
    AnswerCallback(ctx context.Context, callbackID string) error
}

type Button struct {
    Text string  // label
    Data string  // callback_data
}
```

Buttons are a flat slice — each becomes its own row (per layout rule above).

### Dispatcher extension
Add callback-query routing:
```go
func (d *Dispatcher) OnCallback(prefix string, h CallbackHandler)

type CallbackHandler func(ctx context.Context, s Sender, cq *models.CallbackQuery) error
```

`Handle` now also looks at `update.CallbackQuery`:
- Whitelist check on `cq.Message.Chat.ID` (chat from which the keyboard
  was tapped). If `cq.Message` is nil (edge case), fall back to
  `cq.From.ID`.
- If the callback data starts with a registered prefix, invoke that
  handler.
- Else: answer the callback with an empty text (dismiss spinner) and
  drop silently.

### Edge cases
| Case | Handling |
|---|---|
| Callback data malformed | Answer callback empty; log; no user-visible reply. |
| Line in callback not a valid station+line anymore | `EditMessageText` with generic "no departures"-style reply. |
| `QueryDepartures` errors on callback | Edit with `upstreamDownMsg`. |
| Callback from non-whitelisted chat | Answer empty, drop. |
| `cq.Message` is nil | Drop. Telegram only sends this when the original message is too old (>48h); shouldn't happen in normal use. |

### Inputs / outputs
- **New input path:** Telegram `CallbackQuery` updates.
- **New outputs:** `editMessageText`, `answerCallbackQuery`,
  `sendMessage` with `inline_keyboard`.

## Consequences
- `Sender` interface widens — all fakes need the new methods (acceptable;
  only two exist: `tgSender` and the test fake).
- `Dispatcher.Handle` now dispatches on two union cases (Message vs
  CallbackQuery).
- `query_handler.go` grows a picker path; `callback_handler.go` is new.
- No persistent state: callback data carries everything needed. This is
  important because ADR-006's state file doesn't exist yet.
