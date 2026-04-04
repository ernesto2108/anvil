# Graceful Shutdown

**When:** Every production service. HTTP servers, background workers, message consumers.

**Real scenario:** Kubernetes sends SIGTERM, you have 30 seconds to drain connections and finish in-flight work.

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "os/signal"
    "sync/atomic"
    "syscall"
    "time"
)

func main() {
    // Root context canceled on SIGINT or SIGTERM
    ctx, stop := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    var isShuttingDown atomic.Bool

    mux := http.NewServeMux()
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
        if isShuttingDown.Load() {
            w.WriteHeader(http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
    })
    mux.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
        // Use request context -- it will be canceled on shutdown
        select {
        case <-time.After(2 * time.Second):
            fmt.Fprintln(w, "done")
        case <-r.Context().Done():
            http.Error(w, "shutting down", http.StatusServiceUnavailable)
        }
    })

    server := &http.Server{Addr: ":8080", Handler: mux}

    // Start server in background
    go func() {
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            fmt.Printf("server error: %v\n", err)
        }
    }()

    // Start background worker
    workerCtx, workerCancel := context.WithCancel(context.Background())
    workerDone := make(chan struct{})
    go func() {
        defer close(workerDone)
        backgroundWorker(workerCtx)
    }()

    // Wait for shutdown signal
    <-ctx.Done()
    stop() // Allow second Ctrl+C to force-kill

    fmt.Println("shutting down...")
    isShuttingDown.Store(true)

    // 1. Fail readiness probe, wait for load balancer propagation
    time.Sleep(5 * time.Second)

    // 2. Shutdown HTTP server (waits for in-flight requests)
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
    defer cancel()
    if err := server.Shutdown(shutdownCtx); err != nil {
        fmt.Printf("server shutdown error: %v\n", err)
    }

    // 3. Stop background worker and wait for it
    workerCancel()
    <-workerDone

    fmt.Println("shutdown complete")
}

func backgroundWorker(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            fmt.Println("background tick")
        case <-ctx.Done():
            fmt.Println("worker stopping")
            return
        }
    }
}
```

**Common mistake:** Releasing resources (DB connections, caches) immediately on signal. In-flight HTTP handlers still need them. Shut down the HTTP server first (which drains in-flight requests), then close resources.

Source: [Graceful Shutdown in Go: Practical Patterns (VictoriaMetrics)](https://victoriametrics.com/blog/go-graceful-shutdown/)
