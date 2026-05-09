# Riesgos y Deuda Técnica — Anvil

last_updated: 2026-05-08

## Gotchas operativos

### Build tag fts5 obligatorio
- **Dónde:** `Makefile` — todos los targets de build
- **Descripción:** Compilar sin `-tags fts5` produce un binario que no puede abrir la DB (sqlite FTS5 + sqlite-vec requieren el tag). El error falla en runtime, no en compilación.
- **Workaround:** Siempre usar `make build` o `make install` — nunca `go build ./...` a secas

### Binario debe instalarse en $HOME/bin/anvil
- **Dónde:** `Makefile` comentario, `INSTALL_DIR`
- **Descripción:** Los hooks de Claude Code en `~/.claude/settings.json` apuntan a `$HOME/bin/anvil`. Instalar en otra ruta deja los hooks apuntando al binario stale anterior.
- **Workaround:** Usar siempre `make install` (target por defecto)

### Ollama debe estar corriendo para embeddings
- **Dónde:** `internal/memory/ollama/lifecycle.go:56`
- **Descripción:** Si Ollama no está en `http://localhost:11434`, los digests se guardan sin embedding — no aparecen en búsquedas semánticas y el contexto de memoria en futuros runs es incompleto
- **Workaround:** `ollama serve` antes de `anvil run`; o configurar `ANTHROPIC_API_KEY` para usar Haiku como fallback de summarización (pero Haiku no hace embeddings)

### sqlite-vec registro único por proceso
- **Dónde:** `pkg/storage/sqlite.go:19`
- **Descripción:** `sqlite_vec.Auto()` debe llamarse exactamente una vez. Los tests usan `sync.Once` (`vecAutoOnce`) — si un test abre una segunda DB sin pasar por `pkg/storage.OpenDB`, puede fallar con "vec0 module not found"
- **Workaround:** Siempre usar `pkg/storage.OpenDB` para abrir conexiones SQLite; en tests usar `testutil.vecAutoOnce`

### ANVIL_PARENT_RUN_ID propagación manual
- **Dónde:** `internal/cli/emit_translate.go:53`
- **Descripción:** Los subprocesos de agentes deben recibir `ANVIL_PARENT_RUN_ID` via env var para que su telemetría se agrupe al run padre. Si el runner no lo propaga, cada agente crea un run huérfano en el dashboard.
- **Workaround:** `ClaudeRunner.RunAgent` lo propaga automáticamente — solo es problema si se implementa un nuevo runner sin considerar esto

## Deuda técnica

### Archivos candidatos a refactor

| Archivo | Líneas | Razón |
|---------|--------|-------|
| `internal/tui/browse.go` | 851 | TUI monolítica — modelo Bubble Tea con UI, navegación y lógica mezclados |
| `internal/cli/emit_translate.go` | 781 | Traducción de eventos + lógica de provider + detección de session mezcladas |
| `internal/cli/registry.go` | 776 | Switch sobre tipos de ítem en dos lugares (OCP en riesgo) |
| `internal/cli/run.go` | 684 | Setup de runner + provider selection + telemetría + lógica de pipeline en un solo archivo |
| `internal/mcp/context.go` | 632 | Vault Obsidian + búsqueda semántica + formateo de contexto mezclados |
| `internal/orchestrator/executor.go` | 557 | Scheduling + retry + gate + event emission en un solo struct |

### Switch sobre tipos (OCP en riesgo)
- `internal/cli/registry.go:378` — switch sobre `it.Type` para render
- `internal/cli/registry.go:591` — switch sobre `it.Type` para validación
- `internal/memory/transcript/parser.go:140` — switch sobre `b.Type` para parsing de bloques

### TODOs y FIXMEs con impacto

No se detectaron TODOs/FIXMEs de alto impacto en el código fuente analizado (solo referencias a secciones de Markdown en `internal/mcp/context.go`).

## Restricciones conocidas

- **macOS only para dashboard:** `make dashboard-build` requiere `-framework UniformTypeIdentifiers` (linker flag de macOS); el binario `anvil-full` no es portable a Linux/Windows sin cambios en el build
- **CGO requerido para dashboard:** el binario `anvil-full` con UI no puede cross-compilarse sin toolchain CGO del target
- **MCP sobre stdio:** no hay autenticación en el servidor MCP — cualquier proceso con acceso al proceso de Anvil puede invocar todas las tools

## Dependencias frágiles

- **sqlite-vec (CGO):** `github.com/asg017/sqlite-vec-go-bindings` requiere CGO — sin CGO el binario no compila con soporte vectorial. Actualizar la versión requiere verificar compatibilidad de la extensión C.
- **`claude --print` como subprocess:** el runner de Claude Code depende del CLI de Claude Code instalado y en `$PATH`. Cambios en la interfaz CLI de Claude Code (flags, output format) rompen el runner sin aviso.
- **Ollama sin retry configurado:** las llamadas HTTP a Ollama en `internal/memory/ollama/client.go` no tienen retry — si Ollama tiene un hiccup transitorio, el digest falla silenciosamente.

## Áreas sin tests

- `internal/cli/run.go` — el flujo completo de run no tiene tests unitarios (solo integración parcial)
- `internal/deploy/` — tests de integración pero coverage de edge cases limitado en `ownership.go`
- `internal/mcp/tools.go` — el registry de tools no tiene tests de cobertura completa
