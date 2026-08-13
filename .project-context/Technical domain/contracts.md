# Contratos — anvil

last_updated: 2026-08-13

<!-- Bordes del sistema: qué expone, qué consume, qué emite -->

## REST API

anvil no expone una API REST pública de servidor — es una CLI. `pkg/output/rest/client.go` existe como **cliente** REST saliente (consumido internamente), no como servidor expuesto.

### Cliente REST saliente
- **Archivo:** `pkg/output/rest/client.go`
- **Auth:** no confirmado en este scan — gap, revisar en rescan `deep` si se toca este cliente
- **Uso:** salida/reporte de eventos hacia un endpoint externo (uso exacto no confirmado — gap)

## Message Queues / Event Streams

No se detectaron brokers de mensajería (NATS, RabbitMQ, Kafka, Redis Streams) en `go.mod` ni en `docker-compose.*`. anvil no usa mensajería asíncrona externa.

## Servicios externos

### Claude API (Anthropic)
- **Cliente:** `internal/memory/claude/`, invocado desde `internal/cli/run.go`, `dream.go`, `emit_translate.go`, `capture.go`
- **Auth:** `ANTHROPIC_API_KEY` (variable de entorno) — confirmado por el humano: es la única auth externa del sistema
- **Operaciones usadas:** generación de digests/resúmenes de sesión, traducción de eventos (`emit_translate.go`)
- **Timeout configurado:** no confirmado en este scan — gap
- **Fallback:** si `ANTHROPIC_API_KEY` no está seteada, cae a Ollama local (`internal/memory/ollama/`); si tampoco Ollama está disponible, el paso se omite (ver `Technical domain/domain.md` → dominio `memory`, gotcha de digest)

### Ollama (local)
- **Cliente:** `internal/memory/ollama/`
- **Auth:** ninguna — servicio local
- **Operaciones usadas:** fallback de inferencia cuando no hay `ANTHROPIC_API_KEY`

## WebSockets

No se detectó uso de WebSockets en el sondeo grep-first.

## gRPC

`google.golang.org/grpc` v1.74.2 está declarado en `go.mod`, pero no se confirmó su uso concreto (servidor o cliente) en este bootstrap — **gap**, revisar en rescan `deep` si una tarea toca esta dependencia.

## Contratos internos entre dominios

### Server (MCP tools)
- **Definida en:** `internal/mcp/server.go`, `tools.go`
- **Implementada por:** `internal/mcp/*.go` (context.go, orchestration.go, execution.go, inventory.go, utilities.go)
- **Consumida por:** Claude Code, vía el subcomando `anvil mcp-server` (`internal/cli/cli.go`)

### Orchestrator types (DAG/Gate/Executor)
- **Definida en:** `internal/orchestrator/types.go`
- **Implementada por:** `dag.go`, `gate.go`, `executor.go`, `replanner.go`
- **Consumida por:** `internal/cli` (subcomando de pipeline/run), `internal/mcp` (tools de orquestación)

## Modelo de autenticación y autorización

**Confirmado por el humano en este run:** no hay auth entre servicios — anvil es un monolito CLI, no hay servicios separados que autenticarse entre sí. La única autenticación del sistema es saliente, hacia Claude API vía `ANTHROPIC_API_KEY`. Ver `Technical domain/business-rules.md` para el detalle completo del modelo.
