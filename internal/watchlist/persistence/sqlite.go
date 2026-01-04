package persistence

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/emiliopalmerini/treni/internal/watchlist"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Create(ctx context.Context, entity *watchlist.WatchedTrain) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO watched_train (id, train_number, origin_id, origin_name, destination, days_of_week, notes, active, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entity.ID.String(), entity.TrainNumber, entity.OriginID, entity.OriginName,
		entity.Destination, entity.DaysOfWeek, entity.Notes, entity.Active, entity.CreatedAt)
	return err
}

func (r *SQLiteRepository) GetByID(ctx context.Context, id uuid.UUID) (*watchlist.WatchedTrain, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, train_number, origin_id, origin_name, destination, days_of_week, notes, active, created_at
		 FROM watched_train WHERE id = ?`, id.String())

	return scanWatchedTrain(row)
}

func (r *SQLiteRepository) List(ctx context.Context) ([]*watchlist.WatchedTrain, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, train_number, origin_id, origin_name, destination, days_of_week, notes, active, created_at
		 FROM watched_train ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanWatchedTrains(rows)
}

func (r *SQLiteRepository) ListActive(ctx context.Context) ([]*watchlist.WatchedTrain, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, train_number, origin_id, origin_name, destination, days_of_week, notes, active, created_at
		 FROM watched_train WHERE active = 1 ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanWatchedTrains(rows)
}

func (r *SQLiteRepository) Update(ctx context.Context, entity *watchlist.WatchedTrain) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE watched_train SET train_number = ?, origin_id = ?, origin_name = ?, destination = ?,
		 days_of_week = ?, notes = ?, active = ? WHERE id = ?`,
		entity.TrainNumber, entity.OriginID, entity.OriginName, entity.Destination,
		entity.DaysOfWeek, entity.Notes, entity.Active, entity.ID.String())
	return err
}

func (r *SQLiteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM watched_train WHERE id = ?", id.String())
	return err
}

func (r *SQLiteRepository) CreateCheck(ctx context.Context, check *watchlist.TrainCheck) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO train_check (id, watched_id, train_number, delay, status, checked_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		check.ID.String(), check.WatchedID.String(), check.TrainNumber,
		check.Delay, check.Status, check.CheckedAt)
	return err
}

func (r *SQLiteRepository) GetChecksByWatched(ctx context.Context, watchedID uuid.UUID) ([]*watchlist.TrainCheck, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, watched_id, train_number, delay, status, checked_at
		 FROM train_check WHERE watched_id = ? ORDER BY checked_at DESC LIMIT 100`, watchedID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTrainChecks(rows)
}

func (r *SQLiteRepository) GetRecentChecks(ctx context.Context, limit int) ([]*watchlist.TrainCheck, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, watched_id, train_number, delay, status, checked_at
		 FROM train_check ORDER BY checked_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTrainChecks(rows)
}

func scanWatchedTrain(row *sql.Row) (*watchlist.WatchedTrain, error) {
	var entity watchlist.WatchedTrain
	var idStr string
	var daysOfWeek, notes sql.NullString
	if err := row.Scan(&idStr, &entity.TrainNumber, &entity.OriginID, &entity.OriginName,
		&entity.Destination, &daysOfWeek, &notes, &entity.Active, &entity.CreatedAt); err != nil {
		return nil, err
	}
	entity.ID, _ = uuid.Parse(idStr)
	entity.DaysOfWeek = daysOfWeek.String
	entity.Notes = notes.String
	return &entity, nil
}

func scanWatchedTrains(rows *sql.Rows) ([]*watchlist.WatchedTrain, error) {
	var entities []*watchlist.WatchedTrain
	for rows.Next() {
		var entity watchlist.WatchedTrain
		var idStr string
		var daysOfWeek, notes sql.NullString
		if err := rows.Scan(&idStr, &entity.TrainNumber, &entity.OriginID, &entity.OriginName,
			&entity.Destination, &daysOfWeek, &notes, &entity.Active, &entity.CreatedAt); err != nil {
			return nil, err
		}
		entity.ID, _ = uuid.Parse(idStr)
		entity.DaysOfWeek = daysOfWeek.String
		entity.Notes = notes.String
		entities = append(entities, &entity)
	}
	return entities, rows.Err()
}

func scanTrainChecks(rows *sql.Rows) ([]*watchlist.TrainCheck, error) {
	var checks []*watchlist.TrainCheck
	for rows.Next() {
		var check watchlist.TrainCheck
		var idStr, watchedIDStr string
		if err := rows.Scan(&idStr, &watchedIDStr, &check.TrainNumber, &check.Delay,
			&check.Status, &check.CheckedAt); err != nil {
			return nil, err
		}
		check.ID, _ = uuid.Parse(idStr)
		check.WatchedID, _ = uuid.Parse(watchedIDStr)
		checks = append(checks, &check)
	}
	return checks, rows.Err()
}
