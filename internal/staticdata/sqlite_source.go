package staticdata

import (
	"context"
	"database/sql"
)

// SQLiteStationSource provides station data from SQLite.
type SQLiteStationSource struct {
	repo StationRepository
}

// NewSQLiteStationSource creates a new SQLite station source.
func NewSQLiteStationSource(repo StationRepository) *SQLiteStationSource {
	return &SQLiteStationSource{repo: repo}
}

func (s *SQLiteStationSource) Name() string  { return "sqlite" }
func (s *SQLiteStationSource) Priority() int { return 1 } // Highest priority

func (s *SQLiteStationSource) Available(ctx context.Context) bool {
	count, err := s.repo.Count(ctx)
	return err == nil && count > 0
}

func (s *SQLiteStationSource) GetStation(ctx context.Context, id string) (*Station, error) {
	station, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return station, nil
}

func (s *SQLiteStationSource) SearchStations(ctx context.Context, query string) ([]*Station, error) {
	return s.repo.Search(ctx, query)
}

func (s *SQLiteStationSource) ListAllStations(ctx context.Context) ([]*Station, error) {
	return s.repo.List(ctx)
}
