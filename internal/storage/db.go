package storage

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/tursodatabase/go-libsql"
)

type DB struct {
	*sql.DB
}

func New() (*DB, error) {
	url := os.Getenv("TRENI_DATABASE_URL")
	token := os.Getenv("TRENI_AUTH_TOKEN")

	if url == "" {
		return nil, fmt.Errorf("TRENI_DATABASE_URL not set")
	}

	var dsn string
	if token != "" {
		dsn = fmt.Sprintf("%s?authToken=%s", url, token)
	} else {
		dsn = url
	}

	db, err := sql.Open("libsql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{db}, nil
}

func NewLocal(path string) (*DB, error) {
	dsn := fmt.Sprintf("file:%s", path)

	db, err := sql.Open("libsql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{db}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}
