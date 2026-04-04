# Kafka Consumer Patterns

## segmentio/kafka-go Consumer with Manual Commits

```go
reader := kafka.NewReader(kafka.ReaderConfig{
    Brokers:           []string{"broker1:9092", "broker2:9092"},
    GroupID:           "order-processor",
    Topic:             "orders",
    MinBytes:          10e3,              // 10 KB
    MaxBytes:          10e6,              // 10 MB
    MaxWait:           3 * time.Second,
    CommitInterval:    0,                 // manual commit only
    StartOffset:       kafka.LastOffset,
    SessionTimeout:    45 * time.Second,
    HeartbeatInterval: 15 * time.Second,
    RebalanceTimeout:  60 * time.Second,
})
defer reader.Close()

for {
    msg, err := reader.FetchMessage(ctx)
    if err != nil {
        if ctx.Err() != nil {
            break // shutdown
        }
        slog.Error("fetch error", "error", err)
        continue
    }

    if err := process(ctx, msg); err != nil {
        // handle error: retry, DLQ, or skip
        continue
    }

    if err := reader.CommitMessages(ctx, msg); err != nil {
        slog.Error("commit failed", "error", err)
    }
}
```

## confluent-kafka-go Consumer with Topic Map

```go
consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
    "bootstrap.servers":                brokers,
    "group.id":                         groupID,
    "auto.offset.reset":               "latest",
    "group.instance.id":               "instance-" + hostname, // static membership
    "partition.assignment.strategy":    "cooperative-sticky",
    "session.timeout.ms":              45000,
})
if err != nil {
    return fmt.Errorf("create consumer: %w", err)
}

topicHandlers := map[string]func(ctx context.Context, msg []byte) error{
    "orders.created":   handleOrderCreated,
    "orders.cancelled": handleOrderCancelled,
}

topics := make([]string, 0, len(topicHandlers))
for t := range topicHandlers {
    topics = append(topics, t)
}
consumer.SubscribeTopics(topics, nil)
```

## Consumer Rules

- **Manual commits after successful processing** — never auto-commit before processing
- **Use `cooperative-sticky` rebalance** — minimal disruption, unaffected consumers keep processing
- **Static group membership** (`group.instance.id`) — reduces rebalances during rolling deploys
- **One consumer per topic** (Netflix pattern) — simpler maintenance and tuning
- **Set `session.timeout.ms >= 45s`** — high enough to survive GC pauses
- **Always handle `ctx.Done()`** in the consume loop

## Consumer Group Configuration

| Strategy | Behavior | Best For |
|----------|----------|----------|
| `cooperative-sticky` | Minimal disruption, even distribution | Most applications (recommended) |
| `range` | Per-topic range division | Co-partitioned topics |
| `roundrobin` | Even distribution across all topics | Homogeneous workloads |

### Production-Ready Config

```go
// segmentio/kafka-go
reader := kafka.NewReader(kafka.ReaderConfig{
    SessionTimeout:    45 * time.Second,
    HeartbeatInterval: 15 * time.Second,
    RebalanceTimeout:  60 * time.Second,
    CommitInterval:    0,                // manual commits
    StartOffset:       kafka.LastOffset,
})
```

## Message Ordering

- **Within a partition**: strictly ordered
- **Across partitions**: no guarantee

```go
// use entity ID as key — all events for the same entity go to the same partition
writer.WriteMessages(ctx, kafka.Message{
    Key:   []byte(orderID), // hash-based partitioning
    Value: payload,
})
```

**With idempotent producer** (`enable.idempotence=true`): order is preserved even with retries and `max.in.flight=5`.

**Without idempotence**: set `max.in.flight.requests.per.connection=1` (throughput cost).

**Worker pools break ordering** — if you parallelize processing, route same-key messages to the same worker.
