# RabbitMQ Producer Patterns

## Direct Queue Publishing

```go
func (p *Producer) SendMessage(ctx context.Context, queue string, message any) error {
    _, err := p.ch.QueueDeclare(queue, true, false, false, false, nil)
    if err != nil {
        return fmt.Errorf("declare queue: %w", err)
    }

    ctx, cancel := context.WithTimeout(ctx, p.timeout)
    defer cancel()

    body, err := json.Marshal(message)
    if err != nil {
        return fmt.Errorf("marshal message: %w", err)
    }

    confirms := p.ch.NotifyPublish(make(chan amqp.Confirmation, 1))

    err = p.ch.PublishWithContext(ctx,
        "",    // default exchange
        queue, // routing key = queue name
        false, // mandatory
        false, // immediate
        amqp.Publishing{
            ContentType:  "application/json",
            Body:         body,
            DeliveryMode: amqp.Persistent,
            MessageId:    uuid.New().String(), // for idempotency
        },
    )
    if err != nil {
        return fmt.Errorf("publish message: %w", err)
    }

    select {
    case confirm := <-confirms:
        if !confirm.Ack {
            return fmt.Errorf("message nacked by broker")
        }
    case <-ctx.Done():
        return fmt.Errorf("publish confirm timed out: %w", ctx.Err())
    }

    return nil
}
```

## Topic Exchange Publishing

```go
func (p *Producer) SendWithRouting(ctx context.Context, exchange, routingKey string, message any) error {
    err := p.ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
    if err != nil {
        return fmt.Errorf("declare exchange: %w", err)
    }

    ctx, cancel := context.WithTimeout(ctx, p.timeout)
    defer cancel()

    body, err := json.Marshal(message)
    if err != nil {
        return fmt.Errorf("marshal message: %w", err)
    }

    confirms := p.ch.NotifyPublish(make(chan amqp.Confirmation, 1))

    err = p.ch.PublishWithContext(ctx,
        exchange,
        routingKey,
        false,
        false,
        amqp.Publishing{
            ContentType:  "application/json",
            Body:         body,
            DeliveryMode: amqp.Persistent,
        },
    )
    if err != nil {
        return fmt.Errorf("publish message: %w", err)
    }

    select {
    case confirm := <-confirms:
        if !confirm.Ack {
            return fmt.Errorf("message nacked by broker")
        }
    case <-ctx.Done():
        return fmt.Errorf("publish confirm timed out: %w", ctx.Err())
    }

    return nil
}
```

## Producer Rules

- **Always enable publisher confirms** (`ch.Confirm(false)`) — know when the broker accepted the message
- **Always wait for confirmation** — `select` on confirm channel with timeout
- **Always `DeliveryMode: amqp.Persistent`** for production messages
- **Set `MessageId`** for idempotency tracking
- **Use `PublishWithContext`** (not `Publish`) — supports context cancellation
