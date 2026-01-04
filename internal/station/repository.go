package station

import "context"

type StationRepository interface {
	Create(ctx context.Context, entity *Station) error
	GetByID(ctx context.Context, id string) (*Station, error)
	List(ctx context.Context) ([]*Station, error)
	Search(ctx context.Context, query string) ([]*Station, error)
	ListWithCoordinates(ctx context.Context) ([]*Station, error)
	Count(ctx context.Context) (int, error)
	Update(ctx context.Context, entity *Station) error
	Delete(ctx context.Context, id string) error
	Upsert(ctx context.Context, entity *Station) error
}
