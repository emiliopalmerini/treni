package persistence

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/emiliopalmerini/treni/internal/train"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Create(ctx context.Context, entity *train.Train) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO train (id, name) VALUES (?, ?)",
		entity.ID.String(), entity.Name)
	return err
}

func (r *SQLiteRepository) GetByID(ctx context.Context, id uuid.UUID) (*train.Train, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, name FROM train WHERE id = ?",
		id.String())

	var entity train.Train
	var idStr string
	if err := row.Scan(&idStr, &entity.Name); err != nil {
		return nil, err
	}
	entity.ID, _ = uuid.Parse(idStr)
	return &entity, nil
}

func (r *SQLiteRepository) List(ctx context.Context) ([]*train.Train, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name FROM train")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*train.Train
	for rows.Next() {
		var entity train.Train
		var idStr string
		if err := rows.Scan(&idStr, &entity.Name); err != nil {
			return nil, err
		}
		entity.ID, _ = uuid.Parse(idStr)
		entities = append(entities, &entity)
	}
	return entities, rows.Err()
}

func (r *SQLiteRepository) Update(ctx context.Context, entity *train.Train) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE train SET name = ? WHERE id = ?",
		entity.Name, entity.ID.String())
	return err
}

func (r *SQLiteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM train WHERE id = ?",
		id.String())
	return err
}
