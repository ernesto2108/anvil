# PostgreSQL

## Types & Conventions
- Use `UUID` with `uuid_generate_v7()` or `gen_random_uuid()` for primary keys
- Use `TIMESTAMP WITH TIME ZONE` (not `DATETIME`, not `TIMESTAMP` without TZ)
- Use `ENUM` types or `CHECK` constraints for status columns
- Use `TEXT` over `VARCHAR(n)` unless you need the length constraint for business reasons
- `BOOLEAN` is native — use it directly

## Migration Patterns
- `CREATE INDEX CONCURRENTLY` for indexes on large tables (avoids table lock)
- `IF NOT EXISTS` / `IF EXISTS` for idempotent migrations
- `ALTER TABLE ... ADD COLUMN ... DEFAULT <value>` in a single statement (Postgres 11+ rewrites no rows)
- `BEGIN` / `COMMIT` explicit transactions are allowed and encouraged for multi-statement migrations

## Performance Tuning
- Connection pooling: use `pgxpool` or external pooler (PgBouncer) — never open unbounded connections
- `EXPLAIN ANALYZE` to verify index usage before and after changes
- Partial indexes: `CREATE INDEX ... WHERE deleted_at IS NULL` for soft-delete tables
- Covering indexes: `INCLUDE (col)` to avoid heap lookup
- `GROUPING SETS` for multiple aggregations in one pass

## Multi-Tenant
- Row Level Security (RLS) policies for tenant isolation
- `current_setting('app.tenant_id')` for RLS context
- Compound indexes lead with `tenant_id`: `idx_table_tenant_status`

## Go Drivers
- `github.com/jackc/pgx/v5` — preferred (pure Go, connection pooling via `pgxpool`)
- `github.com/lib/pq` — legacy, `database/sql` compatible
- Error translation: `pq.Error` or `pgconn.PgError` codes to domain errors (23505 = unique violation, 23503 = FK violation)
