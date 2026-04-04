# Kafka Producer Patterns

## segmentio/kafka-go Producer

```go
writer := &kafka.Writer{
    Addr:         kafka.TCP("broker1:9092", "broker2:9092"),
    Topic:        "events",
    Balancer:     &kafka.Hash{},        // key-based partitioning
    BatchSize:    100,                   // messages per batch
    BatchBytes:   1048576,               // 1 MB max batch
    BatchTimeout: 10 * time.Millisecond, // flush interval
    RequiredAcks: kafka.RequireAll,      // wait for all ISR
    MaxAttempts:  3,
    Compression:  kafka.Snappy,
    Async:        false,                 // sync for reliability
    AllowAutoTopicCreation: false,
}
defer writer.Close()

err := writer.WriteMessages(ctx, kafka.Message{
    Key:   []byte("order-12345"),
    Value: payload,
    Headers: []kafka.Header{
        {Key: "x-idempotency-key", Value: []byte(idempotencyKey)},
    },
})
if err != nil {
    return fmt.Errorf("write message: %w", err)
}
```

## confluent-kafka-go Producer

```go
producer, err := kafka.NewProducer(&kafka.ConfigMap{
    "bootstrap.servers":  brokerAddress,
    "enable.idempotence": true,
    "acks":               "all",
    "retries":            5,
    "compression.type":   "snappy",
    "linger.ms":          20,
    "batch.size":         32768,
})
if err != nil {
    return fmt.Errorf("create producer: %w", err)
}
defer producer.Close()

// async delivery report
go func() {
    for e := range producer.Events() {
        if m, ok := e.(*kafka.Message); ok && m.TopicPartition.Error != nil {
            slog.Error("delivery failed",
                "topic", *m.TopicPartition.Topic,
                "error", m.TopicPartition.Error,
            )
        }
    }
}()
```

## Producer Rules

- **Always set `acks=all`** — ensures all in-sync replicas acknowledge
- **Enable idempotence** — prevents duplicate messages on retries
- **Use snappy compression** — best balance of speed vs size for Go
- **Set a partition key** for messages that need ordering (entity ID)
- **Never use `fmt.Sprintf` with user input** in topic names
- **Close producers on shutdown** — flushes pending messages
