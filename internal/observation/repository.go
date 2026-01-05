package observation

import (
	"context"
)

type ObservationRepository interface {
	UpsertBatch(ctx context.Context, entities []*TrainObservation) error

	GetGlobalStats(ctx context.Context) (*GlobalStats, error)
	GetStatsByCategory(ctx context.Context) ([]*CategoryStats, error)
	GetStatsByStation(ctx context.Context, stationID string) (*StationStats, error)
	GetStatsByTrain(ctx context.Context, trainNumber int) (*TrainStats, error)

	GetWorstTrains(ctx context.Context, limit int) ([]*TrainStats, error)
	GetWorstStations(ctx context.Context, limit int) ([]*StationStats, error)

	GetRecentObservations(ctx context.Context, limit int) ([]*TrainObservation, error)
	GetRecentByStation(ctx context.Context, stationID string, limit int) ([]*TrainObservation, error)

	GetDelayVariations(ctx context.Context, observationID string) ([]*DelayVariation, error)
}
