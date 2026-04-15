package cli

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	anvilroot "github.com/ernesto2108/anvil"
	"github.com/ernesto2108/anvil/pkg/storage"
)

// openDashboardDB opens the SQLite database, running migrations from the
// filesystem when available, or using the embedded migrations as fallback.
//
// This prevents the "stale binary" problem: when new migrations are added
// to the project but the binary hasn't been rebuilt, the filesystem-based
// migrations ensure the DB schema stays up-to-date.
//
// When forceMigrate is true, any dirty flag in schema_migrations is cleared
// before running migrations. Use this as a last resort when auto-recovery
// fails (e.g., corrupt migration state that recoverDirtyState cannot handle).
func openDashboardDB(dbPath string, forceMigrate bool) (*sql.DB, error) {
	db, err := storage.OpenDB(dbPath)
	if err != nil {
		return nil, err
	}

	if forceMigrate {
		log.Println("storage: --force-migrate: limpiando dirty state manualmente")
		if _, err := db.Exec("UPDATE schema_migrations SET dirty = false"); err != nil {
			log.Printf("storage: --force-migrate: no se pudo limpiar dirty state: %s", err)
		}
	}

	// 1. Try CWD/migrations/ — always fresh during development
	if dir, err := os.Getwd(); err == nil {
		candidate := filepath.Join(dir, "migrations")
		if isDirExists(candidate) {
			if err := storage.RunMigrations(db, candidate); err != nil {
				db.Close()
				return nil, err
			}
			return db, nil
		}
	}

	// 2. Try relative to the executable (installed binaries alongside project)
	if exe, err := os.Executable(); err == nil {
		if real, err := filepath.EvalSymlinks(exe); err == nil {
			candidate := filepath.Join(filepath.Dir(real), "..", "migrations")
			if isDirExists(candidate) {
				if err := storage.RunMigrations(db, candidate); err != nil {
					db.Close()
					return nil, err
				}
				return db, nil
			}
		}
	}

	// 3. Fallback: embedded migrations (binary is self-contained)
	subFS, err := fs.Sub(anvilroot.MigrationsFS, "migrations")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("sub-filesystem de migraciones: %w", err)
	}
	if err := storage.RunMigrationsFS(db, subFS); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func isDirExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}
