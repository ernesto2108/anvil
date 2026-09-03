# Coding Standards — anvil

last_updated: 2026-08-13

## Idioma del código

- **Código fuente:** inglés
- **Comentarios:** inglés
- **Commits:** inglés — Conventional Commits (forzado por `CLAUDE.md` global — regla de "Idioma del contenido de git")
- **Documentación técnica (`.project-context/`, skills, agents):** español

## Naming

### General
- **Variables y funciones:** `camelCase` (privado) / `PascalCase` (exportado) — idiomático Go
- **Constantes:** `PascalCase` para códigos de error exportados (ej. `BadRequestErr`), no `UPPER_SNAKE_CASE` — ver `pkg/errors/error-definition.go`
- **Archivos:** `snake_case.go` (ej. `cmd_migrate.go`, `clean_empty_runs.go`)
- **Tipos / Interfaces / Structs:** `PascalCase`

### Por dominio
- **Subcomandos CLI:** `cmd<Nombre>` como archivo (`cmd_migrate.go`) y función `cmd<Nombre>(...)` en `internal/cli/`
- **Tests:** `<archivo>_test.go` co-ubicado junto al archivo bajo prueba (convención estándar de Go)

## Estructura de carpetas

```
anvil/
├── cmd/anvil/            — entrypoint del binario (main.go), único punto de entrada
├── internal/             — código privado del módulo (no importable externamente)
│   ├── cli/              — subcomandos y despacho de la CLI (internal/cli/cli.go → Run)
│   ├── orchestrator/      — DAG, gates, executor, replanner de pipelines multi-agente
│   ├── mcp/               — servidor MCP embebido (server.go, tools.go, context.go, orchestration.go)
│   ├── memory/            — capa de memoria persistente (claude/, ollama/, haiku/, capture/, transcript/)
│   ├── dashboard/         — modelos y queries del dashboard Wails (dto/, entity/, mapper/, query/, writer/)
│   ├── deploy/            — lógica de deploy/instalación
│   ├── instrumentation/   — telemetría/instrumentación
│   ├── runner/            — ejecución de runs
│   └── tui/               — interfaz TUI (bubbletea)
├── pkg/                  — código reusable, potencialmente importable
│   ├── config/, errors/, fileutil/, frontmatter/, gitutil/, output/, registry/, state/, storage/
├── migrations/           — SQL de golang-migrate (NNNNNN_<nombre>.up/down.sql)
├── skills/, agents/, commands/ — sistema de agentes de Claude Code (solo modificables por agent-designer)
├── vault-template/       — template Obsidian que anvil genera para otros proyectos (NO es la gestión de anvil mismo)
└── frontend/             — frontend del dashboard Wails (React, build embebido para anvil-full)
```

## Reglas de imports / dependencias

- `internal/` no es importable fuera del módulo `github.com/ernesto2108/anvil` (restricción de Go, no de convención)
- `internal/cli` orquesta subcomandos pero delega lógica de dominio a `internal/mcp`, `internal/orchestrator`, `internal/memory`, etc. — no se detectó lógica de negocio pesada directamente en `cli.go`
- El paquete `pkg/errors` centraliza códigos de error (`AnvilMeshCode`) reusados cross-domain

## Linting configurado

| Herramienta | Config | Reglas destacadas |
|---|---|---|
| `go vet` | sin flags custom (`make vet`) | vet estándar de Go |
| — | no se detectó `.golangci.yml`/`.golangci.yaml` en la raíz | gap — no hay linter configurado más allá de `go vet` |

## Model tiers eliminados (2026-09-03)

- El mecanismo de resolución de model tiers (`high`/`medium`/`low`) fue eliminado por completo del deploy. Los agentes ya no declaran `model` en su frontmatter; cada herramienta usa el modelo default de su sesión.
- `pkg/config`: se eliminaron `TierMap`, `TierHigh/Medium/Low`, `IsTier()` y `App.ResolveTier()`. `ProviderConfig.Providers` pasó de `map[string]TierMap` a `[]string` — solo enumera nombres de provider válidos para `anvil provider <name>` y el tool MCP `switch_provider`.
- `anvil.config.yaml`: el bloque `providers:` es ahora una lista simple de nombres (`- claude`, `- opencode`, ...), sin mapeo de modelos. El bloque `permissions:` no cambió.
- `internal/deploy/claude.go`: `adaptClaude` ya no reescribe el campo `model` — el frontmatter del agente se emite verbatim (si un agente aún declarara `model`, pasa sin resolver).
- `internal/deploy/opencode.go`: el template de salida ya no emite la línea `model:`.
- `internal/mcp/inventory.go`: `listAgents` ya no expone el campo `Model` en el inventario.
- `internal/cli/registry.go` y `internal/cli/doctor.go`: se removieron los tres bloques de resolución de tier y el check de "provider tiers" en doctor.

## Patrones prohibidos

- **Build sin `-tags fts5`:** el binario requiere el build tag `fts5` para SQLite FTS5 y `sqlite-vec` — un build sin ese tag no expone la funcionalidad de búsqueda/memoria completa. `make build`/`make install` ya lo incluyen; no remover el tag.

## Patrones de diseño detectados en el código

<!-- Ver Core/patterns.md para el detalle completo por categoría -->

Ver `Core/patterns.md`.
