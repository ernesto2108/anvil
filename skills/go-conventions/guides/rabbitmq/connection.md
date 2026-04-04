# RabbitMQ Connection Management

## Connection + Channel Setup

```go
func NewConnection(cnf RabbitMQConfig) (*amqp.Connection, *amqp.Channel, error) {
    url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
        cnf.User, cnf.Pass, cnf.Host, cnf.Port,
    )

    conn, err := amqp.Dial(url)
    if err != nil {
        return nil, nil, fmt.Errorf("connect to RabbitMQ: %w", err)
    }

    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, nil, fmt.Errorf("open channel: %w", err)
    }

    return conn, ch, nil
}
```

## Auto-Reconnect Pattern

RabbitMQ's Go client does NOT reconnect automatically. Always implement reconnection.

```go
type Client struct {
    conn         *amqp.Connection
    ch           *amqp.Channel
    notifyClose  chan *amqp.Error
    reconnecting bool
    mu           sync.Mutex
}

func (c *Client) monitorConnection(cnf RabbitMQConfig) {
    for {
        reason, ok := <-c.notifyClose
        if !ok {
            return // connection closed normally
        }

        slog.Error("RabbitMQ connection lost", "reason", reason)

        c.mu.Lock()
        if c.reconnecting {
            c.mu.Unlock()
            continue
        }
        c.reconnecting = true
        c.mu.Unlock()

        if err := c.attemptReconnect(cnf); err != nil {
            slog.Error("failed to reconnect after all attempts", "error", err)
            return
        }

        c.mu.Lock()
        c.reconnecting = false
        c.mu.Unlock()
        slog.Info("reconnected to RabbitMQ")
    }
}

func (c *Client) attemptReconnect(cnf RabbitMQConfig) error {
    const maxRetries = 10

    for i := 0; i < maxRetries; i++ {
        time.Sleep(time.Duration(i+1) * time.Second) // linear backoff

        conn, ch, err := NewConnection(cnf)
        if err != nil {
            slog.Error("reconnect attempt failed",
                "attempt", i+1,
                "max", maxRetries,
                "error", err,
            )
            continue
        }

        c.conn = conn
        c.ch = ch
        c.notifyClose = make(chan *amqp.Error)
        c.conn.NotifyClose(c.notifyClose)
        return nil
    }

    return fmt.Errorf("failed to reconnect after %d attempts", maxRetries)
}
```

## Connection Rules

- **Separate connections** for producers and consumers
- **Reuse channels** — do NOT create per-message
- **Use `NotifyClose`** to detect connection loss
- **Enable publisher confirms** for reliable publishing
- **Set QoS/prefetch** on consumer channels
