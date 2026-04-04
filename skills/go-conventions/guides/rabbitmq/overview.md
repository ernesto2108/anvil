# RabbitMQ Overview

## Library Selection

| Library | Type | Best For |
|---------|------|----------|
| `rabbitmq/amqp091-go` | Official client | Full control, production use |
| `wagslane/go-rabbitmq` | Wrapper over amqp091 | Auto-reconnect, simpler API, declarative topology |

**Default choice**: `amqp091-go` — official, well-maintained, full feature support. Handle reconnection manually (the library does NOT auto-reconnect).

## Core Concepts

### Exchange Types

| Type | Routing | Use Case |
|------|---------|----------|
| `direct` | Exact routing key match | Point-to-point, task queues |
| `topic` | Pattern matching (`order.*`, `#.error`) | Event routing by category |
| `fanout` | Broadcast to all bound queues | Notifications, logging |
| `headers` | Header attribute matching | Complex routing rules |

### Durability Rules

- **Exchanges**: always `durable: true`
- **Queues**: always `durable: true` for production
- **Messages**: always `DeliveryMode: amqp.Persistent`
- **These three together** ensure messages survive broker restarts
