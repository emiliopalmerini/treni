package api

import (
	"context"

	"github.com/emiliopalmerini/treni/internal/domain"
)

type TrainClient interface {
	GetTrain(ctx context.Context, trainNumber string) (*domain.Train, error)
	GetStation(ctx context.Context, stationCode string) (*domain.Station, error)
	SearchStation(ctx context.Context, query string) ([]domain.Station, error)
}
