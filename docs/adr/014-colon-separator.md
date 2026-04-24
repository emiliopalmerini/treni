# ADR-014: Switch Query Separator from `>` to `:`

## Status
Accepted (amends ADR-010; touches ADR-013 copy)

## Context
ADR-010 introduced the `<FROM> > <TO>` grammar. In day-to-day use the
`>` character has two ergonomic drawbacks:

- On iOS and most mobile keyboards, `>` sits behind a Shift or
  symbol-page toggle, adding a tap per query.
- In Telegram clients `>` is occasionally interpreted as a
  blockquote/markdown lead character when pasted or forwarded,
  producing visually mangled messages.

`:` is on the primary symbol row of every mobile keyboard we care
about, carries no Telegram formatting meaning, and reads naturally
as "origin: destination" (same shape as `from:to` labels people
already type in tickets and chat).

## Decision

### Grammar
```
QUERY := FROM WS ":" WS TO
FROM  := free text (may be multi-word)
TO    := free text (may be multi-word)
```

Example queries:
- `Desio: Milano`
- `Milano Centrale:Brescia`
- `Desio : Saronno`

Whitespace around the `:` is optional, exactly as `>` was under
ADR-010. Messages that don't contain a `:` fall back to the format
hint.

### Parser change
`parseFromTo` splits on the **first** `:` instead of the first `>`.
All other logic (trim both sides, reject empty FROM/TO) is
unchanged. The split-on-first rule means a colon appearing later in
either side (unusual but possible in free-text station names) lands
in TO; this matches how ADR-010 handled stray `>` characters and is
the simplest rule to explain.

### What this replaces
- ADR-010 grammar separator `>`: removed from the parser, the format
  hint, and all user-facing copy.
- No change to resolution flow, station picker, callback plumbing,
  or `DeparturesVia` semantics (ADR-011/012 logic stays as-is).

### User-facing copy updates
| Location | Before | After |
|---|---|---|
| `formatHint` in `query_handler.go` | `Query format: <FROM> > <TO>. Example: Desio > Milano. /help for more.` | `Query format: <FROM>: <TO>. Example: Desio: Milano. /help for more.` |
| `welcomeText` / `startText` | references to `<FROM> > <TO>` | same sentences with `:` |
| `/help` body (ADR-013) | `<FROM> > <TO>` and `/save <name> <FROM> > <TO>` | `<FROM>: <TO>` and `/save <name> <FROM>: <TO>` |

### Dispatcher alias shim (ADR-013)
ADR-013's text path routes to the query handler when the message
contains `>` and to favorites-lookup otherwise. That check flips
from `strings.Contains(text, ">")` to `strings.Contains(text, ":")`.

Name-shadow consequence: a favorite literally named `foo:bar` would
now be ambiguous with a query. We already forbid `>` in favorite
names (ADR-013); extend that rule to forbid `:` for the same
reason. Existing favorites are stored as `{from, to}` pairs, not as
raw grammar, so on-disk state is unaffected. The validation rule
tightens only for new saves.

### `/save` command body
`/save <name> <FROM> : <TO>` is parsed with the same split-on-first-
`:` rule as the top-level query. The ADR-013 error messages update
their example text from `>` to `:` but keep their structure.

### Callback data format
Unchanged. `q:<STATION_CODE>:<TO>` already uses `:` internally; the
parser splits on the **first two** `:` occurrences, so a TO string
that itself contains a `:` stays in the trailing segment without
ambiguity (same position as before). Truncation bound stays at 40
bytes.

### Backward compatibility
No compatibility window. This is a personal bot with a known set of
chats; the old separator is removed in the same change. Users see
the updated format hint the first time they send a `>` query.

Saved favorites in `state.json` are unaffected (stored as
`{from, to}`, not grammar).

### Edge cases
| Case | Handling |
|---|---|
| No `:` in message | Format hint. |
| Empty FROM or TO | Format hint. |
| Message contains both `:` and `>` | Split on `:` (first occurrence), treat `>` as literal text. Resolution will then fail to find a station, user gets "No station found". |
| Message is a single `:` | Format hint (both sides empty). |
| `/save foo Desio : Milano` | Same as ADR-013 happy path. |
| `/save foo:bar Desio : Milano` | Rejected: name contains `:`. |

### Future work
- None specific to this change. The separator is now on a mobile-
  friendly key and we're done.

## Consequences
- `internal/bot/parser.go`: one-character change (`">"` → `":"`).
- `internal/bot/parser_test.go`: every fixture that uses `>` flips
  to `:`. Add a regression case for `Desio:Milano` (no spaces) and
  for a message containing both `:` and `>` (should parse as FROM=
  `Desio`, TO=`Milano > something` or similar; assert the parse
  result, not its downstream resolution).
- `internal/bot/query_handler.go`: update `formatHint` constant.
- `internal/bot/handlers.go`: update `welcomeText` (or its
  ADR-013 successors `startText` / `helpText`).
- `internal/bot/dispatcher.go`: flip the `>` presence check to `:`
  in the favorites-vs-query alias shim (only relevant after
  ADR-013 lands; sequence this ADR accordingly).
- ADR-013 text: `/save` usage lines and `/help` body update from
  `>` to `:`. The parsing rule ("split on the first separator")
  is unchanged in shape.
- Historical ADRs (010, 011, 012) stay as-is; this ADR notes the
  amendment at the top of 010 when it ships.
