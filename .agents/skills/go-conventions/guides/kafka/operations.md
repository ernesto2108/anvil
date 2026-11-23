# Operaciones Kafka

## Graceful Shutdown

```go
func (c *Consumer) Run(ctx context.Context, handler func(context.Context, kafka.Message) error) {
    ctx, cancel := context.WithCancel(ctx)
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigChan
        slog.Info("shutdown signal received")
        cancel()
    }()

    for {
        select {
        case <-ctx.Done():
            slog.Info("shutting down consumer...")
            if err := c.reader.Close(); err != nil {
                slog.Error("reader close error", "error", err)
            }
            if err := c.dlqWriter.Close(); err != nil {
                slog.Error("DLQ writer close error", "error", err)
            }
            slog.Info("consumer shutdown complete")
            return
        default:
        }

        fetchCtx, fetchCancel := context.WithTimeout(ctx, 10*time.Second)
        msg, err := c.reader.FetchMessage(fetchCtx)
        fetchCancel()

        if err != nil {
            if ctx.Err() != nil {
                break
            }
            continue
        }

        // procesar + commit
        if err := handler(ctx, msg); err != nil {
            slog.Error("processing error", "error", err)
        }

        if err := c.reader.CommitMessages(ctx, msg); err != nil {
            slog.Error("commit error", "error", err)
        }
    }
}
```

---

## Propagación de Trazas con OpenTelemetry

### Adaptador Header Carrier

```go
type KafkaHeaderCarrier []kafka.Header

func (c KafkaHeaderCarrier) Get(key string) string {
    for _, h := range c {
        if h.Key == key {
            return string(h.Value)
        }
    }
    return ""
}

func (c *KafkaHeaderCarrier) Set(key, value string) {
    for i, h := range *c {
        if h.Key == key {
            (*c)[i].Value = []byte(value)
            return
        }
    }
    *c = append(*c, kafka.Header{Key: key, Value: []byte(value)})
}

func (c KafkaHeaderCarrier) Keys() []string {
    keys := make([]string, len(c))
    for i, h := range c {
        keys[i] = h.Key
    }
    return keys
}

func InjectTraceContext(ctx context.Context, msg *kafka.Message) {
    carrier := KafkaHeaderCarrier(msg.Headers)
    otel.GetTextMapPropagator().Inject(ctx, &carrier)
    msg.Headers = carrier
}

func ExtractTraceContext(ctx context.Context, msg kafka.Message) context.Context {
    carrier := KafkaHeaderCarrier(msg.Headers)
    return otel.GetTextMapPropagator().Extract(ctx, carrier)
}
```

### Métricas Clave

| Métrica | Tipo | Descripción |
|--------|------|-------------|
| `kafka_consumer_lag` | Gauge | Mensajes por detrás en la partición |
| `kafka_consumer_messages_total` | Counter | Total de mensajes consumidos |
| `kafka_consumer_errors_total` | Counter | Total de errores de consumo |
| `kafka_producer_messages_total` | Counter | Total de mensajes producidos |
| `kafka_dlq_messages_total` | Counter | Mensajes enviados al DLQ |
| `kafka_consumer_processing_duration` | Histogram | Tiempo de procesamiento de un mensaje |
| `kafka_retry_attempts_total` | Counter | Total de intentos de reintento |

---

## Evolución de Schema

Usa **Protobuf** para servicios Go (mejor tooling y rendimiento). Usa Schema Registry para aplicar compatibilidad.

```protobuf
syntax = "proto3";
package orders;
option go_package = "github.com/myapp/proto/orders";

message OrderEvent {
    string order_id = 1;
    string customer_id = 2;
    double amount = 3;
    OrderStatus status = 4;
    int64 created_at = 5;
    // v2: campos nuevos — compatible hacia atrás
    string currency = 6;
    repeated LineItem items = 7;
}
```

**Reglas de evolución**: agregar campos es seguro, nunca reutilices números de campo, nunca cambies tipos de campo, usa `reserved` para campos eliminados.

---

## Anti-Patrones

| Anti-Patrón | Por Qué Es Malo | Solución |
|-------------|-------------|-----|
| Auto-commit antes de procesar | Mensaje perdido si el procesamiento falla | Commit manual después del éxito |
| `time.Sleep` en el loop de reintento | Bloquea la goroutine, ignora el shutdown | `select` con `time.After` + `ctx.Done()` |
| Ignorar errores del consumer | Pérdida silenciosa de datos | Log + DLQ |
| Un consumer para todos los topics | Difícil de tunear, punto único de falla | Un consumer por topic |
| Sin partition key | Sin garantía de ordenamiento | Usa el ID de entidad como clave |
| Goroutines sin límite por mensaje | Explosión de memoria bajo carga | Worker pool con channel acotado |
| Estrategia de rebalanceo eager | Stop-the-world durante el rebalanceo | `cooperative-sticky` |
| Reintento sin clasificación | Reintentar errores permanentes desperdicia recursos | Clasifica transitorios vs permanentes |
| Sin DLQ | Mensajes veneno bloquean la partición para siempre | Siempre configura un DLQ |
| DLQ sin monitoreo | Los dead letters se acumulan silenciosamente | Alertar sobre profundidad y antigüedad del DLQ |

---

## Matriz de Decisión: Kafka vs Otras Opciones

| Preocupación | Kafka |
|---------|-------|
| **DLQ** | A nivel de aplicación (retry topics + DLT) |
| **Ordenamiento** | Por partición (usa clave) |
| **Exactly-once** | Transacciones + productor idempotente |
| **Backpressure** | pause/resume, max.poll.records, worker pool |
| **Reintento** | Retry topics con TTL o backoff a nivel de app |
| **Schema** | Schema Registry (Protobuf/Avro) |
| **Escala** | Horizontal vía particiones |
| **Mejor para** | Alto throughput, event streaming, replay |
