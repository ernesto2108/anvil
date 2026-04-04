# Kafka Resilience Patterns

## Circuit Breaker

Prevents cascading failures when downstream services are unhealthy.

```go
type State int

const (
    StateClosed   State = iota // normal operation
    StateOpen                   // rejecting all requests
    StateHalfOpen              // testing with limited requests
)

type CircuitBreaker struct {
    mu               sync.Mutex
    state            State
    failureCount     int
    successCount     int
    failureThreshold int
    successThreshold int
    openTimeout      time.Duration
    lastFailure      time.Time
}

var ErrCircuitOpen = errors.New("circuit breaker is open")

func NewCircuitBreaker(failureThreshold, successThreshold int, openTimeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        state:            StateClosed,
        failureThreshold: failureThreshold,
        successThreshold: successThreshold,
        openTimeout:      openTimeout,
    }
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    cb.mu.Lock()
    switch cb.state {
    case StateOpen:
        if time.Since(cb.lastFailure) > cb.openTimeout {
            cb.state = StateHalfOpen
            cb.successCount = 0
        } else {
            cb.mu.Unlock()
            return ErrCircuitOpen
        }
    }
    cb.mu.Unlock()

    err := fn()

    cb.mu.Lock()
    defer cb.mu.Unlock()

    if err != nil {
        cb.failureCount++
        cb.lastFailure = time.Now()
        if cb.failureCount >= cb.failureThreshold {
            cb.state = StateOpen
        }
        return err
    }

    if cb.state == StateHalfOpen {
        cb.successCount++
        if cb.successCount >= cb.successThreshold {
            cb.state = StateClosed
            cb.failureCount = 0
        }
    } else {
        cb.failureCount = 0
    }

    return nil
}
```

**Pipeline**: retry transient errors → circuit breaker trips if downstream is down → DLQ for unrecoverable messages.

---

## Idempotent Consumer

Most production systems use **at-least-once delivery with consumer-side deduplication**.

```go
// schema
// CREATE TABLE processed_messages (
//     message_id   VARCHAR(256) PRIMARY KEY,
//     processed_at TIMESTAMP NOT NULL DEFAULT NOW()
// );
// CREATE INDEX idx_processed_messages_at ON processed_messages (processed_at);

func (ic *IdempotentConsumer) ProcessIdempotently(
    ctx context.Context,
    messageID string,
    handler func(ctx context.Context) error,
) error {
    var exists bool
    err := ic.db.QueryRowContext(ctx,
        `SELECT EXISTS(SELECT 1 FROM processed_messages WHERE message_id = $1)`,
        messageID,
    ).Scan(&exists)
    if err != nil {
        return fmt.Errorf("check processed: %w", err)
    }
    if exists {
        return nil // already processed
    }

    tx, err := ic.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback()

    if err := handler(ctx); err != nil {
        return fmt.Errorf("handle message: %w", err)
    }

    _, err = tx.ExecContext(ctx,
        `INSERT INTO processed_messages (message_id, processed_at)
         VALUES ($1, $2)
         ON CONFLICT (message_id) DO NOTHING`,
        messageID, time.Now().UTC(),
    )
    if err != nil {
        return fmt.Errorf("record processed: %w", err)
    }

    return tx.Commit()
}
```

### Idempotency Key Strategies

| Source | Key Strategy |
|--------|-------------|
| Kafka message | Header `x-idempotency-key` set by producer |
| Event sourcing | `{aggregate_id}:{event_sequence}` |
| External API | `{entity_type}:{entity_id}:{version}` |
| Fallback | `{topic}:{partition}:{offset}` (not ideal — offset can change after rebalance) |

---

## Backpressure

### Worker Pool with Bounded Channel

```go
type WorkerPool struct {
    reader  *kafka.Reader
    workers int
    buffer  int
}

func (wp *WorkerPool) Start(ctx context.Context, handler func(context.Context, kafka.Message) error) {
    messages := make(chan kafka.Message, wp.buffer)

    for i := 0; i < wp.workers; i++ {
        go func(id int) {
            for msg := range messages {
                if err := handler(ctx, msg); err != nil {
                    slog.Error("worker error", "worker", id, "error", err)
                    continue
                }
                if err := wp.reader.CommitMessages(ctx, msg); err != nil {
                    slog.Error("commit error", "worker", id, "error", err)
                }
            }
        }(i)
    }

    for {
        select {
        case <-ctx.Done():
            close(messages)
            return
        default:
        }

        msg, err := wp.reader.FetchMessage(ctx)
        if err != nil {
            if ctx.Err() != nil {
                close(messages)
                return
            }
            slog.Error("fetch error", "error", err)
            continue
        }

        messages <- msg // blocks if workers are busy (backpressure)
    }
}
```

> **Warning**: worker pools break ordering within a partition. Use only when ordering is not required, or route same-key messages to the same worker.

### Pause/Resume (confluent-kafka-go)

```go
// pause specific partitions when buffer is full
consumer.Pause([]kafka.TopicPartition{
    {Topic: &topic, Partition: 0},
})

// resume when buffer drains below threshold
consumer.Resume([]kafka.TopicPartition{
    {Topic: &topic, Partition: 0},
})
```

### Key Config Parameters

| Parameter | Default | Recommendation |
|-----------|---------|----------------|
| `max.poll.records` | 500 | Lower for slow processors |
| `max.poll.interval.ms` | 300000 | Increase for heavy processing |
| `fetch.max.bytes` | 52428800 | Reduce to limit memory |
| `session.timeout.ms` | 45000 | High enough to survive GC pauses |
