# Resource Checklist & Detection

## Complete Resource Checklist

| Resource | Acquire | Release | If You Forget |
|----------|---------|---------|---------------|
| `*sql.Rows` | `db.QueryContext()` | `defer rows.Close()` | Connection pool exhaustion → app deadlock |
| `*sql.Tx` | `db.BeginTx()` | `defer tx.Rollback()` + `tx.Commit()` | Connection pool exhaustion |
| `*sql.Conn` | `db.Conn()` | `defer conn.Close()` | Connection pool exhaustion |
| `http.Response.Body` | `client.Do(req)` | `defer resp.Body.Close()` | File descriptor exhaustion (CLOSE_WAIT) |
| `*os.File` | `os.Open()` | `defer f.Close()` | File descriptor exhaustion |
| `net.Conn` | `net.Dial()` | `defer conn.Close()` | Socket leak |
| `*grpc.ClientConn` | `grpc.Dial()` | `defer conn.Close()` | Connection leak |
| `time.Ticker` | `time.NewTicker()` | `defer ticker.Stop()` | Timer/goroutine leak |
| `context cancel` | `context.WithTimeout()` | `defer cancel()` | Timer goroutine leak |
| Redis client | `redis.NewClient()` | `defer rdb.Close()` | Connection leak |

## Linters That Catch These Automatically

Add to `.golangci.yml`:

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

These catch ~80% of context/cleanup issues at compile time.

## Detection in Production

**Goroutine leaks:**
- Expose `/debug/pprof/goroutine` and monitor count
- Use `runtime.NumGoroutine()` as a Prometheus metric
- Alert when count exceeds 2-3x baseline
- Use `uber-go/goleak` in tests:

```go
func TestNoLeaks(t *testing.T) {
    defer goleak.VerifyNone(t)
    // ... test code ...
}
```

**Connection pool leaks:**
- Export `db.Stats()` to Prometheus
- Alert on `InUse` not returning to baseline after request bursts
- Alert on `WaitCount` steadily increasing

**File descriptor leaks:**
- Monitor `lsof -p <pid> | wc -l` or expose via metrics
- Alert when approaching system limit (`ulimit -n`, typically 1024)
- Symptom: `EMFILE (Too many open files)` — all new connections fail

## Anti-Patterns Found in Production Codebases

These are real patterns that cause production incidents:

| Pattern | Why It's Dangerous | Fix |
|---|---|---|
| `http.Get(url)` | No timeout, no context, hangs forever | `http.NewRequestWithContext(ctx, ...)` |
| `http.DefaultClient.Do(req)` | No timeout configured | Custom client: `&http.Client{Timeout: 15*time.Second}` |
| `db.Query(...)` without context | No cancellation, hangs on slow DB | `db.QueryContext(ctx, ...)` |
| `defer` in a loop | Resources accumulate until function returns | Extract to helper function |
| `resp.Body.Close()` before error check | Nil pointer panic when request fails | Check error first, then defer |
| Missing `rows.Err()` check | Silent mid-iteration failures | Always check after the `for rows.Next()` loop |
| `context.TODO()` in request handlers | No timeout, no cancellation | Use `r.Context()` or derive with timeout |
| Missing pool config on `sql.Open` | Unlimited connections overwhelm DB | Set `MaxOpenConns`, `MaxIdleConns`, lifetimes |
| Not draining response body | TCP connection can't be reused | `io.Copy(io.Discard, resp.Body)` |
