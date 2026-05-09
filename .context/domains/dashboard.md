# Dominio: dashboard

last_updated: 2026-05-08

## Responsabilidad

Capa de lectura (query) y entidades del dominio para el dashboard. Provee acceso tipado a todos los datos de runs, agents, tasks, files, tools, prompts, flows, métricas y errores almacenados en SQLite.

## Archivos clave

```
internal/dashboard/
├── entity/            — entidades de dominio puras (run.go, agent.go, task.go, file.go, tool.go, prompt.go, flow.go, event.go, error_group.go, file_edge.go)
├── query/             — Reader: queries sobre SQLite (runs.go, agents.go, tasks.go, files.go, tools.go, prompts.go, metrics.go, dependencies.go, errors.go, delete.go)
│   └── reader.go      — NewReader(db *sql.DB) — punto de entrada del dominio
├── dto/               — Data Transfer Objects para la capa de presentación (run.go, agent.go, task.go, file.go, tool.go, prompt.go, flow.go, metrics.go, session.go)
├── mapper/            — mapeos entity ↔ dto
├── imports/           — utilidades de importación de datos
└── testutil/          — helpers para tests de integración (vecAutoOnce, DB in-memory)
```

## Flujo principal

```
Query request → query.Reader.Get*/List*(ctx, filters)
→ SQL sobre SQLite (*sql.DB)
→ entity.* → mapper.* → dto.*
→ respuesta al caller (MCP tool, dashboard UI, CLI)
```

## Patrones usados

- **Repository:** `query/reader.go` — `Reader` con múltiples métodos CRUD de lectura; todos reciben `context.Context` y operan sobre `*sql.DB`
- **DTO / Mapper:** separación explícita entity → dto → presentación

## Interfaces públicas

```go
// internal/dashboard/query/reader.go
func NewReader(db *sql.DB) *Reader
// Métodos: GetRun, ListRuns, GetAgent, ListAgents, GetTask, ListTasks,
//          ListFiles, ListTools, ListPrompts, GetMetrics, GetDependencies, ...
```

## Dependencias de este dominio

- `pkg/storage` — `*sql.DB` inyectado vía `NewReader`

## Quién depende de este dominio

- `internal/mcp/` — usa `query.Reader` en todos los handlers de datos
- `internal/cli/` — usa `query.Reader` para mostrar estado, digests, etc.
- `cmd/anvil/` (build tag `dashboard`) — dashboard UI Wails

## Gotchas

- `query_test.go` (830 líneas) requiere sqlite-vec — usar `testutil.vecAutoOnce` para evitar doble registro en el mismo proceso de test
- El dominio es read-only por diseño — las escrituras van por los dominios de instrumentación y CLI, no por el query Reader

## Deuda técnica

- Sin dominio de escritura explícito — los INSERTs están dispersos en `internal/instrumentation/writer/` y `internal/cli/`; considerar un `repository` de escritura separado
