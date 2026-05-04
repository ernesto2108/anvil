// Package testutil provides shared test helpers for dashboard sub-packages.
package testutil

import (
	"database/sql"
	"os"
	"sync"
	"testing"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"

	"github.com/ernesto2108/anvil/pkg/storage"
)

// vecAutoOnce registers sqlite-vec as an autoextension exactly once per test
// process so migration 000033 (CREATE VIRTUAL TABLE … USING vec0) succeeds.
var vecAutoOnce sync.Once

// OpenTestDB creates an in-memory SQLite database with all migrations applied.
// The caller must NOT close the returned *sql.DB — t.Cleanup handles it.
func OpenTestDB(t *testing.T) *sql.DB {
	t.Helper()
	vecAutoOnce.Do(sqlite_vec.Auto)

	db, err := sql.Open("sqlite3", ":memory:?_busy_timeout=5000")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		t.Fatalf("enable foreign_keys: %v", err)
	}

	migrationsDir := findMigrationsDir(t)
	if err := storage.RunMigrations(db, migrationsDir); err != nil {
		db.Close()
		t.Fatalf("run migrations: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func findMigrationsDir(t *testing.T) string {
	t.Helper()
	candidates := []string{
		"../../../migrations",
		"../../migrations",
		"../../../../migrations",
		"migrations",
	}
	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && fi.IsDir() {
			return c
		}
	}
	t.Fatal("could not find migrations directory")
	return ""
}
