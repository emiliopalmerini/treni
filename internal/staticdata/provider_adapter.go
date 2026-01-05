package staticdata

import (
	"context"

	"github.com/emiliopalmerini/treni/internal/station"
)

// StationProviderAdapter adapts CompositeStationProvider to station.StationProvider.
type StationProviderAdapter struct {
	provider *CompositeStationProvider
}

// NewStationProviderAdapter creates a new adapter.
func NewStationProviderAdapter(provider *CompositeStationProvider) *StationProviderAdapter {
	return &StationProviderAdapter{provider: provider}
}

func (a *StationProviderAdapter) SearchStations(ctx context.Context, query string) ([]*station.StationData, *station.DataFreshness, error) {
	results, freshness, err := a.provider.SearchStations(ctx, query)
	if err != nil {
		return nil, nil, err
	}

	stations := make([]*station.StationData, len(results))
	for i, r := range results {
		stations[i] = &station.StationData{
			ID:        r.ID,
			Name:      r.Name,
			Region:    r.Region,
			Latitude:  r.Latitude,
			Longitude: r.Longitude,
			UpdatedAt: r.UpdatedAt,
		}
	}

	var f *station.DataFreshness
	if freshness != nil {
		f = &station.DataFreshness{
			Source:      freshness.Source,
			LastUpdated: freshness.LastUpdated,
			IsStale:     freshness.IsStale,
		}
	}

	return stations, f, nil
}
