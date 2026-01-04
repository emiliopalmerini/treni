package watchlist

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
)

type Service struct {
	repo   WatchlistRepository
	client *viaggiatreno.Client
}

func NewService(repo WatchlistRepository, client *viaggiatreno.Client) *Service {
	return &Service{repo: repo, client: client}
}

func (s *Service) Create(ctx context.Context, entity *WatchedTrain) error {
	entity.ID = uuid.New()
	entity.CreatedAt = time.Now()
	entity.Active = true
	return s.repo.Create(ctx, entity)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*WatchedTrain, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*WatchedTrain, error) {
	return s.repo.List(ctx)
}

func (s *Service) ListActive(ctx context.Context) ([]*WatchedTrain, error) {
	return s.repo.ListActive(ctx)
}

func (s *Service) Update(ctx context.Context, entity *WatchedTrain) error {
	return s.repo.Update(ctx, entity)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	entity, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	entity.Active = active
	return s.repo.Update(ctx, entity)
}

func (s *Service) GetCheckHistory(ctx context.Context, watchedID uuid.UUID) ([]*TrainCheck, error) {
	return s.repo.GetChecksByWatched(ctx, watchedID)
}

func (s *Service) GetRecentChecks(ctx context.Context, limit int) ([]*TrainCheck, error) {
	return s.repo.GetRecentChecks(ctx, limit)
}

// CheckTrain checks the current status of a watched train and records it.
func (s *Service) CheckTrain(ctx context.Context, watchedID uuid.UUID) (*TrainCheck, error) {
	watched, err := s.repo.GetByID(ctx, watchedID)
	if err != nil {
		return nil, err
	}

	matches, err := s.client.CercaNumeroTreno(ctx, fmt.Sprintf("%d", watched.TrainNumber))
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		check := &TrainCheck{
			ID:          uuid.New(),
			WatchedID:   watchedID,
			TrainNumber: watched.TrainNumber,
			Status:      "not_found",
			CheckedAt:   time.Now(),
		}
		if err := s.repo.CreateCheck(ctx, check); err != nil {
			return nil, err
		}
		return check, nil
	}

	match := matches[0]
	status, err := s.client.AndamentoTreno(ctx, match.OriginID, match.Number, match.DepartureTS)
	if err != nil {
		return nil, err
	}

	check := &TrainCheck{
		ID:          uuid.New(),
		WatchedID:   watchedID,
		TrainNumber: watched.TrainNumber,
		CheckedAt:   time.Now(),
	}

	if status == nil {
		check.Status = "unavailable"
	} else {
		check.Delay = status.Delay
		if status.Delay > 0 {
			check.Status = "delayed"
		} else {
			check.Status = "on_time"
		}
	}

	if err := s.repo.CreateCheck(ctx, check); err != nil {
		return nil, err
	}

	return check, nil
}

// CheckAllActive checks all active watched trains.
func (s *Service) CheckAllActive(ctx context.Context) ([]*TrainCheck, error) {
	active, err := s.repo.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	var checks []*TrainCheck
	for _, watched := range active {
		check, err := s.CheckTrain(ctx, watched.ID)
		if err != nil {
			continue
		}
		checks = append(checks, check)
	}

	return checks, nil
}
