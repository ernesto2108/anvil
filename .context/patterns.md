# Patrones de Diseño — Anvil

last_updated: 2026-05-08

## Creacionales

### Factory implícita — clientes AI intercambiables
- **Archivos:** `internal/memory/ollama/client.go:20`, `internal/memory/haiku/client.go:25`
- **Qué construye:** Clientes concretos para Embedder (Ollama) y Summarizer (Claude Haiku/Anthropic)
- **Firma detectada:** `func NewClient(...) *Client` — cada paquete expone un `NewClient` con la misma firma de interfaz
- **Cuándo usar:** Al agregar un nuevo proveedor de embedding o summarización → crear nuevo paquete bajo `internal/memory/<provider>/` con `NewClient()` que implemente `memory.Embedder` o `memory.Summarizer`
- **Anti-pattern:** NO instanciar proveedores directamente fuera de sus paquetes; no mezclar lógica de embed y summarización en el mismo struct

### Functional Options (parcial) — REST client
- **Archivo:** `pkg/output/rest/client.go:19`
- **Tipo option:** `type Options func(r *Option)`
- **Cuándo usar:** Al configurar el cliente REST con timeouts o headers opcionales

### Singleton via sync.Once — sqlite-vec autoextension
- **Archivo:** `pkg/storage/sqlite.go:19`
- **Patrón:** `var sqliteVecOnce sync.Once` — garantiza que `sqlite_vec.Auto()` se llama exactamente una vez por proceso
- **Cuándo usar:** Cualquier inicialización de extensión SQLite global debe pasar por este mecanismo
- **Anti-pattern:** No llamar `sqlite_vec.Auto()` directamente fuera de `pkg/storage` — duplicaría el registro

## Estructurales

### Strategy — Embedder / Summarizer
- **Interfaces:** `internal/memory/digest.go:46` (`Embedder`), `internal/memory/digest.go:51` (`Summarizer`)
- **Qué varía:** El proveedor de embeddings (Ollama) y el proveedor de summarización (Haiku/Anthropic, Ollama)
- **Implementaciones detectadas:** `internal/memory/ollama/` (Embedder + Summarizer), `internal/memory/haiku/` (Summarizer), `internal/memory/claude/` (implícito en CLI)
- **Cuándo agregar nueva impl:** Crear `internal/memory/<provider>/`, implementar las interfaces `Embedder`/`Summarizer`, registrar en los call sites en `internal/cli/run.go` y `internal/cli/digest_from_handoff.go`

### Strategy — AgentRunner
- **Interface:** `internal/orchestrator/types.go:109`
- **Qué varía:** El ejecutor concreto de nodos del DAG
- **Implementaciones detectadas:** `internal/runner/runner.go` (`ClaudeRunner`) — ejecuta `claude --print` como subprocess; `internal/deploy/*.go` (runners por proveedor: `claude.go`, `codex.go`, `gemini.go`, `opencode.go`, `cursor.go`)
- **Cuándo agregar nueva impl:** Implementar `AgentRunner.RunAgent(ctx, node, upstream)` en un nuevo archivo bajo `internal/runner/` o `internal/deploy/`

### Strategy — GateHandler
- **Interface:** `internal/orchestrator/gate.go:19`
- **Qué varía:** Cómo se aprueba/rechaza un gate humano en el DAG
- **Implementaciones detectadas:** `CLIGateHandler` (gate interactivo, `gate.go:32`), `AutoApproveHandler` (bypass para CI, `gate.go:91`)
- **Cuándo usar:** Inyectar la impl correcta según modo de ejecución (interactivo vs CI)

### Repository — Dashboard queries
- **Archivos:** `internal/dashboard/query/` (runs.go, tasks.go, files.go, agents.go, tools.go, prompts.go, metrics.go, dependencies.go, errors.go)
- **Firma detectada:** `func (r *Reader) Get*/List*(ctx context.Context, ...)` — todas las queries pasan `context.Context` y acceden a `*sql.DB`
- **Cuándo usar:** Toda consulta de dashboard debe ir por el `Reader` de `internal/dashboard/query/`
- **Anti-pattern:** No escribir queries SQL directamente en handlers MCP o CLI — ir siempre por el dominio de query

### Observer / Channel — Event Emitter
- **Archivo:** `internal/instrumentation/emitter.go:13`
- **Patrón:** `ch chan Event` con goroutine consumidora; `Start()` arranca el consumer, `Stop()` lo cierra
- **Qué varía:** El `EventWriter` que persiste los eventos (inyectado en el Emitter)
- **Interface:** `internal/instrumentation/events.go:46` (`EventWriter`), `events.go:51` (`Emitter`)

## De comportamiento

### Pipeline / DAG — Orchestrator
- **Archivos:** `internal/orchestrator/dag.go`, `internal/orchestrator/executor.go`
- **Qué varía:** La topología del DAG (definida en archivos YAML de pipeline) y el `AgentRunner` inyectado
- **Flujo:** DAG cargado → topological sort → executor lanza nodos en paralelo respetando dependencias → semáforo de concurrencia (`chan struct{}`) → resultados vía `chan workerResult`
- **Cuándo extender:** Agregar nueva `FailPolicy` o nuevo tipo de nodo → editar `internal/orchestrator/types.go` + `executor.go`

## Go-específicos

### Context propagation
- Todas las funciones de I/O, DB y llamadas a externos reciben `ctx context.Context` como primer argumento
- `ANVIL_PARENT_RUN_ID` y `ANVIL_AGENT_ID` propagados vía env vars a subprocesos (ver `internal/cli/emit_translate.go:53,60`)

### Table-driven tests
- Los tests de integración en `internal/dashboard/query/query_test.go` (830 líneas) y `internal/instrumentation/instrumentation_test.go` (937 líneas) usan tablas de casos
- `vecAutoOnce sync.Once` en archivos de test para evitar doble registro de sqlite-vec en mismo proceso

## Patrones a evitar en este proyecto

- **Global mutable state sin Once:** el patrón `sync.Once` existe precisamente para evitar race conditions en inicialización — no introducir `init()` con efectos secundarios globales
- **HTTP server propio:** no hay framework HTTP y no debe haber — el transporte es MCP stdio
- **Interfaces > 5 métodos:** las interfaces actuales son intencionalmente pequeñas (ISP) — no aglomerar métodos
