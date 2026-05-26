# Proyecto — Anvil

last_updated: 2026-05-08

## Objetivo

Anvil es una CLI + servidor MCP para orquestar pipelines multi-agente de AI (Claude Code, Codex, Gemini, etc.). Captura telemetría de cada run, genera digests semánticos de los outputs, y expone los datos vía MCP a cualquier cliente AI-compatible. Es el "backend de memoria y coordinación" para flujos de desarrollo asistido por AI.

## Restricciones no negociables

- Build tag `fts5` obligatorio — SQLite FTS5 y sqlite-vec son requeridos; build sin el tag produce binario incompleto
- El binario de producción debe instalarse en `$HOME/bin/anvil` — los hooks de Claude Code en `~/.claude/settings.json` apuntan a esa ruta; instalar en otro lugar deja los hooks apuntando a un binario stale
- CGO requerido para el binario con dashboard (`anvil-full`); el binario CLI puro (`anvil`) no requiere CGO
- No hay servidor HTTP público — la superficie de red es MCP sobre stdio y el dashboard Tauri 2 (ventana nativa)
- SQLite como única base de datos — sin PostgreSQL, sin Redis, sin brokers externos

## Stack

| Componente | Tecnología | Versión |
|-----------|-----------|---------|
| Backend / CLI | Go | 1.25.0 |
| Dashboard (opcional) | Tauri 2 + React + Vite (repo separado: anvil-dashboard) | — |
| Base de datos | SQLite + sqlite-vec (vector search) + FTS5 | — |
| Transporte MCP | JSON-RPC 2.0 sobre stdio | — |
| Agentes AI | Claude Code (`claude --print`), Codex, Gemini, OpenCode, Cursor | — |
| Infra | Sin Docker en producción; `docker-compose.test.yml` solo para tests de integración | — |

## Estilo arquitectónico

- **Estilo principal:** Layered monolito con bounded contexts explícitos en `internal/`
- **Capas:** `cmd/anvil/` → `internal/cli/` → dominios (`internal/memory/`, `internal/orchestrator/`, etc.) → `pkg/` (utilidades compartidas)
- **Módulo raíz (Go):** `github.com/ernesto2108/anvil`
- **Convención de paths:** `internal/<domain>/<concern>.go`

## SOLID detectado

| Principio | Estado | Observación |
|-----------|--------|-------------|
| SRP | En riesgo | `internal/tui/browse.go` (851 líneas), `internal/cli/emit_translate.go` (781 líneas), `internal/cli/registry.go` (776 líneas), `internal/cli/run.go` (684 líneas), `internal/mcp/context.go` (632 líneas) |
| OCP | En riesgo | `internal/cli/registry.go:378,591` — switch sobre `it.Type`; agregar tipo requiere editar el switch |
| LSP | No evaluado | — |
| ISP | OK | Interfaces pequeñas (1-3 métodos): `Embedder`, `Summarizer`, `AgentRunner`, `EventSink`, `GateHandler` |
| DIP | Mixto | La mayoría de `NewX` retornan concretos (`*Orchestrator`, `*Reader`, `*Client`); las interfaces existen pero no siempre se inyectan |

## Convenciones establecidas

- Errores envueltos con `fmt.Errorf("domain: %w", err)` — prefijo de paquete siempre
- `context.Context` es siempre el primer argumento en funciones que acceden a I/O o DB
- Tests de integración usan `docker-compose.test.yml` o SQLite en memoria con `vecAutoOnce sync.Once`
- `sync.Once` para registrar `sqlite_vec.Auto()` exactamente una vez por proceso (`pkg/storage/sqlite.go:19`)
- Build tags: `fts5` para CLI, `dashboard production fts5` para binario con UI

## Qué NO introducir

- Brokers de mensajes externos (NATS, RabbitMQ, Kafka) — no existen y la arquitectura no los requiere
- ORM — se usa `database/sql` directo con migraciones (`golang-migrate`)
- Framework HTTP — no hay servidor REST; el único "servidor" es MCP sobre stdio
- Dependencias de paquetes de terceros para logging — se usa `output` propio en `pkg/output`
