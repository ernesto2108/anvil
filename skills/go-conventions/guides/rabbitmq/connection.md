# Gestión de Conexiones RabbitMQ

## Configuración de Connection + Channel

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

## Patrón de Auto-Reconexión

El cliente Go de RabbitMQ NO reconecta automáticamente. Siempre implementa reconexión.

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
            return // conexión cerrada normalmente
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
        time.Sleep(time.Duration(i+1) * time.Second) // backoff lineal

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

## Reglas de Conexión

- **Conexiones separadas** para producers y consumers
- **Reutiliza channels** — NO crees uno por mensaje
- **Usa `NotifyClose`** para detectar pérdida de conexión
- **Habilita publisher confirms** para publicación confiable
- **Configura QoS/prefetch** en los channels de consumer
