# Checklist de Recursos y Detección

## Checklist Completo de Recursos

| Recurso | Adquirir | Liberar | Si lo olvidas |
|----------|---------|---------|---------------|
| `*sql.Rows` | `db.QueryContext()` | `defer rows.Close()` | Agotamiento del pool de conexiones → deadlock en la app |
| `*sql.Tx` | `db.BeginTx()` | `defer tx.Rollback()` + `tx.Commit()` | Agotamiento del pool de conexiones |
| `*sql.Conn` | `db.Conn()` | `defer conn.Close()` | Agotamiento del pool de conexiones |
| `http.Response.Body` | `client.Do(req)` | `defer resp.Body.Close()` | Agotamiento de file descriptors (CLOSE_WAIT) |
| `*os.File` | `os.Open()` | `defer f.Close()` | Agotamiento de file descriptors |
| `net.Conn` | `net.Dial()` | `defer conn.Close()` | Socket leak |
| `*grpc.ClientConn` | `grpc.Dial()` | `defer conn.Close()` | Connection leak |
| `time.Ticker` | `time.NewTicker()` | `defer ticker.Stop()` | Timer/goroutine leak |
| `context cancel` | `context.WithTimeout()` | `defer cancel()` | Timer goroutine leak |
| Redis client | `redis.NewClient()` | `defer rdb.Close()` | Connection leak |

## Linters que Detectan Estos Problemas Automáticamente

Agregar a `.golangci.yml`:

```yaml
linters:
  enable:
    - bodyclose       # unclosed HTTP response bodies
    - rowserrcheck    # rows.Err() not checked after iteration
    - sqlclosecheck   # unclosed sql.Rows and sql.Stmt
    - contextcheck    # context.Background() where parent should propagate
    - noctx           # HTTP requests without context
    - durationcheck   # incorrect time.Duration multiplication
```

Estos detectan ~80% de los problemas de context/cleanup en tiempo de compilación.

## Detección en Producción

**Goroutine leaks:**
- Exponer `/debug/pprof/goroutine` y monitorear el conteo
- Usar `runtime.NumGoroutine()` como métrica de Prometheus
- Alertar cuando el conteo supere 2-3x la línea base
- Usar `uber-go/goleak` en tests:

```go
func TestNoLeaks(t *testing.T) {
    defer goleak.VerifyNone(t)
    // ... test code ...
}
```

**Connection pool leaks:**
- Exportar `db.Stats()` a Prometheus
- Alertar si `InUse` no vuelve a la línea base después de ráfagas de requests
- Alertar si `WaitCount` aumenta constantemente

**File descriptor leaks:**
- Monitorear `lsof -p <pid> | wc -l` o exponer vía métricas
- Alertar cuando se acerque al límite del sistema (`ulimit -n`, típicamente 1024)
- Síntoma: `EMFILE (Too many open files)` — todas las nuevas conexiones fallan

## Anti-Patrones Encontrados en Codebases de Producción

Estos son patrones reales que causan incidentes en producción:

| Patrón | Por qué es peligroso | Corrección |
|---|---|---|
| `http.Get(url)` | Sin timeout, sin contexto, cuelga para siempre | `http.NewRequestWithContext(ctx, ...)` |
| `http.DefaultClient.Do(req)` | Sin timeout configurado | Cliente personalizado: `&http.Client{Timeout: 15*time.Second}` |
| `db.Query(...)` sin contexto | Sin cancelación, cuelga con DB lenta | `db.QueryContext(ctx, ...)` |
| `defer` en un loop | Los recursos se acumulan hasta que la función retorna | Extraer a función helper |
| `resp.Body.Close()` antes de verificar error | Pánico por puntero nil cuando la request falla | Verificar error primero, luego defer |
| Falta verificación de `rows.Err()` | Fallos silenciosos a mitad de iteración | Siempre verificar después del loop `for rows.Next()` |
| `context.TODO()` en request handlers | Sin timeout, sin cancelación | Usar `r.Context()` o derivar con timeout |
| Falta config del pool en `sql.Open` | Conexiones ilimitadas saturan la DB | Configurar `MaxOpenConns`, `MaxIdleConns`, lifetimes |
| No drenar el response body | La conexión TCP no puede reutilizarse | `io.Copy(io.Discard, resp.Body)` |
