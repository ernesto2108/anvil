# Dependencies — anvil

<!-- Grafo de dependencias entre dominios. -->

last_updated: 2026-08-13

## Grafo de dependencias

<!-- Tipo: sync (llamada directa), async (evento/queue), data (FK / esquema compartido) -->

| Dominio | Depende de | Tipo | Notas |
|---------|-----------|------|-------|
| `cli` | `orchestrator` | sync | subcomando de pipeline/run despacha ejecución al DAG |
| `cli` | `memory` | sync | subcomandos `run`, `dream`, `emit-translate`, `capture` invocan la capa de memoria/LLM |
| `cli` | `mcp` | sync | subcomando `mcp-server` levanta el servidor MCP embebido |
| `mcp` | `memory` | sync | tools de contexto (`digestFromHandoff`) llaman directamente a `internal/memory`, `internal/memory/ollama` |
| `mcp` | `orchestrator` | sync | tools de orquestación exponen el DAG/gates como tools MCP |
| `memory` | `pkg/storage` | data | persistencia de digests/transcripts en SQLite (requiere build tag `fts5`) |
| `dashboard` | `pkg/storage` | data | queries de lectura sobre el mismo esquema SQLite que usa `memory` |
| `orchestrator` | — | — | sin dependencias hacia otros dominios internos detectadas (consumido por `cli` y `mcp`, no al revés) |

## Impacto de cambios

Antes de modificar un dominio, consultar la tabla del grafo para identificar los
dominios **downstream** afectados (los que dependen del dominio que vas a tocar).

- Si la dependencia es `sync` → un cambio de contrato rompe a quien depende de inmediato. Ej.: cambiar la firma de una tool en `internal/mcp` rompe el consumo desde Claude Code.
- Si la dependencia es `data` → verificar migraciones y esquema compartido antes de alterar tablas usadas tanto por `memory` como por `dashboard`.

Listar los dominios downstream en el plan de cambio y validar cada uno antes de cerrar.

## Dependencias externas

<!-- Servicios externos que el sistema consume: APIs de terceros, DBs externas, colas -->

| Servicio externo | Tipo | Consumido por | Notas |
|------------------|------|---------------|-------|
| Anthropic Claude API | API REST (via SDK) | `memory` (`internal/memory/claude/`), `cli` (`run.go`, `dream.go`, `emit_translate.go`, `capture.go`) | Auth: `ANTHROPIC_API_KEY`. Sin SLA/rate-limit documentado en el repo — gap |
| Ollama (local) | API REST local | `memory` (`internal/memory/ollama/`) | Fallback sin auth cuando `ANTHROPIC_API_KEY` no está presente; requiere que el usuario tenga Ollama corriendo localmente |
