# Reglas Críticas de Mensajería

Estas reglas aplican tanto a Kafka como a RabbitMQ. Para patrones de implementación completos, cargar la guía específica desde `guides/kafka/` o `guides/rabbitmq/`.

1. **Nunca auto-commit/auto-ack antes de procesar** — siempre ack manual después del procesamiento exitoso
2. **Siempre configurar un DLQ** — los mensajes envenenados bloquearán tu consumer para siempre sin uno
3. **Clasificar errores: transitorios vs permanentes** — solo reintentar errores transitorios, enviar los permanentes al DLQ inmediatamente
4. **Siempre manejar graceful shutdown** — drenar mensajes en vuelo, hacer commit de offsets, cerrar conexiones
5. **Siempre reconectar** — RabbitMQ: reconexión manual con `NotifyClose`. Kafka: manejado por las librerías del cliente
6. **Establecer límites de backpressure** — Kafka: worker pool con channel acotado. RabbitMQ: QoS prefetch
7. **Usar partition keys (Kafka) / single-active-consumer (RabbitMQ)** para garantías de orden
8. **Habilitar publisher confirms (RabbitMQ) / idempotent producer (Kafka)** — nunca fire-and-forget en producción
9. **Monitorear la profundidad y antigüedad del DLQ** — las dead letters se acumulan silenciosamente sin alertas
