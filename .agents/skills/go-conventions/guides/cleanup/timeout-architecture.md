# Arquitectura de Timeouts Multi-Nivel

Los timeouts deben ser por capas: presupuesto total de request → por operación → por intento.

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Level 1: Overall request budget (30s)
    ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
    defer cancel()

    // Level 2: Per-operation budget (10s) — nested under request
    opCtx, opCancel := context.WithTimeout(ctx, 10*time.Second)
    defer opCancel()
    user, err := userService.GetByID(opCtx, userID)

    // Level 3: Per-attempt budget (3s) — inside the service, for retries
    // The service internally does:
    //   attemptCtx, attemptCancel := context.WithTimeout(ctx, 3*time.Second)
    //   defer attemptCancel()
}
```

## Valores por Defecto Recomendados

| Tipo de llamada | Timeout | Notas |
|-----------|---------|-------|
| HTTP client (red de seguridad) | 15-30s | `http.Client{Timeout: ...}` |
| Query de BD (simple) | 5s | `context.WithTimeout(ctx, 5*time.Second)` |
| Query de BD (reporte) | 30-60s | Agregaciones complejas, operaciones en batch |
| Redis | 1-3s | Si Redis es lento, algo está mal |
| gRPC | 5-10s | Los deadlines se propagan automáticamente |
| Request HTTP total | 30-60s | Contexto a nivel de handler |
| Paso de job en background | 5-30s | Por paso, no por job completo |
