# Resumen de RabbitMQ

## Selección de Librería

| Librería | Tipo | Mejor Para |
|---------|------|----------|
| `rabbitmq/amqp091-go` | Cliente oficial | Control total, uso en producción |
| `wagslane/go-rabbitmq` | Wrapper sobre amqp091 | Auto-reconexión, API más simple, topología declarativa |

**Elección por defecto**: `amqp091-go` — oficial, bien mantenida, soporte completo de features. Maneja la reconexión manualmente (la librería NO reconecta automáticamente).

## Conceptos Fundamentales

### Tipos de Exchange

| Tipo | Enrutamiento | Caso de Uso |
|------|---------|----------|
| `direct` | Coincidencia exacta de routing key | Point-to-point, task queues |
| `topic` | Pattern matching (`order.*`, `#.error`) | Enrutamiento de eventos por categoría |
| `fanout` | Broadcast a todas las colas vinculadas | Notificaciones, logging |
| `headers` | Coincidencia por atributos de header | Reglas de enrutamiento complejas |

### Reglas de Durabilidad

- **Exchanges**: siempre `durable: true`
- **Colas**: siempre `durable: true` en producción
- **Mensajes**: siempre `DeliveryMode: amqp.Persistent`
- **Los tres juntos** garantizan que los mensajes sobrevivan reinicios del broker
