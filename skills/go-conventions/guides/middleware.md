# Guía de HTTP Middleware

## El Patrón

Todo middleware tiene la misma firma:

```go
type Middleware func(http.Handler) http.Handler
```

Recibe un handler, retorna un handler. Como capas de una cebolla — cada una envuelve a la siguiente.

```
Request → Recovery → Logging → Auth → Tu Handler → Response
```

## Wrapper de responseWriter

La mayoría de middleware necesita capturar el código de status de la respuesta. El `http.ResponseWriter` del stdlib no lo expone, así que envuélvelo:

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

## Middleware Comunes

### Recovery — captura panics, previene el crash del servidor

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

**Siempre el más externo.** Si el recovery está dentro del logging, un panic en el middleware de logging crashea el servidor.

### Request ID — correlaciona logs de la misma request

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

### Logging — registra cada request con método, path, status y latencia

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

### Auth — valida token, inyecta usuario en el contexto

```go
func AuthMiddleware(auth TokenValidator) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")
            if token == "" {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            // Elimina el prefijo "Bearer "
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

// Helper para extraer el usuario del contexto en los handlers
func UserFromContext(ctx context.Context) (*User, bool) {
    user, ok := ctx.Value(userKey).(*User)
    return user, ok
}
```

### CORS — maneja requests cross-origin

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

## Encadenamiento de Middleware

Aplica en orden: el primero listado = capa más externa.

```go
func Chain(h http.Handler, mws ...Middleware) http.Handler {
    for i := len(mws) - 1; i >= 0; i-- {
        h = mws[i](h)
    }
    return h
}

// Uso — el orden importa:
// 1. Recovery (el más externo — captura todo)
// 2. Request ID (asignar antes del logging)
// 3. Security headers
// 4. Logging (registra después de que el handler completa)
// 5. Metrics
// 6. Auth (antes de la lógica de negocio)
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

Go 1.22 agregó enrutamiento por patrones a la librería estándar — para la mayoría de APIs ya no necesitas Gin, Chi o Echo:

```go
mux := http.NewServeMux()

// Enrutamiento por método + patrón (Go 1.22+)
mux.HandleFunc("GET /users/{id}", handler.GetUser)
mux.HandleFunc("POST /users", handler.CreateUser)
mux.HandleFunc("PUT /users/{id}", handler.UpdateUser)
mux.HandleFunc("DELETE /users/{id}", handler.DeleteUser)

// Acceder a parámetros del path
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id") // Go 1.22+
    // ...
}
```

**Cuándo usar un router de terceros:** grupos de rutas con middleware compartido, matching complejo de paths, middleware por grupo. Si el proyecto ya usa Gin/Chi, sigue la convención del proyecto.

## Testing de Middleware

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
