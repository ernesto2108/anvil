# Guía de Observabilidad

Tres pilares: health checks, métricas, trazas. Combinados con logging estructurado (ver `slog-guide.md`), te dan visibilidad completa de los sistemas en producción.

## Endpoints de Health Check

Requeridos para los probes de liveness y readiness de Kubernetes.

```go
// /healthz — liveness: "¿está vivo el proceso?"
// Si falla, Kubernetes reinicia el pod
mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
})

// /readyz — readiness: "¿puede recibir tráfico?"
// Si falla, Kubernetes deja de enviar tráfico a este pod
mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()

    if err := db.PingContext(ctx); err != nil {
        http.Error(w, "database not ready", http.StatusServiceUnavailable)
        return
    }
    // Agrega más checks: caché, servicios externos, etc.
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
})
```

**Reglas:**
- `/healthz` debe ser rápido y simple — sin dependencias externas
- `/readyz` debe verificar dependencias críticas (DB, caché, message broker)
- Ambos deben responder en menos de 2 segundos
- No expongas detalles internos en las respuestas de health check en producción

## Métricas (Prometheus — Método RED)

RED = Rate, Errors, Duration — las métricas mínimas para cualquier servicio.

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define métricas
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

// Expone el endpoint /metrics
mux.Handle("GET /metrics", promhttp.Handler())

// Instrumenta vía middleware (ver middleware-guide.md)
func MetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

        next.ServeHTTP(wrapped, r)

        duration := time.Since(start).Seconds()
        status := strconv.Itoa(wrapped.status)
        path := r.Pattern // Go 1.22+ — usa el patrón matched, no el path crudo

        httpRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
        httpRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
    })
}
```

**Reglas:**
- Instrumenta en middleware, no en handlers individuales — es una preocupación transversal
- Usa `r.Pattern` (Go 1.22+) no `r.URL.Path` — evita labels de alta cardinalidad
- Sigue RED: tasa de requests, tasa de errores, distribución de duración
- Agrega métricas de negocio con moderación: `orders_created_total`, `payments_processed_total`
- Nunca pongas IDs de usuario o valores de alta cardinalidad en labels

## Trazas (OpenTelemetry)

El distributed tracing sigue una request a través de los servicios.

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("service-name")

// Crea spans en los métodos del servicio
func (s *UserService) Create(ctx context.Context, input CreateUserInput) (*User, error) {
    ctx, span := tracer.Start(ctx, "UserService.Create")
    defer span.End()

    // Agrega atributos para debugging
    span.SetAttributes(attribute.String("user.email", input.Email))

    user, err := s.repo.Insert(ctx, input)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, fmt.Errorf("create user: %w", err)
    }

    return user, nil
}

// El repositorio también crea spans hijo
func (r *UserRepository) Insert(ctx context.Context, input CreateUserInput) (*User, error) {
    ctx, span := tracer.Start(ctx, "UserRepository.Insert")
    defer span.End()

    // Operación DB — el span captura la duración automáticamente
    // ...
}
```

**Integraciones APM existentes:** Si el proyecto usa Elastic APM, el patrón es el mismo:
```go
// Equivalente en Elastic APM
span, ctx := apm.StartSpan(ctx, "UserService.Create", "app")
defer span.End()
```

**Reglas:**
- Siempre propaga `context.Context` — lleva la traza
- Nombra los spans como `Package.Method` para claridad en los trace viewers
- Agrega atributos que ayuden al debugging — pero nunca PII
- Registra errores en spans con `span.RecordError(err)`
- Crea spans en los límites de servicio y operaciones costosas, no en cada función

## Correlación de Logs

Conecta logs, métricas y trazas con identificadores compartidos:

```go
// Agrega trace_id y request_id a cada línea de log
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

Esto te permite saltar de una línea de log → traza → métricas en tu plataforma de observabilidad.

## Graceful Shutdown

Vacía los spans y métricas pendientes antes de que el proceso termine:

```go
func main() {
    // Configura el exporter de trazas
    tp, err := initTracerProvider()
    if err != nil {
        log.Fatal(err)
    }

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    // ... inicia el servidor ...

    <-ctx.Done()

    // Graceful shutdown — vacía la telemetría
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

Sin graceful shutdown, los últimos segundos de spans y métricas se pierden — dificultando el debugging del problema que causó el shutdown.
