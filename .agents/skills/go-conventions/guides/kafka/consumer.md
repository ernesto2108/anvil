# Patrones de Consumer Kafka

## Consumer segmentio/kafka-go con Commits Manuales

```go
reader := kafka.NewReader(kafka.ReaderConfig{
    Brokers:           []string{"broker1:9092", "broker2:9092"},
    GroupID:           "order-processor",
    Topic:             "orders",
    MinBytes:          10e3,              // 10 KB
    MaxBytes:          10e6,              // 10 MB
    MaxWait:           3 * time.Second,
    CommitInterval:    0,                 // solo commit manual
    StartOffset:       kafka.LastOffset,
    SessionTimeout:    45 * time.Second,
    HeartbeatInterval: 15 * time.Second,
    RebalanceTimeout:  60 * time.Second,
})
defer reader.Close()

for {
    msg, err := reader.FetchMessage(ctx)
    if err != nil {
        if ctx.Err() != nil {
            break // shutdown
        }
        slog.Error("fetch error", "error", err)
        continue
    }

    if err := process(ctx, msg); err != nil {
        // maneja el error: reintentar, DLQ, o saltar
        continue
    }

    if err := reader.CommitMessages(ctx, msg); err != nil {
        slog.Error("commit failed", "error", err)
    }
}
```

## Consumer confluent-kafka-go con Topic Map

```go
consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
    "bootstrap.servers":                brokers,
    "group.id":                         groupID,
    "auto.offset.reset":               "latest",
    "group.instance.id":               "instance-" + hostname, // membresía estática
    "partition.assignment.strategy":    "cooperative-sticky",
    "session.timeout.ms":              45000,
})
if err != nil {
    return fmt.Errorf("create consumer: %w", err)
}

topicHandlers := map[string]func(ctx context.Context, msg []byte) error{
    "orders.created":   handleOrderCreated,
    "orders.cancelled": handleOrderCancelled,
}

topics := make([]string, 0, len(topicHandlers))
for t := range topicHandlers {
    topics = append(topics, t)
}
consumer.SubscribeTopics(topics, nil)
```

## Reglas del Consumer

- **Commits manuales después del procesamiento exitoso** — nunca auto-commit antes de procesar
- **Usa rebalanceo `cooperative-sticky`** — disrupción mínima, los consumers no afectados siguen procesando
- **Membresía estática de grupo** (`group.instance.id`) — reduce rebalanceos durante deploys rolling
- **Un consumer por topic** (patrón Netflix) — mantenimiento y tuning más simples
- **Configura `session.timeout.ms >= 45s`** — suficientemente alto para sobrevivir pausas de GC
- **Siempre maneja `ctx.Done()`** en el loop de consume

## Configuración de Consumer Group

| Estrategia | Comportamiento | Mejor Para |
|----------|----------|----------|
| `cooperative-sticky` | Disrupción mínima, distribución uniforme | La mayoría de aplicaciones (recomendado) |
| `range` | División por rango por topic | Topics co-particionados |
| `roundrobin` | Distribución uniforme entre todos los topics | Cargas de trabajo homogéneas |

### Configuración Lista para Producción

```go
// segmentio/kafka-go
reader := kafka.NewReader(kafka.ReaderConfig{
    SessionTimeout:    45 * time.Second,
    HeartbeatInterval: 15 * time.Second,
    RebalanceTimeout:  60 * time.Second,
    CommitInterval:    0,                // commits manuales
    StartOffset:       kafka.LastOffset,
})
```

## Ordenamiento de Mensajes

- **Dentro de una partición**: orden estricto
- **Entre particiones**: sin garantía

```go
// usa el ID de la entidad como clave — todos los eventos de la misma entidad van a la misma partición
writer.WriteMessages(ctx, kafka.Message{
    Key:   []byte(orderID), // particionamiento basado en hash
    Value: payload,
})
```

**Con productor idempotente** (`enable.idempotence=true`): el orden se preserva incluso con reintentos y `max.in.flight=5`.

**Sin idempotencia**: configura `max.in.flight.requests.per.connection=1` (costo en throughput).

**Los worker pools rompen el ordenamiento** — si paralelizas el procesamiento, enruta mensajes con la misma clave al mismo worker.
