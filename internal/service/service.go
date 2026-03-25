package service

import (
	"context"

	"github.com/emiliopalmerini/treni/internal/api"
	"github.com/emiliopalmerini/treni/internal/domain"
)

type Service struct {
	api api.TrainClient
}

func New(api api.TrainClient) *Service {
	return &Service{api: api}
}

// TrainResult holds real-time train data
type TrainResult struct {
	Train *domain.Train
}

// GetTrain returns real-time train data
func (s *Service) GetTrain(ctx context.Context, trainNumber string) (*TrainResult, error) {
	train, err := s.api.GetTrain(ctx, trainNumber)
	if err != nil {
		return nil, err
	}
	return &TrainResult{Train: train}, nil
}

// GetStation returns station data with arrivals and departures
func (s *Service) GetStation(ctx context.Context, stationCode string) (*domain.Station, error) {
	return s.api.GetStation(ctx, stationCode)
}

// SearchStations searches for stations by name
func (s *Service) SearchStations(ctx context.Context, query string) ([]domain.Station, error) {
	return s.api.SearchStation(ctx, query)
}
