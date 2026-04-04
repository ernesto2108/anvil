# RabbitMQ Consumer Patterns

## Basic Consumer with QoS

```go
func (c *Consumer) Consume(queue string, handler func(context.Context, []byte) error) error {
    // limit unacknowledged messages per consumer
    if err := c.ch.Qos(10, 0, false); err != nil {
        return fmt.Errorf("set QoS: %w", err)
    }

    msgs, err := c.ch.Consume(
        queue,
        "",    // consumer tag (auto-generated)
        false, // auto-ack = false (manual ack)
        false, // exclusive
        false, // no-local
        false, // no-wait
        nil,
    )
    if err != nil {
        return fmt.Errorf("consume: %w", err)
    }

    for msg := range msgs {
        ctx := context.Background()

        if err := handler(ctx, msg.Body); err != nil {
            slog.Error("processing failed", "queue", queue, "error", err)
            msg.Nack(false, false) // requeue=false → goes to DLQ via DLX
            continue
        }

        msg.Ack(false)
    }

    return nil
}
```

## Exchange-Bound Consumer

```go
func (c *Consumer) ConsumeFromExchange(exchange, routingKey string, handler func(context.Context, []byte) error) error {
    // declare exchange
    err := c.ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
    if err != nil {
        return fmt.Errorf("declare exchange: %w", err)
    }

    // derive queue name from exchange + routing key
    queueName := fmt.Sprintf("%s-%s-queue", exchange, routingKey)

    q, err := c.ch.QueueDeclare(queueName, true, false, false, false, nil)
    if err != nil {
        return fmt.Errorf("declare queue: %w", err)
    }

    err = c.ch.QueueBind(q.Name, routingKey, exchange, false, nil)
    if err != nil {
        return fmt.Errorf("bind queue: %w", err)
    }

    if err := c.ch.Qos(10, 0, false); err != nil {
        return fmt.Errorf("set QoS: %w", err)
    }

    msgs, err := c.ch.Consume(q.Name, "", false, false, false, false, nil)
    if err != nil {
        return fmt.Errorf("consume: %w", err)
    }

    for msg := range msgs {
        ctx := context.Background()
        if err := handler(ctx, msg.Body); err != nil {
            slog.Error("processing failed", "queue", queueName, "error", err)
            msg.Nack(false, false)
            continue
        }
        msg.Ack(false)
    }

    return nil
}
```

## Consumer Rules

- **Never auto-ack** — always `autoAck: false`, manually `Ack` after successful processing
- **Set QoS/prefetch** — controls backpressure. Start with `prefetch=10`, tune based on processing time
- **`Nack(false, false)`** — `multiple=false, requeue=false` sends to DLX (dead letter exchange)
- **`Nack(false, true)`** — `requeue=true` puts back in queue (use for transient failures with caution — can loop)
- **Handle channel closure** — `range msgs` exits when channel closes, detect and reconnect
