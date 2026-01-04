package journey

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
)

type Service struct {
	repo   JourneyRepository
	client viaggiatreno.Client
}

func NewService(repo JourneyRepository, client viaggiatreno.Client) *Service {
	return &Service{repo: repo, client: client}
}

func (s *Service) Create(ctx context.Context, entity *Journey) error {
	entity.ID = uuid.New()
	entity.RecordedAt = time.Now()
	return s.repo.Create(ctx, entity)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Journey, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*Journey, error) {
	return s.repo.List(ctx)
}

func (s *Service) ListByTrain(ctx context.Context, trainNumber int) ([]*Journey, error) {
	return s.repo.ListByTrain(ctx, trainNumber)
}

func (s *Service) ListByDateRange(ctx context.Context, from, to time.Time) ([]*Journey, error) {
	return s.repo.ListByDateRange(ctx, from, to)
}

func (s *Service) Update(ctx context.Context, entity *Journey) error {
	return s.repo.Update(ctx, entity)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) GetStops(ctx context.Context, journeyID uuid.UUID) ([]*JourneyStop, error) {
	return s.repo.GetStopsByJourney(ctx, journeyID)
}

// RecordFromAPI fetches train status from ViaggiaTreno and records it.
func (s *Service) RecordFromAPI(ctx context.Context, originID, trainNumber string, departureTS int64) (*Journey, error) {
	status, err := s.client.AndamentoTreno(ctx, originID, trainNumber, departureTS)
	if err != nil {
		return nil, err
	}
	if status == nil {
		return nil, nil
	}

	journey := &Journey{
		ID:                 uuid.New(),
		TrainNumber:        status.TrainNumber,
		OriginID:           status.OriginID,
		OriginName:         status.Origin,
		DestinationID:      status.DestinationID,
		DestinationName:    status.Destination,
		ScheduledDeparture: time.UnixMilli(status.DepartureTime),
		Delay:              status.Delay,
		RecordedAt:         time.Now(),
	}

	if err := s.repo.Create(ctx, journey); err != nil {
		return nil, err
	}

	for _, stop := range status.Stops {
		js := &JourneyStop{
			ID:             uuid.New(),
			JourneyID:      journey.ID,
			StationID:      stop.StationID,
			StationName:    stop.StationName,
			ArrivalDelay:   stop.ArrivalDelay,
			DepartureDelay: stop.DepartureDelay,
			Platform:       stop.Platform,
		}
		if stop.ScheduledArrival > 0 {
			t := time.UnixMilli(stop.ScheduledArrival)
			js.ScheduledArrival = t
		}
		if stop.ScheduledDeparture > 0 {
			t := time.UnixMilli(stop.ScheduledDeparture)
			js.ScheduledDeparture = t
		}
		if stop.ActualArrival > 0 {
			t := time.UnixMilli(stop.ActualArrival)
			js.ActualArrival = t
		}
		if stop.ActualDeparture > 0 {
			t := time.UnixMilli(stop.ActualDeparture)
			js.ActualDeparture = t
		}
		if err := s.repo.CreateStop(ctx, js); err != nil {
			return nil, err
		}
	}

	return journey, nil
}
