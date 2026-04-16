# Convention Routing — Guia para prompts de agentes

## El problema

Las convention skills son grandes. `go-conventions` tiene 5,907 lineas en 30+ archivos. Si un agente carga todo, quema ~24K tokens antes de escribir una linea de codigo — y el resultado no mejora.

## La regla

**El orchestrator selecciona los archivos de convenciones. El agente solo lee lo que recibe.**

Los agentes tienen prohibido navegar dispatchers (`SKILL.md`) o decidir que archivos cargar. Si necesitan convenciones que no recibieron, deben pedirlas.

## Como funciona

### Paso 1 — El orchestrator lee el dispatcher (una vez)

```
Read /Users/<user>/projects/anvil/skills/go-conventions/SKILL.md
```

El dispatcher es un indice ligero (~70-160 lineas) con una routing table que mapea tipos de tarea a archivos especificos.

### Paso 2 — Seleccionar archivos segun la tarea

Consultar la routing table y elegir 2-4 archivos relevantes.

### Paso 3 — Pasar al agente en el prompt

Dos formas segun el tamano de la tarea:

**Rutas absolutas** (Medium+):
```
Convention rules (read before implementing):
- /Users/<user>/projects/anvil/skills/go-conventions/rules/coding.md
- /Users/<user>/projects/anvil/skills/go-conventions/rules/database.md
```

**Inline** (Small):
```
Convention rules:
- Errors: wrap con fmt.Errorf("context: %w", err)
- SQL: table-driven scan con sql.Null* para nullable
- Naming: metodos de query empiezan con Get/List/Count
```

## Budget por tamano de tarea

| Tamano | Estrategia | Max archivos | Max lineas |
|--------|-----------|-------------|------------|
| Small (1-5 pts) | Reglas inline en prompt | 0 | ~10 reglas |
| Small (patron complejo) | Rutas absolutas | 1-2 | ~250 |
| Medium (5-8 pts) | Rutas absolutas | 2-4 | ~500 |
| Large (8-13 pts) | Rutas absolutas | 4-6 | ~800 |

## Referencia rapida — archivos por tipo de tarea

### Go (`skills/go-conventions/`)

| Tipo de tarea | Archivos |
|---------------|----------|
| Handler HTTP / servicio | `rules/coding.md` + `rules/architecture.md` |
| SQL / repositorio | `rules/coding.md` + `rules/database.md` |
| Refactor de errores | `rules/coding.md` + `examples/errors.md` |
| Concurrencia | `rules/coding.md` + `guides/concurrency/decision-matrix.md` |
| Middleware | `rules/coding.md` + `guides/middleware.md` |
| Tests (al tester) | `guides/testing/structure-tables.md` + `guides/testing/helpers-mocking.md` |
| Kafka / messaging | `rules/messaging.md` + `guides/kafka/<topico>.md` |

### React (`skills/react-conventions/`)

| Tipo de tarea | Archivos |
|---------------|----------|
| Componente nuevo | `SKILL.md` (seccion de reglas inline) |
| State management | `SKILL.md` + `state-management-guide.md` |
| Performance | `SKILL.md` + `performance-guide.md` |
| Accesibilidad | `SKILL.md` + `accessibility-guide.md` |
| Tests (al tester) | `testing-guide.md` |

### TypeScript (`skills/typescript-conventions/`)

Dispatcher con routing table. Leer `SKILL.md` y elegir por tarea.

### Flutter (`skills/flutter-conventions/`)

| Tipo de tarea | Archivos |
|---------------|----------|
| Widget / UI | `SKILL.md` (reglas inline) |
| State management | `SKILL.md` + `state-management-guide.md` |
| Arquitectura | `architecture-guide.md` |
| Tests (al tester) | `testing-guide.md` |

### Python (`skills/python-conventions/`)

Dispatcher con routing table. Leer `SKILL.md` y elegir por tarea.

### Rust (`skills/rust-conventions/`)

Dispatcher con routing table. Leer `SKILL.md` y elegir por tarea.

### Astro (`skills/astro-conventions/`)

Dispatcher con routing table. Leer `SKILL.md` y elegir por tarea.

## Ejemplo completo de prompt

```
You are the developer agent. Implement the following task.

Complexity: Small (3 pts)
Stack: Go
Mode: normal
Objective: Add a GetRunsByProject method to the query package that returns
all runs for a given project name, ordered by created_at DESC.

Files to change:
- internal/dashboard/query/runs.go (add method)

Convention rules (read before implementing):
- /Users/<user>/projects/anvil/skills/go-conventions/rules/coding.md
- /Users/<user>/projects/anvil/skills/go-conventions/rules/database.md

Context:
- Existing query methods follow the pattern in query/runs.go — scan rows
  into entity structs using table-driven helpers
- The entity is entity.Run (already exists)
- DB is SQLite, accessed via *sql.DB passed to the Query struct
```

## Ejemplo Small con reglas inline

```
You are the developer agent. Implement the following task.

Complexity: Small (2 pts)
Stack: Go
Mode: normal
Objective: Add ProjectName field to the Run entity.

Files to change:
- internal/dashboard/entity/run.go

Convention rules:
- Fields: exported, no pointers for required fields, sql.NullString for optional
- Tags: `json:"snake_case"` + `db:"snake_case"`
- No constructors for entities — direct struct literal
```

## Anti-patrones

| Que NO hacer | Por que |
|--------------|---------|
| "Carga go-conventions" | El agente cargara todo (5,900 lineas) o fallara |
| Pasar el SKILL.md dispatcher al agente | Navegara la routing table y leera de mas |
| No pasar convenciones en Medium+ | El agente preguntara y perdera un round-trip |
| Pasar 8+ archivos de convenciones | Rendimientos decrecientes — mas ruido que senal |
| Inyectar archivos completos en Small | Rutas absolutas son suficientes para 1-2 archivos |
