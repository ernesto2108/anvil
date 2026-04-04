# Resource Cleanup Patterns

## Database Rows — the #1 leak source

```go
// BAD — connection stays checked out forever
rows, err := db.QueryContext(ctx, query)
if err != nil {
    return err
}
// forgot defer rows.Close()
for rows.Next() { ... }

// BAD — using Query for non-SELECT (rows leak)
db.Query("DELETE FROM users WHERE expired = true")
// Query returns *Rows that MUST be closed — use Exec instead

// BAD — defer in a loop (defers stack until function returns)
for _, id := range ids {
    rows, err := db.QueryContext(ctx, "SELECT ... WHERE id = $1", id)
    if err != nil { return err }
    defer rows.Close() // WRONG — won't close until outer function returns
    // all connections held simultaneously
}

// GOOD — extract to function so defer runs per iteration
for _, id := range ids {
    if err := processOne(ctx, db, id); err != nil {
        return err
    }
}

func processOne(ctx context.Context, db *sql.DB, id int) error {
    rows, err := db.QueryContext(ctx, "SELECT ... WHERE id = $1", id)
    if err != nil {
        return err
    }
    defer rows.Close() // runs when processOne returns — correct
    // ...
}
```

## Transactions — defer Rollback as safety net

```go
// BAD — early return without rollback leaks the connection
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return err
}
// ... some operations ...
if condition {
    return nil // LEAKED: tx never committed or rolled back
}
return tx.Commit()

// GOOD — defer Rollback is a no-op after Commit
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback() // safety net: no-op after successful Commit()

// ... operations using tx ...

return tx.Commit() // after this, deferred Rollback does nothing
```

## HTTP Response Body

```go
// BAD — panic if resp is nil on error
resp, err := client.Do(req)
defer resp.Body.Close() // PANIC: resp is nil when err != nil
if err != nil {
    return err
}

// BAD — not closing at all
resp, err := client.Do(req)
if err != nil {
    return err
}
// forgot defer resp.Body.Close()
// file descriptor stuck in CLOSE_WAIT

// GOOD — check error, then defer, then drain for connection reuse
resp, err := client.Do(req)
if err != nil {
    return fmt.Errorf("request failed: %w", err)
}
defer resp.Body.Close()
defer io.Copy(io.Discard, resp.Body) // drain for connection reuse
```

**Why drain?** If you don't fully read the body, the underlying TCP connection can't be reused by the connection pool. `io.Copy(io.Discard, resp.Body)` reads and discards the remaining bytes so the connection returns to the pool.

## Context Cancel Functions

```go
// BAD — context resources never freed
ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
// forgot defer cancel()
// the timer goroutine leaks until the parent context is cancelled

// GOOD — always defer cancel immediately
ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
defer cancel()
```

## Tickers

```go
// BAD — ticker goroutine runs forever
ticker := time.NewTicker(30 * time.Second)
for range ticker.C {
    collectMetrics()
}

// GOOD — stop ticker and use context for shutdown
ticker := time.NewTicker(30 * time.Second)
defer ticker.Stop()

for {
    select {
    case <-ctx.Done():
        return
    case <-ticker.C:
        collectMetrics()
    }
}
```
