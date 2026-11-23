# Guía de Logging Estructurado

## Principios (aplicables a CUALQUIER logger: slog, logrus, zerolog, zap)

Estos principios aplican sin importar qué librería de logging use el proyecto:

1. **Pares clave-valor estructurados** — nunca concatenes strings en los mensajes de log
2. **Logger vía DI** — pasa el logger a través de constructores, nunca uses globals a nivel de package
3. **Nunca registres datos sensibles** — contraseñas, tokens, PII, tarjetas de crédito
4. **Nomenclatura consistente de claves** — usa snake_case para todas las claves de log en el proyecto
5. **Context-aware** — propaga request_id y trace_id a través de todos los logs
6. **Nivel correcto** — Debug (solo desarrollo), Info (operación normal), Warn (recuperable), Error (requiere acción)

## slog (recomendado para proyectos nuevos — stdlib Go 1.21+)

### Configuración

```go
// Producción — salida JSON para parseo por máquinas (Datadog, Elastic, Grafana)
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

// Desarrollo — salida de texto legible por humanos
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))
```

### Patrones de uso

```go
// Logging estructurado básico con helpers tipados
logger.Info("user created",
    slog.String("user_id", u.ID),
    slog.String("email", u.Email),
    slog.Duration("latency", elapsed),
)
// Salida JSON: {"time":"...","level":"INFO","msg":"user created","user_id":"123","email":"john@example.com","latency":"45ms"}

// Logging de errores con valor de error
logger.Error("failed to create user",
    slog.String("email", email),
    slog.Any("error", err),
)

// Logging context-aware — propaga request_id, trace_id
logger.InfoContext(ctx, "processing order", slog.String("order_id", id))
```

### Inyecta el logger vía constructor

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

### LogValuer — controla qué se registra, redacta campos sensibles

```go
// Implementa slog.LogValuer para controlar cómo aparece un tipo en los logs
func (u User) LogValue() slog.Value {
    return slog.GroupValue(
        slog.String("id", u.ID),
        slog.String("email", u.Email),
        // password, token deliberadamente omitidos
    )
}

// Ahora puedes registrar el usuario completo de forma segura
logger.Info("user authenticated", slog.Any("user", user))
// La salida incluye id y email pero NO el password
```

### Agrega contexto de request (patrón middleware)

```go
func RequestIDMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            requestID := r.Header.Get("X-Request-ID")
            if requestID == "" {
                requestID = uuid.NewString()
            }
            // Crea un logger hijo con request_id incorporado
            reqLogger := logger.With(slog.String("request_id", requestID))
            ctx := context.WithValue(r.Context(), loggerKey, reqLogger)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## logrus / zerolog / zap (proyectos existentes)

Si el proyecto ya usa un logger, sigue los mismos principios:

```go
// logrus — campos estructurados
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

Revisa el `go.mod` del proyecto para determinar qué logger está en uso. Sigue los patrones existentes del proyecto — no introduzcas slog en un proyecto que ya usa logrus a menos que estés migrando.

## Anti-patrones

| Patrón | Problema | Solución |
|---|---|---|
| `log.Println("user: " + userID)` | No estructurado, no parseable | Usa pares clave-valor |
| `slog.SetDefault(logger)` como única configuración | Estado global | Pasa vía constructor |
| `logger.Debug(...)` en hot loops | Impacto en rendimiento incluso si está deshabilitado | Verifica `logger.Enabled()` primero |
| Registrar password, token o número de tarjeta | Brecha de seguridad vía logs | Implementa LogValuer, redacta campos |
| Nombres de clave distintos para el mismo concepto | Imposible correlacionar | Estandariza: `user_id` no `userId`/`uid`/`user` |
