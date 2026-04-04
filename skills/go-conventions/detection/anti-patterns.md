# Go Anti-Patterns — Detection Reference

## Passive Detection

When reviewing Go code, scan for these patterns and report using the format:
`[file:line] [severity] [category] anti-pattern-name`

Only report `error` and `warning` by default. Report `suggestion` only when user asks to improve/refactor/optimize.

## Anti-Pattern Table

| Code Pattern | Anti-Pattern | Severity | Category | Fix → Pattern |
|---|---|---|---|---|
| `panic()` outside `main()` | panic-in-library | error | reliability | Return `error` — see Error Handling |
| `init()` doing real work | hidden-init | error | reliability | Constructor injection — see Patterns > Constructor Functions |
| `_ = f.Close()` or error ignored | ignored-error | error | reliability | Handle or log — see Error Handling |
| `var db *sql.DB` at package level | global-mutable-state | error | concurrency | Inject via constructor — see Architecture Rules #7 |
| `setInterval`/ticker without stop | resource-leak | error | memory | `defer ticker.Stop()` in setup |
| `sync.Mutex` protecting channel ops | double-sync | error | concurrency | Use one mechanism — see Concurrency |
| `defer` in loop body | deferred-in-loop | error | memory | Close explicitly in loop body |
| Bare `return err` from raw/untyped errors | unwrapped-error | warning | errors | Map to `errors.New(errors.SomeCode)` or `fmt.Errorf("op: %w", err)` in standalone pkgs — see Error Handling |
| `fmt.Errorf` wrapping an already-mapped `*Errors` | double-wrapped-error | warning | errors | Just `return err` — infra already mapped it — see Error Handling |
| `context.Background()` in handlers | missing-context | warning | reliability | Use `r.Context()` — see Context |
| `context.Context` stored in struct | stored-context | warning | reliability | Pass as first param — see Context |
| `interface{}` / `any` in domain | untyped-domain | warning | types | Concrete types or generics — see domain-entity-guardrails |
| God interface (>5 methods) | god-interface | warning | design | Small interfaces by consumer — see Architecture Rules #2 |
| Circular imports | circular-import | warning | design | Extract shared interface — see Architecture Rules #5 |
| >3 levels of if/else nesting | deep-nesting | warning | readability | Guard clauses — see Patterns > Guard Clauses |
| Long function (>5 params) | param-bloat | warning | readability | Options struct — see Patterns > Functional Options |
| Tests without `t.Run()` subtests | flat-tests | warning | testing | Table-driven with subtests — see Testing Patterns |
| Tests depending on execution order | coupled-tests | warning | testing | Independent setup per test — see Testing Patterns |
| Shared test state between `t.Run` | shared-test-state | warning | testing | Each subtest creates own fixtures |
| `time.Sleep` in tests | sleep-in-tests | warning | testing | Channels, sync, or polling with timeout |
| Missing `t.Helper()` on helpers | missing-t-helper | suggestion | testing | Add `t.Helper()` first line |
| Exporting unused symbols | over-export | suggestion | design | Unexport what's not used externally |
| `log` package instead of `slog` | unstructured-logging | suggestion | observability | Use `slog` (stdlib) |
| String typing for enums | string-enum | suggestion | types | `type Status int` with `iota` |
| `reflect` for simple tasks | unnecessary-reflect | suggestion | performance | Type switches, generics, or concrete code |
| Error strings with capital/period | error-format | suggestion | style | Lowercase, no trailing punctuation |
| Package name `users` or `user_service` | bad-package-name | suggestion | style | Short, singular, no underscores: `user` |
| `math/rand` for tokens/keys/sessions | insecure-random | error | security | Use `crypto/rand` — see security-guide.md |
| `fmt.Sprintf` with user input in SQL | sql-injection | error | security | Parameterized queries with `$N` — see security-guide.md |
| No request body size limit | missing-body-limit | warning | security | `http.MaxBytesReader(w, r.Body, limit)` — see security-guide.md |
| No `/healthz` or `/readyz` endpoint | missing-health-check | warning | observability | Add liveness + readiness — see observability-guide.md |
| `os.Getenv` deep in call stack | scattered-config | warning | design | Load config once in `main()`, inject via constructor |
| Logging PII/secrets (password, token) | logged-secrets | error | security | Implement `LogValuer`, redact fields — see slog-guide.md |
| `strings.TrimSpace` + `== ""` checks in service layer | validation-in-service | warning | architecture | Move to `entity.Validate()` method — see Architecture Rules #8 |
| Service method with >2 field validations before business logic | scattered-validation | warning | architecture | Create input entity with `Validate()` — see Architecture Rules #8 |
| `g.Param("id")` with inline string literal | magic-param-string | warning | architecture | Use `dto.ParamXxx` constant from `dto/constants.go` — see Architecture Rules #9 |
| `g.Query("status")` with inline string literal | magic-query-string | warning | architecture | Use `dto.QueryXxx` constant from `dto/constants.go` — see Architecture Rules #9 |
| TrimSpace + empty check for URL path params in application layer | param-validation-wrong-layer | warning | architecture | Validate path params in handler, not service — see Architecture Rules #9 |
| `errors.WithMessage("foo: " + variable)` string concatenation | concat-in-error-message | warning | style | Use `fmt.Sprintf("foo: %s", variable)` — never concatenate with `+` in error messages |
| `http.Get()`/`http.Post()` (no timeout, no context) | http-no-timeout | error | reliability | `http.NewRequestWithContext(ctx, ...)` — see context-cleanup-guide.md |
| `http.DefaultClient` without Timeout | default-client | warning | reliability | `&http.Client{Timeout: 15*time.Second}` — see context-cleanup-guide.md |
| `db.Query()` without Context | query-no-context | error | reliability | `db.QueryContext(ctx, ...)` — see context-cleanup-guide.md |
| `Query()` for INSERT/UPDATE/DELETE | query-for-exec | error | memory | Use `ExecContext()` — Query returns rows that must be closed |
| Missing `rows.Close()` after QueryContext | unclosed-rows | error | memory | `defer rows.Close()` immediately after error check |
| Missing `rows.Err()` after iteration loop | unchecked-rows-err | warning | reliability | Always check `rows.Err()` after `for rows.Next()` |
| `defer` in a loop body | defer-in-loop-cleanup | error | memory | Extract to helper function — see context-cleanup-guide.md |
| Missing `sql.DB` pool config (MaxOpenConns) | no-pool-config | warning | reliability | Set MaxOpenConns, MaxIdleConns, ConnMaxLifetime |
| `resp.Body.Close()` before error check | close-before-check | error | crashes | Check error first, then `defer resp.Body.Close()` |
| `func AsError(err error, target interface{}) bool { return errors.As(err, &target) }` | errors-as-double-pointer | error | crashes | `target` is already a pointer inside `interface{}` — `&target` creates `*interface{}` which `errors.As` cannot unwrap. Use `errors.As(err, &customErr)` directly at the call site, never wrap it |
| External service error discarded: `if err != nil { return errors.New(domainErr) }` without logging `err` | swallowed-external-error | warning | observability | Log the original error before returning domain error: `log.Error("service X failed", log.WithError(err))` — otherwise debugging is impossible |
| Service interface method receives >1 raw `string` parameter (e.g., `GetStats(ctx, tenantID string, status string)`) | service-accepts-raw-strings | warning | architecture | Create a request/filter entity struct with `Validate()`. Service receives the entity, not raw strings. See `examples/service-contracts.md` — matches `DTO → Entity.ToBusiness() → Service(entity) → entity.Validate()` flow |
| Type named `*DTO` inside `domain/entities/` package | dto-in-domain | warning | naming | `DTO` belongs at transport boundaries (`http/dto/`, `psql/dto/`). Domain uses descriptive names: `Detail`, `Summary`, `Filter` — see Architecture Rules #11 |
| Domain aggregate struct duplicates fields from child entities instead of composing them | flattened-aggregate | warning | architecture | Compose with existing entity types: `OrderDetail{Order, []Item}` not field-by-field copy — see Architecture Rules #10 |
| Persistence DTO struct (in `infrastructure/output/persistencia/*/dto/`) uses plain `string`, `int`, or `time.Time` instead of `sql.Null*` types | dto-without-sql-null | warning | architecture | ALL fields in persistence DTOs must use `sql.NullString`, `sql.NullInt64`, `sql.NullTime`, etc. Mapper's `ToBusiness()` extracts actual values. Plain types cause silent zero-value bugs on NULL from JOINs, COALESCE, or GROUPING SETS |
