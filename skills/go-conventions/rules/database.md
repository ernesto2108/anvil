# Database Patterns

These patterns reflect the real DB access layer used across projects:

- **Query functions in `queries/` package** — each query is a function returning `(string, []any, error)`. Complex queries use `strings.Builder` with parameterized `$N` placeholders
- **Persistence DTOs** — structs with `sql.Null*` for ALL fields (`NullString`, `NullInt64`, `NullFloat64`, `NullTime`), separate from domain entities. Live in `dto/` or `persistence/` package
- **Mappers** — `ToBusiness()` method on single DTO, `NewToBusiness()` batch function for slices. Extract `.String`, `.Int64`, `.Time` etc. from `sql.Null*` fields
- **Repository struct** — holds `client` (custom DB interface) + `timeout time.Duration`. Every method calls `context.WithTimeout(ctx, r.timeout)` with deferred cancel
- **DB interface** — custom `PostgresSql` interface wrapping `*sql.DB` with own `Rows` interface for testability. Never depend on `*sql.DB` directly in repositories
- **Error translation** — `PostgresError(err)` translates `pq.Error` codes to domain errors (duplicate key → conflict, foreign key → not found, etc.)
- **Transactions** — `BeginTx` + deferred rollback + explicit commit at the end. Rollback is no-op after commit
- **Two DTO layers** — HTTP/input DTOs (json tags, binding tags) vs persistence/output DTOs (sql.Null* fields). Never mix them

Repository method flow: `WithTimeout → query() → execute → scan into DTO → map to domain`

See `examples/good-patterns.md` for complete code examples.

## Migrations — `golang-migrate`

Use `github.com/golang-migrate/migrate/v4` for all database migrations. Never write ad-hoc `CREATE TABLE` or schema DDL in application code.

- **Embed migrations** with `//go:embed migrations/*.sql` so they ship inside the binary
- **File naming:** `<number>_<action>_<target>.up.sql` / `.down.sql` (e.g., `000001_create_users.up.sql`)
- **One change per pair** — a single up/down pair does one logical thing
- **DBA agent owns migration files** — the developer does NOT create or modify `.up.sql`/`.down.sql` files. If your task needs schema changes, tell the orchestrator to invoke the DBA first
- **Engine-specific rules** — see `/db-engines` for PRAGMAs, driver setup, and migration quirks per engine
