package storage_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/ernesto2108/anvil/pkg/storage"
)

func TestOpenDB(t *testing.T) {
	t.Run("creates db file with WAL and foreign keys", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "sub", "test.db")

		db, err := storage.OpenDB(dbPath)
		if err != nil {
			t.Fatalf("OpenDB: %v", err)
		}
		defer db.Close()

		// File must exist with 0600 permissions.
		info, err := os.Stat(dbPath)
		if err != nil {
			t.Fatalf("stat db file: %v", err)
		}
		if perm := info.Mode().Perm(); perm&0077 != 0 {
			t.Errorf("permisos: esperado 0600, obtuvo %04o", perm)
		}

		// WAL journal mode should be active.
		var journalMode string
		if err := db.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
			t.Fatalf("PRAGMA journal_mode: %v", err)
		}
		if journalMode != "wal" {
			t.Errorf("journal_mode: esperado %q, obtuvo %q", "wal", journalMode)
		}

		// foreign_keys must be ON.
		var fk int
		if err := db.QueryRow("PRAGMA foreign_keys").Scan(&fk); err != nil {
			t.Fatalf("PRAGMA foreign_keys: %v", err)
		}
		if fk != 1 {
			t.Errorf("foreign_keys: esperado 1, obtuvo %d", fk)
		}

		// Basic write/read should work.
		if _, err := db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY)"); err != nil {
			t.Fatalf("CREATE TABLE: %v", err)
		}
		if _, err := db.Exec("INSERT INTO test (id) VALUES (1)"); err != nil {
			t.Fatalf("INSERT: %v", err)
		}
		var id int
		if err := db.QueryRow("SELECT id FROM test").Scan(&id); err != nil {
			t.Fatalf("SELECT: %v", err)
		}
		if id != 1 {
			t.Errorf("id: esperado 1, obtuvo %d", id)
		}
	})

	t.Run("opens existing db without error", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "existing.db")

		db1, err := storage.OpenDB(dbPath)
		if err != nil {
			t.Fatalf("OpenDB (1st): %v", err)
		}
		if _, err := db1.Exec("CREATE TABLE marker (v TEXT)"); err != nil {
			t.Fatalf("CREATE TABLE: %v", err)
		}
		db1.Close()

		db2, err := storage.OpenDB(dbPath)
		if err != nil {
			t.Fatalf("OpenDB (2nd): %v", err)
		}
		defer db2.Close()

		var name string
		err = db2.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='marker'").Scan(&name)
		if err != nil {
			t.Fatalf("table marker not found after reopen: %v", err)
		}
	})
}

func TestRunMigrations(t *testing.T) {
	t.Run("applies filesystem migrations", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.db")

		// Create a minimal migration pair.
		migDir := filepath.Join(dir, "migrations")
		if err := os.Mkdir(migDir, 0755); err != nil {
			t.Fatal(err)
		}
		up := "CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT NOT NULL);\n"
		down := "DROP TABLE IF EXISTS items;\n"
		os.WriteFile(filepath.Join(migDir, "000001_create_items.up.sql"), []byte(up), 0644)
		os.WriteFile(filepath.Join(migDir, "000001_create_items.down.sql"), []byte(down), 0644)

		db, err := storage.OpenDB(dbPath)
		if err != nil {
			t.Fatalf("OpenDB: %v", err)
		}
		defer db.Close()

		if err := storage.RunMigrations(db, migDir); err != nil {
			t.Fatalf("RunMigrations: %v", err)
		}

		// Table must exist.
		var name string
		if err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='items'").Scan(&name); err != nil {
			t.Fatalf("table items not found: %v", err)
		}

		// Running again should be idempotent (ErrNoChange is swallowed).
		if err := storage.RunMigrations(db, migDir); err != nil {
			t.Fatalf("RunMigrations (idempotent): %v", err)
		}
	})
}

func TestRunMigrationsFS(t *testing.T) {
	t.Run("applies embedded migrations", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test.db")

		memFS := fstest.MapFS{
			"000001_create_widgets.up.sql":   &fstest.MapFile{Data: []byte("CREATE TABLE widgets (id INTEGER PRIMARY KEY);\n")},
			"000001_create_widgets.down.sql": &fstest.MapFile{Data: []byte("DROP TABLE IF EXISTS widgets;\n")},
		}

		db, err := storage.OpenDB(dbPath)
		if err != nil {
			t.Fatalf("OpenDB: %v", err)
		}
		defer db.Close()

		if err := storage.RunMigrationsFS(db, fs.FS(memFS)); err != nil {
			t.Fatalf("RunMigrationsFS: %v", err)
		}

		var name string
		if err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='widgets'").Scan(&name); err != nil {
			t.Fatalf("table widgets not found: %v", err)
		}
	})
}
