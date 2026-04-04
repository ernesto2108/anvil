# Coding Rules

## Error Handling

### Centralized error catalog (`pkg/errors`)

Use a centralized `pkg/errors` package to define all domain error codes with their HTTP/gRPC mappings. Each error is a typed struct (`*Errors`) with `Code`, `Message`, `HttpErrorCode`, and `GrpcErrorCode`. The error middleware extracts `*Errors` via `errors.As` and responds with the correct status code automatically.

```go
// pkg/errors/error-definition.go
const (
    BadRequestErr   DomainErrorCode = "BAD_REQUEST_ERR"
    UnauthorizedErr DomainErrorCode = "UNAUTHORIZED_ERR"
    NotFoundErr     DomainErrorCode = "NOT_FOUND_ERR"
)

// Each code maps to an Errors struct with HTTP code, gRPC code, and message
```

### Error flow by layer

**Infrastructure** — maps external errors (DB, cache, HTTP) to domain error codes:

```go
// good: infrastructure maps raw errors to domain errors
switch {
case err == nil:
    return user.ToBusiness(), nil
case utils.IsError(err, sql.ErrNoRows):
    return entities.User{}, errors.New(errors.UserNotFoundErr)
default:
    return entities.User{}, errors.New(errors.ScanErr)
}
```

**Application** — uses domain errors for business conditions, propagates infra errors as-is:

```go
// good: infra already mapped the error — just propagate
user, err := s.db.GetByEmail(ctx, request.Email)
if err != nil {
    return err  // already *errors.Errors from infra layer
}

// good: business condition — create a new domain error
if !match {
    return errors.New(errors.UnauthorizedErr)
}

// good: unexpected error from a utility — map to InternalErr
hash, err := utils.VerifyPasswordHash(stored, input)
if err != nil {
    return errors.New(errors.InternalErr)
}
```

**Handler/HTTP** — passes errors to gin context, middleware handles the rest:

```go
// good: handler just forwards errors — middleware resolves HTTP codes
user, err := h.svc.SignIn(ctx, req)
if err != nil {
    g.Errors = append(g.Errors, g.Error(err))
    return
}
```

### When to use `fmt.Errorf` vs `errors.New`

| Situation | Use | Why |
|-----------|-----|-----|
| Known business/domain condition | `errors.New(errors.SomeCode)` | Middleware maps code → HTTP status automatically |
| Error already mapped by lower layer | `return err` | Don't re-wrap what's already a `*Errors` |
| Truly unexpected internal failure | `errors.New(errors.InternalErr)` | Don't leak internal details to the client |
| Standalone library / pkg with no error catalog | `fmt.Errorf("context: %w", err)` | No domain errors available, wrap for traceability |
| Error message with dynamic values | `errors.WithMessage(fmt.Sprintf("msg: %s", v))` | Always use `fmt.Sprintf` for messages with variables — never string concatenation with `+` |

**Anti-pattern**: wrapping an already-mapped `*Errors` with `fmt.Errorf` — it adds noise and can confuse middleware that checks error types.

**Anti-pattern**: `errors.WithMessage("prefix: " + variable)` — use `fmt.Sprintf("prefix: %s", variable)` instead. String concatenation in error messages is harder to read and inconsistent with Go idioms.

```go
// bad: GetByEmail already returns *Errors (UserNotFoundErr or ScanErr)
user, err := s.db.GetByEmail(ctx, req.Email)
if err != nil {
    return fmt.Errorf("sign in: %w", err)  // wraps already-typed error
}

// good: just propagate
if err != nil {
    return err
}
```

- Never `panic` in library/application code — only in `main()` for unrecoverable bootstrap failures
- Check errors immediately — never defer error checking

## Naming

- **Receivers**: short, 1-2 letter, consistent across methods (`func (s *Server)`, not `func (server *Server)`)
- **Interfaces**: verb-based, small (`Reader`, `Validator`, not `UserServiceInterface`)
- **Constructors**: `NewXxx` returns concrete type, not interface
- **Unexported by default** — only export what the package API needs
- **Acronyms**: all caps (`HTTP`, `ID`, `URL`), not `Http`, `Id`, `Url`
- **Package names**: short, lowercase, no underscores, singular (`user`, not `users` or `user_service`)

## Context & Resource Cleanup

- Always first parameter: `func DoThing(ctx context.Context, ...)`
- Set timeouts on every external call (HTTP, DB, Redis, gRPC)
- Never store `context.Context` in structs
- Never use `http.Get()` / `http.Post()` / `http.DefaultClient` — create client with `Timeout`
- Always `defer rows.Close()` immediately after `QueryContext` error check
- Always `defer resp.Body.Close()` immediately after `client.Do()` error check
- Always `defer cancel()` after `context.WithTimeout` / `context.WithCancel`
- Configure `sql.DB` pool: `MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`
- See `guides/cleanup/` for full patterns, multi-level timeouts, and connection pool config

## Concurrency

- Protect shared state with `sync.Mutex` or channels — choose one per resource
- Prefer channels for coordination, mutexes for state protection
- Always handle goroutine lifecycle — no fire-and-forget
- Use `errgroup` for parallel work with error propagation
- Run tests with `-race` flag always
