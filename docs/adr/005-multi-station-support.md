# ADR-005: Multi-Station Support

## Status
Accepted

## Context
Currently the dashboard supports a single saved station via the `station` cookie. Users who commute through multiple stations (e.g., home station + office station) must clear and re-pick each time.

## Decision
Support up to 5 saved stations with a tab-based switcher on the dashboard.

### Cookie format
- Cookie name: `stations`
- Value: pipe-separated `code:name` pairs, e.g. `S01700:MILANO+CENTRALE|S08409:ROMA+TERMINI`
- Names are URL-encoded
- Max 5 entries; adding a 6th is rejected

### New/changed endpoints
| Endpoint | Method | Change |
|---|---|---|
| `GET /` | Home | Parse `stations` cookie, load first station, render multi-station dashboard |
| `GET /api/pick-station?code=X&name=Y` | PickStation | Append to cookie (reject if 5). Redirect home |
| `GET /api/remove-station?code=X` | RemoveStation | Remove one station from cookie. If last, redirect to picker |
| `GET /api/clear-station` | ClearStation | Removed — replaced by RemoveStation |
| `GET /api/station-content/{code}` | StationContent | New. Returns the departures/arrivals tabs + board partial for HTMX swap |

### Dashboard UI
```
[Milano Centrale ×] [Roma Termini ×] [+]    ← station tabs
[Departures] [Arrivals]                      ← board tabs
┌──────────────────────────────────┐
│  departure/arrival table         │
└──────────────────────────────────┘
```

- Clicking a station tab calls `/api/station-content/{code}` and swaps the content area
- "+" opens the station picker (with a back link to dashboard)
- "×" calls `/api/remove-station?code=X`
- Active station tab is highlighted

### Picker changes
- When stations already exist, picker shows "Back to dashboard" link
- After picking a station, user returns to dashboard with new station active

### Edge cases
- No stations cookie → show picker (same as today)
- Station code in cookie no longer valid → skip it, show remaining stations
- All stations invalid → show picker
- Cookie has duplicate codes → deduplicate on read
- Adding a station already in the list → move it to the end (no duplicate)

## Consequences
- Old `station` cookie is ignored (no backward compat)
- Single cookie, no DB involvement
- Max cookie size for 5 stations: well under 4KB limit
