package persistence

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/emiliopalmerini/treni/internal/database/sqlc"
	"github.com/emiliopalmerini/treni/internal/watchlist"
)

type SQLiteRepository struct {
	q *sqlc.Queries
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{q: sqlc.New(db)}
}

func (r *SQLiteRepository) Create(ctx context.Context, entity *watchlist.WatchedTrain) error {
	return r.q.CreateWatchedTrain(ctx, sqlc.CreateWatchedTrainParams{
		ID:          entity.ID.String(),
		TrainNumber: int64(entity.TrainNumber),
		OriginID:    entity.OriginID,
		OriginName:  entity.OriginName,
		Destination: entity.Destination,
		DaysOfWeek:  strPtr(entity.DaysOfWeek),
		Notes:       strPtr(entity.Notes),
		Active:      boolToInt64Ptr(entity.Active),
		CreatedAt:   entity.CreatedAt,
	})
}

func (r *SQLiteRepository) GetByID(ctx context.Context, id uuid.UUID) (*watchlist.WatchedTrain, error) {
	row, err := r.q.GetWatchedTrainByID(ctx, id.String())
	if err != nil {
		return nil, err
	}
	return sqlcWatchedTrainToWatchedTrain(row), nil
}

func (r *SQLiteRepository) List(ctx context.Context) ([]*watchlist.WatchedTrain, error) {
	rows, err := r.q.ListWatchedTrains(ctx)
	if err != nil {
		return nil, err
	}
	return sqlcWatchedTrainsToWatchedTrains(rows), nil
}

func (r *SQLiteRepository) ListActive(ctx context.Context) ([]*watchlist.WatchedTrain, error) {
	rows, err := r.q.ListActiveWatchedTrains(ctx)
	if err != nil {
		return nil, err
	}
	return sqlcWatchedTrainsToWatchedTrains(rows), nil
}

func (r *SQLiteRepository) Update(ctx context.Context, entity *watchlist.WatchedTrain) error {
	return r.q.UpdateWatchedTrain(ctx, sqlc.UpdateWatchedTrainParams{
		ID:          entity.ID.String(),
		TrainNumber: int64(entity.TrainNumber),
		OriginID:    entity.OriginID,
		OriginName:  entity.OriginName,
		Destination: entity.Destination,
		DaysOfWeek:  strPtr(entity.DaysOfWeek),
		Notes:       strPtr(entity.Notes),
		Active:      boolToInt64Ptr(entity.Active),
	})
}

func (r *SQLiteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteWatchedTrain(ctx, id.String())
}

func (r *SQLiteRepository) CreateCheck(ctx context.Context, check *watchlist.TrainCheck) error {
	return r.q.CreateTrainCheck(ctx, sqlc.CreateTrainCheckParams{
		ID:          check.ID.String(),
		WatchedID:   check.WatchedID.String(),
		TrainNumber: int64(check.TrainNumber),
		Delay:       ptr(int64(check.Delay)),
		Status:      check.Status,
		CheckedAt:   check.CheckedAt,
	})
}

func (r *SQLiteRepository) GetChecksByWatched(ctx context.Context, watchedID uuid.UUID) ([]*watchlist.TrainCheck, error) {
	rows, err := r.q.GetTrainChecksByWatched(ctx, watchedID.String())
	if err != nil {
		return nil, err
	}
	return sqlcTrainChecksToTrainChecks(rows), nil
}

func (r *SQLiteRepository) GetRecentChecks(ctx context.Context, limit int) ([]*watchlist.TrainCheck, error) {
	rows, err := r.q.GetRecentTrainChecks(ctx, int64(limit))
	if err != nil {
		return nil, err
	}
	return sqlcTrainChecksToTrainChecks(rows), nil
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

func boolToInt64Ptr(b bool) *int64 {
	if b {
		return ptr(int64(1))
	}
	return ptr(int64(0))
}

func int64ToBool(p *int64) bool {
	if p == nil {
		return false
	}
	return *p == 1
}

func sqlcWatchedTrainToWatchedTrain(row sqlc.WatchedTrain) *watchlist.WatchedTrain {
	id, _ := uuid.Parse(row.ID)
	return &watchlist.WatchedTrain{
		ID:          id,
		TrainNumber: int(row.TrainNumber),
		OriginID:    row.OriginID,
		OriginName:  row.OriginName,
		Destination: row.Destination,
		DaysOfWeek:  deref(row.DaysOfWeek),
		Notes:       deref(row.Notes),
		Active:      int64ToBool(row.Active),
		CreatedAt:   row.CreatedAt,
	}
}

func sqlcWatchedTrainsToWatchedTrains(rows []sqlc.WatchedTrain) []*watchlist.WatchedTrain {
	trains := make([]*watchlist.WatchedTrain, len(rows))
	for i, row := range rows {
		trains[i] = sqlcWatchedTrainToWatchedTrain(row)
	}
	return trains
}

func sqlcTrainChecksToTrainChecks(rows []sqlc.TrainCheck) []*watchlist.TrainCheck {
	checks := make([]*watchlist.TrainCheck, len(rows))
	for i, row := range rows {
		id, _ := uuid.Parse(row.ID)
		watchedID, _ := uuid.Parse(row.WatchedID)
		checks[i] = &watchlist.TrainCheck{
			ID:          id,
			WatchedID:   watchedID,
			TrainNumber: int(row.TrainNumber),
			Delay:       int(deref(row.Delay)),
			Status:      row.Status,
			CheckedAt:   row.CheckedAt,
		}
	}
	return checks
}
