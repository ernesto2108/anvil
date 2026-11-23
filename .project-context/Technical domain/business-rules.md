# Business Rules — anvil

<!-- Invariantes de negocio que cruzan dominios. -->

last_updated: 2026-08-13

## Invariantes globales

### Build tag `fts5` obligatorio
- **Regla:** todo build de producción/instalación (`make build`, `make install`, `make dashboard-build`) debe incluir el tag `fts5`
- **Dónde se aplica:** `Makefile` (targets `build`, `install`, `dashboard-build`)
- **Por qué:** sin el tag, `sqlite_vec` (búsqueda vectorial/FTS5) no queda registrado — la capa de memoria/búsqueda queda degradada silenciosamente — confirmado como restricción no negociable por el humano

### `sqlite_vec.Auto()` se registra una sola vez por proceso
- **Regla:** la extensión SQLite vec se inicializa exactamente una vez, sin importar cuántas conexiones se abran
- **Dónde se aplica:** `pkg/storage/sqlite.go:15,43` (`sync.Once`)
- **Por qué:** re-registrar el driver de SQLite múltiples veces produce errores o comportamiento indefinido

## Reglas por dominio

### memory
- **Fallback silencioso sin proveedor LLM:** si no hay `ANTHROPIC_API_KEY` ni Ollama saludable, el digest de terminal se omite (no falla el run)
  - Dónde: `internal/cli/run.go:552`
  - Por qué: no bloquear el flujo principal de la CLI por falta de un proveedor de inferencia opcional

### orchestrator
- **Un DAG debe resolverse antes de ejecutarse:** el executor no ejecuta tasks fuera del orden resuelto por `dag.go`
  - Dónde: `internal/orchestrator/executor.go`, `dag.go`
  - Por qué: garantizar que las dependencias entre tasks del pipeline se respeten

## Reglas cross-dominio

### mcp ↔ memory: todo handoff cerrado debe generar memoria
- **Regla:** cuando una tarea se cierra vía `/task-complete` (Claude Code) sin pasar por el runner de la CLI, `digestFromHandoff` debe generar el digest manualmente para no perder ese cierre en la capa de memoria
- **Dominios involucrados:** `mcp`, `memory`
- **Dónde se aplica:** `internal/mcp/context.go` (comentario explícito sobre el propósito de `digestFromHandoff`)
- **Por qué:** cerrar el gap entre tareas cerradas vía Claude Code directo y tareas cerradas vía el runner propio de anvil

## Modelo de autenticación y autorización

<!-- Confirmado explícitamente por el humano en este run -->

### Autenticación entre servicios
- **Mecanismo interno:** ninguno — anvil es un monolito CLI, no hay servicios separados en runtime
- **Razón:** confirmado por el humano: el dashboard Wails (`anvil-full`) es un modo de build del mismo binario, no un proceso separado que requiera autenticarse
- **Servicios que requieren auth interna:** ninguno

### Autenticación hacia el exterior
- **Mecanismo externo:** `ANTHROPIC_API_KEY` (variable de entorno) hacia Claude API
- **Header utilizado:** gestionado internamente por el SDK/cliente de Anthropic usado en `internal/memory/claude/` — no se expone header custom en el código de aplicación
- **Quién valida:** Anthropic (servicio externo) — anvil solo la lee de entorno y falla/degrada si está ausente

### Reglas de autorización
- No hay reglas de autorización entre componentes internos — todo el binario corre con los permisos del usuario local que lo ejecuta.
