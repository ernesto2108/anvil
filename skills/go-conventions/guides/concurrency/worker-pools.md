# Worker Pool (Bounded Concurrency)

**When:** You have many tasks but need to limit concurrent execution (API rate limits, DB connection pools, memory constraints).

**Real scenario:** Processing 10,000 database records with at most 20 concurrent HTTP calls.

## Using errgroup.SetLimit (preferred for new code)

```go
package main

import (
    "context"
    "fmt"
    "time"

    "golang.org/x/sync/errgroup"
)

type Record struct {
    ID   int
    Name string
}

func processRecord(ctx context.Context, r Record) error {
    // Simulate external API call
    select {
    case <-time.After(100 * time.Millisecond):
        fmt.Printf("processed record %d\n", r.ID)
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func processAll(ctx context.Context, records []Record) error {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(20) // At most 20 concurrent goroutines

    for _, r := range records {
        r := r // capture loop variable (not needed in Go 1.22+)
        g.Go(func() error {
            return processRecord(ctx, r)
        })
    }

    return g.Wait() // Returns first error; cancels ctx on error
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    records := make([]Record, 100)
    for i := range records {
        records[i] = Record{ID: i, Name: fmt.Sprintf("item-%d", i)}
    }

    if err := processAll(ctx, records); err != nil {
        fmt.Printf("failed: %v\n", err)
    }
}
```

## Using channels (classic pattern)

```go
func workerPool(ctx context.Context, jobs []Record, numWorkers int) error {
    jobsCh := make(chan Record)
    errCh := make(chan error, 1) // buffered: first error wins
    var wg sync.WaitGroup

    // Start fixed number of workers
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobsCh {
                if err := processRecord(ctx, job); err != nil {
                    select {
                    case errCh <- err: // send first error
                    default: // already have an error
                    }
                    return
                }
            }
        }()
    }

    // Send jobs
    go func() {
        defer close(jobsCh)
        for _, job := range jobs {
            select {
            case jobsCh <- job:
            case <-ctx.Done():
                return
            }
        }
    }()

    wg.Wait()
    close(errCh)
    return <-errCh
}
```

**Common mistake:** Spawning one goroutine per item without bounds. 10,000 goroutines making HTTP calls will exhaust file descriptors and overwhelm downstream services. Always bound concurrency.
