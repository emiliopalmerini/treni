package service

import (
	"context"
	"database/sql"

	"github.com/emiliopalmerini/treni/internal/api"
	"github.com/emiliopalmerini/treni/internal/domain"
	"github.com/emiliopalmerini/treni/internal/storage/sqlc"
)

type Service struct {
	api     api.TrainClient
	queries *sqlc.Queries
}

func New(api api.TrainClient, queries *sqlc.Queries) *Service {
	return &Service{
		api:     api,
		queries: queries,
	}
}

// TrainResult combines real-time data with historical stats
type TrainResult struct {
	Train *domain.Train
	Stats *domain.TrainStats
}

// GetTrain returns real-time train data combined with historical stats if available
func (s *Service) GetTrain(ctx context.Context, trainNumber string) (*TrainResult, error) {
	train, err := s.api.GetTrain(ctx, trainNumber)
	if err != nil {
		return nil, err
	}

	result := &TrainResult{Train: train}

	// Try to get historical stats (don't fail if not available)
	if s.queries != nil {
		stats, err := s.queries.GetTrainStats(ctx, trainNumber)
		if err == nil && stats.TotalTrips > 0 {
			result.Stats = mapTrainStats(stats)
		}
	}

	return result, nil
}

// GetStation returns station data with arrivals and departures
func (s *Service) GetStation(ctx context.Context, stationCode string) (*domain.Station, error) {
	return s.api.GetStation(ctx, stationCode)
}

// SearchStations searches for stations by name
func (s *Service) SearchStations(ctx context.Context, query string) ([]domain.Station, error) {
	return s.api.SearchStation(ctx, query)
}

// GetTrainStats returns historical statistics for a train
func (s *Service) GetTrainStats(ctx context.Context, trainNumber string) (*domain.TrainStats, error) {
	if s.queries == nil {
		return nil, nil
	}

	stats, err := s.queries.GetTrainStats(ctx, trainNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return mapTrainStats(stats), nil
}

// Helper functions

func mapTrainStats(s sqlc.GetTrainStatsRow) *domain.TrainStats {
	totalTrips := int(s.TotalTrips)
	onTimeTrips := int(nullFloat(s.OnTimeTrips))

	var onTimeRate float64
	if totalTrips > 0 {
		onTimeRate = float64(onTimeTrips) / float64(totalTrips)
	}

	return &domain.TrainStats{
		TrainNumber:    s.TrainNumber,
		TotalTrips:     totalTrips,
		OnTimeTrips:    onTimeTrips,
		DelayedTrips:   int(nullFloat(s.DelayedTrips)),
		CancelledTrips: int(nullFloat(s.CancelledTrips)),
		AverageDelay:   nullFloat(s.AverageDelay),
		MaxDelay:       interfaceToInt(s.MaxDelay),
		OnTimeRate:     onTimeRate,
	}
}

func nullFloat(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0
}

func interfaceToInt(v interface{}) int {
	switch val := v.(type) {
	case int64:
		return int(val)
	case int:
		return val
	case float64:
		return int(val)
	default:
		return 0
	}
}
