# Dominio — anvil

last_updated: 2026-08-13

<!-- anvil no tiene un único dominio; el repo se organiza en bounded contexts bajo internal/.
     Cada sección de abajo sigue la estructura del template domain.tmpl.md. -->

## Dominio: cli

### Responsabilidad
Punto de entrada y dispatch de todos los subcomandos de la CLI (`anvil <subcomando> [args]`).

### Archivos clave
```
internal/cli/
├── cli.go              — Run(args) — dispatch principal, detecta subcomando
├── cmd_migrate.go       — subcomando de migraciones
├── run.go               — subcomando run (dispara agentes/pipelines), selección de proveedor LLM
├── registry.go          — subcomando de registry (928 líneas — candidato a split)
├── emit_translate.go    — traducción de eventos emitidos (781 líneas)
├── deploy.go, doctor.go, stats.go, status.go, ... — un archivo por subcomando
```

### Flujo principal
```
anvil <cmd> <args> → cli.Run() → switch temprano (comandos sin config) o config.Load() → cmd<Nombre>(cfg, args)
```

### Patrones usados
- **Subcommand Dispatch:** `internal/cli/cli.go` — ver `Core/patterns.md`

### Dependencias de este dominio
- `pkg/config`, `pkg/state`, `pkg/output`, `pkg/gitutil`, `pkg/fileutil` — utilidades transversales
- `internal/mcp`, `internal/orchestrator`, `internal/memory` — lógica de dominio delegada

### Gotchas
- Algunos subcomandos (`emit`, `verify-direct`, `clean-empty-runs`, `capture`, `mcp-server`) se atienden en un switch temprano en `cli.go` **antes** de cargar `config.Load` — para poder correr fuera de un directorio de proyecto con `anvil.yaml`. Nuevo subcomando que no necesite config de repo debe seguir este mismo patrón.

### Deuda técnica
- `internal/cli/registry.go` (928 líneas) y `internal/cli/emit_translate.go` (781 líneas) — candidatos a refactor por tamaño.

---

## Dominio: orchestrator

### Responsabilidad
Orquesta pipelines multi-agente: resuelve el orden de ejecución (DAG), aplica puntos de aprobación (gates) y replanifica ante fallos.

### Archivos clave
```
internal/orchestrator/
├── dag.go        — grafo de dependencias entre tasks
├── gate.go       — puntos de decisión/aprobación humana o automática
├── executor.go   — ejecución del DAG resuelto (557 líneas)
├── replanner.go  — reajuste de plan ante fallos
└── types.go      — tipos e interfaces compartidas
```

### Patrones usados
- **DAG + Gate + Executor:** ver `Core/patterns.md`

### Dependencias de este dominio
- Tipos compartidos en `types.go`; consumido por `internal/cli` (subcomando `run`/pipeline) y `internal/mcp` (tools de orquestación)

### Deuda técnica
- `executor.go` (557 líneas) — candidato a revisión de responsabilidad única.

---

## Dominio: memory

### Responsabilidad
Captura, persiste y sirve memoria de sesión (transcripts, digests) usando SQLite (FTS5 + sqlite-vec) y múltiples backends de inferencia.

### Archivos clave
```
internal/memory/
├── capture/     — captura de eventos/transcripts en vivo
├── claude/      — cliente hacia Claude API (Anthropic)
├── ollama/      — cliente hacia Ollama local (fallback sin API key)
├── haiku/       — uso específico de modelo Haiku para tareas ligeras (digest de terminal)
└── transcript/  — parseo/almacenamiento de transcripts de sesión
```

### Flujo principal
```
Evento de sesión → capture/ → digest (claude/haiku si ANTHROPIC_API_KEY, si no ollama/) → persistencia SQLite (pkg/storage)
```

### Dependencias de este dominio
- `pkg/storage` (SQLite, requiere build tag `fts5`)
- `ANTHROPIC_API_KEY` (externo, opcional) — ver `Technical domain/dependencies.md`

### Gotchas
- Si ni `ANTHROPIC_API_KEY` ni Ollama están disponibles, el digest de terminal se omite silenciosamente (`internal/cli/run.go:552`) — no falla, pero pierde ese resumen.

---

## Dominio: mcp

### Responsabilidad
Servidor MCP embebido en el propio binario — expone operaciones de memoria, orquestación e inventario como tools invocables por Claude Code.

### Archivos clave
```
internal/mcp/
├── server.go         — bootstrap del servidor MCP
├── tools.go          — registro de tools
├── context.go        — tools relacionadas a contexto/handoff (632 líneas)
├── orchestration.go  — tools de orquestación (499 líneas)
├── execution.go, inventory.go, utilities.go
```

### Interfaces públicas
```go
func (s *Server) <tool>(ctx context.Context, args map[string]any) (string, error)
```

### Dependencias de este dominio
- `internal/memory`, `internal/memory/ollama` (ver import en `context.go`)

### Deuda técnica
- `context.go` (632 líneas) — candidato a split por responsabilidad (parseo de handoff, digest, tools de contexto mezclados).

---

## Dominio: dashboard

### Responsabilidad
Modelos, queries y writers para el dashboard Wails (`anvil-full`) que visualiza runs, agentes, archivos y flujos.

### Archivos clave
```
internal/dashboard/
├── entity/    — agent.go, event.go, file.go, file_edge.go, flow.go, prompt.go, run.go, task.go, tool.go
├── dto/       — data transfer objects hacia el frontend
├── mapper/    — mapeo entity → dto
├── query/     — queries de lectura sobre SQLite
├── writer/    — escritura de eventos/datos
└── testutil/  — utilidades de test compartidas
```

### Gotchas
- El dashboard (`anvil-full`) es un **modo de build** del mismo binario (`go build -tags "dashboard production fts5"`), no un servicio separado — confirmado explícitamente por el humano en este scan.

---

## Dominio: deploy / instrumentation / runner / tui

### Responsabilidad
- `deploy/` — lógica de instalación/deploy del binario y artefactos relacionados.
- `instrumentation/` — telemetría/instrumentación interna.
- `runner/` — ejecución de runs (probablemente el motor detrás de `anvil run`).
- `tui/` — interfaz de terminal interactiva (bubbletea) — `browse.go` (851 líneas) / `browse_test.go` (977 líneas).

**Gap:** estos cuatro dominios no fueron inspeccionados en profundidad en este bootstrap (sondeo grep-first, presupuesto de líneas) — profundizar en un rescan `deep` si una tarea los toca directamente.
