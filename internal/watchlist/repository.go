package watchlist

import (
	"context"

	"github.com/google/uuid"
)

type WatchlistRepository interface {
	Create(ctx context.Context, entity *WatchedTrain) error
	GetByID(ctx context.Context, id uuid.UUID) (*WatchedTrain, error)
	List(ctx context.Context) ([]*WatchedTrain, error)
	ListActive(ctx context.Context) ([]*WatchedTrain, error)
	Update(ctx context.Context, entity *WatchedTrain) error
	Delete(ctx context.Context, id uuid.UUID) error

	CreateCheck(ctx context.Context, check *TrainCheck) error
	GetChecksByWatched(ctx context.Context, watchedID uuid.UUID) ([]*TrainCheck, error)
	GetRecentChecks(ctx context.Context, limit int) ([]*TrainCheck, error)
}
