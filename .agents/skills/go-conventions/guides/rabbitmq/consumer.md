# Patrones de Consumer RabbitMQ

## Consumer Básico con QoS

```go
func (c *Consumer) Consume(queue string, handler func(context.Context, []byte) error) error {
    // limita los mensajes no reconocidos por consumer
    if err := c.ch.Qos(10, 0, false); err != nil {
        return fmt.Errorf("set QoS: %w", err)
    }

    msgs, err := c.ch.Consume(
        queue,
        "",    // consumer tag (auto-generado)
        false, // auto-ack = false (ack manual)
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
            msg.Nack(false, false) // requeue=false → va al DLQ vía DLX
            continue
        }

        msg.Ack(false)
    }

    return nil
}
```

## Consumer Vinculado a Exchange

```go
func (c *Consumer) ConsumeFromExchange(exchange, routingKey string, handler func(context.Context, []byte) error) error {
    // declara el exchange
    err := c.ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
    if err != nil {
        return fmt.Errorf("declare exchange: %w", err)
    }

    // deriva el nombre de la cola del exchange + routing key
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

## Reglas del Consumer

- **Nunca auto-ack** — siempre `autoAck: false`, `Ack` manual después del procesamiento exitoso
- **Configura QoS/prefetch** — controla el backpressure. Comienza con `prefetch=10`, ajusta según el tiempo de procesamiento
- **`Nack(false, false)`** — `multiple=false, requeue=false` envía al DLX (dead letter exchange)
- **`Nack(false, true)`** — `requeue=true` devuelve el mensaje a la cola (úsalo para fallas transitorias con precaución — puede generar bucles)
- **Maneja el cierre del channel** — `range msgs` termina cuando el channel se cierra, detecta y reconecta
