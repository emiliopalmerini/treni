package staticdata

import (
	"context"
	"time"

	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
)

// APIStationSource provides station data from the ViaggiaTreno API.
type APIStationSource struct {
	client viaggiatreno.Client
}

// NewAPIStationSource creates a new API station source.
func NewAPIStationSource(client viaggiatreno.Client) *APIStationSource {
	return &APIStationSource{client: client}
}

func (a *APIStationSource) Name() string  { return "api" }
func (a *APIStationSource) Priority() int { return 10 } // Lower priority than SQLite

func (a *APIStationSource) Available(ctx context.Context) bool {
	return true
}

func (a *APIStationSource) GetStation(ctx context.Context, id string) (*Station, error) {
	results, err := a.client.AutocompletaStazione(ctx, id)
	if err != nil {
		return nil, ErrSourceUnavailable
	}

	for _, r := range results {
		if r.ID == id {
			return &Station{
				ID:        r.ID,
				Name:      r.Name,
				UpdatedAt: time.Now(),
			}, nil
		}
	}
	return nil, ErrNotFound
}

func (a *APIStationSource) SearchStations(ctx context.Context, query string) ([]*Station, error) {
	results, err := a.client.AutocompletaStazione(ctx, query)
	if err != nil {
		return nil, ErrSourceUnavailable
	}

	stations := make([]*Station, len(results))
	now := time.Now()
	for i, r := range results {
		stations[i] = &Station{
			ID:        r.ID,
			Name:      r.Name,
			UpdatedAt: now,
		}
	}
	return stations, nil
}

func (a *APIStationSource) ListAllStations(ctx context.Context) ([]*Station, error) {
	var allStations []*Station
	now := time.Now()

	for region := 1; region <= 22; region++ {
		stations, err := a.client.ElencoStazioni(ctx, region)
		if err != nil {
			continue
		}

		for _, rs := range stations {
			allStations = append(allStations, &Station{
				ID:        rs.ID,
				Name:      rs.Name,
				Region:    rs.Region,
				Latitude:  rs.Latitude,
				Longitude: rs.Longitude,
				UpdatedAt: now,
			})
		}
	}

	if len(allStations) == 0 {
		return nil, ErrSourceUnavailable
	}

	return allStations, nil
}
