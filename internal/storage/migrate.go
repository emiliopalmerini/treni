package storage

import (
	"embed"
	"fmt"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func (db *DB) Migrate() error {
	content, err := migrationsFS.ReadFile("migrations/001_init.up.sql")
	if err != nil {
		return fmt.Errorf("read migration file: %w", err)
	}

	_, err = db.Exec(string(content))
	if err != nil {
		return fmt.Errorf("execute migration: %w", err)
	}

	return nil
}
