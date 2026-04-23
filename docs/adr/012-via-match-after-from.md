# ADR-012: Only Match Stops That Come After FROM

## Status
Accepted

## Context
ADR-011 matches `TO` against any stop on the train's route. That's
too loose: the same physical S9 line runs Saronno↔Albairate in both
directions and stops at Milano in between. From Desio:
- Trains heading *toward* Milano have Milano **after** Desio in the
  stop list — correct match.
- Trains heading *away* from Milano have Milano **before** Desio in
  the stop list — should not match, but currently does.

## Decision

### Rule
A train matches `FROM > TO` iff `TO` (case-insensitive substring)
matches one of:
- `departure.Destination` (the terminus, unchanged), **or**
- `train.Stops[i].StationName` for some `i > fromIdx`, where
  `fromIdx` is the index of FROM in the stop list.

FROM is located in the stop list by `StationCode` equality (robust
against stop-name spelling quirks).

### FROM not found in stops
If the FROM station code doesn't appear in the returned stop list
(unusual; API quirk or a train that skips the queried station),
fall back to terminus-match only. Don't silently match every stop
in both directions, which was the bug.

### Inputs / outputs
`DeparturesVia` signature unchanged (`ctx, stationCode, viaMatch,
window`). Only internal logic changes.

### Edge cases
| Case | Handling |
|---|---|
| FROM code matches twice in stops (loop service?) | Use the first occurrence; stops after it are still downstream. |
| GetTrain succeeds but returns empty `Stops` | Terminus-only match (same as ADR-011 fallback). |

### Inputs into the matcher
Plumb `fromCode` into the per-train worker (previously only the TO
needle was passed).

## Consequences
- One small logic change in `DeparturesVia`; signature unchanged.
- New service test covers the "TO appears before FROM" case that
  should no longer match.
- Supersedes ADR-011's "any stop" rule with a more correct "any
  stop downstream" rule.
