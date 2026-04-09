package store

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
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

// RunMigrationsFS aplica todas las migraciones pendientes al esquema del dashboard
// leyendo los archivos .sql desde un filesystem embebido (fs.FS), típicamente un
// embed.FS incluido en el binario. migrations debe contener los archivos SQL en su
// raíz (usa fs.Sub antes de llamar si el embed incluye un subdirectorio).
// Retorna nil si no hay cambios pendientes (migrate.ErrNoChange).
//
// IMPORTANTE: no llama m.Close() para evitar cerrar el *sql.DB que administra el
// caller. Los drivers creados con WithInstance son propiedad del caller.
func RunMigrationsFS(db *sql.DB, migrations fs.FS) error {
	srcDriver, err := iofs.New(migrations, ".")
	if err != nil {
		return fmt.Errorf("dashboard/store: crear source driver iofs: %w", err)
	}

	dbDriver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("dashboard/store: crear database driver de migraciones: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", srcDriver, "sqlite3", dbDriver)
	if err != nil {
		return fmt.Errorf("dashboard/store: inicializar migrate: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("dashboard/store: aplicar migraciones: %w", err)
	}

	return nil
}
