# Propagación de Contexto y la Regla de Oro

Las dos fuentes más comunes de incidentes en producción con Go: conexiones colgadas por falta de contexto/timeouts, y fugas de recursos por falta de llamadas a Close(). Ambas son silenciosas — funcionan bien en dev, luego explotan bajo carga.

## La Regla de Oro

```go
resource, err := acquire()
if err != nil {
    return err
}
defer resource.Close() // ALWAYS immediately after error check
```

Verificar el error PRIMERO, luego defer close. Nunca hacer defer antes de verificar el error (pánico por puntero nil). Nunca omitir el defer (fuga de recursos en cualquier ruta de error).

---

## Llamadas HTTP Client

```go
// BAD — hangs forever if server is unresponsive
resp, err := http.Get(url)

// BAD — http.DefaultClient has no timeout
resp, err := http.DefaultClient.Do(req)

// GOOD — context with timeout + custom client
client := &http.Client{Timeout: 15 * time.Second}

req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
if err != nil {
    return fmt.Errorf("create request: %w", err)
}

resp, err := client.Do(req)
if err != nil {
    return fmt.Errorf("fetch %s: %w", url, err)
}
defer resp.Body.Close()
```

**Reglas:**
- Nunca usar `http.Get()`, `http.Post()`, `http.Head()` — no tienen timeout
- Nunca usar `http.DefaultClient` en producción — crear un cliente con `Timeout`
- Siempre usar `http.NewRequestWithContext(ctx, ...)` — propaga la cancelación
- El `http.Client.Timeout` es una red de seguridad; los timeouts de contexto por llamada son el control primario

## Queries de Base de Datos

```go
// BAD — no context, no cancellation, hangs on slow DB or network partition
rows, err := db.Query("SELECT * FROM users WHERE active = true")

// GOOD — context-aware, cancels if request times out
rows, err := db.QueryContext(ctx, "SELECT * FROM users WHERE active = true")
if err != nil {
    return fmt.Errorf("query users: %w", err)
}
defer rows.Close() // CRITICAL — without this, connection stays checked out

// Iterate and check for errors
var users []User
for rows.Next() {
    var u User
    if err := rows.Scan(&u.ID, &u.Name); err != nil {
        return fmt.Errorf("scan user: %w", err)
    }
    users = append(users, u)
}
if err := rows.Err(); err != nil { // always check iteration errors
    return fmt.Errorf("iterate users: %w", err)
}
```

**Reglas:**
- Siempre usar `QueryContext`, `ExecContext`, `QueryRowContext` — nunca las versiones sin contexto
- Siempre `defer rows.Close()` después de `QueryContext`
- Siempre verificar `rows.Err()` después del loop — captura fallos a mitad de iteración
- Usar `ExecContext` para INSERT/UPDATE/DELETE — nunca `Query` (rows filtradas)

## Redis

```go
// BAD — shared timeout budget for sequential operations
ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
defer cancel()
val1, _ := rdb.Get(ctx, "key1").Result() // takes 2.5s
val2, _ := rdb.Get(ctx, "key2").Result() // only 0.5s left!

// GOOD — per-operation timeout
func getFromRedis(ctx context.Context, rdb *redis.Client, key string) (string, error) {
    opCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    val, err := rdb.Get(opCtx, key).Result()
    if err != nil {
        return "", fmt.Errorf("redis get %s: %w", key, err)
    }
    return val, nil
}
```

**Importante:** Cuando el deadline de un contexto Redis se excede, el cliente debe cerrar esa conexión (no puede reutilizarse de forma segura). Esto fuerza un nuevo handshake TCP + TLS, que puede generar cascada: timeouts → churning de conexiones → más timeouts. Configurar timeouts de lectura/escritura en el propio cliente Redis como capa de seguridad adicional.

## gRPC

```go
// BAD — no deadline, can hang forever
resp, err := client.GetUser(context.Background(), &pb.UserRequest{Id: id})

// GOOD — deadline propagates automatically to server via gRPC metadata
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
resp, err := client.GetUser(ctx, &pb.UserRequest{Id: id})
```

gRPC propaga los deadlines a través de los metadatos — el servidor ve el tiempo restante disponible automáticamente.

## AWS SDK

```go
// ACCEPTABLE for initialization only
conf, err := config.LoadDefaultConfig(context.Background())

// BETTER — timeout even during init (fails fast if AWS is unreachable)
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
conf, err := config.LoadDefaultConfig(ctx)
```
