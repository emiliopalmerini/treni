package observation

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
)

type VoyageService interface {
	EnsureVoyageForTrain(ctx context.Context, trainNumber int, originID string, departureTime time.Time) (voyageID uuid.UUID, err error)
	UpdateVoyageStop(ctx context.Context, voyageID uuid.UUID, stationID string, arrivalDelay, departureDelay int, platform string) error
}

type Service struct {
	repo          ObservationRepository
	voyageService VoyageService
}

func NewService(repo ObservationRepository, voyageService VoyageService) *Service {
	return &Service{
		repo:          repo,
		voyageService: voyageService,
	}
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
			OriginID:         d.OriginID,
			OriginName:       d.Origin,
			DestinationID:    d.DestinationID,
			DestinationName:  d.Destination,
			ScheduledTime:    time.UnixMilli(d.DepartureTime),
			Delay:            d.Delay,
			Platform:         d.EffectivePlatform(),
			CirculationState: d.CirculationState,
		}
		entities = append(entities, entity)

		// Ensure voyage exists and update stop in background
		go s.ensureAndUpdateVoyage(context.Background(), d.TrainNumber, d.OriginID, time.UnixMilli(d.DepartureTime), stationID, 0, d.Delay, d.EffectivePlatform())
	}

	if err := s.repo.UpsertBatch(ctx, entities); err != nil {
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
			OriginID:         a.OriginID,
			OriginName:       a.Origin,
			DestinationID:    a.DestinationID,
			DestinationName:  a.Destination,
			ScheduledTime:    time.UnixMilli(a.ArrivalTime),
			Delay:            a.Delay,
			Platform:         a.EffectivePlatform(),
			CirculationState: a.CirculationState,
		}
		entities = append(entities, entity)

		// Ensure voyage exists and update stop in background
		go s.ensureAndUpdateVoyage(context.Background(), a.TrainNumber, a.OriginID, time.UnixMilli(a.ArrivalTime), stationID, a.Delay, 0, a.EffectivePlatform())
	}

	if err := s.repo.UpsertBatch(ctx, entities); err != nil {
		log.Printf("failed to record arrivals: %v", err)
	}
}

func (s *Service) ensureAndUpdateVoyage(ctx context.Context, trainNumber int, originID string, departureTime time.Time, stationID string, arrivalDelay, departureDelay int, platform string) {
	if s.voyageService == nil {
		return
	}

	voyageID, err := s.voyageService.EnsureVoyageForTrain(ctx, trainNumber, originID, departureTime)
	if err != nil {
		log.Printf("failed to ensure voyage for train %d: %v", trainNumber, err)
		return
	}

	if err := s.voyageService.UpdateVoyageStop(ctx, voyageID, stationID, arrivalDelay, departureDelay, platform); err != nil {
		log.Printf("failed to update voyage stop for train %d at station %s: %v", trainNumber, stationID, err)
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

func (s *Service) GetDelayVariations(ctx context.Context, observationID string) ([]*DelayVariation, error) {
	return s.repo.GetDelayVariations(ctx, observationID)
}
