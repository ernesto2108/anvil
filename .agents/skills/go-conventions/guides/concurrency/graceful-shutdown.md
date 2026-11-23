# Graceful Shutdown

**Cuándo:** Todo servicio en producción. Servidores HTTP, workers de fondo, consumidores de mensajes.

**Escenario real:** Kubernetes envía SIGTERM, tienes 30 segundos para drenar conexiones y terminar el trabajo en vuelo.

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
    // Contexto raíz cancelado en SIGINT o SIGTERM
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
        // Usa el contexto de la request -- se cancela al hacer shutdown
        select {
        case <-time.After(2 * time.Second):
            fmt.Fprintln(w, "done")
        case <-r.Context().Done():
            http.Error(w, "shutting down", http.StatusServiceUnavailable)
        }
    })

    server := &http.Server{Addr: ":8080", Handler: mux}

    // Inicia el servidor en background
    go func() {
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            fmt.Printf("server error: %v\n", err)
        }
    }()

    // Inicia el worker de fondo
    workerCtx, workerCancel := context.WithCancel(context.Background())
    workerDone := make(chan struct{})
    go func() {
        defer close(workerDone)
        backgroundWorker(workerCtx)
    }()

    // Espera la señal de shutdown
    <-ctx.Done()
    stop() // Permite que un segundo Ctrl+C fuerce el cierre

    fmt.Println("shutting down...")
    isShuttingDown.Store(true)

    // 1. Falla el readiness probe, espera la propagación del load balancer
    time.Sleep(5 * time.Second)

    // 2. Apaga el servidor HTTP (espera a que terminen las requests en vuelo)
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
    defer cancel()
    if err := server.Shutdown(shutdownCtx); err != nil {
        fmt.Printf("server shutdown error: %v\n", err)
    }

    // 3. Detiene el worker de fondo y espera a que termine
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

**Error común:** Liberar recursos (conexiones DB, cachés) inmediatamente al recibir la señal. Los handlers HTTP en vuelo todavía los necesitan. Primero apaga el servidor HTTP (que drena las requests en vuelo), luego cierra los recursos.

Fuente: [Graceful Shutdown in Go: Practical Patterns (VictoriaMetrics)](https://victoriametrics.com/blog/go-graceful-shutdown/)
