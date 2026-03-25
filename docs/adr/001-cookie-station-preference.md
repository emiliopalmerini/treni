# ADR-001: Cookie-based Station Preference and Station Picker

## Status
Accepted

## Context
The app requires multiple clicks to reach a station's departure board. Users who check the same station daily (commuters) need zero-click access after initial setup.

## Decision
Use a server-readable HTTP cookie (`station=CODE:Name`) to persist the user's preferred station. The `Home` handler reads this cookie and either renders the departure dashboard (cookie present) or a station picker (cookie absent).

### Inputs
- Cookie `station` with format `CODE:Name` (URL-encoded)
- Station search query (for picker mode)
- `code` and `name` query params for `/api/pick-station`

### Outputs
- `StationPickerPage` rendered when no valid cookie exists
- `DashboardPage` rendered when cookie is valid (ADR-002)
- `PickerSearchResults` partial for HTMX search in picker mode
- Cookie set/cleared via `/api/pick-station` and `/api/clear-station`

### Edge Cases
- Invalid or malformed cookie value → treat as missing, show picker
- Station code in cookie no longer valid (API error) → show picker with error hint
- Station names with special characters (accents, periods) → URL-encode in cookie value
- Cookie expiry: 1 year, refreshed on each pick

### Error Conditions
- ViaggiaTreno API unreachable when loading dashboard → show error page
- Empty search query → return empty results (existing behavior)

## Changes
- `web/handlers/handlers.go` — Modify `Home`, `Search`; add `PickStation`, `ClearStation`
- `web/templates/home.templ` — Replace `HomePage` with `StationPickerPage`, `PickerSearchResults`
- `cmd/trenid/main.go` — Register `/api/pick-station`, `/api/clear-station`

## Consequences
- First visit requires one interaction (search + click)
- Every subsequent visit renders the board server-side in the initial response — no JS round-trip
- Cookie is ~50 bytes, negligible overhead
- `/station/{code}` deep links continue to work independently
