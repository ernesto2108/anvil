# Reglas de Código

## Manejo de Errores

### Catálogo centralizado de errores (`pkg/errors`)

Usar un paquete `pkg/errors` centralizado para definir todos los códigos de error de dominio con sus mappings HTTP/gRPC. Cada error es un struct tipado (`*Errors`) con `Code`, `Message`, `HttpErrorCode`, y `GrpcErrorCode`. El middleware de errores extrae `*Errors` vía `errors.As` y responde con el código de estado correcto automáticamente.

```go
// pkg/errors/error-definition.go
const (
    BadRequestErr   DomainErrorCode = "BAD_REQUEST_ERR"
    UnauthorizedErr DomainErrorCode = "UNAUTHORIZED_ERR"
    NotFoundErr     DomainErrorCode = "NOT_FOUND_ERR"
)

// Each code maps to an Errors struct with HTTP code, gRPC code, and message
```

### Flujo de errores por capa

**Infrastructure** — mapea errores externos (DB, cache, HTTP) a códigos de error de dominio:

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

**Application** — usa errores de dominio para condiciones de negocio, propaga errores de infra tal cual:

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

**Handler/HTTP** — pasa errores al contexto de gin, el middleware se encarga del resto:

```go
// good: handler just forwards errors — middleware resolves HTTP codes
user, err := h.svc.SignIn(ctx, req)
if err != nil {
    g.Errors = append(g.Errors, g.Error(err))
    return
}
```

### Cuándo usar `fmt.Errorf` vs `errors.New`

| Situación | Usar | Por qué |
|-----------|------|---------|
| Condición de negocio/dominio conocida | `errors.New(errors.SomeCode)` | El middleware mapea el código → estado HTTP automáticamente |
| Error ya mapeado por una capa inferior | `return err` | No re-envolver lo que ya es `*Errors` |
| Fallo interno verdaderamente inesperado | `errors.New(errors.InternalErr)` | No filtrar detalles internos al cliente |
| Librería standalone / pkg sin catálogo de errores | `fmt.Errorf("context: %w", err)` | Sin errores de dominio disponibles, envolver para trazabilidad |
| Mensaje de error con valores dinámicos | `errors.WithMessage(fmt.Sprintf("msg: %s", v))` | Siempre usar `fmt.Sprintf` para mensajes con variables — nunca concatenación de strings con `+` |

**Anti-patrón**: envolver un `*Errors` ya mapeado con `fmt.Errorf` — agrega ruido y puede confundir al middleware que verifica tipos de error.

**Anti-patrón**: `errors.WithMessage("prefix: " + variable)` — usar `fmt.Sprintf("prefix: %s", variable)` en su lugar. La concatenación de strings en mensajes de error es más difícil de leer e inconsistente con los idiomas de Go.

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

- Nunca `panic` en código de librería/aplicación — solo en `main()` para fallos de bootstrap irrecuperables
- Verificar errores inmediatamente — nunca diferir la verificación de errores

## Nomenclatura

- **Receptores**: cortos, 1-2 letras, consistentes entre métodos (`func (s *Server)`, no `func (server *Server)`)
- **Interfaces**: basadas en verbos, pequeñas (`Reader`, `Validator`, no `UserServiceInterface`)
- **Constructores**: `NewXxx` retorna tipo concreto, no interfaz
- **No exportado por defecto** — exportar solo lo que la API del paquete necesita
- **Acrónimos**: todo en mayúsculas (`HTTP`, `ID`, `URL`), no `Http`, `Id`, `Url`
- **Nombres de paquetes**: cortos, lowercase, sin guiones bajos, singular (`user`, no `users` ni `user_service`)

## Context y Limpieza de Recursos

- Siempre primer parámetro: `func DoThing(ctx context.Context, ...)`
- Establecer timeouts en cada llamada externa (HTTP, DB, Redis, gRPC)
- Nunca almacenar `context.Context` en structs
- Nunca usar `http.Get()` / `http.Post()` / `http.DefaultClient` — crear cliente con `Timeout`
- Siempre `defer rows.Close()` inmediatamente después de la verificación de error de `QueryContext`
- Siempre `defer resp.Body.Close()` inmediatamente después de la verificación de error de `client.Do()`
- Siempre `defer cancel()` después de `context.WithTimeout` / `context.WithCancel`
- Configurar el pool de `sql.DB`: `MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`
- Ver `guides/cleanup/` para patrones completos, timeouts multi-nivel, y configuración del connection pool

## Concurrencia

- Proteger estado compartido con `sync.Mutex` o channels — elegir uno por recurso
- Preferir channels para coordinación, mutexes para protección de estado
- Siempre manejar el ciclo de vida de goroutines — sin fire-and-forget
- Usar `errgroup` para trabajo paralelo con propagación de errores
- Ejecutar tests siempre con la flag `-race`
