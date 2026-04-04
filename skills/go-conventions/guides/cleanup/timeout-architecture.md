# Multi-Level Timeout Architecture

Timeouts should be layered: overall request → per-operation → per-attempt.

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

## Recommended Defaults

| Call Type | Timeout | Notes |
|-----------|---------|-------|
| HTTP client (safety net) | 15-30s | `http.Client{Timeout: ...}` |
| Database query (simple) | 5s | `context.WithTimeout(ctx, 5*time.Second)` |
| Database query (report) | 30-60s | Complex aggregations, batch operations |
| Redis | 1-3s | If Redis is slow, something is broken |
| gRPC | 5-10s | Deadlines propagate automatically |
| Overall HTTP request | 30-60s | Handler-level context |
| Background job step | 5-30s | Per-step, not per-job |
