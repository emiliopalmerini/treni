package api

import (
	"context"
	"time"

	"github.com/emiliopalmerini/treni/internal/domain"
)

type TrainClient interface {
	GetTrain(ctx context.Context, trainNumber string) (*domain.Train, error)
	GetStation(ctx context.Context, stationCode string) (*domain.Station, error)
	SearchStation(ctx context.Context, query string) ([]domain.Station, error)
	// GetDepartures fetches the departure board from stationCode pivoted
	// around at. The underlying API returns a window of trains around that
	// moment; callers filter further.
	GetDepartures(ctx context.Context, stationCode string, at time.Time) ([]domain.Departure, error)
}
