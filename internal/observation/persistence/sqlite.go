package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/emiliopalmerini/treni/internal/database/nullable"
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

func (r *SQLiteRepository) UpsertBatch(ctx context.Context, entities []*observation.TrainObservation) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := r.q.WithTx(tx)
	now := time.Now()

	for _, entity := range entities {
		existing, err := q.GetObservationByKey(ctx, sqlc.GetObservationByKeyParams{
			TrainNumber:     int64(entity.TrainNumber),
			StationID:       entity.StationID,
			ObservationType: string(entity.ObservationType),
			ScheduledDate:   nullable.StrPtr(entity.ScheduledTime.Format("2006-01-02")),
		})

		var previousDelay *int64
		if err == nil {
			previousDelay = existing.Delay
		}

		result, err := q.UpsertObservation(ctx, observationToUpsertParams(entity))
		if err != nil {
			return err
		}

		if previousDelay != nil && *previousDelay != int64(entity.Delay) {
			if err := q.CreateDelayVariation(ctx, sqlc.CreateDelayVariationParams{
				ID:            uuid.New().String(),
				ObservationID: result.ID,
				RecordedAt:    now,
				Delay:         int64(entity.Delay),
			}); err != nil {
				return err
			}
		} else if previousDelay == nil {
			if err := q.CreateDelayVariation(ctx, sqlc.CreateDelayVariationParams{
				ID:            uuid.New().String(),
				ObservationID: result.ID,
				RecordedAt:    now,
				Delay:         int64(entity.Delay),
			}); err != nil {
				return err
			}
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
		AverageDelay:      nullable.ToFloat64(row.AverageDelay),
		OnTimeCount:       int(nullable.Deref(row.OnTimeCount)),
		CancelledCount:    int(nullable.Deref(row.CancelledCount)),
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
			Category:         nullable.Deref(row.Category),
			ObservationCount: int(row.ObservationCount),
			AverageDelay:     nullable.ToFloat64(row.AverageDelay),
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
		AverageDelay:     nullable.ToFloat64(row.AverageDelay),
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
		Category:         nullable.Deref(row.Category),
		OriginID:         nullable.Deref(row.OriginID),
		OriginName:       nullable.Deref(row.OriginName),
		DestinationID:    nullable.Deref(row.DestinationID),
		DestinationName:  nullable.Deref(row.DestinationName),
		ObservationCount: int(row.ObservationCount),
		AverageDelay:     nullable.ToFloat64(row.AverageDelay),
		MaxDelay:         nullable.ToInt(row.MaxDelay),
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
			Category:         nullable.Deref(row.Category),
			OriginID:         nullable.Deref(row.OriginID),
			OriginName:       nullable.Deref(row.OriginName),
			DestinationID:    nullable.Deref(row.DestinationID),
			DestinationName:  nullable.Deref(row.DestinationName),
			ObservationCount: int(row.ObservationCount),
			AverageDelay:     nullable.ToFloat64(row.AverageDelay),
			MaxDelay:         nullable.ToInt(row.MaxDelay),
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
			AverageDelay:     nullable.ToFloat64(row.AverageDelay),
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
	return recentRowsToObservations(rows), nil
}

func (r *SQLiteRepository) GetRecentByStation(ctx context.Context, stationID string, limit int) ([]*observation.TrainObservation, error) {
	rows, err := r.q.GetRecentObservationsByStation(ctx, sqlc.GetRecentObservationsByStationParams{
		StationID: stationID,
		Limit:     int64(limit),
	})
	if err != nil {
		return nil, err
	}
	return recentByStationRowsToObservations(rows), nil
}

func (r *SQLiteRepository) GetDelayVariations(ctx context.Context, observationID string) ([]*observation.DelayVariation, error) {
	rows, err := r.q.GetDelayVariationsByObservation(ctx, observationID)
	if err != nil {
		return nil, err
	}

	variations := make([]*observation.DelayVariation, len(rows))
	for i, row := range rows {
		id, _ := uuid.Parse(row.ID)
		obsID, _ := uuid.Parse(row.ObservationID)
		variations[i] = &observation.DelayVariation{
			ID:            id,
			ObservationID: obsID,
			RecordedAt:    row.RecordedAt,
			Delay:         int(row.Delay),
		}
	}
	return variations, nil
}

func observationToUpsertParams(entity *observation.TrainObservation) sqlc.UpsertObservationParams {
	return sqlc.UpsertObservationParams{
		ID:               entity.ID.String(),
		ObservedAt:       entity.ObservedAt,
		StationID:        entity.StationID,
		StationName:      entity.StationName,
		ObservationType:  string(entity.ObservationType),
		TrainNumber:      int64(entity.TrainNumber),
		TrainCategory:    nullable.StrPtr(entity.TrainCategory),
		OriginID:         nullable.StrPtr(entity.OriginID),
		OriginName:       nullable.StrPtr(entity.OriginName),
		DestinationID:    nullable.StrPtr(entity.DestinationID),
		DestinationName:  nullable.StrPtr(entity.DestinationName),
		ScheduledTime:    nullable.TimePtrFromValue(entity.ScheduledTime),
		ScheduledDate:    nullable.StrPtr(entity.ScheduledTime.Format("2006-01-02")),
		Delay:            nullable.Ptr(int64(entity.Delay)),
		Platform:         nullable.StrPtr(entity.Platform),
		CirculationState: nullable.Ptr(int64(entity.CirculationState)),
	}
}

func recentRowsToObservations(rows []sqlc.GetRecentObservationsRow) []*observation.TrainObservation {
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
			TrainCategory:    nullable.Deref(row.TrainCategory),
			OriginID:         nullable.Deref(row.OriginID),
			OriginName:       nullable.Deref(row.OriginName),
			DestinationID:    nullable.Deref(row.DestinationID),
			DestinationName:  nullable.Deref(row.DestinationName),
			ScheduledTime:    nullable.Deref(row.ScheduledTime),
			Delay:            int(nullable.Deref(row.Delay)),
			Platform:         nullable.Deref(row.Platform),
			CirculationState: int(nullable.Deref(row.CirculationState)),
		}
	}
	return observations
}

func recentByStationRowsToObservations(rows []sqlc.GetRecentObservationsByStationRow) []*observation.TrainObservation {
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
			TrainCategory:    nullable.Deref(row.TrainCategory),
			OriginID:         nullable.Deref(row.OriginID),
			OriginName:       nullable.Deref(row.OriginName),
			DestinationID:    nullable.Deref(row.DestinationID),
			DestinationName:  nullable.Deref(row.DestinationName),
			ScheduledTime:    nullable.Deref(row.ScheduledTime),
			Delay:            int(nullable.Deref(row.Delay)),
			Platform:         nullable.Deref(row.Platform),
			CirculationState: int(nullable.Deref(row.CirculationState)),
		}
	}
	return observations
}
