# Rate Limiting

**Cuándo:** Al llamar APIs externas con límites de tasa, o al procesar eventos sin saturar sistemas downstream.

**Escenario real:** Llamar a una API de terceros que permite 10 requests por segundo.

## Usando golang.org/x/time/rate (token bucket)

```go
package main

import (
    "context"
    "fmt"
    "time"

    "golang.org/x/sync/errgroup"
    "golang.org/x/time/rate"
)

func callExternalAPI(ctx context.Context, itemID int) error {
    fmt.Printf("[%s] calling API for item %d\n", time.Now().Format("15:04:05.000"), itemID)
    return nil
}

func processWithRateLimit(ctx context.Context, itemIDs []int) error {
    limiter := rate.NewLimiter(rate.Limit(10), 1) // 10 por segundo, burst de 1

    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(5) // También limita la concurrencia

    for _, id := range itemIDs {
        id := id
        g.Go(func() error {
            // Espera un token del rate limiter
            if err := limiter.Wait(ctx); err != nil {
                return err
            }
            return callExternalAPI(ctx, id)
        })
    }

    return g.Wait()
}
```

## Usando time.Ticker (tasa fija simple)

```go
func processAtFixedRate(ctx context.Context, items []string) error {
    ticker := time.NewTicker(100 * time.Millisecond) // 10 por segundo
    defer ticker.Stop()

    for _, item := range items {
        select {
        case <-ticker.C:
            if err := process(ctx, item); err != nil {
                return fmt.Errorf("process %s: %w", item, err)
            }
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    return nil
}
```

**Error común:** Olvidar `ticker.Stop()`. Los tickers que no se detienen generan un goroutine leak y un timer interno. Siempre `defer ticker.Stop()`.
