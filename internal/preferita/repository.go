package preferita

import "context"

type Repository interface {
	List(ctx context.Context) ([]*Preferita, error)
	Add(ctx context.Context, entity *Preferita) error
	Remove(ctx context.Context, stationID string) error
	Exists(ctx context.Context, stationID string) (bool, error)
}
