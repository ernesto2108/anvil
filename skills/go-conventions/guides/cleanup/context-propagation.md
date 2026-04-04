# Context Propagation & The Golden Rule

The two most common sources of production incidents in Go: hanging connections from missing context/timeouts, and resource leaks from missing Close() calls. Both are silent — they work fine in dev, then explode under load.

## The Golden Rule

```go
resource, err := acquire()
if err != nil {
    return err
}
defer resource.Close() // ALWAYS immediately after error check
```

Error check FIRST, then defer close. Never defer before checking the error (nil pointer panic). Never skip the defer (resource leak on any error path).

---

## HTTP Client Calls

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

**Rules:**
- Never use `http.Get()`, `http.Post()`, `http.Head()` — they have no timeout
- Never use `http.DefaultClient` in production — create a client with `Timeout`
- Always use `http.NewRequestWithContext(ctx, ...)` — propagates cancellation
- The `http.Client.Timeout` is a safety net; per-call context timeouts are the primary control

## Database Queries

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

**Rules:**
- Always use `QueryContext`, `ExecContext`, `QueryRowContext` — never the non-context versions
- Always `defer rows.Close()` after `QueryContext`
- Always check `rows.Err()` after the loop — it catches mid-iteration failures
- Use `ExecContext` for INSERT/UPDATE/DELETE — never `Query` (leaked rows)

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

**Important:** When a Redis context deadline is exceeded, the client must close that connection (it can't be safely reused). This forces a new TCP + TLS handshake, which can cascade: timeouts → connection churn → more timeouts. Configure read/write timeouts on the Redis client itself as a separate safety layer.

## gRPC

```go
// BAD — no deadline, can hang forever
resp, err := client.GetUser(context.Background(), &pb.UserRequest{Id: id})

// GOOD — deadline propagates automatically to server via gRPC metadata
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
resp, err := client.GetUser(ctx, &pb.UserRequest{Id: id})
```

gRPC propagates deadlines through metadata — the server sees the remaining time budget automatically.

## AWS SDK

```go
// ACCEPTABLE for initialization only
conf, err := config.LoadDefaultConfig(context.Background())

// BETTER — timeout even during init (fails fast if AWS is unreachable)
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
conf, err := config.LoadDefaultConfig(ctx)
```
