# Patrones de Producer Kafka

## Producer segmentio/kafka-go

```go
writer := &kafka.Writer{
    Addr:         kafka.TCP("broker1:9092", "broker2:9092"),
    Topic:        "events",
    Balancer:     &kafka.Hash{},        // particionamiento basado en clave
    BatchSize:    100,                   // mensajes por batch
    BatchBytes:   1048576,               // batch máximo de 1 MB
    BatchTimeout: 10 * time.Millisecond, // intervalo de flush
    RequiredAcks: kafka.RequireAll,      // espera todos los ISR
    MaxAttempts:  3,
    Compression:  kafka.Snappy,
    Async:        false,                 // sync para confiabilidad
    AllowAutoTopicCreation: false,
}
defer writer.Close()

err := writer.WriteMessages(ctx, kafka.Message{
    Key:   []byte("order-12345"),
    Value: payload,
    Headers: []kafka.Header{
        {Key: "x-idempotency-key", Value: []byte(idempotencyKey)},
    },
})
if err != nil {
    return fmt.Errorf("write message: %w", err)
}
```

## Producer confluent-kafka-go

```go
producer, err := kafka.NewProducer(&kafka.ConfigMap{
    "bootstrap.servers":  brokerAddress,
    "enable.idempotence": true,
    "acks":               "all",
    "retries":            5,
    "compression.type":   "snappy",
    "linger.ms":          20,
    "batch.size":         32768,
})
if err != nil {
    return fmt.Errorf("create producer: %w", err)
}
defer producer.Close()

// reporte de entrega asíncrono
go func() {
    for e := range producer.Events() {
        if m, ok := e.(*kafka.Message); ok && m.TopicPartition.Error != nil {
            slog.Error("delivery failed",
                "topic", *m.TopicPartition.Topic,
                "error", m.TopicPartition.Error,
            )
        }
    }
}()
```

## Reglas del Producer

- **Siempre configura `acks=all`** — asegura que todas las réplicas en sincronía reconozcan
- **Habilita idempotencia** — previene mensajes duplicados en reintentos
- **Usa compresión snappy** — mejor balance entre velocidad y tamaño para Go
- **Establece una partition key** para mensajes que necesiten ordenamiento (ID de entidad)
- **Nunca uses `fmt.Sprintf` con input de usuario** en nombres de topics
- **Cierra producers en el shutdown** — vacía los mensajes pendientes
