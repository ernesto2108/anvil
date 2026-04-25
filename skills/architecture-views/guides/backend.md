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
asyncapi: "3.0.0"
info:
  title: <Service> Events
  version: 1.0.0

servers:
  production:
    host: <broker-host>
    protocol: kafka  # o nats, amqp, sqs, etc.

channels:
  <channelName>:
    address: <topic-or-queue-name>
    messages:
      <EventName>:
        payload:
          type: object
          required: [eventId, occurredAt]
          properties:
            eventId:
              type: string
              format: uuid
              description: Clave de idempotencia
            occurredAt:
              type: string
              format: date-time
            # ... campos de dominio

operations:
  publish<EventName>:
    action: send
    channel:
      $ref: "#/channels/<channelName>"
  consume<EventName>:
    action: receive
    channel:
      $ref: "#/channels/<channelName>"
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

## Variables de entorno — incluir si aplica

<!-- Listar TODAS las env vars nuevas que esta feature introduce. -->
<!-- El developer las agrega al .env.example del proyecto. -->

| Variable | Ejemplo | Descripción | Secreto |
|---|---|---|---|
| `<SERVICE>_<PROPERTY>` | `valor-ejemplo` | Qué configura | Sí / No |

<!-- Ejemplo:
| `KAFKA_BROKERS` | `localhost:9092` | Brokers de Kafka, separados por coma | No |
| `KAFKA_SASL_PASSWORD` | `change-me` | Password SASL para Kafka | Sí |
| `REDIS_URL` | `redis://localhost:6379/0` | URL de conexión a Redis | No |
| `STRIPE_API_KEY` | `sk_test_...` | API key de Stripe | Sí |
-->
```

### Convenciones de naming

- **Casing:** `SCREAMING_SNAKE_CASE` siempre
- **Prefijo por servicio:** `DB_*`, `REDIS_*`, `KAFKA_*`, `AWS_*`, `OTEL_*`
- **URLs completas:** sufijo `_URL` (ej. `DATABASE_URL`, `REDIS_URL`, `AMQP_URL`)
- **Componentes separados:** `_HOST`, `_PORT`, `_NAME`, `_USER` cuando se necesita granularidad
- **Secretos:** sufijo `_SECRET`, `_KEY`, `_PASSWORD`, `_TOKEN` — nunca `_PASS`
- **Booleanos:** `ENABLE_*` o `*_ENABLED` con valores `true`/`false`
- **APIs externas:** `<SERVICE>_API_KEY`, `<SERVICE>_API_URL`

### Nombres estándar por infraestructura

| Infraestructura | Variables estándar |
|---|---|
| **Base de datos** | `DATABASE_URL`, `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`, `DB_SSL_MODE`, `DB_MAX_OPEN_CONNS` |
| **Redis** | `REDIS_URL`, `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB` |
| **Kafka** | `KAFKA_BROKERS`, `KAFKA_GROUP_ID`, `KAFKA_TOPIC_PREFIX`, `KAFKA_SASL_USERNAME`, `KAFKA_SASL_PASSWORD` |
| **NATS** | `NATS_URL`, `NATS_TOKEN`, `NATS_CLUSTER_ID` |
| **RabbitMQ** | `AMQP_URL`, `RABBITMQ_HOST`, `RABBITMQ_PORT`, `RABBITMQ_USER`, `RABBITMQ_PASSWORD`, `RABBITMQ_VHOST` |
| **AWS** | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`, `S3_BUCKET`, `SQS_QUEUE_URL`, `SNS_TOPIC_ARN` |
| **GCP** | `GOOGLE_APPLICATION_CREDENTIALS`, `GOOGLE_CLOUD_PROJECT`, `GCS_BUCKET` |
| **SMTP** | `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_FROM` |
| **Auth/JWT** | `JWT_SECRET`, `JWT_EXPIRY`, `JWT_ISSUER`, `OAUTH_CLIENT_ID`, `OAUTH_CLIENT_SECRET` |
| **HTTP server** | `PORT`, `HOST`, `APP_ENV`, `LOG_LEVEL` |
| **Observabilidad** | `OTEL_SERVICE_NAME`, `OTEL_EXPORTER_OTLP_ENDPOINT`, `SENTRY_DSN` |

**Regla:** usar los nombres estándar de la tabla cuando existan. No inventar nombres propios para infra conocida (ej. no usar `CACHE_ADDR` cuando el estándar es `REDIS_URL`).

### `.env.example`

El developer DEBE agregar al `.env.example` del proyecto todas las env vars nuevas con valores placeholder:

```env
# Feature: <TASK-ID>
KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=my-service
KAFKA_SASL_PASSWORD=change-me
```

**Reglas del `.env.example`:**
- Se commitea a git — es documentación de configuración requerida
- Valores placeholder que muestren el formato esperado, nunca secrets reales
- Agrupar por feature o servicio con comentarios
- `.env` (con valores reales) va en `.gitignore` — nunca en el repo

### Mapping a Docker / Kubernetes

| Herramienta | Config no-sensible | Secrets |
|---|---|---|
| **docker-compose** | `environment:` o `env_file: .env` | `secrets:` de Docker Swarm, o `.env` excluido de imagen |
| **K8s ConfigMap** | `envFrom: configMapRef` | `envFrom: secretRef` o `secretKeyRef` individual |

ConfigMap/Secret names en `kebab-case` (ej. `app-config`). Keys internas en `SCREAMING_SNAKE_CASE`.

```

## Estrategia de persistencia

- **Concurrencia:** ...
- **Idempotencia:** clave de idempotencia, ventana de deduplicación
- **Reintentos / backoff:** política, límite de intentos
- **Manejo de fallos:** qué pasa si el downstream no responde
```

## Archivos de spec ejecutables (OBLIGATORIO para tareas Medium+)

Los fragmentos OpenAPI/AsyncAPI/proto en este documento son **borradores**. El arquitecto DEBE también generar los archivos ejecutables que el developer y el tester consumen:

### Ubicación de archivos

| Tipo | Ubicación | Formato |
|---|---|---|
| REST | `api/openapi.yaml` | OpenAPI 3.1.0 |
| Eventos | `api/asyncapi.yaml` | AsyncAPI 3.0.0 |
| gRPC | `proto/<package>/v1/<service>.proto` | Protocol Buffers 3 |

**Monorepo:** `services/<svc>/api/` para OpenAPI/AsyncAPI. `proto/` compartido en raíz con subdirectorios que reflejan el package name.

### Reglas de archivos spec

- **El archivo spec es la fuente de verdad** — el fragmento en `architecture-backend.md` es documentación de diseño, el archivo en `api/` o `proto/` es el contrato ejecutable
- Si el archivo ya existe → **extenderlo** con los nuevos endpoints/eventos/RPCs. No sobrescribirlo
- Si el archivo no existe → **crearlo** con la estructura mínima (info, servers, los contratos de esta tarea)
- El developer implementa contra el archivo spec, no contra el fragmento del markdown
- El tester usa el spec para informar los tests de Hurl (assertions de schema) y Schemathesis (fuzzing)
- El spec se versiona en git junto al código — los cambios al spec son parte del PR

### Validación

```bash
# OpenAPI
npx @redocly/cli lint api/openapi.yaml

# AsyncAPI
asyncapi validate api/asyncapi.yaml

# Proto (buf)
buf lint proto/
```

### Cuándo NO generar archivos spec

- **Tareas Small** que no tocan endpoints, eventos ni RPCs → no aplica
- **Tauri commands / IPC** → no hay estándar de spec ejecutable; documentar solo en `architecture-backend.md`
- **Endpoints internos sin consumidores externos** → evaluar; si solo lo consume el frontend del mismo repo, el spec es recomendado pero no obligatorio

## Reglas

- Usar SOLO las secciones que apliquen — omitir secciones vacías completamente
- OpenAPI es la fuente de verdad para contratos REST; AsyncAPI para eventos — no prosa
- Cada schema de evento necesita un `eventId` (clave de idempotencia) y `occurredAt`
- Garantías de entrega, orden y estrategia de DLQ son obligatorias para cualquier sección async
- La taxonomía de errores debe clasificar errores como retryable vs fatal — no solo códigos HTTP
- Los diagramas de secuencia muestran happy path + path de fallo principal
- Si existe frontend, los schemas REST/command aquí son canónicos — frontend los deriva de estos
