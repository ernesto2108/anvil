# Resumen de Kafka

## Selección de Librería

| Escenario | Librería |
|----------|---------|
| Go puro, dev local sencillo, sin CGo | `segmentio/kafka-go` |
| Máximo rendimiento, ecosistema Confluent, transacciones | `confluent-kafka-go` |
| Codebase existente con Sarama, necesita mock testing | `IBM/sarama` |
| Proyecto nuevo, propósito general | `segmentio/kafka-go` |
| Requerimientos financieros/críticos de exactly-once | `confluent-kafka-go` |

## Contexto de la Industria

- **Netflix**: 2+ billones de msgs/día. Un consumer por topic. Invirtió tempranamente en trazabilidad de mensajes (Inca)
- **Uber**: Billones de msgs/día. Construyó uForwarder (consumer proxy push-based). Control de flujo a nivel de partición, no pausa total. Clusters federados (~150 nodos cada uno)
- **LinkedIn**: 7+ billones de msgs/día, 100+ clusters, 100k+ topics. Separa topics por categoría (commands, events, logs, metrics). El particionamiento como primitiva de escalado

**Conclusiones clave**: consumers de responsabilidad única, paralelismo basado en particiones, invertir en observabilidad, separar topics por categoría de mensaje.
