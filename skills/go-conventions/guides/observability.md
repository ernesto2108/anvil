# Observability Guide

Three pillars: health checks, metrics, tracing. Combined with structured logging (see `slog-guide.md`), these give you full visibility into production systems.

## Health Check Endpoints

Required for Kubernetes liveness and readiness probes.

```go
// /healthz — liveness: "is the process alive?"
// If this fails, Kubernetes restarts the pod
mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
})

// /readyz — readiness: "can it handle traffic?"
// If this fails, Kubernetes stops sending traffic to this pod
mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()

    if err := db.PingContext(ctx); err != nil {
        http.Error(w, "database not ready", http.StatusServiceUnavailable)
        return
    }
    // Add more checks: cache, external services, etc.
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
})
```

**Rules:**
- `/healthz` should be fast and simple — no external dependencies
- `/readyz` should check critical dependencies (DB, cache, message broker)
- Both should respond within 2 seconds
- Don't expose internal details in health check responses in production

## Metrics (Prometheus — RED Method)

RED = Rate, Errors, Duration — the minimum metrics for any service.

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define metrics
var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests by method, path, and status",
        },
        []string{"method", "path", "status"},
    )

    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
)

// Expose /metrics endpoint
mux.Handle("GET /metrics", promhttp.Handler())

// Instrument via middleware (see middleware-guide.md)
func MetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

        next.ServeHTTP(wrapped, r)

        duration := time.Since(start).Seconds()
        status := strconv.Itoa(wrapped.status)
        path := r.Pattern // Go 1.22+ — use matched pattern, not raw path

        httpRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
        httpRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
    })
}
```

**Rules:**
- Instrument middleware, not individual handlers — cross-cutting concern
- Use `r.Pattern` (Go 1.22+) not `r.URL.Path` — avoids high-cardinality labels
- Track RED: request rate, error rate, duration distribution
- Add business metrics sparingly: `orders_created_total`, `payments_processed_total`
- Never put user IDs or high-cardinality values in labels

## Tracing (OpenTelemetry)

Distributed tracing follows a request across services.

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("service-name")

// Create spans in service methods
func (s *UserService) Create(ctx context.Context, input CreateUserInput) (*User, error) {
    ctx, span := tracer.Start(ctx, "UserService.Create")
    defer span.End()

    // Add attributes for debugging
    span.SetAttributes(attribute.String("user.email", input.Email))

    user, err := s.repo.Insert(ctx, input)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, fmt.Errorf("create user: %w", err)
    }

    return user, nil
}

// Repository also creates child spans
func (r *UserRepository) Insert(ctx context.Context, input CreateUserInput) (*User, error) {
    ctx, span := tracer.Start(ctx, "UserRepository.Insert")
    defer span.End()

    // DB operation — span captures duration automatically
    // ...
}
```

**Existing APM integrations:** If the project uses Elastic APM, the pattern is the same:
```go
// Elastic APM equivalent
span, ctx := apm.StartSpan(ctx, "UserService.Create", "app")
defer span.End()
```

**Rules:**
- Always propagate `context.Context` — it carries the trace
- Name spans as `Package.Method` for clarity in trace viewers
- Add attributes that help debugging — but never PII
- Record errors on spans with `span.RecordError(err)`
- Create spans at service boundaries and expensive operations, not every function

## Log Correlation

Connect logs, metrics, and traces with shared identifiers:

```go
// Add trace_id and request_id to every log line
func (s *UserService) Create(ctx context.Context, input CreateUserInput) (*User, error) {
    ctx, span := tracer.Start(ctx, "UserService.Create")
    defer span.End()

    s.logger.InfoContext(ctx, "creating user",
        slog.String("email", input.Email),
        slog.String("trace_id", span.SpanContext().TraceID().String()),
    )
    // ...
}
```

This lets you jump from a log line → trace → metrics in your observability platform.

## Graceful Shutdown

Flush pending spans and metrics before the process exits:

```go
func main() {
    // Setup tracing exporter
    tp, err := initTracerProvider()
    if err != nil {
        log.Fatal(err)
    }

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    // ... start server ...

    <-ctx.Done()

    // Graceful shutdown — flush telemetry
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := tp.Shutdown(shutdownCtx); err != nil {
        logger.Error("failed to shutdown tracer", slog.Any("error", err))
    }
    if err := srv.Shutdown(shutdownCtx); err != nil {
        logger.Error("failed to shutdown server", slog.Any("error", err))
    }
}
```

Without graceful shutdown, the last few seconds of spans and metrics are lost — making it harder to debug the issue that caused the shutdown.
