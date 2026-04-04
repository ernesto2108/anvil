# Concurrency Decision Matrix

When to use which concurrency primitive:

| Scenario | Use | Why |
|---|---|---|
| Protect shared state (cache, counter, config) | `sync.Mutex` / `sync.RWMutex` | Simplest, fastest for guarding data |
| Coordinate goroutines, pass data between stages | Channels | Natural for producer-consumer, pipelines |
| Parallel tasks that can fail | `errgroup.Group` | Combines WaitGroup + error propagation + context cancellation |
| Wait for N goroutines to finish (no errors) | `sync.WaitGroup` | Lightweight, no error handling needed |
| Atomic counter or flag | `sync/atomic` | Lock-free, fastest for single values |
| One-time initialization | `sync.Once` | Thread-safe lazy init |
| Concurrent map access | `sync.Map` or `map` + `sync.RWMutex` | `sync.Map` for read-heavy with stable keys; `map` + mutex for everything else |
| Rate limiting | `time.Ticker` + channel or `golang.org/x/time/rate` | Token bucket for API rate limits |
| Timeout / cancellation propagation | `context.Context` | Always -- it is the standard cancellation mechanism |

## Quick Decision Flow

```
Need to share data between goroutines?
  YES -> Are goroutines passing ownership of data?
           YES -> Channel
           NO  -> Is it read-heavy, write-rare?
                    YES -> sync.RWMutex
                    NO  -> sync.Mutex
  NO  -> Are goroutines doing parallel work?
           YES -> Can they fail?
                    YES -> errgroup
                    NO  -> sync.WaitGroup
           NO  -> Do you need timeout/cancellation?
                    YES -> context.WithTimeout / context.WithCancel
```

Source: [Go Wiki: Mutex or Channel](https://go.dev/wiki/MutexOrChannel) -- "Use whichever is most expressive and simple for your problem."
