# Timeout and Cancellation with Context

**When:** Every external call (HTTP, DB, gRPC, file I/O) in production code.

**Real scenario:** HTTP handler that calls 3 microservices, must respond within 5 seconds total.

```go
func handleOrder(w http.ResponseWriter, r *http.Request) {
    // Inherit request context (canceled when client disconnects)
    ctx := r.Context()

    // Add overall timeout for this handler
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel() // Always defer cancel

    g, ctx := errgroup.WithContext(ctx)

    var (
        user    User
        inv     Inventory
        pricing Price
    )

    g.Go(func() error {
        var err error
        user, err = fetchUser(ctx, r.FormValue("user_id"))
        return err
    })
    g.Go(func() error {
        var err error
        inv, err = checkInventory(ctx, r.FormValue("item_id"))
        return err
    })
    g.Go(func() error {
        var err error
        pricing, err = getPrice(ctx, r.FormValue("item_id"))
        return err
    })

    if err := g.Wait(); err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            http.Error(w, "request timed out", http.StatusGatewayTimeout)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Use user, inv, pricing to build response...
    json.NewEncoder(w).Encode(map[string]any{
        "user": user.Name, "available": inv.Qty, "price": pricing.Amount,
    })
}
```

## Context rules (from [Go blog: Context](https://go.dev/blog/context)):
- Pass context as the first parameter to every function on the call path
- Never store context in a struct
- Use `context.Background()` only in `main()`, `init()`, and top-level test setup
- Always `defer cancel()` after `WithTimeout` / `WithCancel` / `WithDeadline`
- Use unexported key types for context values to prevent collisions

```go
// Type-safe context values
type ctxKey int
const requestIDKey ctxKey = 0

func WithRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey, id)
}

func RequestID(ctx context.Context) string {
    id, _ := ctx.Value(requestIDKey).(string)
    return id
}
```

**Common mistake:** Calling cancel on a context you received as a parameter. Only the creator of a derived context should call its cancel function. Sub-operations should return errors, not cancel parent contexts.
