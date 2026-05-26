# Dominio: mcp

last_updated: 2026-05-08

## Responsabilidad

Implementar el servidor MCP (Model Context Protocol) de Anvil sobre JSON-RPC 2.0 / stdio. Expone internals de Anvil (runs, memoria, pipelines, skills, contexto de vault) como herramientas invocables por cualquier cliente AI-compatible.

## Archivos clave

```
internal/mcp/
├── server.go          — Server struct, Serve(), loop JSON-RPC sobre stdin/stdout
├── tools.go           — buildRegistry(): define y registra todas las MCP tools
├── orchestration.go   — handlers de tools de orquestación de pipelines
├── context.go         — handlers de tools de contexto (vault, Obsidian, memoria) (632 líneas)
├── execution.go       — handlers de ejecución de agentes individuales
├── inventory.go       — handlers de inventory (skills, agents, files)
├── utilities.go       — helpers compartidos entre handlers
```

## Flujo principal

```
Cliente AI → stdin (JSON-RPC request)
→ server.Serve() deserializa request
→ buildRegistry() lookup por method name
→ handler específico (context, orchestration, execution, inventory)
→ respuesta JSON-RPC → stdout
```

## Patrones usados

- **Registry / dispatch:** `buildRegistry()` en `tools.go` — mapa de nombre de tool a handler function
- **Facade:** el servidor MCP es una facade sobre todos los dominios internos de Anvil

## Dependencias de este dominio

- `internal/memory` — búsqueda de digests, embeddings
- `internal/dashboard/query` — Reader para datos de runs/tasks/agents
- `internal/orchestrator` — tipos de pipeline
- `pkg/config` — `config.App` compartido

## Variables de entorno relevantes

- `ANVIL_VAULT_PATH` — path al vault Obsidian; si está vacío, las tools de vault retornan error manejado

## Gotchas

- `context.go` (632 líneas) mezcla lógica de vault Obsidian, búsqueda semántica y formateo de contexto — candidato a split
- El servidor no tiene autenticación — cualquier proceso con acceso a stdin/stdout puede invocar todas las tools
- Los tests de orquestación (`orchestration_test.go`, 898 líneas) son los más grandes del repo — setup complejo con DB in-memory

## Deuda técnica

- `context.go` (632 líneas) — mezcla múltiples concerns de contexto; refactor a subpaquetes recomendado
