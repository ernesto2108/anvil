# Message Queues / Event Stores

## Cuándo Usar

Event streaming, desacoplamiento de microservicios, procesamiento asíncrono, event sourcing. No son databases en sentido estricto, pero el DBA a menudo gestiona su schema (formato de mensajes) y operaciones.

## Opciones y Cuándo Elegir

| Motor | Cuándo elegir | Notas |
|---|---|---|
| **Kafka** | Event streaming, event sourcing, alta durabilidad, reprocessing | Log distribuido. Retención por tiempo/tamaño. Consumers pueden re-leer |
| **RabbitMQ** | Task queues, routing complejo, RPC patterns | Broker tradicional. Mensaje se elimina al ack. Más simple que Kafka |
| **NATS** | Ultra-low latency, microservices messaging | JetStream para persistencia. Core NATS = at-most-once delivery |

## Conceptos Clave

### Kafka
- **Topic**: stream de eventos con nombre (equivalente a tabla)
- **Partition**: subdivisión de topic para paralelismo. Orden garantizado DENTRO de una partition
- **Consumer Group**: grupo de consumers que se reparten partitions. Un mensaje va a UN consumer del grupo
- **Offset**: posición del consumer en la partition. Permite re-leer desde cualquier punto
- **Key**: determina la partition (hash). Mensajes con misma key van a misma partition → orden garantizado

### RabbitMQ
- **Exchange**: router de mensajes (direct, topic, fanout, headers)
- **Queue**: buffer de mensajes. Consumers leen de queues
- **Binding**: regla que conecta exchange a queue
- **Acknowledgment**: consumer confirma procesamiento. Sin ack → re-delivery

### NATS
- **Subject**: tema de publicación (equivalente a topic, con wildcards `*` y `>`)
- **JetStream**: capa de persistencia sobre NATS Core
- **Consumer**: durable (persiste offset) o ephemeral

## Schema de Mensajes — La Responsabilidad del DBA

### Schema Registry (CRÍTICO para equipos)

El Schema Registry versionea y valida el formato de mensajes. Sin él, productores y consumidores rompen silenciosamente con cambios de schema.

| Herramienta | Compatibilidad | Formatos |
|---|---|---|
| **Confluent Schema Registry** | Kafka nativo | Avro, Protobuf, JSON Schema |
| **Karapace** | Open-source, compatible con Confluent | Avro, Protobuf, JSON Schema |
| **AWS Glue Schema Registry** | AWS managed | Avro, JSON Schema |

### Modos de compatibilidad

| Modo | Regla | Cuándo usar |
|---|---|---|
| **BACKWARD** | Nuevo consumer puede leer mensajes viejos | Default recomendado |
| **FORWARD** | Viejo consumer puede leer mensajes nuevos | Cuando no controlas los consumers |
| **FULL** | Ambas direcciones | Máxima seguridad, más restrictivo |
| **NONE** | Sin validación | Solo desarrollo |

### Ejemplo de evolución de schema (Avro)
```json
// v1
{"type": "record", "name": "OrderCreated", "fields": [
  {"name": "order_id", "type": "string"},
  {"name": "amount", "type": "double"}
]}

// v2 — BACKWARD compatible (campo nuevo con default)
{"type": "record", "name": "OrderCreated", "fields": [
  {"name": "order_id", "type": "string"},
  {"name": "amount", "type": "double"},
  {"name": "currency", "type": "string", "default": "USD"}
]}

// BREAKING — NO compatible (campo eliminado sin default)
// El DBA debe RECHAZAR este cambio
```

## Convenciones de Naming

### Topics / Subjects
```
# Kafka: {domain}.{entity}.{event_type}
orders.order.created
payments.payment.processed
users.user.updated

# NATS: {domain}.{entity}.{event_type} (con dots como separadores)
orders.order.created
orders.order.>            # wildcard para todos los eventos de order

# RabbitMQ exchanges: {domain}.{entity}
# RabbitMQ queues: {service}.{consumer_purpose}
exchange: orders.order
queue: billing.process-payment
```

### Campos del mensaje
```json
{
  "event_id": "uuid",           // idempotency key
  "event_type": "order.created", // tipo explícito
  "event_version": 2,           // versión del schema
  "timestamp": "ISO8601",       // cuándo ocurrió
  "source": "orders-service",   // quién lo publicó
  "data": { ... }               // payload del evento
}
```

## Migraciones de Schema de Mensajes

### Flujo del DBA para cambios de schema

1. **Proponer cambio** — definir nuevo schema (Avro/Protobuf/JSON Schema)
2. **Validar compatibilidad** — verificar contra Schema Registry en modo BACKWARD
3. **Si es compatible** → registrar nueva versión, producers actualizan
4. **Si NO es compatible** → crear nuevo topic (versionado)
   ```
   orders.order.created.v1  → viejo (deprecar gradualmente)
   orders.order.created.v2  → nuevo
   ```
5. **Dual-publish** durante transición: publicar en v1 y v2
6. **Cutover** cuando todos los consumers migraron a v2

### Cambios seguros (BACKWARD compatible)
- Agregar campo con default
- Agregar campo opcional (nullable)
- Agregar nuevo enum value al final

### Cambios BREAKING (requieren nuevo topic)
- Eliminar campo
- Cambiar tipo de campo
- Renombrar campo
- Cambiar semántica de campo existente

## Pitfalls de Producción

| # | Pitfall | Consecuencia | Prevención |
|---|---|---|---|
| 1 | Sin Schema Registry | Productores y consumers rompen silenciosamente | Schema Registry desde día 1 |
| 2 | Consumer lag acumulado | OOM o pérdida de datos por retention | Monitorear `kafka-consumer-groups --describe` |
| 3 | Mal número de partitions al crear topic | Difícil cambiar después (rebalanceo) | Estimar throughput futuro. 3-12 partitions es razonable para empezar |
| 4 | RabbitMQ queues sin TTL ni límite | Memoria agotada si consumer muere | `x-message-ttl`, `x-max-length` en queue |
| 5 | No manejar Dead Letter Queue | Mensajes fallidos desaparecen | DLQ + monitoreo de DLQ depth |
| 6 | No idempotencia en consumers | Duplicados procesados múltiples veces | Idempotency key (`event_id`) + dedup store |
| 7 | Kafka: un solo consumer group para todo | Un consumer lento frena todo | Consumer groups separados por servicio/propósito |

## Optimización de Rendimiento

### Kafka Producers
- `linger.ms=5-20`: agrupa mensajes antes de enviar — más throughput
- `batch.size=16384-65536`: tamaño de batch en bytes
- `compression.type=lz4` o `snappy`: reduce red y disco
- `acks=all` para durabilidad, `acks=1` para throughput

### Kafka Consumers
- Paralelismo = número de partitions (un consumer por partition max)
- `fetch.min.bytes` + `fetch.max.wait.ms` para batching de reads
- `enable.auto.commit=false` + commit manual para control preciso

### RabbitMQ
- `prefetch_count`: cuántos mensajes puede recibir un consumer sin ack (10-50 típico)
- Publisher confirms para garantizar que el broker recibió el mensaje
- Lazy queues para cargas con picos: almacena en disco en lugar de RAM

## Drivers por Stack

| Stack | Kafka | RabbitMQ | NATS |
|---|---|---|---|
| Go | `segmentio/kafka-go` o `confluent-kafka-go` | `rabbitmq/amqp091-go` | `nats-io/nats.go` |
| TypeScript | `kafkajs` | `amqplib` | `nats.js` |
| Python | `confluent-kafka-python` | `aio-pika` (async) o `pika` | `nats-py` |
| Rust | `rdkafka` | `lapin` | `async-nats` |

**Notas:**
- `confluent-kafka-go` requiere librdkafka (C) — más rápido pero más difícil de cross-compile
- `segmentio/kafka-go` es Go puro — más simple de compilar, suficiente para la mayoría de casos
