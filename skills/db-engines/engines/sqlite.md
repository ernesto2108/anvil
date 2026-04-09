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

## Go Driver
- `github.com/mattn/go-sqlite3` — the only production-grade `database/sql` driver (requires CGO)
- Connection string: `file:<path>?_journal_mode=WAL&_foreign_keys=on&_busy_timeout=5000`
- Single writer pattern: open one `*sql.DB` with `SetMaxOpenConns(1)` for writes
- For read-heavy workloads: separate `*sql.DB` for reads with `SetMaxOpenConns(max(4, runtime.NumCPU()))`
