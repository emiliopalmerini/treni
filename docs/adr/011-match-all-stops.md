# ADR-011: Match TO Against Every Stop, Not Just the Terminus

## Status
Accepted

## Context
ADR-010 filters departures where the **terminus** name contains TO.
That's too narrow: a train terminating at Albairate but stopping at
Milano Porta Garibaldi is a valid answer to `Desio > Milano`, and the
current filter drops it.

## Decision

### Semantics
Rename `DeparturesFromTo` → `DeparturesVia`. A departure matches if any
of these is true:
1. `departure.Destination` contains TO (case-insensitive).
2. Any stop in the train's full stop list has a `StationName` that
   contains TO (case-insensitive).

### Fetch strategy
For the candidate list (departures at FROM, within window), fan out
`GetTrain(trainNumber)` concurrently with a worker pool (cap 8). For
each train:
- If the `GetTrain` call succeeds → run the any-stop match.
- If it fails (timeout, 500, parse error) → fall back to the old
  terminus match (rule 1 only). Don't drop on failure; don't let one
  bad call veto the result.

Each worker has a 10-second budget; the outer query's context still
applies.

### Ordering
Preserve the original ascending `ScheduledTime` order, matching
ADR-010 behavior. Concurrency is purely for throughput.

### Inputs / outputs (contract)
```go
func (s *Service) DeparturesVia(
    ctx context.Context, stationCode, viaMatch string, window time.Duration,
) ([]domain.Departure, error)
```

`viaMatch` is the raw TO string from the user (trimmed, any case,
possibly truncated by callback_data). Treated as a case-insensitive
substring.

### Interface change
`api.TrainClient` already has `GetTrain`. No interface change. The
viaggiatreno client's `GetTrain` already returns `*domain.Train` with
`.Stops[].StationName`.

### Edge cases
| Case | Handling |
|---|---|
| `GetTrain` returns empty `Stops` | Terminus-only match. |
| Departure has empty `TrainNumber` | Terminus-only match (can't fetch). |
| All `GetTrain` calls timeout | Whole query still returns trains that matched on terminus. User sees a partial, best-effort answer; no error reply. |
| Upstream `GetDepartures` error | Error propagated (handler shows generic reply). |

### Performance
- 10–30 candidate trains per query in a typical 60-min window.
- `GetTrain` is two HTTP calls (FindTrainOrigin + andamentoTreno).
- With 8-way parallelism and a 10-second per-train timeout, worst case
  ~10s wall-clock. Typical case 2–4s.
- No caching in this ADR. Add if empirically warranted.

### What stays
- Parser, handlers, station picker, callback plumbing all unchanged.
- Only the `DeparturesFromTo` → `DeparturesVia` rename + semantic change.
- Format and empty-state messages unchanged (they reference FROM/TO).

## Consequences
- Service tests expand to cover fan-out + failure fallback.
- One more service method (rename) so callers update.
- Query latency goes up; acceptable given the UX gain.
