# MySQL

## Types & Conventions
- No native `UUID` — use `CHAR(36)` or `BINARY(16)` with app-level conversion
- Use `DATETIME` not `TIMESTAMP` (TIMESTAMP has 2038 problem and auto-updates)
- Use `VARCHAR(n)` with explicit lengths — MySQL enforces them
- `BOOLEAN` is alias for `TINYINT(1)` — 0/1, not true/false
- `TEXT` / `LONGTEXT` for large strings — `VARCHAR` max is 65,535 bytes

## Migration Patterns
- `ALTER TABLE ... ALGORITHM=INPLACE` when possible (avoids table copy)
- No `CREATE INDEX CONCURRENTLY` — index creation locks the table
- `IF NOT EXISTS` / `IF EXISTS` supported
- `ALTER TABLE ... ADD COLUMN` supports `AFTER <column>` for column ordering
- Online DDL: for large tables, consider `pt-online-schema-change` or `gh-ost`

## Performance Tuning
- InnoDB is the default and correct engine — never use MyISAM
- `innodb_buffer_pool_size` = 70-80% of available RAM for dedicated DB servers
- `EXPLAIN FORMAT=JSON` for detailed query analysis
- Covering indexes with `INCLUDE` not supported — put all needed columns in the index itself
- `FORCE INDEX(idx_name)` as last resort for wrong index selection

## Multi-Tenant
- No native RLS — enforce tenant isolation at application/query level
- Compound indexes: `(tenant_id, ...)` for all tenant-scoped queries
- Consider schema-per-tenant for strong isolation (adds operational complexity)

## Go Drivers
- `github.com/go-sql-driver/mysql` — standard `database/sql` driver
- Connection string: `user:pass@tcp(host:port)/dbname?parseTime=true&multiStatements=true`
- `parseTime=true` is MANDATORY — without it, `DATETIME` columns return `[]byte` instead of `time.Time`
- Error translation: `*mysql.MySQLError` with error numbers (1062 = duplicate, 1452 = FK violation)
