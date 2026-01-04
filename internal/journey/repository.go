package journey

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type JourneyRepository interface {
	Create(ctx context.Context, entity *Journey) error
	GetByID(ctx context.Context, id uuid.UUID) (*Journey, error)
	List(ctx context.Context) ([]*Journey, error)
	ListByTrain(ctx context.Context, trainNumber int) ([]*Journey, error)
	ListByDateRange(ctx context.Context, from, to time.Time) ([]*Journey, error)
	Update(ctx context.Context, entity *Journey) error
	Delete(ctx context.Context, id uuid.UUID) error

	CreateStop(ctx context.Context, stop *JourneyStop) error
	GetStopsByJourney(ctx context.Context, journeyID uuid.UUID) ([]*JourneyStop, error)
}
