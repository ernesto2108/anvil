# Concurrency Anti-Patterns and Corrections

## Anti-Pattern 1: Goroutine Leak (No Way to Stop)

```go
// BAD: goroutine runs forever, no cancellation mechanism
func startPoller(url string) {
    go func() {
        for {
            resp, _ := http.Get(url)
            resp.Body.Close()
            time.Sleep(30 * time.Second)
        }
    }()
}
```

```go
// GOOD: goroutine respects context cancellation
func startPoller(ctx context.Context, url string) {
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
                resp, err := http.DefaultClient.Do(req)
                if err != nil {
                    continue
                }
                resp.Body.Close()
            case <-ctx.Done():
                return
            }
        }
    }()
}
```

**Detection:** Monitor goroutine count with `runtime.NumGoroutine()` or expose via pprof. If it grows over time, you have a leak.

## Anti-Pattern 2: Race Condition (Shared State Without Sync)

```go
// BAD: concurrent writes to map -- will panic or corrupt
var cache = make(map[string]int)

func handler(w http.ResponseWriter, r *http.Request) {
    key := r.URL.Query().Get("key")
    cache[key]++ // DATA RACE
    fmt.Fprintf(w, "%d", cache[key])
}
```

```go
// GOOD: protect with mutex
var (
    mu    sync.Mutex
    cache = make(map[string]int)
)

func handler(w http.ResponseWriter, r *http.Request) {
    key := r.URL.Query().Get("key")
    mu.Lock()
    cache[key]++
    count := cache[key]
    mu.Unlock()
    fmt.Fprintf(w, "%d", count)
}
```

**Detection:** Always run tests with `-race`: `go test -race ./...`. Run your service with `-race` in staging. The race detector finds races at runtime with ~2-10x overhead.

## Anti-Pattern 3: Channel Deadlock (Unbuffered Misuse)

```go
// BAD: deadlock -- unbuffered channel, nobody reading
func main() {
    ch := make(chan int)
    ch <- 42     // blocks forever: no goroutine reading
    fmt.Println(<-ch)
}
```

```go
// GOOD: send in a goroutine, or use buffered channel
func main() {
    ch := make(chan int, 1) // buffered: send won't block
    ch <- 42
    fmt.Println(<-ch)

    // OR: send in a goroutine
    ch2 := make(chan int)
    go func() { ch2 <- 42 }()
    fmt.Println(<-ch2)
}
```

**Rule:** Never send on an unbuffered channel in the same goroutine that reads from it. Unbuffered channels require a concurrent reader.

## Anti-Pattern 4: Over-Synchronization (Mutex + Channel Together)

```go
// BAD: using both mutex AND channel to protect the same data
type Counter struct {
    mu    sync.Mutex
    ch    chan int
    count int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    c.count++
    c.ch <- c.count // Why? Pick one mechanism
    c.mu.Unlock()
}
```

```go
// GOOD: pick one mechanism
// Option A: mutex only (simpler for state protection)
type Counter struct {
    mu    sync.Mutex
    count int
}

func (c *Counter) Increment() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
    return c.count
}

// Option B: channel only (if you need to stream updates)
type Counter struct {
    inc chan struct{}
    val chan int
}

func NewCounter() *Counter {
    c := &Counter{
        inc: make(chan struct{}),
        val: make(chan int),
    }
    go func() {
        count := 0
        for range c.inc {
            count++
            c.val <- count
        }
    }()
    return c
}
```

## Anti-Pattern 5: Forgetting to Drain Channels

```go
// BAD: producer goroutine leaks because nobody reads remaining values
func search(ctx context.Context, query string) (string, error) {
    results := make(chan string, 3)

    go func() { results <- searchBackend1(query) }() // may block forever
    go func() { results <- searchBackend2(query) }()
    go func() { results <- searchBackend3(query) }()

    return <-results, nil // Takes first result, abandons others
}
```

```go
// GOOD: use context cancellation + ensure goroutines can exit
func search(ctx context.Context, query string) (string, error) {
    ctx, cancel := context.WithCancel(ctx)
    defer cancel() // Signals all goroutines to stop

    results := make(chan string, 3) // buffered: goroutines won't block

    search := func(fn func(context.Context, string) string) {
        select {
        case results <- fn(ctx, query):
        case <-ctx.Done():
        }
    }

    go func() { search(searchBackend1) }()
    go func() { search(searchBackend2) }()
    go func() { search(searchBackend3) }()

    select {
    case r := <-results:
        return r, nil
    case <-ctx.Done():
        return "", ctx.Err()
    }
}
```

## Anti-Pattern 6: Context Not Propagated

```go
// BAD: creates new background context, ignoring caller's timeout/cancellation
func GetUser(ctx context.Context, id int) (*User, error) {
    // Ignores the ctx parameter entirely!
    resp, err := http.Get(fmt.Sprintf("/users/%d", id))
    // ...
}
```

```go
// GOOD: propagate context through the entire call chain
func GetUser(ctx context.Context, id int) (*User, error) {
    req, err := http.NewRequestWithContext(ctx, "GET",
        fmt.Sprintf("/users/%d", id), nil)
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("fetch user %d: %w", id, err)
    }
    defer resp.Body.Close()
    // ...
}
```

**Rule:** If a function accepts `context.Context`, pass it to every downstream call that supports it. This includes `http.NewRequestWithContext`, `db.QueryContext`, `grpc` calls, etc. Using `http.Get` or `db.Query` (without context) inside a context-aware function defeats the purpose.

## Production Checklist

Before shipping concurrent Go code:

- [ ] Every goroutine has a way to stop (context, done channel, or channel close)
- [ ] Every channel send uses `select` with `ctx.Done()` (or is guaranteed to be consumed)
- [ ] Every channel is closed by exactly one sender when done
- [ ] `context.Context` is propagated to all external calls (HTTP, DB, gRPC)
- [ ] `defer cancel()` follows every `WithTimeout` / `WithCancel`
- [ ] Shared maps are protected by `sync.RWMutex` or use `sync.Map`
- [ ] Worker pools bound concurrency to a reasonable limit
- [ ] `go test -race ./...` passes
- [ ] Goroutine count is monitored in production (pprof or metrics)
- [ ] Panics in goroutines are recovered (or use `errgroup` which propagates errors)
- [ ] Tickers are stopped with `defer ticker.Stop()`
- [ ] Graceful shutdown handles SIGTERM with timeout

## Sources

- [Go Concurrency Patterns: Pipelines and Cancellation](https://go.dev/blog/pipelines)
- [Go Concurrency Patterns: Context](https://go.dev/blog/context)
- [Go Wiki: Mutex or Channel?](https://go.dev/wiki/MutexOrChannel)
- [errgroup package documentation](https://pkg.go.dev/golang.org/x/sync/errgroup)
- [Graceful Shutdown in Go: Practical Patterns (VictoriaMetrics)](https://victoriametrics.com/blog/go-graceful-shutdown/)
- [7 Common Concurrency Pitfalls in Go](https://cristiancurteanu.com/7-common-concurrency-pitfalls-in-go-and-how-to-avoid-them/)
- [Channels vs Mutexes in Go](https://dev.to/gkoos/channels-vs-mutexes-in-go-the-big-showdown-338n)
- [7 Powerful Golang Concurrency Patterns (2025)](https://cristiancurteanu.com/7-powerful-golang-concurrency-patterns-that-will-transform-your-code-in-2025/)
- [Go Concurrency Patterns: Practical Guide (2026)](https://www.sachith.co.uk/go-concurrency-patterns-practical-guide-mar-11-2026/)
- [How to Write Bug-Free Goroutines in Go](https://itnext.io/how-to-write-bug-free-goroutines-in-go-golang-59042b1b63fb)
- [Why You Should Use errgroup.WithContext() in Server Handlers](https://www.fullstory.com/blog/why-errgroup-withcontext-in-golang-server-handlers/)
- [Worker Pool Pattern in Go](https://corentings.dev/blog/go-pattern-worker/)
- [Effective Go](https://go.dev/doc/effective_go)
