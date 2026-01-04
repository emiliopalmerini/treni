package observation

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
)

type Service struct {
	repo ObservationRepository
}

func NewService(repo ObservationRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RecordDepartures(ctx context.Context, stationID, stationName string, departures []viaggiatreno.Departure) {
	if len(departures) == 0 {
		return
	}

	now := time.Now()
	entities := make([]*TrainObservation, 0, len(departures))

	for _, d := range departures {
		entity := &TrainObservation{
			ID:               uuid.New(),
			ObservedAt:       now,
			StationID:        stationID,
			StationName:      stationName,
			ObservationType:  ObservationTypeDeparture,
			TrainNumber:      d.TrainNumber,
			TrainCategory:    d.CategoryDesc,
			OriginName:       d.Origin,
			DestinationName:  d.Destination,
			ScheduledTime:    time.UnixMilli(d.DepartureTime),
			Delay:            d.Delay,
			Platform:         d.EffectivePlatform(),
			CirculationState: d.CirculationState,
		}
		entities = append(entities, entity)
	}

	if err := s.repo.CreateBatch(ctx, entities); err != nil {
		log.Printf("failed to record departures: %v", err)
	}
}

func (s *Service) RecordArrivals(ctx context.Context, stationID, stationName string, arrivals []viaggiatreno.Arrival) {
	if len(arrivals) == 0 {
		return
	}

	now := time.Now()
	entities := make([]*TrainObservation, 0, len(arrivals))

	for _, a := range arrivals {
		entity := &TrainObservation{
			ID:               uuid.New(),
			ObservedAt:       now,
			StationID:        stationID,
			StationName:      stationName,
			ObservationType:  ObservationTypeArrival,
			TrainNumber:      a.TrainNumber,
			TrainCategory:    a.CategoryDesc,
			OriginName:       a.Origin,
			DestinationName:  a.Destination,
			ScheduledTime:    time.UnixMilli(a.ArrivalTime),
			Delay:            a.Delay,
			Platform:         a.EffectivePlatform(),
			CirculationState: a.CirculationState,
		}
		entities = append(entities, entity)
	}

	if err := s.repo.CreateBatch(ctx, entities); err != nil {
		log.Printf("failed to record arrivals: %v", err)
	}
}

func (s *Service) GetGlobalStats(ctx context.Context) (*GlobalStats, error) {
	return s.repo.GetGlobalStats(ctx)
}

func (s *Service) GetStatsByCategory(ctx context.Context) ([]*CategoryStats, error) {
	return s.repo.GetStatsByCategory(ctx)
}

func (s *Service) GetStatsByStation(ctx context.Context, stationID string) (*StationStats, error) {
	return s.repo.GetStatsByStation(ctx, stationID)
}

func (s *Service) GetStatsByTrain(ctx context.Context, trainNumber int) (*TrainStats, error) {
	return s.repo.GetStatsByTrain(ctx, trainNumber)
}

func (s *Service) GetWorstTrains(ctx context.Context, limit int) ([]*TrainStats, error) {
	return s.repo.GetWorstTrains(ctx, limit)
}

func (s *Service) GetWorstStations(ctx context.Context, limit int) ([]*StationStats, error) {
	return s.repo.GetWorstStations(ctx, limit)
}

func (s *Service) GetRecentObservations(ctx context.Context, limit int) ([]*TrainObservation, error) {
	return s.repo.GetRecentObservations(ctx, limit)
}

func (s *Service) GetRecentByStation(ctx context.Context, stationID string, limit int) ([]*TrainObservation, error) {
	return s.repo.GetRecentByStation(ctx, stationID, limit)
}
