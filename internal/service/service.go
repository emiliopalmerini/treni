package service

import (
	"context"
	"database/sql"
	"time"

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

// TrainRanking represents a train in rankings
type TrainRanking struct {
	TrainNumber string
	Category    string
	Origin      string
	Destination string
	TripCount   int
	AvgDelay    float64
	MaxDelay    int
	OnTimeRate  float64
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
	station, err := s.api.GetStation(ctx, stationCode)
	if err != nil {
		return nil, err
	}

	return station, nil
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

// GetDelayHistory returns historical delay records for a train
func (s *Service) GetDelayHistory(ctx context.Context, trainNumber string) ([]domain.DelayRecord, error) {
	if s.queries == nil {
		return nil, nil
	}

	records, err := s.queries.GetDelayRecordsByTrain(ctx, trainNumber)
	if err != nil {
		return nil, err
	}

	result := make([]domain.DelayRecord, len(records))
	for i, r := range records {
		result[i] = domain.DelayRecord{
			ID:            r.ID,
			TrainNumber:   r.TrainNumber,
			TrainCategory: nullString(r.TrainCategory),
			Origin:        r.Origin,
			Destination:   r.Destination,
			Date:          r.Date,
			Delay:         int(r.Delay),
			Cancelled:     nullBool(r.Cancelled),
			Source:        nullString(r.Source),
			RecordedAt:    nullTime(r.RecordedAt),
		}
	}

	return result, nil
}

// GetMostDelayedTrains returns the most delayed trains in the given period
func (s *Service) GetMostDelayedTrains(ctx context.Context, days, limit int) ([]TrainRanking, error) {
	if s.queries == nil {
		return nil, nil
	}

	to := time.Now()
	from := to.AddDate(0, 0, -days)

	rows, err := s.queries.GetMostDelayedTrains(ctx, sqlc.GetMostDelayedTrainsParams{
		FromDate:   from,
		ToDate:     to,
		LimitCount: int64(limit),
	})
	if err != nil {
		return nil, err
	}

	result := make([]TrainRanking, len(rows))
	for i, r := range rows {
		result[i] = TrainRanking{
			TrainNumber: r.TrainNumber,
			Category:    nullString(r.TrainCategory),
			Origin:      r.Origin,
			Destination: r.Destination,
			TripCount:   int(r.TripCount),
			AvgDelay:    nullFloat(r.AvgDelay),
			MaxDelay:    interfaceToInt(r.MaxDelay),
		}
	}

	return result, nil
}

// GetMostReliableTrains returns the most reliable trains in the given period
func (s *Service) GetMostReliableTrains(ctx context.Context, days, limit int) ([]TrainRanking, error) {
	if s.queries == nil {
		return nil, nil
	}

	to := time.Now()
	from := to.AddDate(0, 0, -days)

	rows, err := s.queries.GetMostReliableTrains(ctx, sqlc.GetMostReliableTrainsParams{
		FromDate:   from,
		ToDate:     to,
		LimitCount: int64(limit),
	})
	if err != nil {
		return nil, err
	}

	result := make([]TrainRanking, len(rows))
	for i, r := range rows {
		result[i] = TrainRanking{
			TrainNumber: r.TrainNumber,
			Category:    nullString(r.TrainCategory),
			Origin:      r.Origin,
			Destination: r.Destination,
			TripCount:   int(r.TripCount),
			AvgDelay:    nullFloat(r.AvgDelay),
			OnTimeRate:  float64(r.OnTimeRate),
		}
	}

	return result, nil
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

func nullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func nullBool(nb sql.NullBool) bool {
	if nb.Valid {
		return nb.Bool
	}
	return false
}

func nullFloat(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0
}

func nullTime(nt sql.NullTime) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return time.Time{}
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
