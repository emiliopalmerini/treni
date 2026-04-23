# ADR-010: Pivot from `<LINE> <STATION>` to `<FROM> > <TO>`

## Status
Accepted (supersedes ADR-007 and ADR-009)

## Context
ADR-007 shipped `S9 Desio`-style queries that filter a station's
departure board by `TrainCategory`. Smoke-testing against live
ViaggiaTreno data shows `TrainCategory` values don't match the
intuitive line codes users know (`S9`, `RV`, etc.) reliably enough.
The UX target is still "get to my train fast"; the fix is to reframe
the query from line-centric to origin/destination-centric.

ADR-009's direction picker was necessary only because the line-based
query returned trains heading both ways. With an explicit destination
filter that collapses.

## Decision

### Grammar
```
QUERY := FROM WS ">" WS TO
FROM  := free text (may be multi-word)
TO    := free text (may be multi-word)
```

Example queries:
- `Desio > Milano`
- `Milano Centrale > Brescia`
- `Desio > Saronno`

Whitespace around the `>` is optional. Messages that don't contain a
`>` fall back to the format hint.

### Resolution flow
1. Split the message on the first `>` into `(from, to)`. Trim both.
   Reject if either side is empty.
2. `SearchStations(from)` → 0/1/many handling unchanged from ADR-008:
   - 0 → `"No station found for 'X'."`
   - 1 → proceed with that station.
   - 2–5 → inline keyboard picker. Callback data carries `to` so the
     pick resumes the query.
3. Fetch departures at the resolved station.
4. Filter where `strings.Contains(strings.ToLower(departure.Destination), strings.ToLower(to))`.
5. Filter to the time window (60 min, same as before).
6. Sort ascending. Format and reply.

### Callback data format
The station picker's callback now carries the TO string:
```
q:<STATION_CODE>:<TO>
```

TO is truncated to ~40 bytes to stay comfortably under Telegram's
64-byte callback_data limit. Callers use substring match, so
truncation only loses specificity — a truncated `"Milano C"` still
matches any `"Milano Centrale"` or `"Milano Cadorna"` terminus.

### What this replaces / removes
- `ADR-007` line filter (`TrainCategory` equality): removed.
- `ADR-009` direction picker: removed. The TO side of the query encodes
  direction.
- Callback prefix `d:` and `NewDirectionHandler`: deleted with their tests.
- `QueryDepartures` service method: replaced by `DeparturesFromTo`.

### What stays
- ADR-006 bot skeleton + whitelist.
- ADR-008 station picker, with callback data updated to carry TO.
- `.env` loader and `TIME_OVERRIDE`.
- 60-min window (configurable window is still a future ADR).

### Reply format (unchanged except the header)
```
Desio → Milano (next 60 min, now 08:00)

08:14         bin 2   RV  → Milano Centrale
08:29 +1'     bin 2   S   → Milano Porta Garibaldi
```

Train category still appears per row so the user can eyeball which
service they're taking, but it's informational, not a query dimension.

### Edge cases
| Case | Handling |
|---|---|
| No `>` in message | format hint. |
| Empty FROM or TO | format hint. |
| TO matches zero departures in window | `"No trains from <From> toward '<to>' in the next 60 min (now HH:MM)."` |
| TO matches departures whose destination *contains* an unrelated substring (e.g. TO="a" matches everything) | Accepted. The user can refine the query. |

### Future work
- Match against the train's full stop list (not just terminus), so
  `Desio > Milano` also surfaces trains that terminate at Albairate
  but stop at Milano. Requires a `GetTrain` call per candidate, or
  the ViaggiaTreno `soluzioniViaggioNew` endpoint.
- Configurable window (`Desio > Milano 2h`).

## Consequences
- Parser, service, handlers all change.
- Test files reshuffled: direction tests deleted; query tests updated
  to the new grammar; new tests cover `FROM > TO` with substring match.
- The existing ADR-007 and ADR-009 docs stay as historical record;
  this ADR marks them superseded at the top.
