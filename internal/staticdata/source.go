package staticdata

import "context"

// DataSource represents a single data source.
type DataSource interface {
	Name() string
	Priority() int // Lower is higher priority
	Available(ctx context.Context) bool
}

// StationSource provides station data from a specific source.
type StationSource interface {
	DataSource
	GetStation(ctx context.Context, id string) (*Station, error)
	SearchStations(ctx context.Context, query string) ([]*Station, error)
	ListAllStations(ctx context.Context) ([]*Station, error)
}
