# ADR-002: Delay-focused Departure Dashboard

## Status
Accepted

## Context
The current station page treats all departures equally. The primary use case is checking delays — delayed trains should be visually prominent.

## Decision
Enhance the departure/arrival board with row-level delay severity styling and increase the auto-refresh rate from 60s to 30s. Add a `DashboardPage` template that wraps the enhanced board for the homepage.

### Inputs
- `*domain.Station` with populated `Departures` and `Arrivals`
- Each departure/arrival's `Delay` (int, minutes) and `Status` (TrainStatus)

### Outputs
- `DashboardPage(station)` — homepage template with station header, change-station button, departure board
- Enhanced `DeparturesPartial` / `ArrivalsPartial` with row-level CSS classes
- `delayRowClass` helper function

### Delay Severity Classes
| Condition | Class | Visual |
|-----------|-------|--------|
| On time (0 min) | `row-on-time` | Muted opacity |
| Low (1-5 min) | `row-delay-low` | Yellow left border |
| Medium (6-15 min) | `row-delay-mid` | Orange left border, bolder |
| High (16+ min) | `row-delay-high` | Red left border, red tint |
| Cancelled | `row-cancelled` | Red bg, strikethrough |

### Edge Cases
- Negative delay (train early) → treat as on-time
- Zero departures → "No departures at this time" message (existing)

## Changes
- `web/templates/home.templ` — Add `DashboardPage`
- `web/templates/station.templ` — Add row classes to `DeparturesPartial`/`ArrivalsPartial`, 30s refresh
- `web/templates/components.templ` — Add `delayRowClass` helper

## Consequences
- Delayed trains are immediately visible without reading individual cells
- 30s refresh increases API calls by 2x but ViaggiaTreno has no rate limits for reasonable usage
- Existing `/station/{code}` page also benefits from the enhanced board
