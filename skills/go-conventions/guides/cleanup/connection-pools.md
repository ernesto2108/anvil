# Connection Pool Configuration (sql.DB)

```go
db, err := sql.Open("postgres", dsn)
if err != nil {
    log.Fatal(err)
}

// NEVER leave defaults in production
db.SetMaxOpenConns(25)                  // default: 0 (unlimited) — DANGEROUS
db.SetMaxIdleConns(25)                  // default: 2 — too low, constant reconnects
db.SetConnMaxLifetime(5 * time.Minute)  // default: 0 (forever) — stale after failover
db.SetConnMaxIdleTime(5 * time.Minute)  // releases idle conns when load drops
```

| Setting | Default | Risk |
|---------|---------|------|
| MaxOpenConns | 0 (unlimited) | Overwhelms database |
| MaxIdleConns | 2 | Reconnection overhead under load |
| ConnMaxLifetime | 0 (forever) | Stale connections after DB failover |
| ConnMaxIdleTime | 0 (forever) | Idle connections waste resources |

**Rules:**
- `MaxIdleConns` <= `MaxOpenConns` (enforced automatically)
- Setting `MaxOpenConns` too low causes app deadlock (goroutines wait like a semaphore)
- Monitor `db.Stats()`: alert on `WaitCount` increasing or `InUse` near `MaxOpenConns`

## Monitor Connection Pool Health

```go
stats := db.Stats()
// stats.InUse            — connections currently checked out
// stats.Idle             — connections sitting idle
// stats.WaitCount        — times a goroutine had to wait for a connection
// stats.WaitDuration     — total time spent waiting

// Alert on:
// - WaitCount increasing → pool too small
// - InUse near MaxOpenConns → approaching exhaustion
// - InUse not returning to ~0 after request burst → CONNECTION LEAK
```

## Background Operations (Go 1.21+)

```go
// When you need fire-and-forget without parent cancellation
func handler(w http.ResponseWriter, req *http.Request) {
    // Preserves values (trace_id, user) but detaches cancel signal
    bgCtx := context.WithoutCancel(req.Context())
    go sendAnalytics(bgCtx, event)
}
```

Use sparingly — most operations should respect parent cancellation.
