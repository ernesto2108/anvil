package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// RunMigrations applies all pending migrations from the filesystem directory
// at migrationsPath. Returns nil if there are no pending changes.
//
// If the database is in a dirty state (failed migration), this function
// attempts auto-recovery by forcing the version back to the last clean state.
func RunMigrations(db *sql.DB, migrationsPath string) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("storage: crear database driver de migraciones: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"sqlite3", driver,
	)
	if err != nil {
		return fmt.Errorf("storage: inicializar migrate: %w", err)
	}

	if err := recoverDirtyState(m); err != nil {
		return fmt.Errorf("storage: recuperar dirty state: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("storage: aplicar migraciones: %w", err)
	}

	return nil
}

// RunMigrationsFS applies all pending migrations reading .sql files from an
// fs.FS (typically an embed.FS). The migrations must be at the root of the
// provided filesystem — use fs.Sub before calling if needed.
//
// Does not call m.Close() to avoid closing the *sql.DB owned by the caller.
//
// If the database is in a dirty state (failed migration), this function
// attempts auto-recovery by forcing the version back to the last clean state.
func RunMigrationsFS(db *sql.DB, migrations fs.FS) error {
	srcDriver, err := iofs.New(migrations, ".")
	if err != nil {
		return fmt.Errorf("storage: crear source driver iofs: %w", err)
	}

	dbDriver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("storage: crear database driver de migraciones: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", srcDriver, "sqlite3", dbDriver)
	if err != nil {
		return fmt.Errorf("storage: inicializar migrate: %w", err)
	}

	if err := recoverDirtyState(m); err != nil {
		return fmt.Errorf("storage: recuperar dirty state: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("storage: aplicar migraciones: %w", err)
	}

	return nil
}

// recoverDirtyState checks if the DB is in a dirty migration state and
// attempts recovery by forcing the version back to the last clean state.
// A dirty state occurs when a migration fails mid-execution; subsequent
// m.Up() calls refuse to run until the state is resolved.
func recoverDirtyState(m *migrate.Migrate) error {
	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return nil // fresh DB, no migrations applied yet
		}
		return fmt.Errorf("leer version de migración: %w", err)
	}
	if !dirty {
		return nil
	}

	// Force rollback to previous clean version.
	targetVersion := int(version) - 1
	log.Printf("storage: dirty state detectado en version %d, forzando rollback a version %d", version, targetVersion)

	if targetVersion < 0 {
		targetVersion = -1 // special value: no version applied
	}

	if err := m.Force(targetVersion); err != nil {
		return fmt.Errorf("forzar rollback a version %d: %w", targetVersion, err)
	}

	log.Printf("storage: dirty state resuelto, continuando con migraciones")
	return nil
}

// SyncSchemaVersionFile reads the current migration version from the database
// and writes it to the SCHEMA_VERSION file inside migrationsDir. This keeps
// the file in sync after new migrations are applied (useful in dev/CI).
func SyncSchemaVersionFile(db *sql.DB, migrationsDir string) error {
	var version int
	err := db.QueryRow("SELECT version FROM schema_migrations LIMIT 1").Scan(&version)
	if err != nil {
		return fmt.Errorf("storage: leer version de schema_migrations: %w", err)
	}

	versionFile := filepath.Join(migrationsDir, "SCHEMA_VERSION")
	newContent := strconv.Itoa(version) + "\n"

	existing, err := os.ReadFile(versionFile)
	if err == nil && strings.TrimSpace(string(existing)) == strconv.Itoa(version) {
		return nil // already up to date
	}

	if err := os.WriteFile(versionFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("storage: escribir SCHEMA_VERSION: %w", err)
	}

	log.Printf("storage: SCHEMA_VERSION actualizado a %d", version)
	return nil
}
