# Messaging Critical Rules

These rules apply to both Kafka and RabbitMQ. For full implementation patterns, load the specific guide from `guides/kafka/` or `guides/rabbitmq/`.

1. **Never auto-commit/auto-ack before processing** — always manual ack after successful processing
2. **Always configure a DLQ** — poison messages will block your consumer forever without one
3. **Classify errors: transient vs permanent** — only retry transient errors, send permanent to DLQ immediately
4. **Always handle graceful shutdown** — drain in-flight messages, commit offsets, close connections
5. **Always reconnect** — RabbitMQ: manual reconnect with `NotifyClose`. Kafka: handled by client libraries
6. **Set backpressure limits** — Kafka: worker pool with bounded channel. RabbitMQ: QoS prefetch
7. **Use partition keys (Kafka) / single-active-consumer (RabbitMQ)** for ordering guarantees
8. **Enable publisher confirms (RabbitMQ) / idempotent producer (Kafka)** — never fire-and-forget in production
9. **Monitor DLQ depth and age** — dead letters pile up silently without alerts
