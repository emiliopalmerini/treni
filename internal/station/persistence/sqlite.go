package persistence

import (
	"context"
	"database/sql"

	"github.com/emiliopalmerini/treni/internal/station"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Create(ctx context.Context, entity *station.Station) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO station (id, name, region, latitude, longitude, is_favorite)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		entity.ID, entity.Name, entity.Region, entity.Latitude, entity.Longitude, entity.IsFavorite)
	return err
}

func (r *SQLiteRepository) GetByID(ctx context.Context, id string) (*station.Station, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, region, latitude, longitude, is_favorite
		 FROM station WHERE id = ?`, id)

	var entity station.Station
	if err := row.Scan(&entity.ID, &entity.Name, &entity.Region, &entity.Latitude, &entity.Longitude, &entity.IsFavorite); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *SQLiteRepository) List(ctx context.Context) ([]*station.Station, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, region, latitude, longitude, is_favorite
		 FROM station ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*station.Station
	for rows.Next() {
		var entity station.Station
		if err := rows.Scan(&entity.ID, &entity.Name, &entity.Region, &entity.Latitude, &entity.Longitude, &entity.IsFavorite); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}
	return entities, rows.Err()
}

func (r *SQLiteRepository) Search(ctx context.Context, query string) ([]*station.Station, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, region, latitude, longitude, is_favorite
		 FROM station WHERE name LIKE ? ORDER BY name LIMIT 20`,
		"%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*station.Station
	for rows.Next() {
		var entity station.Station
		if err := rows.Scan(&entity.ID, &entity.Name, &entity.Region, &entity.Latitude, &entity.Longitude, &entity.IsFavorite); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}
	return entities, rows.Err()
}

func (r *SQLiteRepository) ListFavorites(ctx context.Context) ([]*station.Station, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, region, latitude, longitude, is_favorite
		 FROM station WHERE is_favorite = 1 ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*station.Station
	for rows.Next() {
		var entity station.Station
		if err := rows.Scan(&entity.ID, &entity.Name, &entity.Region, &entity.Latitude, &entity.Longitude, &entity.IsFavorite); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}
	return entities, rows.Err()
}

func (r *SQLiteRepository) ListWithCoordinates(ctx context.Context) ([]*station.Station, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, region, latitude, longitude, is_favorite
		 FROM station WHERE latitude != 0 AND longitude != 0`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*station.Station
	for rows.Next() {
		var entity station.Station
		if err := rows.Scan(&entity.ID, &entity.Name, &entity.Region, &entity.Latitude, &entity.Longitude, &entity.IsFavorite); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}
	return entities, rows.Err()
}

func (r *SQLiteRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM station").Scan(&count)
	return count, err
}

func (r *SQLiteRepository) Update(ctx context.Context, entity *station.Station) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE station SET name = ?, region = ?, latitude = ?, longitude = ?, is_favorite = ?
		 WHERE id = ?`,
		entity.Name, entity.Region, entity.Latitude, entity.Longitude, entity.IsFavorite, entity.ID)
	return err
}

func (r *SQLiteRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM station WHERE id = ?", id)
	return err
}

func (r *SQLiteRepository) Upsert(ctx context.Context, entity *station.Station) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO station (id, name, region, latitude, longitude, is_favorite)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		 name = excluded.name,
		 region = excluded.region,
		 latitude = excluded.latitude,
		 longitude = excluded.longitude`,
		entity.ID, entity.Name, entity.Region, entity.Latitude, entity.Longitude, entity.IsFavorite)
	return err
}
