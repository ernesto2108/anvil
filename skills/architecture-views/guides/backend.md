# Template: architecture-backend.md

Inspired by: Stripe spec-driven design + bflorat Application View.

**Generate when:** backend work is involved.

## Template

```markdown
# Arquitectura Backend — <TASK-ID>

## Patrones de comunicación usados

<!-- List which patterns this feature uses. Include ONLY sections below that apply. -->
- [ ] REST API
- [ ] Eventos async (Kafka / RabbitMQ / SQS)
- [ ] gRPC
- [ ] WebSockets / SSE
- [ ] Webhooks
- [ ] Tauri commands (desktop IPC)

---

## Contratos REST (OpenAPI) — incluir si aplica

<!-- Executable spec — YAML fragment. -->

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

<!-- Use AsyncAPI format for Kafka topics, RabbitMQ exchanges, SQS queues, etc. -->

```yaml
asyncapi: "2.6.0"
channels:
  <topic-or-queue-name>:
    publish:        # producer side
      message:
        $ref: "#/components/messages/<EventName>"
    subscribe:      # consumer side
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
            description: Idempotency key
          occurredAt:
            type: string
            format: date-time
          # ... domain fields
```

**Garantías de entrega:** at-most-once / at-least-once / exactly-once  
**Orden:** global / por partition key / sin garantía  
**Idempotencia:** cómo el consumidor detecta duplicados (eventId, dedup window)  
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

<!-- Ports & adapters: what the system does, not how -->
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

## Rules

- Use ONLY the sections that apply — omit empty sections entirely
- OpenAPI is source of truth for REST contracts; AsyncAPI for events — not prose
- Every event schema needs an `eventId` (idempotency key) and `occurredAt`
- Delivery guarantees, ordering, and DLQ strategy are mandatory for any async section
- Error taxonomy must classify errors as retryable vs fatal — not just HTTP codes
- Sequence diagrams show happy path + primary failure path
- If frontend exists, REST/command schemas here are canonical — frontend derives from these
