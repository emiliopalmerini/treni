package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/emiliopalmerini/treni/internal/database/sqlc"
	"github.com/emiliopalmerini/treni/internal/staticdata"
)

type SQLiteMetadataRepository struct {
	q *sqlc.Queries
}

func NewSQLiteMetadataRepository(db *sql.DB) *SQLiteMetadataRepository {
	return &SQLiteMetadataRepository{q: sqlc.New(db)}
}

func (r *SQLiteMetadataRepository) Get(ctx context.Context, entityType string) (*staticdata.ImportMetadata, error) {
	row, err := r.q.GetImportMetadata(ctx, entityType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, staticdata.ErrNotFound
		}
		return nil, err
	}

	return &staticdata.ImportMetadata{
		EntityType:   row.EntityType,
		LastImport:   row.LastImport,
		RecordCount:  int(deref(row.RecordCount)),
		DurationMs:   deref(row.ImportDurationMs),
		Status:       deref(row.Status),
		ErrorMessage: deref(row.ErrorMessage),
	}, nil
}

func (r *SQLiteMetadataRepository) Upsert(ctx context.Context, meta *staticdata.ImportMetadata) error {
	return r.q.UpsertImportMetadata(ctx, sqlc.UpsertImportMetadataParams{
		EntityType:       meta.EntityType,
		LastImport:       meta.LastImport,
		RecordCount:      ptr(int64(meta.RecordCount)),
		ImportDurationMs: ptr(meta.DurationMs),
		Status:           ptr(meta.Status),
		ErrorMessage:     ptr(meta.ErrorMessage),
	})
}

func (r *SQLiteMetadataRepository) ShouldRefresh(ctx context.Context, entityType string, maxAge time.Duration) (bool, error) {
	meta, err := r.Get(ctx, entityType)
	if err != nil {
		if err == staticdata.ErrNotFound {
			return true, nil
		}
		return false, err
	}

	return time.Since(meta.LastImport) > maxAge, nil
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
