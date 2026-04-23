package service

import (
	"context"
	"sort"
	"strings"
	"sync"
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

const (
	viaWorkers       = 8
	perTrainTimeout  = 10 * time.Second
)

// DeparturesVia returns departures from stationCode whose train either
// terminates at a station containing viaMatch OR passes through any stop
// containing viaMatch (case-insensitive substring). Departures with a
// GetTrain failure fall back to terminus-only matching.
func (s *Service) DeparturesVia(ctx context.Context, stationCode, viaMatch string, window time.Duration) ([]domain.Departure, error) {
	now := s.now()
	deps, err := s.api.GetDepartures(ctx, stationCode, now)
	if err != nil {
		return nil, err
	}
	cutoff := now.Add(window)
	needle := strings.ToLower(viaMatch)

	// Step 1: candidates within the window.
	var candidates []domain.Departure
	for _, d := range deps {
		if d.ScheduledTime.Before(now) || d.ScheduledTime.After(cutoff) {
			continue
		}
		candidates = append(candidates, d)
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	// Step 2: fan out GetTrain, decide inclusion for each.
	type decision struct {
		index   int
		include bool
	}
	decisions := make(chan decision, len(candidates))

	sem := make(chan struct{}, viaWorkers)
	var wg sync.WaitGroup
	for i, d := range candidates {
		wg.Add(1)
		go func(i int, d domain.Departure) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			decisions <- decision{index: i, include: trainMatchesVia(ctx, s.api, d, needle)}
		}(i, d)
	}
	wg.Wait()
	close(decisions)

	include := make([]bool, len(candidates))
	for dec := range decisions {
		include[dec.index] = dec.include
	}

	out := make([]domain.Departure, 0, len(candidates))
	for i, d := range candidates {
		if include[i] {
			out = append(out, d)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ScheduledTime.Before(out[j].ScheduledTime)
	})
	return out, nil
}

func trainMatchesVia(ctx context.Context, client api.TrainClient, d domain.Departure, needle string) bool {
	terminusMatches := strings.Contains(strings.ToLower(d.Destination), needle)

	if d.TrainNumber == "" {
		return terminusMatches
	}

	tctx, cancel := context.WithTimeout(ctx, perTrainTimeout)
	defer cancel()
	train, err := client.GetTrain(tctx, d.TrainNumber)
	if err != nil || train == nil {
		return terminusMatches
	}
	for _, stop := range train.Stops {
		if strings.Contains(strings.ToLower(stop.StationName), needle) {
			return true
		}
	}
	return terminusMatches
}
