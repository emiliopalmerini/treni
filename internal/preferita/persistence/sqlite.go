package persistence

import (
	"context"
	"database/sql"

	"github.com/emiliopalmerini/treni/internal/database/sqlc"
	"github.com/emiliopalmerini/treni/internal/preferita"
)

type SQLiteRepository struct {
	q *sqlc.Queries
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{q: sqlc.New(db)}
}

func (r *SQLiteRepository) List(ctx context.Context) ([]*preferita.Preferita, error) {
	rows, err := r.q.ListPreferite(ctx)
	if err != nil {
		return nil, err
	}

	entities := make([]*preferita.Preferita, len(rows))
	for i, row := range rows {
		entities[i] = &preferita.Preferita{
			StationID: row.StationID,
			Name:      row.Name,
		}
	}
	return entities, nil
}

func (r *SQLiteRepository) Add(ctx context.Context, entity *preferita.Preferita) error {
	return r.q.AddPreferita(ctx, sqlc.AddPreferitaParams{
		StationID: entity.StationID,
		Name:      entity.Name,
	})
}

func (r *SQLiteRepository) Remove(ctx context.Context, stationID string) error {
	return r.q.RemovePreferita(ctx, stationID)
}

func (r *SQLiteRepository) Exists(ctx context.Context, stationID string) (bool, error) {
	count, err := r.q.PreferitaExists(ctx, stationID)
	return count > 0, err
}
