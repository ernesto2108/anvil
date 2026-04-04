# RabbitMQ Dead Letter Queue & Retry Patterns

## Dead Letter Queue (DLQ) — Native DLX

RabbitMQ has **native DLQ support** via Dead Letter Exchanges (DLX). Messages are dead-lettered when:

- Consumer sends `Nack` or `Reject` with `requeue=false`
- Message TTL expires
- Queue max-length is exceeded

### DLQ Setup

```go
func SetupQueueWithDLQ(ch *amqp.Channel, mainQueue, dlxExchange, dlqQueue string) error {
    // 1. declare dead-letter exchange
    if err := ch.ExchangeDeclare(dlxExchange, "direct", true, false, false, false, nil); err != nil {
        return fmt.Errorf("declare DLX: %w", err)
    }

    // 2. declare dead-letter queue
    if _, err := ch.QueueDeclare(dlqQueue, true, false, false, false, nil); err != nil {
        return fmt.Errorf("declare DLQ: %w", err)
    }

    // 3. bind DLQ to DLX
    if err := ch.QueueBind(dlqQueue, dlqQueue, dlxExchange, false, nil); err != nil {
        return fmt.Errorf("bind DLQ: %w", err)
    }

    // 4. declare main queue with DLX arguments
    args := amqp.Table{
        "x-dead-letter-exchange":    dlxExchange,
        "x-dead-letter-routing-key": dlqQueue,
    }
    if _, err := ch.QueueDeclare(mainQueue, true, false, false, false, args); err != nil {
        return fmt.Errorf("declare main queue: %w", err)
    }

    return nil
}
```

### Consumer with DLQ

```go
func ConsumeWithDLQ(ch *amqp.Channel, queue string, handler func(amqp.Delivery) error) error {
    if err := ch.Qos(10, 0, false); err != nil {
        return fmt.Errorf("set QoS: %w", err)
    }

    msgs, err := ch.Consume(queue, "", false, false, false, false, nil)
    if err != nil {
        return fmt.Errorf("consume: %w", err)
    }

    for msg := range msgs {
        if err := handler(msg); err != nil {
            slog.Error("processing failed, routing to DLQ", "error", err)
            // nack with requeue=false → DLX routes to DLQ automatically
            msg.Nack(false, false)
            continue
        }
        msg.Ack(false)
    }

    return nil
}
```

---

## Retry with TTL Queues

RabbitMQ supports retry via **TTL-based redelivery** through chained queues. When a message's TTL expires in a retry queue, it dead-letters back to the main exchange.

### Retry Topology

```
main-exchange → main-queue
                  ↓ (nack)
              retry-exchange-1 → retry-queue-1 (TTL: 1s) → back to main-exchange
                                   ↓ (nack)
                               retry-exchange-2 → retry-queue-2 (TTL: 5s) → back to main-exchange
                                                    ↓ (nack)
                                                dlx-exchange → dlq-queue
```

### Setup

```go
type RetryLevel struct {
    Name  string
    Delay time.Duration
}

type RetryTopology struct {
    MainExchange string
    MainQueue    string
    DLXExchange  string
    DLQQueue     string
    Levels       []RetryLevel
}

func SetupRetryTopology(ch *amqp.Channel, t RetryTopology) error {
    // main exchange
    if err := ch.ExchangeDeclare(t.MainExchange, "direct", true, false, false, false, nil); err != nil {
        return fmt.Errorf("declare main exchange: %w", err)
    }

    // DLX + DLQ
    if err := ch.ExchangeDeclare(t.DLXExchange, "direct", true, false, false, false, nil); err != nil {
        return fmt.Errorf("declare DLX: %w", err)
    }
    if _, err := ch.QueueDeclare(t.DLQQueue, true, false, false, false, nil); err != nil {
        return fmt.Errorf("declare DLQ: %w", err)
    }
    if err := ch.QueueBind(t.DLQQueue, t.DLQQueue, t.DLXExchange, false, nil); err != nil {
        return fmt.Errorf("bind DLQ: %w", err)
    }

    // retry queues — each has TTL, dead-letters back to main exchange
    for _, level := range t.Levels {
        retryExchange := fmt.Sprintf("%s.retry.%s", t.MainExchange, level.Name)
        retryQueue := fmt.Sprintf("%s.retry.%s", t.MainQueue, level.Name)

        if err := ch.ExchangeDeclare(retryExchange, "direct", true, false, false, false, nil); err != nil {
            return fmt.Errorf("declare retry exchange %s: %w", retryExchange, err)
        }

        args := amqp.Table{
            "x-message-ttl":             int64(level.Delay / time.Millisecond),
            "x-dead-letter-exchange":    t.MainExchange,
            "x-dead-letter-routing-key": t.MainQueue,
        }

        if _, err := ch.QueueDeclare(retryQueue, true, false, false, false, args); err != nil {
            return fmt.Errorf("declare retry queue %s: %w", retryQueue, err)
        }

        if err := ch.QueueBind(retryQueue, retryQueue, retryExchange, false, nil); err != nil {
            return fmt.Errorf("bind retry queue %s: %w", retryQueue, err)
        }
    }

    // main queue — dead-letters to DLX after all retries exhausted
    mainArgs := amqp.Table{
        "x-dead-letter-exchange": t.DLXExchange,
    }
    if _, err := ch.QueueDeclare(t.MainQueue, true, false, false, false, mainArgs); err != nil {
        return fmt.Errorf("declare main queue: %w", err)
    }
    if err := ch.QueueBind(t.MainQueue, t.MainQueue, t.MainExchange, false, nil); err != nil {
        return fmt.Errorf("bind main queue: %w", err)
    }

    return nil
}
```

### Example Usage

```go
topology := RetryTopology{
    MainExchange: "orders",
    MainQueue:    "orders.process",
    DLXExchange:  "orders.dlx",
    DLQQueue:     "orders.dlq",
    Levels: []RetryLevel{
        {Name: "1s", Delay: 1 * time.Second},
        {Name: "5s", Delay: 5 * time.Second},
        {Name: "30s", Delay: 30 * time.Second},
    },
}

if err := SetupRetryTopology(ch, topology); err != nil {
    log.Fatal(err)
}
```

### Retry with x-death Header

RabbitMQ adds an `x-death` header each time a message is dead-lettered. Use it to track retry count:

```go
func getRetryCount(msg amqp.Delivery) int {
    xDeath, ok := msg.Headers["x-death"]
    if !ok {
        return 0
    }

    deaths, ok := xDeath.([]interface{})
    if !ok || len(deaths) == 0 {
        return 0
    }

    firstDeath, ok := deaths[0].(amqp.Table)
    if !ok {
        return 0
    }

    count, ok := firstDeath["count"].(int64)
    if !ok {
        return 0
    }

    return int(count)
}
```
