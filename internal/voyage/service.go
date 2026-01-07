package voyage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
)

type Service struct {
	repo        VoyageRepository
	trainClient viaggiatreno.Client
}

func NewService(repo VoyageRepository, trainClient viaggiatreno.Client) *Service {
	return &Service{
		repo:        repo,
		trainClient: trainClient,
	}
}

func (s *Service) EnsureVoyageForTrain(ctx context.Context, trainNumber int, originID string, departureTime time.Time) (uuid.UUID, error) {
	scheduledDate := departureTime.Format("2006-01-02")

	existing, err := s.repo.FindByKey(ctx, trainNumber, originID, scheduledDate)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to find voyage: %w", err)
	}

	if existing != nil {
		return existing.ID, nil
	}

	trainStatus, err := s.trainClient.AndamentoTreno(ctx, originID, fmt.Sprintf("%d", trainNumber), departureTime.UnixMilli())
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to fetch train status: %w", err)
	}

	voyage, err := s.createFromTrainStatus(ctx, trainStatus, scheduledDate)
	if err != nil {
		return uuid.Nil, err
	}
	return voyage.ID, nil
}

func (s *Service) UpdateVoyageStop(ctx context.Context, voyageID uuid.UUID, stationID string, arrivalDelay, departureDelay int, platform string) error {
	stop, err := s.repo.FindStopByStation(ctx, voyageID, stationID)
	if err != nil {
		return fmt.Errorf("failed to find stop: %w", err)
	}

	if stop == nil {
		return fmt.Errorf("stop not found for station %s in voyage %s", stationID, voyageID)
	}

	now := time.Now()
	stop.ArrivalDelay = arrivalDelay
	stop.DepartureDelay = departureDelay
	stop.Platform = platform
	stop.LastObservationAt = &now

	if stop.ScheduledArrival != nil {
		actual := stop.ScheduledArrival.Add(time.Duration(arrivalDelay) * time.Minute)
		stop.ActualArrival = &actual
	}
	if stop.ScheduledDeparture != nil {
		actual := stop.ScheduledDeparture.Add(time.Duration(departureDelay) * time.Minute)
		stop.ActualDeparture = &actual
	}

	if err := s.repo.UpdateStop(ctx, stop); err != nil {
		return fmt.Errorf("failed to update stop: %w", err)
	}

	voyage, err := s.repo.GetByID(ctx, voyageID)
	if err != nil {
		return fmt.Errorf("failed to get voyage: %w", err)
	}
	voyage.UpdatedAt = now

	return s.repo.Update(ctx, voyage)
}

func (s *Service) GetVoyageWithStops(ctx context.Context, voyageID uuid.UUID) (*VoyageWithStops, error) {
	return s.repo.GetVoyageWithStops(ctx, voyageID)
}

func (s *Service) GetVoyagesByTrain(ctx context.Context, trainNumber int, limit int) ([]*Voyage, error) {
	return s.repo.GetVoyagesByTrain(ctx, trainNumber, limit)
}

func (s *Service) GetRecentVoyages(ctx context.Context, limit int) ([]*Voyage, error) {
	return s.repo.GetRecentVoyages(ctx, limit)
}

func (s *Service) createFromTrainStatus(ctx context.Context, status *viaggiatreno.TrainStatus, scheduledDate string) (*Voyage, error) {
	now := time.Now()
	voyageID := uuid.New()

	voyage := &Voyage{
		ID:                 voyageID,
		TrainNumber:        status.TrainNumber,
		TrainCategory:      status.Category,
		OriginID:           status.OriginID,
		OriginName:         status.Origin,
		DestinationID:      status.DestinationID,
		DestinationName:    status.Destination,
		ScheduledDate:      scheduledDate,
		ScheduledDeparture: time.UnixMilli(status.DepartureTime),
		CirculationState:   status.CirculationState,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.repo.Create(ctx, voyage); err != nil {
		return nil, fmt.Errorf("failed to create voyage: %w", err)
	}

	stops := make([]*VoyageStop, 0, len(status.Stops))
	for i, apiStop := range status.Stops {
		stop := &VoyageStop{
			ID:             uuid.New(),
			VoyageID:       voyageID,
			StationID:      apiStop.StationID,
			StationName:    apiStop.StationName,
			StopSequence:   i + 1,
			StopType:       apiStop.StopType,
			ArrivalDelay:   apiStop.ArrivalDelay,
			DepartureDelay: apiStop.DepartureDelay,
			Platform:       apiStop.EffectivePlatform(),
			IsSuppressed:   apiStop.ActualStopType == 3,
		}

		if apiStop.ScheduledArrival > 0 {
			t := time.UnixMilli(apiStop.ScheduledArrival)
			stop.ScheduledArrival = &t
		}
		if apiStop.ScheduledDeparture > 0 {
			t := time.UnixMilli(apiStop.ScheduledDeparture)
			stop.ScheduledDeparture = &t
		}
		if apiStop.ActualArrival > 0 {
			t := time.UnixMilli(apiStop.ActualArrival)
			stop.ActualArrival = &t
		}
		if apiStop.ActualDeparture > 0 {
			t := time.UnixMilli(apiStop.ActualDeparture)
			stop.ActualDeparture = &t
		}

		stops = append(stops, stop)
	}

	if err := s.repo.CreateStopsBatch(ctx, stops); err != nil {
		return nil, fmt.Errorf("failed to create stops: %w", err)
	}

	log.Printf("Created voyage %s for train %d with %d stops", voyageID, status.TrainNumber, len(stops))
	return voyage, nil
}
