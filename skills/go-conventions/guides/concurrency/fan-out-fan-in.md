# Fan-Out / Fan-In

**When:** You have N independent items to process and want to parallelize across workers, then collect all results.

**Real scenario:** Fetching enrichment data from an external API for 1000 records.

```go
package main

import (
    "context"
    "fmt"
    "sync"
)

// Stage 1: Generator -- produces work items
func generate(ctx context.Context, items []string) <-chan string {
    out := make(chan string)
    go func() {
        defer close(out)
        for _, item := range items {
            select {
            case out <- item:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

// Stage 2: Worker -- processes one item at a time
func process(ctx context.Context, in <-chan string) <-chan string {
    out := make(chan string)
    go func() {
        defer close(out)
        for item := range in {
            // Simulate API call or computation
            result := "processed:" + item
            select {
            case out <- result:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

// Fan-in: merge multiple channels into one
func merge(ctx context.Context, channels ...<-chan string) <-chan string {
    out := make(chan string)
    var wg sync.WaitGroup

    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan string) {
            defer wg.Done()
            for val := range c {
                select {
                case out <- val:
                case <-ctx.Done():
                    return
                }
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()
    return out
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    items := []string{"a", "b", "c", "d", "e", "f"}
    in := generate(ctx, items)

    // Fan-out: 3 workers reading from the same channel
    w1 := process(ctx, in)
    w2 := process(ctx, in)
    w3 := process(ctx, in)

    // Fan-in: merge results
    for result := range merge(ctx, w1, w2, w3) {
        fmt.Println(result)
    }
}
```

**Common mistake:** Forgetting `select` with `ctx.Done()` on channel sends. Without it, if the consumer stops reading, producers block forever (goroutine leak). Every channel send in a goroutine should be wrapped in a select with a cancellation case.
