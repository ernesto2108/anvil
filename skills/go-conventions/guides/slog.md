# Structured Logging Guide

## Principles (apply to ANY logger: slog, logrus, zerolog, zap)

These principles apply regardless of which logging library the project uses:

1. **Structured key-value pairs** — never concatenate strings into log messages
2. **Logger via DI** — pass logger through constructors, never use package-level globals
3. **Never log sensitive data** — passwords, tokens, PII, credit cards
4. **Consistent key naming** — use snake_case for all log keys across the project
5. **Context-aware** — propagate request_id and trace_id through all logs
6. **Right level** — Debug (development only), Info (normal ops), Warn (recoverable), Error (action needed)

## slog (recommended for new projects — Go 1.21+ stdlib)

### Setup

```go
// Production — JSON output for machine parsing (Datadog, Elastic, Grafana)
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

// Development — human-readable text output
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))
```

### Usage patterns

```go
// Basic structured logging with typed helpers
logger.Info("user created",
    slog.String("user_id", u.ID),
    slog.String("email", u.Email),
    slog.Duration("latency", elapsed),
)
// JSON output: {"time":"...","level":"INFO","msg":"user created","user_id":"123","email":"john@example.com","latency":"45ms"}

// Error logging with error value
logger.Error("failed to create user",
    slog.String("email", email),
    slog.Any("error", err),
)

// Context-aware logging — propagates request_id, trace_id
logger.InfoContext(ctx, "processing order", slog.String("order_id", id))
```

### Inject logger via constructor

```go
type UserService struct {
    repo   UserRepository
    logger *slog.Logger
}

func NewUserService(repo UserRepository, logger *slog.Logger) *UserService {
    return &UserService{
        repo:   repo,
        logger: logger,
    }
}
```

### LogValuer — control what gets logged, redact sensitive fields

```go
// Implement slog.LogValuer to control how a type appears in logs
func (u User) LogValue() slog.Value {
    return slog.GroupValue(
        slog.String("id", u.ID),
        slog.String("email", u.Email),
        // password, token deliberately omitted
    )
}

// Now you can safely log the whole user
logger.Info("user authenticated", slog.Any("user", user))
// Output includes id and email but NOT password
```

### Add request context (middleware pattern)

```go
func RequestIDMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            requestID := r.Header.Get("X-Request-ID")
            if requestID == "" {
                requestID = uuid.NewString()
            }
            // Create child logger with request_id baked in
            reqLogger := logger.With(slog.String("request_id", requestID))
            ctx := context.WithValue(r.Context(), loggerKey, reqLogger)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## logrus / zerolog / zap (existing projects)

If the project already uses a logger, follow the same principles:

```go
// logrus — structured fields
logger.WithFields(logrus.Fields{
    "user_id": userID,
    "email":   email,
    "latency": elapsed,
}).Info("user created")

// zerolog — zero-allocation
log.Info().
    Str("user_id", userID).
    Str("email", email).
    Dur("latency", elapsed).
    Msg("user created")
```

Check the project's `go.mod` to determine which logger is in use. Follow the project's existing patterns — don't introduce slog into a project that already uses logrus unless migrating.

## Anti-patterns

| Pattern | Problem | Fix |
|---|---|---|
| `log.Println("user: " + userID)` | Unstructured, unparseable | Use key-value pairs |
| `slog.SetDefault(logger)` as only setup | Global state | Pass via constructor |
| `logger.Debug(...)` in hot loops | Performance hit even if disabled | Check `logger.Enabled()` first |
| Logging password, token, or card number | Security breach via logs | Implement LogValuer, redact fields |
| Different key names for same concept | Impossible to correlate | Standardize: `user_id` not `userId`/`uid`/`user` |
