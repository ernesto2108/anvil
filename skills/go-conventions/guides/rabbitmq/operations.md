# RabbitMQ Operations

## Backpressure

### QoS Prefetch

The primary backpressure mechanism in RabbitMQ:

```go
// limit unacknowledged messages per consumer
ch.Qos(
    10,    // prefetch count — max unacked messages
    0,     // prefetch size (0 = no limit)
    false, // per-consumer (not per-channel)
)
```

| Processing Time | Recommended Prefetch |
|----------------|---------------------|
| < 10ms | 50-100 |
| 10-100ms | 10-30 |
| 100ms-1s | 5-10 |
| > 1s | 1-5 |

### Single Active Consumer (Ordering)

When you need ordering guarantees with multiple consumer instances:

```go
args := amqp.Table{
    "x-single-active-consumer": true,
}
q, _ := ch.QueueDeclare("ordered-queue", true, false, false, false, args)
```

Only one consumer receives messages at a time. If it fails, RabbitMQ fails over to the next consumer.

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
                    return // channel closed
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

    // cancel consumer on channel (stops deliveries)
    if err := c.channel.Cancel("", false); err != nil {
        slog.Error("channel cancel error", "error", err)
    }

    // wait for in-flight messages to finish
    c.wg.Wait()

    if err := c.channel.Close(); err != nil {
        slog.Error("channel close error", "error", err)
    }
    if err := c.conn.Close(); err != nil {
        slog.Error("connection close error", "error", err)
    }

    slog.Info("shutdown complete")
}

// usage in main
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

## Message Processing with Retry and Ack

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

    // all retries exhausted — nack to DLX
    slog.Error("all retries exhausted, routing to DLQ",
        "queue", msg.RoutingKey,
        "error", err,
    )
    msg.Nack(false, false) // requeue=false → DLX
}
```

---

## Anti-Patterns

| Anti-Pattern | Why It's Bad | Fix |
|-------------|-------------|-----|
| Auto-ack enabled | Message lost if processing fails | `autoAck: false`, manual `Ack` |
| No QoS/prefetch set | Consumer gets flooded, OOM | Set `Qos(10, 0, false)` |
| Creating channel per message | Performance bottleneck, broker pressure | Reuse channels |
| No reconnection logic | Consumer dies silently on network blip | `NotifyClose` + reconnect loop |
| `Nack(false, true)` without retry limit | Infinite requeue loop | Use DLX or track retry count via `x-death` |
| Sharing connection for producer + consumer | Blocked by slow consumer | Separate connections |
| Non-durable queues in production | Messages lost on broker restart | `durable: true` + `Persistent` delivery |
| No publisher confirms | Silent message loss | `ch.Confirm(false)` + wait for confirmation |
| Ignoring `x-death` header | No visibility into retry count | Parse `x-death` for retry decisions |
| No DLQ configured | Failed messages disappear or loop | Always set `x-dead-letter-exchange` |

---

## Decision Matrix: RabbitMQ vs Other Options

| Concern | RabbitMQ |
|---------|----------|
| **DLQ** | Native (DLX + DLQ) — best-in-class |
| **Ordering** | Per-queue with single consumer (`x-single-active-consumer`) |
| **Exactly-once** | Publisher confirms + consumer dedup |
| **Backpressure** | QoS prefetch |
| **Retry** | TTL queues with DLX chain (native) |
| **Schema** | Application-layer |
| **Scale** | Horizontal via queue sharding |
| **Best for** | Task queues, routing, request-reply, complex topology |
