package staticdata

import (
	"context"
	"log"
	"sort"
	"time"
)

// CompositeStationProvider orchestrates multiple station sources with fallback.
type CompositeStationProvider struct {
	sources       []StationSource
	stalenessAge  time.Duration
	writebackRepo StationRepository
}

// NewCompositeStationProvider creates a provider with multiple sources.
// Sources are tried in priority order (lower priority number = higher priority).
// The writebackRepo is used to persist API results back to SQLite.
func NewCompositeStationProvider(
	stalenessAge time.Duration,
	writebackRepo StationRepository,
	sources ...StationSource,
) *CompositeStationProvider {
	// Sort by priority (ascending)
	sorted := make([]StationSource, len(sources))
	copy(sorted, sources)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority() < sorted[j].Priority()
	})

	return &CompositeStationProvider{
		sources:       sorted,
		stalenessAge:  stalenessAge,
		writebackRepo: writebackRepo,
	}
}

func (c *CompositeStationProvider) GetStation(ctx context.Context, id string) (*Station, *DataFreshness, error) {
	var lastErr error

	for _, source := range c.sources {
		if !source.Available(ctx) {
			continue
		}

		station, err := source.GetStation(ctx, id)
		if err == nil && station != nil {
			freshness := c.buildFreshness(source.Name(), station.UpdatedAt)

			// Write-through: if from API, persist to SQLite
			if source.Name() == "api" && c.writebackRepo != nil {
				go func(s *Station) {
					if err := c.writebackRepo.Upsert(context.Background(), s); err != nil {
						log.Printf("failed to write-through station %s: %v", s.ID, err)
					}
				}(station)
			}

			return station, freshness, nil
		}

		if err != nil && err != ErrNotFound {
			lastErr = err
			log.Printf("source %s failed for GetStation(%s): %v", source.Name(), id, err)
		}
	}

	if lastErr != nil {
		return nil, nil, lastErr
	}
	return nil, nil, ErrNotFound
}

func (c *CompositeStationProvider) SearchStations(ctx context.Context, query string) ([]*Station, *DataFreshness, error) {
	var lastErr error

	for _, source := range c.sources {
		if !source.Available(ctx) {
			continue
		}

		stations, err := source.SearchStations(ctx, query)
		if err == nil && len(stations) > 0 {
			var latestUpdate time.Time
			for _, s := range stations {
				if s.UpdatedAt.After(latestUpdate) {
					latestUpdate = s.UpdatedAt
				}
			}
			freshness := c.buildFreshness(source.Name(), latestUpdate)

			// Write-through: if from API, persist to SQLite
			if source.Name() == "api" && c.writebackRepo != nil {
				go func(stns []*Station) {
					for _, s := range stns {
						if err := c.writebackRepo.Upsert(context.Background(), s); err != nil {
							log.Printf("failed to write-through station %s: %v", s.ID, err)
						}
					}
				}(stations)
			}

			return stations, freshness, nil
		}

		if err != nil && err != ErrNotFound {
			lastErr = err
			log.Printf("source %s failed for SearchStations(%s): %v", source.Name(), query, err)
		}
	}

	if lastErr != nil {
		return nil, nil, lastErr
	}
	return []*Station{}, &DataFreshness{Source: "none"}, nil
}

func (c *CompositeStationProvider) ListAllStations(ctx context.Context) ([]*Station, *DataFreshness, error) {
	var lastErr error

	for _, source := range c.sources {
		if !source.Available(ctx) {
			continue
		}

		stations, err := source.ListAllStations(ctx)
		if err == nil && len(stations) > 0 {
			var latestUpdate time.Time
			for _, s := range stations {
				if s.UpdatedAt.After(latestUpdate) {
					latestUpdate = s.UpdatedAt
				}
			}
			freshness := c.buildFreshness(source.Name(), latestUpdate)

			return stations, freshness, nil
		}

		if err != nil && err != ErrNotFound {
			lastErr = err
			log.Printf("source %s failed for ListAllStations: %v", source.Name(), err)
		}
	}

	if lastErr != nil {
		return nil, nil, lastErr
	}
	return []*Station{}, &DataFreshness{Source: "none"}, nil
}

func (c *CompositeStationProvider) buildFreshness(source string, updatedAt time.Time) *DataFreshness {
	return &DataFreshness{
		Source:      source,
		LastUpdated: updatedAt,
		IsStale:     time.Since(updatedAt) > c.stalenessAge,
	}
}
