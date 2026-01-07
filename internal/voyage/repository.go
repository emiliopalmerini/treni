package voyage

import (
	"context"

	"github.com/google/uuid"
)

type VoyageRepository interface {
	Create(ctx context.Context, voyage *Voyage) error
	GetByID(ctx context.Context, id uuid.UUID) (*Voyage, error)
	FindByKey(ctx context.Context, trainNumber int, originID string, scheduledDate string) (*Voyage, error)
	Update(ctx context.Context, voyage *Voyage) error

	CreateStop(ctx context.Context, stop *VoyageStop) error
	CreateStopsBatch(ctx context.Context, stops []*VoyageStop) error
	UpdateStop(ctx context.Context, stop *VoyageStop) error
	GetStopsByVoyage(ctx context.Context, voyageID uuid.UUID) ([]*VoyageStop, error)
	FindStopByStation(ctx context.Context, voyageID uuid.UUID, stationID string) (*VoyageStop, error)

	GetVoyageWithStops(ctx context.Context, voyageID uuid.UUID) (*VoyageWithStops, error)
	GetVoyagesByTrain(ctx context.Context, trainNumber int, limit int) ([]*Voyage, error)
	GetVoyagesByDate(ctx context.Context, date string, limit int) ([]*Voyage, error)
	GetRecentVoyages(ctx context.Context, limit int) ([]*Voyage, error)
}
