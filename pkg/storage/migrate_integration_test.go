//go:build integration

package storage_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/ernesto2108/anvil/pkg/storage"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// migrationsDir returns the absolute path to the project's migrations/ folder.
func migrationsDir(t *testing.T) string {
	t.Helper()
	// In CI the working dir is /app (repo root).
	// Locally, go test runs from pkg/storage/, so we walk up.
	candidates := []string{
		"migrations",
		filepath.Join("..", "..", "migrations"),
	}
	for _, c := range candidates {
		abs, _ := filepath.Abs(c)
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			return abs
		}
	}
	t.Fatal("cannot locate migrations/ directory")
	return ""
}

// openTestDB creates a temp SQLite database ready for migrations.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := storage.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// tableExists checks whether a table exists in the database.
func tableExists(t *testing.T, db *sql.DB, table string) bool {
	t.Helper()
	var name string
	err := db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
	).Scan(&name)
	return err == nil
}

// schemaVersion reads the current migration version from schema_migrations.
func schemaVersion(t *testing.T, db *sql.DB) uint {
	t.Helper()
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		t.Fatalf("sqlite3 driver: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsDir(t), "sqlite3", driver)
	if err != nil {
		t.Fatalf("migrate instance: %v", err)
	}
	v, _, err := m.Version()
	if err != nil {
		t.Fatalf("m.Version: %v", err)
	}
	return v
}

// expectedTables are all tables that should exist after all migrations run.
var expectedTables = []string{
	"runs", "agents", "files", "events",
	"tool_uses", "tasks", "prompts",
	"error_groups", "error_group_runs", "error_group_history",
}

func TestIntegration_AllMigrationsUp(t *testing.T) {
	db := openTestDB(t)
	migDir := migrationsDir(t)

	if err := storage.RunMigrations(db, migDir); err != nil {
		t.Fatalf("RunMigrations up: %v", err)
	}

	for _, table := range expectedTables {
		if !tableExists(t, db, table) {
			t.Errorf("table %q should exist after full migration", table)
		}
	}

	v := schemaVersion(t, db)
	if v != 20 {
		t.Errorf("schema version: expected 20, got %d", v)
	}
}

func TestIntegration_UpThenDownToZero(t *testing.T) {
	db := openTestDB(t)
	migDir := migrationsDir(t)

	if err := storage.RunMigrations(db, migDir); err != nil {
		t.Fatalf("RunMigrations up: %v", err)
	}

	// Migrate all the way down.
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		t.Fatalf("sqlite3 driver: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+migDir, "sqlite3", driver)
	if err != nil {
		t.Fatalf("migrate instance: %v", err)
	}
	if err := m.Down(); err != nil {
		t.Fatalf("m.Down: %v", err)
	}

	for _, table := range expectedTables {
		if tableExists(t, db, table) {
			t.Errorf("table %q should NOT exist after full rollback", table)
		}
	}
}

func TestIntegration_PartialThenFull(t *testing.T) {
	db := openTestDB(t)
	migDir := migrationsDir(t)

	// Migrate up to version 10 only.
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		t.Fatalf("sqlite3 driver: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+migDir, "sqlite3", driver)
	if err != nil {
		t.Fatalf("migrate instance: %v", err)
	}
	if err := m.Migrate(10); err != nil {
		t.Fatalf("m.Migrate(10): %v", err)
	}

	// Tables from early migrations should exist.
	if !tableExists(t, db, "runs") {
		t.Fatal("table runs should exist at version 10")
	}
	// Tables from later migrations should NOT exist yet.
	if tableExists(t, db, "error_groups") {
		t.Fatal("table error_groups should NOT exist at version 10")
	}

	// Now apply all remaining migrations via RunMigrations.
	if err := storage.RunMigrations(db, migDir); err != nil {
		t.Fatalf("RunMigrations (partial→full): %v", err)
	}

	for _, table := range expectedTables {
		if !tableExists(t, db, table) {
			t.Errorf("table %q should exist after full migration", table)
		}
	}
}

func TestIntegration_DirtyStateRecovery(t *testing.T) {
	db := openTestDB(t)
	migDir := migrationsDir(t)

	// Apply migrations up to version 5.
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		t.Fatalf("sqlite3 driver: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+migDir, "sqlite3", driver)
	if err != nil {
		t.Fatalf("migrate instance: %v", err)
	}
	if err := m.Migrate(5); err != nil {
		t.Fatalf("m.Migrate(5): %v", err)
	}

	// Simulate a dirty state at version 6 (migration failed mid-execution).
	if _, err := db.Exec("UPDATE schema_migrations SET version = 6, dirty = true"); err != nil {
		t.Fatalf("simular dirty state: %v", err)
	}

	// RunMigrations should recover from dirty state and apply all remaining.
	if err := storage.RunMigrations(db, migDir); err != nil {
		t.Fatalf("RunMigrations after dirty state: %v", err)
	}

	for _, table := range expectedTables {
		if !tableExists(t, db, table) {
			t.Errorf("table %q should exist after dirty recovery + migration", table)
		}
	}

	v := schemaVersion(t, db)
	if v != 20 {
		t.Errorf("schema version: expected 20, got %d", v)
	}
}
