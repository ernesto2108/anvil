package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations aplica todas las migraciones pendientes al esquema del dashboard.
// migrationsPath es la ruta al directorio con los archivos .sql (e.g. "migrations").
// Retorna nil si no hay cambios pendientes (migrate.ErrNoChange).
func RunMigrations(db *sql.DB, migrationsPath string) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("dashboard/store: crear database driver de migraciones: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"sqlite3", driver,
	)
	if err != nil {
		return fmt.Errorf("dashboard/store: inicializar migrate: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("dashboard/store: aplicar migraciones: %w", err)
	}

	return nil
}
