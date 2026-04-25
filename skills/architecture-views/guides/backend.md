# Template: architecture-backend.md

Inspirado en: diseño spec-driven de Stripe + bflorat Application View.

**Generar cuando:** hay trabajo de backend involucrado.

## Template

```markdown
# Arquitectura Backend — <TASK-ID>

## Patrones de comunicación usados

<!-- Listar qué patrones usa esta feature. Incluir SOLO las secciones abajo que apliquen. -->
- [ ] REST API
- [ ] Eventos async (Kafka / RabbitMQ / SQS)
- [ ] gRPC
- [ ] WebSockets / SSE
- [ ] Webhooks
- [ ] Tauri commands (desktop IPC)

---

## Contratos REST (OpenAPI) — incluir si aplica

<!-- Spec ejecutable — fragmento YAML. -->

```yaml
openapi: "3.1.0"
paths:
  /api/v1/<resource>:
    post:
      summary: ...
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/<RequestDTO>"
      responses:
        "201":
          description: ...
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/<ResponseDTO>"
        "400":
          $ref: "#/components/responses/ValidationError"
components:
  schemas:
    <RequestDTO>:
      type: object
      required: [field1]
      properties:
        field1:
          type: string
```

---

## Contratos de eventos / mensajes (AsyncAPI) — incluir si aplica

<!-- Usar formato AsyncAPI para Kafka topics, RabbitMQ exchanges, SQS queues, etc. -->

```yaml
asyncapi: "2.6.0"
channels:
  <topic-or-queue-name>:
    publish:        # lado productor
      message:
        $ref: "#/components/messages/<EventName>"
    subscribe:      # lado consumidor
      message:
        $ref: "#/components/messages/<EventName>"
components:
  messages:
    <EventName>:
      payload:
        type: object
        required: [eventId, occurredAt]
        properties:
          eventId:
            type: string
            description: Clave de idempotencia
          occurredAt:
            type: string
            format: date-time
          # ... campos de dominio
```

**Garantías de entrega:** at-most-once / at-least-once / exactly-once
**Orden:** global / por partition key / sin garantía
**Idempotencia:** cómo el consumidor detecta duplicados (eventId, ventana de dedup)
**Dead letter:** qué pasa si el consumer falla N veces

---

## Contratos gRPC — incluir si aplica

```proto
service <ServiceName> {
  rpc <MethodName> (<RequestMsg>) returns (<ResponseMsg>);
  rpc <StreamMethod> (<RequestMsg>) returns (stream <ResponseMsg>);
}

message <RequestMsg> {
  string field1 = 1;
  int32 field2 = 2;
}
```

---

## Contratos Tauri commands (desktop IPC) — incluir si aplica

```yaml
commands:
  <command_name>:
    params:
      field1: Type
    returns: Vec<DtoType>
    notes: "..."
```

---

## Casos de uso

<!-- Ports & adapters: qué hace el sistema, no cómo -->
- ...

## Comportamiento runtime

### <Flujo principal>

```mermaid
sequenceDiagram
  ...
```

## Taxonomía de errores

| Código / tipo | Retryable | Descripción | Cuándo ocurre |
|---|---|---|---|

## Estrategia de persistencia

- **Concurrencia:** ...
- **Idempotencia:** clave de idempotencia, ventana de deduplicación
- **Reintentos / backoff:** política, límite de intentos
- **Manejo de fallos:** qué pasa si el downstream no responde
```

## Reglas

- Usar SOLO las secciones que apliquen — omitir secciones vacías completamente
- OpenAPI es la fuente de verdad para contratos REST; AsyncAPI para eventos — no prosa
- Cada schema de evento necesita un `eventId` (clave de idempotencia) y `occurredAt`
- Garantías de entrega, orden y estrategia de DLQ son obligatorias para cualquier sección async
- La taxonomía de errores debe clasificar errores como retryable vs fatal — no solo códigos HTTP
- Los diagramas de secuencia muestran happy path + path de fallo principal
- Si existe frontend, los schemas REST/command aquí son canónicos — frontend los deriva de estos
