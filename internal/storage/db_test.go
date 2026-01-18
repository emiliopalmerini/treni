package storage

import (
	"os"
	"testing"
)

func TestIntegrationTursoConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	url := os.Getenv("TURSO_DATABASE_URL")
	token := os.Getenv("TURSO_AUTH_TOKEN")

	if url == "" || token == "" {
		t.Skip("TURSO_DATABASE_URL and TURSO_AUTH_TOKEN not set")
	}

	db, err := New()
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// Test query
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if result != 1 {
		t.Errorf("expected 1, got %d", result)
	}
}

func TestLocalConnection(t *testing.T) {
	db, err := NewLocal(":memory:")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// Test query
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if result != 1 {
		t.Errorf("expected 1, got %d", result)
	}
}
