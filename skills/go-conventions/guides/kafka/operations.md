# Kafka Operations

## Graceful Shutdown

```go
func (c *Consumer) Run(ctx context.Context, handler func(context.Context, kafka.Message) error) {
    ctx, cancel := context.WithCancel(ctx)
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigChan
        slog.Info("shutdown signal received")
        cancel()
    }()

    for {
        select {
        case <-ctx.Done():
            slog.Info("shutting down consumer...")
            if err := c.reader.Close(); err != nil {
                slog.Error("reader close error", "error", err)
            }
            if err := c.dlqWriter.Close(); err != nil {
                slog.Error("DLQ writer close error", "error", err)
            }
            slog.Info("consumer shutdown complete")
            return
        default:
        }

        fetchCtx, fetchCancel := context.WithTimeout(ctx, 10*time.Second)
        msg, err := c.reader.FetchMessage(fetchCtx)
        fetchCancel()

        if err != nil {
            if ctx.Err() != nil {
                break
            }
            continue
        }

        // process + commit
        if err := handler(ctx, msg); err != nil {
            slog.Error("processing error", "error", err)
        }

        if err := c.reader.CommitMessages(ctx, msg); err != nil {
            slog.Error("commit error", "error", err)
        }
    }
}
```

---

## OpenTelemetry Trace Propagation

### Header Carrier Adapter

```go
type KafkaHeaderCarrier []kafka.Header

func (c KafkaHeaderCarrier) Get(key string) string {
    for _, h := range c {
        if h.Key == key {
            return string(h.Value)
        }
    }
    return ""
}

func (c *KafkaHeaderCarrier) Set(key, value string) {
    for i, h := range *c {
        if h.Key == key {
            (*c)[i].Value = []byte(value)
            return
        }
    }
    *c = append(*c, kafka.Header{Key: key, Value: []byte(value)})
}

func (c KafkaHeaderCarrier) Keys() []string {
    keys := make([]string, len(c))
    for i, h := range c {
        keys[i] = h.Key
    }
    return keys
}

func InjectTraceContext(ctx context.Context, msg *kafka.Message) {
    carrier := KafkaHeaderCarrier(msg.Headers)
    otel.GetTextMapPropagator().Inject(ctx, &carrier)
    msg.Headers = carrier
}

func ExtractTraceContext(ctx context.Context, msg kafka.Message) context.Context {
    carrier := KafkaHeaderCarrier(msg.Headers)
    return otel.GetTextMapPropagator().Extract(ctx, carrier)
}
```

### Key Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `kafka_consumer_lag` | Gauge | Messages behind in partition |
| `kafka_consumer_messages_total` | Counter | Total messages consumed |
| `kafka_consumer_errors_total` | Counter | Total consumption errors |
| `kafka_producer_messages_total` | Counter | Total messages produced |
| `kafka_dlq_messages_total` | Counter | Messages sent to DLQ |
| `kafka_consumer_processing_duration` | Histogram | Time to process a message |
| `kafka_retry_attempts_total` | Counter | Total retry attempts |

---

## Schema Evolution

Use **Protobuf** for Go services (best tooling and performance). Use Schema Registry for compatibility enforcement.

```protobuf
syntax = "proto3";
package orders;
option go_package = "github.com/myapp/proto/orders";

message OrderEvent {
    string order_id = 1;
    string customer_id = 2;
    double amount = 3;
    OrderStatus status = 4;
    int64 created_at = 5;
    // v2: new fields — backward compatible
    string currency = 6;
    repeated LineItem items = 7;
}
```

**Evolution rules**: adding fields is safe, never reuse field numbers, never change field types, use `reserved` for removed fields.

---

## Anti-Patterns

| Anti-Pattern | Why It's Bad | Fix |
|-------------|-------------|-----|
| Auto-commit before processing | Message lost if processing fails | Manual commit after success |
| `time.Sleep` in retry loop | Blocks goroutine, ignores shutdown | `select` with `time.After` + `ctx.Done()` |
| Ignoring consumer errors | Silent data loss | Log + DLQ |
| One consumer for all topics | Hard to tune, single point of failure | One consumer per topic |
| No partition key | No ordering guarantee | Use entity ID as key |
| Unbounded goroutines per message | Memory explosion under load | Worker pool with bounded channel |
| Eager rebalance strategy | Stop-the-world during rebalance | `cooperative-sticky` |
| Retry without classification | Retrying permanent errors wastes resources | Classify transient vs permanent |
| No DLQ | Poison messages block partition forever | Always configure a DLQ |
| DLQ without monitoring | Dead letters pile up silently | Alert on DLQ depth and age |

---

## Decision Matrix: Kafka vs Other Options

| Concern | Kafka |
|---------|-------|
| **DLQ** | Application-layer (retry topics + DLT) |
| **Ordering** | Per-partition (use key) |
| **Exactly-once** | Transactions + idempotent producer |
| **Backpressure** | pause/resume, max.poll.records, worker pool |
| **Retry** | Retry topics with TTL or app-layer backoff |
| **Schema** | Schema Registry (Protobuf/Avro) |
| **Scale** | Horizontal via partitions |
| **Best for** | High throughput, event streaming, replay |
