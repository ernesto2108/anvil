# Patrones de Diseño — anvil

last_updated: 2026-08-13

<!-- Este archivo se construye por inferencia estructural, no por nombres.
     Un patrón puede llamarse de cualquier manera o no tener nombre explícito.
     La firma del código es lo que importa. -->

## Creacionales

### Lazy Singleton (sync.Once) — inicialización de extensión SQLite
- **Archivo:** `pkg/storage/sqlite.go:15` (declaración `sqliteVecOnce`), uso en línea 43
- **Qué construye:** garantiza que `sqlite_vec.Auto()` se registre exactamente una vez por proceso, sin importar cuántas conexiones/goroutines abran la DB
- **Firma detectada:** `var sqliteVecOnce sync.Once` + `sqliteVecOnce.Do(sqlite_vec.Auto)`
- **Cuándo usar:** al registrar extensiones globales de SQLite (drivers, funciones custom) que no deben re-registrarse
- **Anti-pattern:** NO llamar `sqlite_vec.Auto()` directamente fuera de este `Once` — duplicaría el registro del driver

### Provider Fallback Chain — selección de motor de inferencia
- **Archivo:** `internal/cli/run.go:447-453` y `:545-552`
- **Qué construye:** selecciona el proveedor de LLM disponible en cascada: Claude API (si `ANTHROPIC_API_KEY` está seteada) → Ollama (si está saludable) → skip silencioso si ninguno está disponible
- **Firma detectada:** `if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" { ... }` seguido de fallback a Ollama
- **Cuándo usar:** al agregar un nuevo backend de inferencia — debe insertarse en la cadena de fallback, no reemplazar la existente
- **Anti-pattern:** NO asumir que `ANTHROPIC_API_KEY` siempre está presente — todo código que llama a Claude debe manejar el caso ausente

## Estructurales

### Registry Client (HTTP fetch + JSON decode) — `pkg/registry/registry.go`
- **Archivo:** `pkg/registry/registry.go:29` (`func Fetch(url string) (*Index, error)`)
- **Qué encapsula:** descarga y parseo de un índice remoto de agentes/skills publicados (`Entry`, `Index`)
- **Firma detectada:** cliente HTTP con timeout fijo (`&http.Client{Timeout: 15 * time.Second}`), retorna error envuelto con `fmt.Errorf("...: %w", err)`
- **Cuándo usar:** al agregar nuevas fuentes de datos remotas para el registry de agentes/skills

### MCP Server como fachada de subsistemas — `internal/mcp/`
- **Archivo:** `internal/mcp/server.go`, `tools.go`, `context.go`, `orchestration.go`, `execution.go`, `inventory.go`, `utilities.go`
- **Qué encapsula:** expone operaciones de memoria, orquestación e inventario del proyecto como tools MCP consumibles por Claude Code, delegando a `internal/memory` y `internal/orchestrator`
- **Firma detectada:** métodos `(s *Server) <tool>(ctx context.Context, args map[string]any) (string, error)` — ej. `digestFromHandoff` en `internal/mcp/context.go`
- **Cuándo usar:** al exponer una nueva capability del CLI como tool invocable desde Claude Code

## De comportamiento

### DAG + Gate + Executor — orquestación de pipelines multi-agente
- **Archivo:** `internal/orchestrator/dag.go`, `gate.go`, `executor.go`, `replanner.go`
- **Qué varía:** la estrategia de ejecución de un pipeline (orden de tareas vía DAG, puntos de aprobación vía Gate, reintento/replanificación vía Replanner) es intercambiable sin tocar el resto
- **Implementaciones detectadas:** `dag.go` (grafo de dependencias entre tasks), `gate.go` (puntos de decisión/aprobación), `executor.go` (ejecución del DAG resuelto), `replanner.go` (reajuste de plan ante fallos)
- **Cuándo agregar nueva impl:** al soportar un nuevo tipo de gate o estrategia de replanificación, extender las interfaces existentes en `types.go` en vez de bifurcar el executor

### Subcommand Dispatch — `internal/cli/cli.go`
- **Archivo:** `internal/cli/cli.go:16` (`func Run(args []string)`)
- **Qué varía:** el subcomando ejecutado (`emit`, `verify-direct`, `clean-empty-runs`, `capture`, `mcp-server`, y el resto vía `cfg`) — un `switch cmd` temprano atiende comandos que no requieren config de repo, y el resto cae al flujo con `config.Load`
- **Implementaciones detectadas:** un archivo por subcomando en `internal/cli/` (`cmd_migrate.go`, `deploy.go`, `doctor.go`, `stats.go`, etc.)
- **Cuándo agregar nueva impl:** agregar un nuevo archivo `cmd_<nombre>.go` con `cmd<Nombre>(cfg *config.App, args []string)` y registrar el case en `cli.go`; si el comando no necesita config de repo, agregarlo al switch temprano

## Go-específicos

### Functional Options
- **Archivo:** no se detectó el patrón `type Option func(*Config)` en el código de aplicación (`grep -rn "type Option func"` sin resultados fuera de `pkg/errors/error-options.go`, que usa una variante propia `Option func(o *option)`)
- **Tipo option:** `pkg/errors/error-options.go` — `type Option func(o *option)`, usado para construir errores `AnvilMeshCode` con metadata opcional
- **Cuándo usar:** al construir errores de dominio con campos opcionales (mensaje, causa, código HTTP)

### Table-driven tests
- **Archivo:** generalizado en el repo (`*_test.go` co-ubicados) — ej. `internal/orchestrator/dag_test.go`, `internal/mcp/orchestration_test.go`
- **Cuándo usar:** convención estándar para tests de Go en este repo — seguir el mismo estilo al agregar tests nuevos

## TypeScript-específicos

<!-- El frontend del dashboard (frontend/) existe pero no fue inspeccionado en este scan bootstrap — gap: pendiente de escaneo dedicado si se toca ese dominio -->

## Patrones a evitar en este proyecto

- **Microservicios / service mesh:** anvil es intencionalmente un monolito CLI (`cmd/anvil`) — no introducir separación de servicios ni `service-map.yaml` multi-servicio; el dashboard Wails (`anvil-full`) es un modo de build del mismo binario, no un servicio independiente.
