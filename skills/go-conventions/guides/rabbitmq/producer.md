# Patrones de Producer RabbitMQ

## Publicación Directa a Cola

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
        "",    // exchange por defecto
        queue, // routing key = nombre de la cola
        false, // mandatory
        false, // immediate
        amqp.Publishing{
            ContentType:  "application/json",
            Body:         body,
            DeliveryMode: amqp.Persistent,
            MessageId:    uuid.New().String(), // para idempotencia
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

## Publicación con Topic Exchange

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

## Reglas del Producer

- **Siempre habilita publisher confirms** (`ch.Confirm(false)`) — saber cuándo el broker aceptó el mensaje
- **Siempre espera la confirmación** — `select` en el canal de confirmación con timeout
- **Siempre `DeliveryMode: amqp.Persistent`** para mensajes en producción
- **Configura `MessageId`** para rastreo de idempotencia
- **Usa `PublishWithContext`** (no `Publish`) — soporta cancelación de contexto
