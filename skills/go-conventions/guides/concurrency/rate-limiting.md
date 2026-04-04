# Rate Limiting

**When:** Calling external APIs with rate limits, processing events without overwhelming downstream systems.

**Real scenario:** Calling a third-party API that allows 10 requests per second.

## Using golang.org/x/time/rate (token bucket)

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
    limiter := rate.NewLimiter(rate.Limit(10), 1) // 10 per second, burst of 1

    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(5) // Also bound concurrency

    for _, id := range itemIDs {
        id := id
        g.Go(func() error {
            // Wait for rate limiter token
            if err := limiter.Wait(ctx); err != nil {
                return err
            }
            return callExternalAPI(ctx, id)
        })
    }

    return g.Wait()
}
```

## Using time.Ticker (simple fixed-rate)

```go
func processAtFixedRate(ctx context.Context, items []string) error {
    ticker := time.NewTicker(100 * time.Millisecond) // 10 per second
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

**Common mistake:** Forgetting `ticker.Stop()`. Tickers that are not stopped leak a goroutine and a timer internally. Always `defer ticker.Stop()`.
