# Proyecto — anvil

last_updated: 2026-08-13
task_tool: ""

## Objetivo

anvil es un sistema de gestión de agentes de IA en Go: define agentes, skills y commands una sola vez y los despliega a múltiples targets (Claude Code, OpenCode, Gemini, Codex) vía `deploy_agents`. Orquesta pipelines multi-agente sobre proyectos externos usando DAG y gates, captura y persiste memoria de sesión (transcripts, digests) en SQLite con FTS5/embeddings vectoriales, y expone un servidor MCP embebido para que otros clientes consuman esa memoria y esas capacidades de orquestación. Mantiene un sistema de contexto vivo (`.project-context/`) por proyecto, con un dashboard opcional (Wails + React) para visualizar runs y agentes.

Confirmado por el humano el 2026-08-13.

## Restricciones no negociables

- Build tag `fts5` obligatorio al compilar (`go build -tags fts5 ...`) — sin él, SQLite FTS5 y `sqlite-vec` no quedan disponibles y la capa de memoria/búsqueda queda degradada.
- anvil es un monolito CLI único (`cmd/anvil`) — el dashboard Wails (`anvil-full`) es un modo de build alterno del mismo binario, no un servicio separado; no introducir separación de servicios ni `service-map.yaml` multi-servicio.
- No commitear `.env`/credenciales reales — `ANTHROPIC_API_KEY` se lee de entorno, nunca hardcodeada.

## Stack

| Componente | Tecnología | Versión |
|-----------|-----------|---------|
| Lenguaje principal | Go | 1.25.0 |
| Base de datos | SQLite (`mattn/go-sqlite3` + `sqlite-vec-go-bindings`) | driver v1.14.42 / sqlite-vec v0.1.6 |
| Migraciones | golang-migrate/migrate/v4 | v4.19.1 |
| TUI | charmbracelet/bubbletea + bubbles + lipgloss | v1.3.10 / v1.0.0 / v1.1.0 |
| RPC | google.golang.org/grpc | v1.74.2 |
| Dashboard (opcional) | Wails v2 + React (frontend/) | — |
| LLM providers | Anthropic Claude API (ANTHROPIC_API_KEY), Ollama local (fallback) | — |
| Infra | ninguna — CLI local, sin contenedores en runtime (Docker solo usado para validar migraciones en CI) | — |

## Estilo arquitectónico

- **Estilo principal:** Monolito CLI modular — un único binario (`cmd/anvil`) con subcomandos despachados en `internal/cli`, delegando a paquetes de dominio (`internal/orchestrator`, `internal/memory`, `internal/mcp`, `internal/dashboard`, `internal/deploy`, `internal/instrumentation`, `internal/tui`, `internal/runner`).
- **Capas:** `cmd/anvil` (entrypoint) → `internal/cli` (dispatch de subcomandos) → paquetes de dominio en `internal/*` → `pkg/storage` (SQLite) / `pkg/config`, `pkg/output`, etc. (utilidades transversales)
- **Módulo raíz (Go):** `github.com/ernesto2108/anvil`
- **Convención de paths:** `internal/<dominio>/<archivo>.go`, un archivo por subcomando en `internal/cli/cmd_<nombre>.go` o `<nombre>.go`

## SOLID detectado

| Principio | Estado | Observación |
|-----------|--------|-------------|
| SRP | En riesgo | Archivos grandes detectados: `internal/cli/registry.go` (928 líneas), `internal/cli/emit_translate.go` (781), `internal/cli/run.go` (762), `internal/mcp/context.go` (632), `internal/orchestrator/executor.go` (557) — candidatos a revisión de responsabilidad única |
| OCP | OK / No evaluado a fondo | El dispatch de subcomandos (`internal/cli/cli.go`) se extiende agregando archivos nuevos, sin modificar lógica existente en la mayoría de los casos |
| LSP | No evaluado | Sin evidencia de jerarquías de tipos profundas que amerite evaluación |
| ISP | OK / No evaluado | No se detectaron interfaces grandes (>5 métodos) en el sondeo grep-first |
| DIP | En riesgo | `internal/cli` importa directamente `pkg/config`, `pkg/state`, `pkg/output`, `pkg/gitutil`, `pkg/fileutil` — acoplamiento a paquetes concretos sin capa de interfaces intermedia visible |

## Convenciones establecidas

- Un archivo por subcomando CLI en `internal/cli/`
- Tests co-ubicados (`*_test.go`) junto al archivo bajo prueba
- `go test -race ./...` como comando canónico de test (`make test`)
- Errores de dominio centralizados en `pkg/errors` (`AnvilMeshCode`)

## Qué NO introducir

- Microservicios o separación de `anvil` en múltiples repos/servicios — decisión confirmada: monolito CLI.
- `service-map.yaml` con múltiples servicios — no aplica a este repo.
- Regla de gobernanza de `agents/*.md`/`skills/*/SKILL.md`/`commands/*.md` (solo modificables por `agent-designer`) — **vive únicamente en `CLAUDE.md` global**, por decisión explícita del humano no se duplica aquí.

## Estrategia de migraciones

- **Herramienta:** golang-migrate (`github.com/golang-migrate/migrate/v4` v4.19.1)
- **Directorio:** `migrations/` (formato `NNNNNN_<nombre>.up.sql` / `.down.sql`)
- **Runner:** manual vía subcomando propio del CLI
- **Cómo correr:** subcomando implementado en `internal/cli/cmd_migrate.go` (integración en `pkg/storage/migrate.go`)
- **Cómo hacer rollback:** vía golang-migrate estándar (`.down.sql` correspondiente) — comando exacto no confirmado explícitamente en este scan, inferido de la librería usada
- **Quién aprueba:** no documentado explícitamente — se asume revisión de PR estándar
- **Validación en CI:** `docker-compose.test.yml` + `docker/test-migrations.dockerfile` corren las migraciones contra un contenedor de prueba
- **Restricciones:**
  - Build tag `fts5` no aplica a migraciones SQL puras, pero sí al binario que las ejecuta si usa la capa de storage completa
