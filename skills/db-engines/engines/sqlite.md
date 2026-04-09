# SQLite

## Limitations
- `ALTER TABLE` can only `ADD COLUMN` — no DROP, no RENAME (before 3.35), no ALTER TYPE
- No `ENUM` type — use `CHECK` constraints: `status TEXT CHECK(status IN ('active','inactive'))`
- No concurrent index creation — no `CONCURRENTLY` option
- No native `UUID` type — store as `TEXT`
- No `TIMESTAMP` type — store as `TEXT` in RFC3339 format
- `AUTOINCREMENT` is optional and slower than `INTEGER PRIMARY KEY` (which auto-increments by default)
- For schema changes beyond ADD COLUMN: **4-step pattern** — create new table, copy data, drop old, rename

## Production PRAGMAs (MANDATORY on connection open)

Apply these in order on every new connection:

```sql
PRAGMA journal_mode = WAL;          -- concurrent reads, non-blocking writes
PRAGMA synchronous = NORMAL;        -- safe in WAL mode, avoids FSYNC on every write
PRAGMA busy_timeout = 5000;         -- wait up to 5s on lock contention instead of failing
PRAGMA foreign_keys = ON;           -- enforce FK constraints (OFF by default!)
PRAGMA cache_size = -65536;         -- 64MB page cache (default ~2MB is too small)
PRAGMA temp_store = MEMORY;         -- temp tables/indices in memory
PRAGMA mmap_size = 268435456;       -- 256MB memory-mapped I/O for read performance
```

## WAL Mode Details
- WAL enables concurrent reads while a write is in progress — critical for any multi-goroutine or multi-process access
- Auto-checkpoint at 1000 pages (default) — sufficient for CLI/desktop tools
- For high-write workloads: `PRAGMA wal_autocheckpoint = <pages>` or manual `PRAGMA wal_checkpoint(TRUNCATE)` on graceful shutdown
- Always checkpoint before backup (`sqlite3 .backup`)
- WAL mode persists across connections — set it once, it stays

## Data Rotation Patterns
- Use `ON DELETE CASCADE` on all child FKs so a single `DELETE FROM parent` cleans children
- Count-based rotation: `DELETE FROM parent WHERE id NOT IN (SELECT id FROM parent ORDER BY created_at DESC LIMIT ?)` — predictable DB size
- Run `PRAGMA optimize` periodically (on app shutdown) to maintain query planner stats
- `VACUUM` reclaims space after large deletes — but locks the DB, so run during maintenance windows or on shutdown

## Migration Quirks
- Migrations must NOT contain explicit `BEGIN` / `COMMIT` — `golang-migrate` and most tools wrap each file in an implicit transaction
- `CREATE TABLE IF NOT EXISTS` for idempotent first migration
- `ON DELETE CASCADE` must be declared at table creation — cannot be added later via ALTER
- Include `CREATE INDEX` statements in the same migration as the table they index
- Foreign keys are NOT enforced unless `PRAGMA foreign_keys = ON` — this is per-connection, not persistent
- When using `golang-migrate` with `WithInstance(db, ...)`: do NOT call `m.Close()` — it closes the shared `*sql.DB` connection. Let the caller manage the connection lifecycle
- When tables have FK relationships, order migrations so parent tables are created first (e.g., `000001_create_runs`, `000002_create_agents` where agents references runs)

## Migration sources: `iofs` vs `file://` (CRITICAL for shipped binaries)

`golang-migrate` supports multiple sources. The choice matters a lot for distribution.

### `file://` source
- Reads `.sql` files from disk at runtime
- Works only when migrations live on the host filesystem — typically during tests or local dev
- **Fails in any binary distributed without the repo** — the user does not have `./migrations/` on their machine

```go
m, err := migrate.NewWithDatabaseInstance(
    "file://"+migrationsPath, "sqlite3", driver,
)
```

### `iofs` source (required for shipped binaries)
- Reads migrations from any `fs.FS` — including `embed.FS`
- The migrations become part of the binary via `//go:embed`
- This is the ONLY source that works for a binary the end-user downloads

```go
import (
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/sqlite3"
    "github.com/golang-migrate/migrate/v4/source/iofs"
)

// migrations is an embed.FS rooted at the migrations directory
src, err := iofs.New(migrations, ".")
if err != nil {
    return fmt.Errorf("create iofs source: %w", err)
}

driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
if err != nil {
    return fmt.Errorf("create sqlite3 driver: %w", err)
}

m, err := migrate.NewWithInstance("iofs", src, "sqlite3", driver)
if err != nil {
    return fmt.Errorf("create migrate instance: %w", err)
}
```

### Rule — design for iofs from day one
If the store will ever be distributed in a binary (CLI tool, desktop app, server), **use `iofs` from the first migration**. Offering BOTH sources is fine (tests can keep using `file://` for ergonomics), but the default production path must be `iofs`.

**Recommended pattern — two constructors sharing a helper:**
```go
// Shared DB setup — no duplication
func openDB(dbPath string) (*sql.DB, error) { /* ... */ }

// Production: embedded migrations
func NewFS(dbPath string, migrations fs.FS, subPath string, maxRuns int) (*Store, error) {
    db, err := openDB(dbPath)
    if err != nil {
        return nil, err
    }
    sub, err := fs.Sub(migrations, subPath)
    if err != nil {
        _ = db.Close()
        return nil, fmt.Errorf("resolve sub-fs: %w", err)
    }
    if err := RunMigrationsFS(db, sub); err != nil {
        _ = db.Close()
        return nil, err
    }
    return &Store{db: db}, nil
}

// Tests / dev: disk migrations
func New(dbPath, migrationsPath string, maxRuns int) (*Store, error) { /* uses file:// */ }
```

### Anti-pattern — `file://` in production code
If the CLI command opens the store with `store.New(dbPath, filepath.Join(home, ".app", "migrations"), ...)`, the binary will fail at first run for any user that does not have a separately-installed migrations directory. This is a **design gap**, not a bug — catch it in architecture review, not at runtime.

## Go Driver
- `github.com/mattn/go-sqlite3` — the only production-grade `database/sql` driver (requires CGO)
- Connection string: `file:<path>?_journal_mode=WAL&_foreign_keys=on&_busy_timeout=5000`
- Single writer pattern: open one `*sql.DB` with `SetMaxOpenConns(1)` for writes
- For read-heavy workloads: separate `*sql.DB` for reads with `SetMaxOpenConns(max(4, runtime.NumCPU()))`
