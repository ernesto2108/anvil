# Dominio: cli

last_updated: 2026-05-08

## Responsabilidad

Punto de entrada de todos los subcomandos de `anvil`. Orquesta el flujo completo de un run: preflight, construcción del DAG, ejecución con runner, captura de telemetría, generación de digests y gestión de la DB.

## Archivos clave

```
internal/cli/
├── run.go              — comando `anvil run`: preflight, DAG, runner, telemetría (684 líneas)
├── emit.go             — comando `anvil emit`: recibe eventos de hooks de Claude Code
├── emit_translate.go   — traducción de eventos MCP/hooks al schema de instrumentación (781 líneas)
├── registry.go         — gestión del registry de agentes/skills (776 líneas)
├── pipeline.go         — carga y lanza pipelines YAML
├── digest_from_handoff.go — genera Digest desde archivo .handoff/*.md
├── digests.go          — gestión de digests (list, show, search)
├── capture.go          — captura de sesiones de Claude Code
├── migrations.go       — runner de migraciones al inicio
├── provider.go         — selección del proveedor AI (Ollama vs Haiku)
├── preflight.go        — validaciones antes de lanzar el run
├── presets.go          — gestión de presets de pipeline
├── status.go           — `anvil status` — estado del run actual
├── init.go             — `anvil init` — setup inicial del proyecto
├── mcp_server.go       — `anvil mcp` — lanza el servidor MCP
├── doctor.go           — `anvil doctor` — diagnóstico del entorno
└── targets.go          — gestión de targets (Claude, OpenCode, Gemini, Codex)
```

## Flujo principal

```
anvil run <task> → preflight.Check()
→ provider.Select() (Ollama o Haiku según env)
→ memory.search() para contexto de runs previos
→ runner.New(workDir, model, runID, task)
→ orchestrator.executor.Run(dag, runner, emitter)
→ capture.Monitor() para transcript en background
→ memory.capture.orchestrator para digest al finalizar
```

## Patrones usados

- **switch sobre tipo:** `registry.go:378,591` — OCP en riesgo; agregar nuevo tipo de ítem requiere editar el switch
- **sync.Once:** `emit_translate.go:26` (analyzerOnce) — inicialización lazy del analizador de transcripts
- **Strategy:** selección de provider en `provider.go`

## Dependencias de este dominio

- `internal/orchestrator` — DAG + executor
- `internal/runner` — ClaudeRunner
- `internal/memory` — store, search, summarizer, embedder
- `internal/instrumentation` — Emitter para telemetría
- `internal/deploy` — runners de agentes por proveedor
- `pkg/config`, `pkg/storage`, `pkg/errors`, `pkg/output`

## Quién depende de este dominio

- `cmd/anvil/` — main package; registra todos los subcomandos CLI

## Gotchas

- `ANVIL_SKIP_EMIT=1` deshabilita el emit de eventos (`emit.go:32`); útil en tests pero no debe usarse en producción
- `ANVIL_PARENT_RUN_ID` debe estar seteado en subprocesos para que la telemetría se agrupe correctamente — `emit_translate.go:53`
- Las migraciones corren en cada inicio del CLI (`migrations.go`) — si la DB está corrupta, el CLI falla al arrancar

## Deuda técnica

- `emit_translate.go` (781 líneas) y `registry.go` (776 líneas) — ambos superan 300 líneas con responsabilidades mezcladas; candidatos a split
- `run.go` (684 líneas) — mezcla setup de runner, gestión de telemetría y lógica de provider selection
