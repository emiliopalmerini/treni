package observation

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
)

// Service handles observation business logic.
type Service struct {
	repo     ObservationRepository
	notifier ObservationNotifier
	wg       sync.WaitGroup
	mu       sync.Mutex
	closed   bool
}

// NewService creates a new observation service.
func NewService(repo ObservationRepository, notifier ObservationNotifier) *Service {
	if notifier == nil {
		notifier = &NoopNotifier{}
	}
	return &Service{
		repo:     repo,
		notifier: notifier,
	}
}

// RecordDepartures records train departure observations.
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

		// Notify listeners in background with proper goroutine management
		s.notifyAsync(ObservationEvent{
			TrainNumber:    d.TrainNumber,
			OriginID:       d.OriginID,
			DepartureTime:  time.UnixMilli(d.DepartureTime),
			StationID:      stationID,
			ArrivalDelay:   0,
			DepartureDelay: d.Delay,
			Platform:       d.EffectivePlatform(),
		})
	}

	if err := s.repo.UpsertBatch(ctx, entities); err != nil {
		log.Printf("failed to record departures: %v", err)
	}
}

// RecordArrivals records train arrival observations.
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

		// Notify listeners in background with proper goroutine management
		s.notifyAsync(ObservationEvent{
			TrainNumber:    a.TrainNumber,
			OriginID:       a.OriginID,
			DepartureTime:  time.UnixMilli(a.ArrivalTime),
			StationID:      stationID,
			ArrivalDelay:   a.Delay,
			DepartureDelay: 0,
			Platform:       a.EffectivePlatform(),
		})
	}

	if err := s.repo.UpsertBatch(ctx, entities); err != nil {
		log.Printf("failed to record arrivals: %v", err)
	}
}

// notifyAsync sends an observation event to the notifier in a background goroutine.
// The goroutine is tracked via WaitGroup for graceful shutdown.
func (s *Service) notifyAsync(event ObservationEvent) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.wg.Add(1)
	s.mu.Unlock()

	go func() {
		defer s.wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		s.notifier.OnObservation(ctx, event)
	}()
}

// Shutdown waits for all background goroutines to complete.
// Call this during application shutdown to ensure clean termination.
func (s *Service) Shutdown() {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
	s.wg.Wait()
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
