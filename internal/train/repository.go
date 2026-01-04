package train

import (
	"context"

	"github.com/google/uuid"
)

type TrainRepository interface {
	Create(ctx context.Context, entity *Train) error
	GetByID(ctx context.Context, id uuid.UUID) (*Train, error)
	List(ctx context.Context) ([]*Train, error)
	Update(ctx context.Context, entity *Train) error
	Delete(ctx context.Context, id uuid.UUID) error
}
