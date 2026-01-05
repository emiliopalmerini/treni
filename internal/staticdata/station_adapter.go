package staticdata

import (
	"context"

	"github.com/emiliopalmerini/treni/internal/station"
)

// StationRepositoryAdapter adapts station.StationRepository to staticdata.StationRepository.
type StationRepositoryAdapter struct {
	repo station.StationRepository
}

// NewStationRepositoryAdapter creates a new adapter.
func NewStationRepositoryAdapter(repo station.StationRepository) *StationRepositoryAdapter {
	return &StationRepositoryAdapter{repo: repo}
}

func (a *StationRepositoryAdapter) GetByID(ctx context.Context, id string) (*Station, error) {
	s, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return stationToStaticData(s), nil
}

func (a *StationRepositoryAdapter) Search(ctx context.Context, query string) ([]*Station, error) {
	stations, err := a.repo.Search(ctx, query)
	if err != nil {
		return nil, err
	}
	return stationsToStaticData(stations), nil
}

func (a *StationRepositoryAdapter) List(ctx context.Context) ([]*Station, error) {
	stations, err := a.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	return stationsToStaticData(stations), nil
}

func (a *StationRepositoryAdapter) Count(ctx context.Context) (int, error) {
	return a.repo.Count(ctx)
}

func (a *StationRepositoryAdapter) Upsert(ctx context.Context, entity *Station) error {
	return a.repo.Upsert(ctx, staticDataToStation(entity))
}

func stationToStaticData(s *station.Station) *Station {
	return &Station{
		ID:        s.ID,
		Name:      s.Name,
		Region:    s.Region,
		Latitude:  s.Latitude,
		Longitude: s.Longitude,
		UpdatedAt: s.UpdatedAt,
	}
}

func stationsToStaticData(stations []*station.Station) []*Station {
	result := make([]*Station, len(stations))
	for i, s := range stations {
		result[i] = stationToStaticData(s)
	}
	return result
}

func staticDataToStation(s *Station) *station.Station {
	return &station.Station{
		ID:        s.ID,
		Name:      s.Name,
		Region:    s.Region,
		Latitude:  s.Latitude,
		Longitude: s.Longitude,
		UpdatedAt: s.UpdatedAt,
	}
}
