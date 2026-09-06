# Anti-Patrones de Concurrencia y Correcciones

## Anti-Patrón 1: Goroutine Leak (Sin Forma de Detenerla)

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

**Detección:** Monitorear el conteo de goroutines con `runtime.NumGoroutine()` o exponer vía pprof. Si crece con el tiempo, hay una fuga.

## Anti-Patrón 2: Race Condition (Estado Compartido Sin Sincronización)

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

**Detección:** Siempre ejecutar tests con `-race`: `go test -race ./...`. Ejecutar el servicio con `-race` en staging. El race detector encuentra carreras en tiempo de ejecución con ~2-10x de overhead.

## Anti-Patrón 3: Channel Deadlock (Mal Uso de Canal sin Buffer)

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

**Regla:** Nunca enviar en un canal sin buffer en la misma goroutine que lee de él. Los canales sin buffer requieren un reader concurrente.

## Anti-Patrón 4: Sobre-Sincronización (Mutex + Channel Juntos)

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

## Anti-Patrón 5: Olvidar Drenar Canales

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

## Anti-Patrón 6: Contexto No Propagado

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

**Regla:** Si una función acepta `context.Context`, pasarlo a cada llamada downstream que lo soporte. Esto incluye `http.NewRequestWithContext`, `db.QueryContext`, llamadas `grpc`, etc. Usar `http.Get` o `db.Query` (sin contexto) dentro de una función con contexto anula el propósito.

## Checklist de Producción

Antes de publicar código Go concurrente:

- [ ] Cada goroutine tiene una forma de detenerse (contexto, done channel, o cierre de canal)
- [ ] Cada envío a canal usa `select` con `ctx.Done()` (o está garantizado de ser consumido)
- [ ] Cada canal es cerrado por exactamente un sender cuando termina
- [ ] `context.Context` se propaga a todas las llamadas externas (HTTP, DB, gRPC)
- [ ] `defer cancel()` sigue cada `WithTimeout` / `WithCancel`
- [ ] Los mapas compartidos están protegidos por `sync.RWMutex` o usan `sync.Map`
- [ ] Los worker pools limitan la concurrencia a un número razonable
- [ ] `go test -race ./...` pasa
- [ ] El conteo de goroutines se monitorea en producción (pprof o métricas)
- [ ] Los pánicos en goroutines están recuperados (o usar `errgroup` que propaga errores)
- [ ] Los Tickers se detienen con `defer ticker.Stop()`
- [ ] El shutdown graceful maneja SIGTERM con timeout

## Fuentes

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
