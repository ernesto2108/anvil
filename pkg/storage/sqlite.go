package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"
)

// sqliteVecOnce ensures sqlite_vec.Auto() is called exactly once per process.
// Auto() registers an autoextension hook so every connection opened afterward
// has the vec0 module available — calling it multiple times is harmless but
// cheaper to gate.
var sqliteVecOnce sync.Once

// OpenDB creates the database file (with restrictive permissions) and opens a
// *sql.DB with WAL journal mode and foreign keys enabled. The caller is
// responsible for closing the returned *sql.DB.
func OpenDB(dbPath string) (*sql.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("storage: crear directorio de base de datos: %w", err)
	}

	// Pre-create the file with 0600 permissions to avoid a window where the
	// file exists with umask-derived permissions (typically 0644).
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		f, err := os.OpenFile(dbPath, os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			return nil, fmt.Errorf("storage: crear archivo de base de datos: %w", err)
		}
		_ = f.Close()
	}

	// Register sqlite-vec as an autoextension before opening the connection.
	// Once Auto() runs, every sql.Open("sqlite3", …) in the process gets vec0
	// available — needed by the digests vector search path.
	sqliteVecOnce.Do(sqlite_vec.Auto)

	// busy_timeout avoids SQLITE_BUSY when anvil-emit writes concurrently.
	dsn := dbPath + "?_busy_timeout=5000&_journal_mode=WAL"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("storage: abrir base de datos: %w", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("storage: configurar foreign_keys=ON: %w", err)
	}

	// Allow up to 4 concurrent connections for parallel reads.
	// SQLite WAL mode supports concurrent readers. Writes still serialize
	// via SQLite's internal locking + busy_timeout.
	db.SetMaxOpenConns(4)

	// Recycle idle connections quickly so each new query re-reads the WAL
	// index and sees rows written by external processes (anvil emit).
	db.SetConnMaxIdleTime(2 * time.Second)
	db.SetMaxIdleConns(2)

	return db, nil
}
