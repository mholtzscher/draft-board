package database

import (
	"database/sql"
	"testing"
)

// NewTestDB creates an in-memory SQLite database for testing
func NewTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}

	if err := RunMigrations(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

// CloseTestDB closes the test database
func CloseTestDB(t *testing.T, db *sql.DB) {
	t.Helper()
	if err := db.Close(); err != nil {
		t.Errorf("failed to close test database: %v", err)
	}
}
