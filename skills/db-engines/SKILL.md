---
name: db-engines
description: Engine-specific knowledge for PostgreSQL, SQLite, and MySQL. PRAGMAs, limitations, best practices, drivers, connection tuning, and migration quirks per engine. Load before writing migrations or optimizing queries.
---

# Database Engine Reference

> Engine-specific rules, limitations, and best practices. The DBA agent loads this skill to adapt migrations and schema decisions to the target engine.

## When to Use

- Before writing any migration — detect the engine first, then load the relevant reference
- When optimizing queries — engine-specific techniques differ
- When setting up a new database — connection tuning, PRAGMAs, pooling

## Engine Detection

Detect the engine from project signals (check in order):

1. Existing migration files — syntax reveals engine (e.g., `SERIAL` = Postgres, `AUTOINCREMENT` = SQLite)
2. Driver imports in code — `pq` / `pgx` = Postgres, `go-sqlite3` = SQLite, `go-sql-driver/mysql` = MySQL
3. Connection strings in config — `postgres://`, `file:*.db`, `mysql://`
4. `docker-compose.yml` or infra files — service image names
5. Architect's design doc — may specify engine choice

## Engine References

Load ONLY the engine(s) relevant to the current task:

- **PostgreSQL** → read `engines/postgresql.md`
- **SQLite** → read `engines/sqlite.md`
- **MySQL** → read `engines/mysql.md`

Each reference covers: types, limitations, migration patterns, performance tuning, multi-tenant patterns, and Go drivers.

## Migration Tooling by Language

| Language | Tool | Notes |
|----------|------|-------|
| Go | `github.com/golang-migrate/migrate/v4` | Embed with `//go:embed`, use `iofs` source driver. See `/go-conventions` for setup code |
| Node.js | `knex` or `prisma migrate` | |
| Python | `alembic` (SQLAlchemy) or `django.db.migrations` | |
| Rust | `sqlx migrate` or `diesel migrations` | |

## Quick Reference — Engine Comparison

| Feature | PostgreSQL | SQLite | MySQL |
|---------|-----------|--------|-------|
| UUID native | Yes | No (TEXT) | No (CHAR(36)) |
| ENUM type | Yes | No (CHECK) | Yes |
| ALTER TABLE DROP COLUMN | Yes | No (4-step) | Yes |
| Concurrent index | Yes | No | No |
| RLS | Yes | No | No |
| FK enforcement | Always on | Per-connection PRAGMA | Always on (InnoDB) |
| Transactions in migrations | Explicit ok | Implicit only | Explicit ok |
| Max concurrent writers | Many | One | Many |
| Recommended Go driver | pgx/v5 | mattn/go-sqlite3 | go-sql-driver/mysql |
