package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/emiliopalmerini/treni/internal/database/sqlc"
	"github.com/emiliopalmerini/treni/internal/observation"
)

type SQLiteRepository struct {
	db *sql.DB
	q  *sqlc.Queries
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db, q: sqlc.New(db)}
}

func (r *SQLiteRepository) Create(ctx context.Context, entity *observation.TrainObservation) error {
	return r.q.CreateObservation(ctx, observationToParams(entity))
}

func (r *SQLiteRepository) CreateBatch(ctx context.Context, entities []*observation.TrainObservation) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := r.q.WithTx(tx)
	for _, entity := range entities {
		if err := q.CreateObservation(ctx, observationToParams(entity)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *SQLiteRepository) GetGlobalStats(ctx context.Context) (*observation.GlobalStats, error) {
	row, err := r.q.GetGlobalStats(ctx)
	if err != nil {
		return nil, err
	}

	stats := &observation.GlobalStats{
		TotalObservations: int(row.TotalObservations),
		AverageDelay:      toFloat64(row.AverageDelay),
		OnTimeCount:       int(deref(row.OnTimeCount)),
		CancelledCount:    int(deref(row.CancelledCount)),
	}

	if stats.TotalObservations > 0 {
		stats.OnTimePercentage = float64(stats.OnTimeCount) / float64(stats.TotalObservations) * 100
	}

	return stats, nil
}

func (r *SQLiteRepository) GetStatsByCategory(ctx context.Context) ([]*observation.CategoryStats, error) {
	rows, err := r.q.GetStatsByCategory(ctx)
	if err != nil {
		return nil, err
	}

	stats := make([]*observation.CategoryStats, len(rows))
	for i, row := range rows {
		stats[i] = &observation.CategoryStats{
			Category:         deref(row.Category),
			ObservationCount: int(row.ObservationCount),
			AverageDelay:     toFloat64(row.AverageDelay),
			OnTimePercentage: float64(row.OnTimePercentage),
		}
	}
	return stats, nil
}

func (r *SQLiteRepository) GetStatsByStation(ctx context.Context, stationID string) (*observation.StationStats, error) {
	row, err := r.q.GetStatsByStation(ctx, stationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &observation.StationStats{
		StationID:        row.StationID,
		StationName:      row.StationName,
		ObservationCount: int(row.ObservationCount),
		AverageDelay:     toFloat64(row.AverageDelay),
		OnTimePercentage: float64(row.OnTimePercentage),
	}, nil
}

func (r *SQLiteRepository) GetStatsByTrain(ctx context.Context, trainNumber int) (*observation.TrainStats, error) {
	row, err := r.q.GetStatsByTrain(ctx, int64(trainNumber))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &observation.TrainStats{
		TrainNumber:      int(row.TrainNumber),
		Category:         deref(row.Category),
		OriginName:       deref(row.OriginName),
		DestinationName:  deref(row.DestinationName),
		ObservationCount: int(row.ObservationCount),
		AverageDelay:     toFloat64(row.AverageDelay),
		MaxDelay:         toInt(row.MaxDelay),
		OnTimePercentage: float64(row.OnTimePercentage),
	}, nil
}

func (r *SQLiteRepository) GetWorstTrains(ctx context.Context, limit int) ([]*observation.TrainStats, error) {
	rows, err := r.q.GetWorstTrains(ctx, int64(limit))
	if err != nil {
		return nil, err
	}

	stats := make([]*observation.TrainStats, len(rows))
	for i, row := range rows {
		stats[i] = &observation.TrainStats{
			TrainNumber:      int(row.TrainNumber),
			Category:         deref(row.Category),
			OriginName:       deref(row.OriginName),
			DestinationName:  deref(row.DestinationName),
			ObservationCount: int(row.ObservationCount),
			AverageDelay:     toFloat64(row.AverageDelay),
			MaxDelay:         toInt(row.MaxDelay),
			OnTimePercentage: float64(row.OnTimePercentage),
		}
	}
	return stats, nil
}

func (r *SQLiteRepository) GetWorstStations(ctx context.Context, limit int) ([]*observation.StationStats, error) {
	rows, err := r.q.GetWorstStations(ctx, int64(limit))
	if err != nil {
		return nil, err
	}

	stats := make([]*observation.StationStats, len(rows))
	for i, row := range rows {
		stats[i] = &observation.StationStats{
			StationID:        row.StationID,
			StationName:      row.StationName,
			ObservationCount: int(row.ObservationCount),
			AverageDelay:     toFloat64(row.AverageDelay),
			OnTimePercentage: float64(row.OnTimePercentage),
		}
	}
	return stats, nil
}

func (r *SQLiteRepository) GetRecentObservations(ctx context.Context, limit int) ([]*observation.TrainObservation, error) {
	rows, err := r.q.GetRecentObservations(ctx, int64(limit))
	if err != nil {
		return nil, err
	}
	return sqlcObservationsToObservations(rows), nil
}

func (r *SQLiteRepository) GetRecentByStation(ctx context.Context, stationID string, limit int) ([]*observation.TrainObservation, error) {
	rows, err := r.q.GetRecentObservationsByStation(ctx, sqlc.GetRecentObservationsByStationParams{
		StationID: stationID,
		Limit:     int64(limit),
	})
	if err != nil {
		return nil, err
	}
	return sqlcObservationsToObservations(rows), nil
}

func ptr[T any](v T) *T {
	return &v
}

func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	case int:
		return float64(val)
	default:
		return 0
	}
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int64:
		return int(val)
	case int:
		return val
	default:
		return 0
	}
}

func observationToParams(entity *observation.TrainObservation) sqlc.CreateObservationParams {
	return sqlc.CreateObservationParams{
		ID:               entity.ID.String(),
		ObservedAt:       entity.ObservedAt,
		StationID:        entity.StationID,
		StationName:      entity.StationName,
		ObservationType:  string(entity.ObservationType),
		TrainNumber:      int64(entity.TrainNumber),
		TrainCategory:    strPtr(entity.TrainCategory),
		OriginID:         strPtr(entity.OriginID),
		OriginName:       strPtr(entity.OriginName),
		DestinationID:    strPtr(entity.DestinationID),
		DestinationName:  strPtr(entity.DestinationName),
		ScheduledTime:    timePtr(entity.ScheduledTime),
		Delay:            ptr(int64(entity.Delay)),
		Platform:         strPtr(entity.Platform),
		CirculationState: ptr(int64(entity.CirculationState)),
	}
}

func sqlcObservationsToObservations(rows []sqlc.TrainObservation) []*observation.TrainObservation {
	observations := make([]*observation.TrainObservation, len(rows))
	for i, row := range rows {
		id, _ := uuid.Parse(row.ID)
		observations[i] = &observation.TrainObservation{
			ID:               id,
			ObservedAt:       row.ObservedAt,
			StationID:        row.StationID,
			StationName:      row.StationName,
			ObservationType:  observation.ObservationType(row.ObservationType),
			TrainNumber:      int(row.TrainNumber),
			TrainCategory:    deref(row.TrainCategory),
			OriginID:         deref(row.OriginID),
			OriginName:       deref(row.OriginName),
			DestinationID:    deref(row.DestinationID),
			DestinationName:  deref(row.DestinationName),
			ScheduledTime:    deref(row.ScheduledTime),
			Delay:            int(deref(row.Delay)),
			Platform:         deref(row.Platform),
			CirculationState: int(deref(row.CirculationState)),
		}
	}
	return observations
}
