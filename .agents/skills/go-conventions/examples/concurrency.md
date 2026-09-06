# Ejemplos de Concurrencia

## Bien: Shutdown Graceful con errgroup

```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    srv := newServer()

    g, gCtx := errgroup.WithContext(ctx)

    g.Go(func() error {
        return srv.ListenAndServe()
    })

    g.Go(func() error {
        <-gCtx.Done()
        shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        return srv.Shutdown(shutCtx)
    })

    if err := g.Wait(); err != nil && !errors.Is(err, http.ErrServerClosed) {
        slog.Error("server error", "err", err)
        os.Exit(1)
    }
}
```

**Por qué:** Shutdown limpio, propagación correcta de errores, sin goroutine leaks.
