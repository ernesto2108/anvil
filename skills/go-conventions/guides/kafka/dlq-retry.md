# Kafka Dead Letter Queue & Retry Patterns

## Dead Letter Queue (DLQ)

Kafka has **no native DLQ**. Implement at the application layer.

### Pattern: Retry Topics + Dead Letter Topic

```
main-topic
  → main-topic.retry-1  (1s delay)
  → main-topic.retry-2  (5s delay)
  → main-topic.retry-3  (30s delay)
  → main-topic.dlt      (dead letter topic)
```

### DLQ Message Structure

```go
type DLQMessage struct {
    OriginalTopic     string          `json:"original_topic"`
    OriginalPartition int             `json:"original_partition"`
    OriginalOffset    int64           `json:"original_offset"`
    OriginalKey       string          `json:"original_key"`
    Payload           json.RawMessage `json:"payload"`
    Error             string          `json:"error"`
    FailedAt          time.Time       `json:"failed_at"`
    RetryCount        int             `json:"retry_count"`
}
```

### DLQ Producer

```go
type DLQProducer struct {
    writer *kafka.Writer
}

func NewDLQProducer(brokers []string, dlqTopic string) *DLQProducer {
    return &DLQProducer{
        writer: &kafka.Writer{
            Addr:         kafka.TCP(brokers...),
            Topic:        dlqTopic,
            Balancer:     &kafka.LeastBytes{},
            RequiredAcks: kafka.RequireAll,
            MaxAttempts:  3,
        },
    }
}

func (d *DLQProducer) Send(ctx context.Context, msg kafka.Message, originalErr error, retryCount int) error {
    dlqMsg := DLQMessage{
        OriginalTopic:     msg.Topic,
        OriginalPartition: msg.Partition,
        OriginalOffset:    msg.Offset,
        OriginalKey:       string(msg.Key),
        Payload:           msg.Value,
        Error:             originalErr.Error(),
        FailedAt:          time.Now().UTC(),
        RetryCount:        retryCount,
    }

    body, err := json.Marshal(dlqMsg)
    if err != nil {
        return fmt.Errorf("marshal DLQ message: %w", err)
    }

    return d.writer.WriteMessages(ctx, kafka.Message{
        Key:   msg.Key,
        Value: body,
        Headers: []kafka.Header{
            {Key: "x-original-topic", Value: []byte(msg.Topic)},
            {Key: "x-error-type", Value: []byte("processing_failure")},
            {Key: "x-failed-at", Value: []byte(time.Now().UTC().Format(time.RFC3339))},
        },
    })
}

func (d *DLQProducer) Close() error {
    return d.writer.Close()
}
```

### When to Use DLQs

- Schema mismatches — producer sends a format the consumer can't deserialize
- Malformed payloads — corrupted or incomplete data
- Business rule violations — valid format but semantically invalid
- Permanent downstream failures

### When NOT to Use DLQs

- When strict message ordering is required (DLQ breaks order)
- When automated retries are sufficient for transient failures
- When messages have no recovery path — log and drop instead

### DLQ Operational Framework

1. **Inspect and discard** — fundamentally invalid messages (corrupted payloads)
2. **Fix data and replay** — recoverable failures requiring manual intervention
3. **Escalate operationally** — system-level signals (broken deployments, dependency changes)

---

## Retry Patterns

### Failure Classification

Always distinguish transient from permanent failures:

```go
type RetriableError struct {
    Err error
}

func (e *RetriableError) Error() string { return e.Err.Error() }
func (e *RetriableError) Unwrap() error { return e.Err }

func classifyError(err error) error {
    // schema/deserialization errors are permanent
    var syntaxErr *json.SyntaxError
    if errors.As(err, &syntaxErr) {
        return err // non-retriable — goes to DLQ
    }

    // network/timeout errors are transient
    var netErr net.Error
    if errors.As(err, &netErr) && netErr.Timeout() {
        return &RetriableError{Err: err}
    }

    // default: treat as transient
    return &RetriableError{Err: err}
}
```

### Exponential Backoff with Jitter

```go
type RetryConfig struct {
    MaxAttempts     int
    InitialInterval time.Duration
    MaxInterval     time.Duration
    Multiplier      float64
}

func DefaultRetryConfig() RetryConfig {
    return RetryConfig{
        MaxAttempts:     4,
        InitialInterval: 1 * time.Second,
        MaxInterval:     30 * time.Second,
        Multiplier:      2.0,
    }
}

func calculateBackoff(attempt int, initial, max time.Duration) time.Duration {
    backoff := time.Duration(float64(initial) * math.Pow(2.0, float64(attempt)))
    if backoff > max {
        backoff = max
    }
    // jitter: +/- 20%
    jitter := time.Duration(float64(backoff) * 0.2)
    return backoff - jitter + time.Duration(rand.Int63n(int64(2*jitter)))
}
```

### Process with Retry + DLQ

```go
func (mp *MessageProcessor) ProcessWithRetry(
    ctx context.Context,
    msg kafka.Message,
    handler func(kafka.Message) error,
) error {
    var lastErr error

    for attempt := 1; attempt <= mp.config.MaxAttempts; attempt++ {
        if err := handler(msg); err == nil {
            return nil
        } else {
            lastErr = err
        }

        // non-retriable errors go straight to DLQ
        var retriable *RetriableError
        if !errors.As(lastErr, &retriable) {
            return mp.dlq.Send(ctx, msg, lastErr, attempt)
        }

        backoff := calculateBackoff(attempt, mp.config.InitialInterval, mp.config.MaxInterval)
        select {
        case <-time.After(backoff):
        case <-ctx.Done():
            return fmt.Errorf("retry cancelled: %w", ctx.Err())
        }
    }

    return mp.dlq.Send(ctx, msg, lastErr, mp.config.MaxAttempts)
}
```

---

## Poison Message Detection

A poison message causes the consumer to fail repeatedly, blocking the entire partition.

```go
type PoisonDetector struct {
    mu          sync.Mutex
    failCounts  map[string]int // key: topic:partition:offset
    maxFailures int
    dlq         *DLQProducer
}

func (pd *PoisonDetector) Handle(ctx context.Context, msg kafka.Message, err error) Action {
    key := fmt.Sprintf("%s:%d:%d", msg.Topic, msg.Partition, msg.Offset)

    pd.mu.Lock()
    pd.failCounts[key]++
    count := pd.failCounts[key]
    pd.mu.Unlock()

    if count >= pd.maxFailures {
        slog.Warn("poison message detected", "key", key, "failures", count)
        if dlqErr := pd.dlq.Send(ctx, msg, err, count); dlqErr != nil {
            return ActionPause // pause if DLQ fails
        }
        pd.mu.Lock()
        delete(pd.failCounts, key)
        pd.mu.Unlock()
        return ActionSkip
    }

    return ActionRetry
}

type Action int

const (
    ActionRetry Action = iota
    ActionSkip
    ActionPause
)
```
