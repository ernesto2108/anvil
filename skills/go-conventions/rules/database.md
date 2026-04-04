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
