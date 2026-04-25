# Operaciones RabbitMQ

## Backpressure

### QoS Prefetch

El mecanismo principal de backpressure en RabbitMQ:

```go
// limita los mensajes no reconocidos por consumer
ch.Qos(
    10,    // prefetch count — máximo de mensajes sin ack
    0,     // prefetch size (0 = sin límite)
    false, // por consumer (no por channel)
)
```

| Tiempo de Procesamiento | Prefetch Recomendado |
|----------------|---------------------|
| < 10ms | 50-100 |
| 10-100ms | 10-30 |
| 100ms-1s | 5-10 |
| > 1s | 1-5 |

### Single Active Consumer (Ordenamiento)

Cuando necesitas garantías de ordenamiento con múltiples instancias de consumer:

```go
args := amqp.Table{
    "x-single-active-consumer": true,
}
q, _ := ch.QueueDeclare("ordered-queue", true, false, false, false, args)
```

Solo un consumer recibe mensajes a la vez. Si falla, RabbitMQ hace failover al siguiente consumer.

---

## Graceful Shutdown

```go
type Consumer struct {
    conn    *amqp.Connection
    channel *amqp.Channel
    done    chan struct{}
    wg      sync.WaitGroup
}

func (c *Consumer) Start(queue string, handler func(amqp.Delivery) error) error {
    msgs, err := c.channel.Consume(queue, "", false, false, false, false, nil)
    if err != nil {
        return fmt.Errorf("consume: %w", err)
    }

    c.wg.Add(1)
    go func() {
        defer c.wg.Done()
        for {
            select {
            case msg, ok := <-msgs:
                if !ok {
                    return // channel cerrado
                }
                if err := handler(msg); err != nil {
                    slog.Error("handler error", "error", err)
                    msg.Nack(false, false)
                    continue
                }
                msg.Ack(false)

            case <-c.done:
                slog.Info("draining remaining messages...")
                for msg := range msgs {
                    if err := handler(msg); err != nil {
                        msg.Nack(false, false)
                        continue
                    }
                    msg.Ack(false)
                }
                return
            }
        }
    }()

    return nil
}

func (c *Consumer) Shutdown() {
    slog.Info("initiating graceful shutdown...")

    close(c.done)

    // cancela el consumer en el channel (detiene las entregas)
    if err := c.channel.Cancel("", false); err != nil {
        slog.Error("channel cancel error", "error", err)
    }

    // espera a que terminen los mensajes en vuelo
    c.wg.Wait()

    if err := c.channel.Close(); err != nil {
        slog.Error("channel close error", "error", err)
    }
    if err := c.conn.Close(); err != nil {
        slog.Error("connection close error", "error", err)
    }

    slog.Info("shutdown complete")
}

// uso en main
func main() {
    consumer, err := NewConsumer(config)
    if err != nil {
        log.Fatal(err)
    }

    consumer.Start("orders", orderHandler)

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    consumer.Shutdown()
}
```

---

## Health Checks

```go
func (c *Client) CheckHealth(ctx context.Context) (string, error) {
    if c.conn == nil || c.conn.IsClosed() {
        return "down", nil
    }
    if c.ch == nil {
        return "down", nil
    }
    return "up", nil
}
```

---

## Procesamiento de Mensajes con Reintento y Ack

```go
const maxRetries = 3

func (c *Consumer) processMessage(ctx context.Context, msg amqp.Delivery, handler func(context.Context, []byte) error) {
    start := time.Now()

    var err error
    for attempt := 0; attempt < maxRetries; attempt++ {
        err = handler(ctx, msg.Body)
        if err == nil {
            msg.Ack(false)
            slog.Info("message processed",
                "queue", msg.RoutingKey,
                "duration", time.Since(start),
            )
            return
        }

        slog.Error("processing attempt failed",
            "attempt", attempt+1,
            "max", maxRetries,
            "error", err,
        )

        if attempt < maxRetries-1 {
            time.Sleep(time.Duration(attempt+1) * time.Second)
        }
    }

    // todos los reintentos agotados — nack al DLX
    slog.Error("all retries exhausted, routing to DLQ",
        "queue", msg.RoutingKey,
        "error", err,
    )
    msg.Nack(false, false) // requeue=false → DLX
}
```

---

## Anti-Patrones

| Anti-Patrón | Por Qué Es Malo | Solución |
|-------------|-------------|-----|
| Auto-ack habilitado | Mensaje perdido si el procesamiento falla | `autoAck: false`, `Ack` manual |
| Sin QoS/prefetch configurado | El consumer se satura, OOM | Configura `Qos(10, 0, false)` |
| Crear channel por mensaje | Cuello de botella de rendimiento, presión en el broker | Reutiliza channels |
| Sin lógica de reconexión | El consumer muere silenciosamente ante un blip de red | `NotifyClose` + loop de reconexión |
| `Nack(false, true)` sin límite de reintentos | Bucle infinito de requeue | Usa DLX o rastrea el conteo de reintentos vía `x-death` |
| Compartir conexión para producer + consumer | Bloqueado por consumer lento | Conexiones separadas |
| Colas no durables en producción | Mensajes perdidos al reiniciar el broker | `durable: true` + entrega `Persistent` |
| Sin publisher confirms | Pérdida silenciosa de mensajes | `ch.Confirm(false)` + esperar confirmación |
| Ignorar el header `x-death` | Sin visibilidad del conteo de reintentos | Parsea `x-death` para decisiones de reintento |
| Sin DLQ configurado | Los mensajes fallidos desaparecen o se repiten en bucle | Siempre configura `x-dead-letter-exchange` |

---

## Matriz de Decisión: RabbitMQ vs Otras Opciones

| Preocupación | RabbitMQ |
|---------|----------|
| **DLQ** | Nativo (DLX + DLQ) — el mejor de su clase |
| **Ordenamiento** | Por cola con consumer único (`x-single-active-consumer`) |
| **Exactly-once** | Publisher confirms + deduplicación en el consumer |
| **Backpressure** | QoS prefetch |
| **Reintento** | Colas TTL con cadena DLX (nativo) |
| **Schema** | A nivel de aplicación |
| **Escala** | Horizontal vía queue sharding |
| **Mejor para** | Task queues, enrutamiento, request-reply, topología compleja |
