---
name: go-conventions
description: Convenciones y estándares de código para backend en Go. Usar cuando se escriba código Go, se revisen patrones Go, o el usuario mencione "go conventions", "idiomatic Go", "error handling in Go", "Go best practices", "Go testing patterns", o cuando se trabaje con archivos .go.
---

# Go Conventions

> **IMPORTANTE:** Este archivo es un despachador ligero. NO cargues todos los archivos referenciados a la vez. Lee la tabla de enrutamiento abajo, identifica qué archivos son relevantes para la tarea actual, y carga SOLO esos usando la herramienta Read. Cada archivo pesa ~3-5KB. Cargar archivos innecesarios desperdicia tokens de contexto.

## Stack y Filosofía

- **Go stdlib primero** — agrega dependencias solo cuando la stdlib es genuinamente insuficiente
- **Simplicidad sobre inteligencia** — si necesita un comentario para explicarse, simplificarlo
- **Explícito sobre implícito** — sin magia, sin efectos secundarios en `init()`, sin estado global
- **Los errores son valores** — manéjalos, no los ocultes
- **Composición sobre herencia** — embeber, no extender

## Señales de Alerta (siempre detener el trabajo)

- `panic()` fuera de `main()` → error
- `init()` haciendo trabajo real → error
- Errores ignorados (`_ = f()`) → error
- Estado global mutable → error
- Fugas de recursos (tickers sin cerrar, deferred en loops) → error

## Detección de Anti-Patrones

**Detección pasiva:** Al revisar código Go, carga `detection/anti-patterns.md` y escanea los patrones `error` y `warning`. Reporta como `[file:line] [severity] [category] anti-pattern-name`.

**Detección activa:** Cuando el usuario pide "improve", "refactor", "optimize" o "clean" — reporta también patrones de nivel `suggestion` y propone correcciones referenciando la regla o guía relevante.

## Qué Cargar

Carga **solo** los archivos relevantes para la tarea actual:

### Reglas (referencia rápida, ~2-3KB cada una)

| Trabajando en... | Cargar |
|---|---|
| Manejo de errores, naming, context, concurrencia básica | `rules/coding.md` |
| Imports, DTOs, validación, DI | `rules/architecture.md` |
| SQL, transacciones, repositories | `rules/database.md` |
| Reglas críticas de Kafka/RabbitMQ | `rules/messaging.md` |
| Functional options, constructores, guard clauses | `rules/patterns.md` |

### Guías (patrones detallados con código, ~3-5KB cada una)

| Trabajando en... | Cargar |
|---|---|
| Qué primitiva de concurrencia usar | `guides/concurrency/decision-matrix.md` |
| Patrón fan-out/fan-in | `guides/concurrency/fan-out-fan-in.md` |
| Worker pools, concurrencia acotada | `guides/concurrency/worker-pools.md` |
| Etapas de pipeline | `guides/concurrency/pipelines.md` |
| Timeouts, cancelación de context | `guides/concurrency/timeout-cancellation.md` |
| Rate limiting | `guides/concurrency/rate-limiting.md` |
| Graceful shutdown (HTTP, workers) | `guides/concurrency/graceful-shutdown.md` |
| Acceso concurrente a map | `guides/concurrency/concurrent-map.md` |
| Pub/sub event broadcasting | `guides/concurrency/pub-sub.md` |
| Anti-patrones de concurrencia + checklist | `guides/concurrency/anti-patterns.md` |
| Context en HTTP client, DB, Redis, gRPC | `guides/cleanup/context-propagation.md` |
| Diseño de timeout multi-nivel | `guides/cleanup/timeout-architecture.md` |
| Rows, transacciones, HTTP body, tickers | `guides/cleanup/resource-cleanup.md` |
| Configuración del pool sql.DB, monitoreo | `guides/cleanup/connection-pools.md` |
| Checklist de recursos, linters, detección en producción | `guides/cleanup/detection-checklist.md` |
| Estructura de tests, tests basados en tabla | `guides/testing/structure-tables.md` |
| Test helpers, mocking con interfaces | `guides/testing/helpers-mocking.md` |
| Testing de HTTP handlers (Gin) | `guides/testing/http-handlers.md` |
| Testing de repositories (mock rows) | `guides/testing/repositories.md` |
| Fixtures, testdata, tests de integración | `guides/testing/fixtures-integration.md` |
| Cobertura, benchmarks | `guides/testing/coverage-benchmarks.md` |
| Kafka overview, selección de librería | `guides/kafka/overview.md` |
| Patrones de productor Kafka | `guides/kafka/producer.md` |
| Consumer Kafka, grupos, ordenamiento | `guides/kafka/consumer.md` |
| Kafka DLQ, retry, mensajes envenenados | `guides/kafka/dlq-retry.md` |
| Kafka circuit breaker, idempotencia, backpressure | `guides/kafka/resilience.md` |
| Kafka shutdown, tracing, schema, anti-patrones | `guides/kafka/operations.md` |
| RabbitMQ overview, exchanges, durabilidad | `guides/rabbitmq/overview.md` |
| Conexión RabbitMQ, auto-reconexión | `guides/rabbitmq/connection.md` |
| Productor RabbitMQ, confirms | `guides/rabbitmq/producer.md` |
| Consumer RabbitMQ, QoS, ack/nack | `guides/rabbitmq/consumer.md` |
| RabbitMQ DLX/DLQ, cadenas de retry con TTL | `guides/rabbitmq/dlq-retry.md` |
| Backpressure RabbitMQ, shutdown, anti-patrones | `guides/rabbitmq/operations.md` |
| Logging estructurado (slog) | `guides/slog.md` |
| Health checks, Prometheus, OpenTelemetry | `guides/observability.md` |
| Composición de middleware HTTP | `guides/middleware.md` |
| SQL injection, crypto, validación de input | `guides/security.md` |
| Reglas `//go:embed`, build tags, builds de desktop con Wails, CGO_LDFLAGS | `guides/embed-and-desktop-builds.md` |

### Detección y Checklists

| Cuándo... | Cargar |
|---|---|
| Revisión de código | `detection/anti-patterns.md` |
| Antes de escribir código Go | `checklists/pre.md` |
| Después de escribir código Go | `checklists/post.md` |

### Ejemplos (patrones buenos y malos por dominio, ~2-3KB cada uno)

| Trabajando en... | Cargar |
|---|---|
| Patrones de manejo de errores | `examples/errors.md` |
| Arquitectura, interfaces, DI | `examples/architecture.md` |
| Patrones de testing | `examples/testing.md` |
| Base de datos, repositories, DTOs | `examples/database.md` |
| Validación de entidades | `examples/validation.md` |
| Concurrencia, shutdown | `examples/concurrency.md` |
| Flujo completo Handler → Service → Repo, reglas de wrapping de errores | `examples/service-contracts.md` |

## Gate Post-Implementación

Después de CUALQUIER cambio de código en archivos `.go`, invoca la skill `/lint` antes de considerar la tarea terminada.
