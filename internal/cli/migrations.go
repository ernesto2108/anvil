package cli

import (
	"os"
	"path/filepath"

	anvilroot "github.com/ernesto2108/anvil"
	"github.com/ernesto2108/anvil/internal/dashboard/store"
)

// openStorePreferFilesystem opens the SQLite store using filesystem-based
// migrations when available (CWD/migrations/ or relative to the executable).
// Falls back to the embedded migrations compiled into the binary.
//
// This prevents the "stale binary" problem: when new migrations are added
// to the project but the binary hasn't been rebuilt, the filesystem-based
// migrations ensure the DB schema stays up-to-date.
func openStorePreferFilesystem(dbPath string) (*store.SQLiteStore, error) {
	// 1. Try CWD/migrations/ — always fresh during development
	if dir, err := os.Getwd(); err == nil {
		candidate := filepath.Join(dir, "migrations")
		if isDirExists(candidate) {
			return store.New(dbPath, candidate, 0)
		}
	}

	// 2. Try relative to the executable (installed binaries alongside project)
	if exe, err := os.Executable(); err == nil {
		if real, err := filepath.EvalSymlinks(exe); err == nil {
			candidate := filepath.Join(filepath.Dir(real), "..", "migrations")
			if isDirExists(candidate) {
				return store.New(dbPath, candidate, 0)
			}
		}
	}

	// 3. Fallback: embedded migrations (binary is self-contained)
	return store.NewFS(dbPath, anvilroot.MigrationsFS, "migrations", 0)
}

func isDirExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}
