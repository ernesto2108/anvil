# Pipeline (Stage1 -> Stage2 -> Stage3)

**When:** Data flows through sequential transformation stages, each potentially concurrent.

**Real scenario:** ETL pipeline: read CSV rows -> validate/transform -> batch insert to database.

```go
package main

import (
    "context"
    "fmt"
    "strings"
)

type RawRow struct {
    Line int
    Data string
}

type ValidRow struct {
    Line int
    Name string
    Age  int
}

type InsertResult struct {
    Line int
    OK   bool
}

// Stage 1: Read raw data
func readRows(ctx context.Context, lines []string) <-chan RawRow {
    out := make(chan RawRow)
    go func() {
        defer close(out)
        for i, line := range lines {
            select {
            case out <- RawRow{Line: i + 1, Data: line}:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

// Stage 2: Validate and transform (can run multiple workers)
func validate(ctx context.Context, in <-chan RawRow) <-chan ValidRow {
    out := make(chan ValidRow)
    go func() {
        defer close(out)
        for row := range in {
            parts := strings.SplitN(row.Data, ",", 2)
            if len(parts) != 2 {
                continue // skip invalid rows
            }
            select {
            case out <- ValidRow{Line: row.Line, Name: parts[0]}: // simplified
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

// Stage 3: Batch insert
func batchInsert(ctx context.Context, in <-chan ValidRow) <-chan InsertResult {
    out := make(chan InsertResult)
    go func() {
        defer close(out)
        for row := range in {
            // Simulate DB insert
            select {
            case out <- InsertResult{Line: row.Line, OK: true}:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    lines := []string{"Alice,30", "Bob,25", "bad-line", "Carol,28"}

    // Wire the pipeline: read -> validate -> insert
    raw := readRows(ctx, lines)
    valid := validate(ctx, raw)
    results := batchInsert(ctx, valid)

    for r := range results {
        fmt.Printf("line %d: ok=%v\n", r.Line, r.OK)
    }
}
```

**Key rule from the Go blog:** "Stages close their outbound channels when all send operations are done. Stages keep receiving from inbound channels until those channels are closed or senders are unblocked." (Source: [Go Concurrency Patterns: Pipelines](https://go.dev/blog/pipelines))

**Common mistake:** Not closing channels. If stage 2 never closes its output channel, stage 3 blocks on `range` forever. Every goroutine that owns a channel must `defer close(out)`.
