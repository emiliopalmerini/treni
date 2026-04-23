package service

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/emiliopalmerini/treni/internal/api"
	"github.com/emiliopalmerini/treni/internal/domain"
)

type Service struct {
	api api.TrainClient
	now func() time.Time
}

func New(api api.TrainClient) *Service {
	return NewWithClock(api, time.Now)
}

func NewWithClock(api api.TrainClient, now func() time.Time) *Service {
	return &Service{api: api, now: now}
}

// TrainResult holds real-time train data
type TrainResult struct {
	Train *domain.Train
}

func (s *Service) GetTrain(ctx context.Context, trainNumber string) (*TrainResult, error) {
	train, err := s.api.GetTrain(ctx, trainNumber)
	if err != nil {
		return nil, err
	}
	return &TrainResult{Train: train}, nil
}

func (s *Service) GetStation(ctx context.Context, stationCode string) (*domain.Station, error) {
	return s.api.GetStation(ctx, stationCode)
}

func (s *Service) SearchStations(ctx context.Context, query string) ([]domain.Station, error) {
	return s.api.SearchStation(ctx, query)
}

func (s *Service) Now() time.Time {
	return s.now()
}

// DeparturesFromTo returns departures from stationCode whose terminus
// (Destination) contains toMatch as a case-insensitive substring and whose
// ScheduledTime is within [now, now+window]. Sorted ascending by time.
func (s *Service) DeparturesFromTo(ctx context.Context, stationCode, toMatch string, window time.Duration) ([]domain.Departure, error) {
	now := s.now()
	deps, err := s.api.GetDepartures(ctx, stationCode, now)
	if err != nil {
		return nil, err
	}
	cutoff := now.Add(window)
	toLower := strings.ToLower(toMatch)

	var out []domain.Departure
	for _, d := range deps {
		if !strings.Contains(strings.ToLower(d.Destination), toLower) {
			continue
		}
		if d.ScheduledTime.Before(now) || d.ScheduledTime.After(cutoff) {
			continue
		}
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ScheduledTime.Before(out[j].ScheduledTime)
	})
	return out, nil
}
