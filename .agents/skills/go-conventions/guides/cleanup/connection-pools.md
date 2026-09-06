# Configuración del Pool de Conexiones (sql.DB)

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

| Configuración | Por defecto | Riesgo |
|---------|---------|------|
| MaxOpenConns | 0 (ilimitado) | Satura la base de datos |
| MaxIdleConns | 2 | Overhead de reconexión bajo carga |
| ConnMaxLifetime | 0 (para siempre) | Conexiones obsoletas después de failover de DB |
| ConnMaxIdleTime | 0 (para siempre) | Las conexiones inactivas desperdician recursos |

**Reglas:**
- `MaxIdleConns` <= `MaxOpenConns` (aplicado automáticamente)
- Configurar `MaxOpenConns` demasiado bajo causa deadlock en la app (las goroutines esperan como un semáforo)
- Monitorear `db.Stats()`: alertar si `WaitCount` aumenta o `InUse` se acerca a `MaxOpenConns`

## Monitorear la Salud del Pool de Conexiones

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

## Operaciones en Background (Go 1.21+)

```go
// When you need fire-and-forget without parent cancellation
func handler(w http.ResponseWriter, req *http.Request) {
    // Preserves values (trace_id, user) but detaches cancel signal
    bgCtx := context.WithoutCancel(req.Context())
    go sendAnalytics(bgCtx, event)
}
```

Usar con moderación — la mayoría de las operaciones deberían respetar la cancelación del padre.
