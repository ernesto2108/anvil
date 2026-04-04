# HTTP Middleware Guide

## The Pattern

Every middleware has the same signature:

```go
type Middleware func(http.Handler) http.Handler
```

Receives a handler, returns a handler. Like layers of an onion — each wraps the next.

```
Request → Recovery → Logging → Auth → Your Handler → Response
```

## responseWriter Wrapper

Most middleware needs to capture the response status code. The stdlib `http.ResponseWriter` doesn't expose it, so wrap it:

```go
type responseWriter struct {
    http.ResponseWriter
    status int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.status = code
    rw.ResponseWriter.WriteHeader(code)
}
```

## Common Middleware

### Recovery — catch panics, prevent server crash

```go
func RecoveryMiddleware(logger *slog.Logger) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if rec := recover(); rec != nil {
                    logger.Error("panic recovered",
                        slog.Any("panic", rec),
                        slog.String("stack", string(debug.Stack())),
                        slog.String("path", r.URL.Path),
                    )
                    http.Error(w, "internal server error", http.StatusInternalServerError)
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}
```

**Always outermost.** If recovery is inside logging, a panic in logging middleware crashes the server.

### Request ID — correlate logs for the same request

```go
func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.NewString()
        }
        ctx := context.WithValue(r.Context(), requestIDKey, requestID)
        w.Header().Set("X-Request-ID", requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Logging — log every request with method, path, status, latency

```go
func LoggingMiddleware(logger *slog.Logger) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

            next.ServeHTTP(wrapped, r)

            logger.Info("request",
                slog.String("method", r.Method),
                slog.String("path", r.URL.Path),
                slog.Int("status", wrapped.status),
                slog.Duration("latency", time.Since(start)),
                slog.String("remote", r.RemoteAddr),
            )
        })
    }
}
```

### Auth — validate token, inject user into context

```go
func AuthMiddleware(auth TokenValidator) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")
            if token == "" {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            // Strip "Bearer " prefix
            token = strings.TrimPrefix(token, "Bearer ")

            user, err := auth.Validate(r.Context(), token)
            if err != nil {
                http.Error(w, "invalid token", http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), userKey, user)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Helper to extract user from context in handlers
func UserFromContext(ctx context.Context) (*User, bool) {
    user, ok := ctx.Value(userKey).(*User)
    return user, ok
}
```

### CORS — handle cross-origin requests

```go
func CORSMiddleware(allowedOrigins []string) Middleware {
    allowed := make(map[string]bool, len(allowedOrigins))
    for _, o := range allowedOrigins {
        allowed[o] = true
    }

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")
            if allowed[origin] {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
                w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
                w.Header().Set("Access-Control-Max-Age", "86400")
            }

            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

## Chaining Middleware

Apply in order: first listed = outermost layer.

```go
func Chain(h http.Handler, mws ...Middleware) http.Handler {
    for i := len(mws) - 1; i >= 0; i-- {
        h = mws[i](h)
    }
    return h
}

// Usage — order matters:
// 1. Recovery (outermost — catches everything)
// 2. Request ID (assign before logging)
// 3. Security headers
// 4. Logging (logs after handler completes)
// 5. Metrics
// 6. Auth (before business logic)
mux := http.NewServeMux()
mux.HandleFunc("GET /users/{id}", handler.GetUser)
mux.HandleFunc("POST /users", handler.CreateUser)

wrapped := Chain(mux,
    RecoveryMiddleware(logger),
    RequestIDMiddleware,
    SecurityHeadersMiddleware,
    LoggingMiddleware(logger),
    MetricsMiddleware,
    AuthMiddleware(auth),
)

srv := &http.Server{Addr: ":8080", Handler: wrapped}
```

## stdlib ServeMux (Go 1.22+)

Go 1.22 added pattern routing to the standard library — for most APIs you no longer need Gin, Chi, or Echo:

```go
mux := http.NewServeMux()

// Method + pattern routing (Go 1.22+)
mux.HandleFunc("GET /users/{id}", handler.GetUser)
mux.HandleFunc("POST /users", handler.CreateUser)
mux.HandleFunc("PUT /users/{id}", handler.UpdateUser)
mux.HandleFunc("DELETE /users/{id}", handler.DeleteUser)

// Access path parameters
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id") // Go 1.22+
    // ...
}
```

**When to use a third-party router:** route groups with shared middleware, complex path matching, middleware per-group. If the project already uses Gin/Chi, follow the project convention.

## Testing Middleware

```go
func TestLoggingMiddleware(t *testing.T) {
    var buf bytes.Buffer
    logger := slog.New(slog.NewJSONHandler(&buf, nil))

    handler := LoggingMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    req := httptest.NewRequest("GET", "/test", nil)
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Errorf("got status %d, want 200", rec.Code)
    }
    if !strings.Contains(buf.String(), `"path":"/test"`) {
        t.Error("expected path in log output")
    }
}
```
