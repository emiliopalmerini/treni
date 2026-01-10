package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/emiliopalmerini/treni/internal/database/nullable"
	"github.com/emiliopalmerini/treni/internal/database/sqlc"
	"github.com/emiliopalmerini/treni/internal/station"
)

type SQLiteRepository struct {
	q *sqlc.Queries
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{q: sqlc.New(db)}
}

func (r *SQLiteRepository) Create(ctx context.Context, entity *station.Station) error {
	updatedAt := entity.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}
	return r.q.CreateStation(ctx, sqlc.CreateStationParams{
		ID:        entity.ID,
		Name:      entity.Name,
		Region:    nullable.Ptr(int64(entity.Region)),
		Latitude:  nullable.Ptr(entity.Latitude),
		Longitude: nullable.Ptr(entity.Longitude),
		UpdatedAt: nullable.Ptr(updatedAt),
	})
}

func (r *SQLiteRepository) GetByID(ctx context.Context, id string) (*station.Station, error) {
	row, err := r.q.GetStationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return rowToStation(row), nil
}

func (r *SQLiteRepository) List(ctx context.Context) ([]*station.Station, error) {
	rows, err := r.q.ListStations(ctx)
	if err != nil {
		return nil, err
	}
	return listRowsToStations(rows), nil
}

func (r *SQLiteRepository) Search(ctx context.Context, query string) ([]*station.Station, error) {
	rows, err := r.q.SearchStations(ctx, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	return searchRowsToStations(rows), nil
}

func (r *SQLiteRepository) ListWithCoordinates(ctx context.Context) ([]*station.Station, error) {
	rows, err := r.q.ListStationsWithCoordinates(ctx)
	if err != nil {
		return nil, err
	}
	return coordRowsToStations(rows), nil
}

func (r *SQLiteRepository) Count(ctx context.Context) (int, error) {
	count, err := r.q.CountStations(ctx)
	return int(count), err
}

func (r *SQLiteRepository) Update(ctx context.Context, entity *station.Station) error {
	updatedAt := entity.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}
	return r.q.UpdateStation(ctx, sqlc.UpdateStationParams{
		ID:        entity.ID,
		Name:      entity.Name,
		Region:    nullable.Ptr(int64(entity.Region)),
		Latitude:  nullable.Ptr(entity.Latitude),
		Longitude: nullable.Ptr(entity.Longitude),
		UpdatedAt: nullable.Ptr(updatedAt),
	})
}

func (r *SQLiteRepository) Delete(ctx context.Context, id string) error {
	return r.q.DeleteStation(ctx, id)
}

func (r *SQLiteRepository) Upsert(ctx context.Context, entity *station.Station) error {
	updatedAt := entity.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}
	return r.q.UpsertStation(ctx, sqlc.UpsertStationParams{
		ID:        entity.ID,
		Name:      entity.Name,
		Region:    nullable.Ptr(int64(entity.Region)),
		Latitude:  nullable.Ptr(entity.Latitude),
		Longitude: nullable.Ptr(entity.Longitude),
		UpdatedAt: nullable.Ptr(updatedAt),
	})
}

func rowToStation(row sqlc.GetStationByIDRow) *station.Station {
	return &station.Station{
		ID:        row.ID,
		Name:      row.Name,
		Region:    int(nullable.Deref(row.Region)),
		Latitude:  nullable.Deref(row.Latitude),
		Longitude: nullable.Deref(row.Longitude),
		UpdatedAt: nullable.Deref(row.UpdatedAt),
	}
}

func listRowsToStations(rows []sqlc.ListStationsRow) []*station.Station {
	stations := make([]*station.Station, len(rows))
	for i, row := range rows {
		stations[i] = &station.Station{
			ID:        row.ID,
			Name:      row.Name,
			Region:    int(nullable.Deref(row.Region)),
			Latitude:  nullable.Deref(row.Latitude),
			Longitude: nullable.Deref(row.Longitude),
			UpdatedAt: nullable.Deref(row.UpdatedAt),
		}
	}
	return stations
}

func searchRowsToStations(rows []sqlc.SearchStationsRow) []*station.Station {
	stations := make([]*station.Station, len(rows))
	for i, row := range rows {
		stations[i] = &station.Station{
			ID:        row.ID,
			Name:      row.Name,
			Region:    int(nullable.Deref(row.Region)),
			Latitude:  nullable.Deref(row.Latitude),
			Longitude: nullable.Deref(row.Longitude),
			UpdatedAt: nullable.Deref(row.UpdatedAt),
		}
	}
	return stations
}

func coordRowsToStations(rows []sqlc.ListStationsWithCoordinatesRow) []*station.Station {
	stations := make([]*station.Station, len(rows))
	for i, row := range rows {
		stations[i] = &station.Station{
			ID:        row.ID,
			Name:      row.Name,
			Region:    int(nullable.Deref(row.Region)),
			Latitude:  nullable.Deref(row.Latitude),
			Longitude: nullable.Deref(row.Longitude),
			UpdatedAt: nullable.Deref(row.UpdatedAt),
		}
	}
	return stations
}
