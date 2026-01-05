package staticdata

import (
	"context"
	"time"
)

// StationProvider provides station data from the best available source.
type StationProvider interface {
	GetStation(ctx context.Context, id string) (*Station, *DataFreshness, error)
	SearchStations(ctx context.Context, query string) ([]*Station, *DataFreshness, error)
	ListAllStations(ctx context.Context) ([]*Station, *DataFreshness, error)
}

// StationRepository defines storage operations for stations.
type StationRepository interface {
	GetByID(ctx context.Context, id string) (*Station, error)
	Search(ctx context.Context, query string) ([]*Station, error)
	List(ctx context.Context) ([]*Station, error)
	Count(ctx context.Context) (int, error)
	Upsert(ctx context.Context, entity *Station) error
}

// ImportMetadataRepository tracks import status.
type ImportMetadataRepository interface {
	Get(ctx context.Context, entityType string) (*ImportMetadata, error)
	Upsert(ctx context.Context, meta *ImportMetadata) error
	ShouldRefresh(ctx context.Context, entityType string, maxAge time.Duration) (bool, error)
}
