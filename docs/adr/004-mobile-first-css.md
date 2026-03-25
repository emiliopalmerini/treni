# ADR-004: Mobile-first CSS Redesign

## Status
Accepted

## Context
The app is primarily used on a phone to check station delays. The current CSS is desktop-first with a 640px breakpoint. It needs to be mobile-first with larger touch targets and delay-focused visual design.

## Decision
Rewrite the CSS mobile-first: base styles target phones, media queries add desktop enhancements.

## Changes
- `web/static/css/main.css`:
  - Delay row severity styles (borders, backgrounds, opacity)
  - Station picker centered layout
  - Dashboard header (flex, change-station button)
  - Mobile board: reduced padding, larger tap targets, card-based rows on narrow screens
  - Remove unused hero, quick-links, analytics CSS

## Consequences
- Better phone experience for the primary use case
- Desktop still works but is no longer the primary design target
