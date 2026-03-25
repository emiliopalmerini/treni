# ADR-003: Navigation and Route Cleanup

## Status
Accepted

## Context
The analytics page and multi-link navigation add complexity that doesn't serve the core use case (checking delays quickly). Removing them simplifies the app.

## Decision
Remove the analytics feature entirely and simplify the navigation to just the brand link.

## Changes
- `web/templates/layout.templ` — Remove nav-links div, keep only brand link
- `cmd/trenid/main.go` — Remove `/analytics`, `/api/analytics/delayed`, `/api/analytics/reliable` routes
- `web/handlers/handlers.go` — Remove `Analytics`, `DelayedRankings`, `ReliableRankings` handlers
- `web/templates/analytics.templ` — Delete file

## Consequences
- Simpler nav, fewer routes, less code to maintain
- Analytics data collection (storage layer) remains intact — can be re-added later if needed
- Service methods `GetMostDelayedTrains`/`GetMostReliableTrains` remain available
