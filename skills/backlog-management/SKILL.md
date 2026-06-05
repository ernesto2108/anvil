---
name: backlog-management
description: Reglas de gestión del backlog local (sprint-current.md, board.md, dashboard.md, transiciones de estado, regla de los 3 lugares, jerarquía de trabajo, milestones, IDs y formato de tarea). Usada por cualquier agente que actualice el backlog.
---

# Gestión del Backlog

Esta skill cubre una única responsabilidad: **gestión del backlog local** — formato de `sprint-current.md`, `board.md`, `dashboard.md`, transiciones de estado, la regla de los 3 lugares, jerarquía de trabajo (PROJECT → MILESTONE → US → TASKS), milestones, formato de IDs y formato de tarea.

La escritura de tasks atómicas a partir del spec vive en la skill `task-writer`.

## Gestión del backlog

### Jerarquía de trabajo

```
PROJECT (el repo/producto — ej. Anvil, Dashboard)
  └── MILESTONE (milestone de entrega — ej. MVP, v1.0, v2.0)
       └── US (User Stories / Features — cada PRD es una US)
            └── TASKS (tareas técnicas — backend, frontend, DB, tests)
                 └── SUB TASKS (pasos dentro de una tarea — rastreados en handoff, no en backlog)
```

- **PROJECT** — implícito desde el repo. No se rastrea en el backlog.
- **MILESTONE** — grupos de features con un objetivo de entrega compartido. Rastreado en `## 9. Milestones y Timeline` del PRD y en el frontmatter de la tarea.
- **US** — cada PRD representa una User Story o Feature.
- **TASKS** — la descomposición de una US en ítems técnicos. Filas en `sprint-current.md`.
- **SUB TASKS** — pasos de implementación dentro de una tarea. En archivos `.handoff/`.

### Gestión de milestones

Los milestones se definen en `## 9. Milestones y Timeline` del PRD y se propagan a cada tarea.

**En sprint-current.md:** agrupar por milestone usando encabezados de sección:

```
| | **── 🎯 MVP ──** | | | | | |
| PROJ-FEAT-001 | Create auth flow | P0 | feat | developer | 5 | my-service |
| PROJ-FEAT-002 | Auth UI | P0 | feat | developer | 5 | my-web |
| | **── 🎯 v1.0 ──** | | | | | |
| PROJ-FEAT-010 | Add analytics | P1 | feat | developer | 3 | my-service |
```

**En el frontmatter de la tarea:** incluir `milestone: <name>` para que Dataview pueda agrupar/filtrar.

### Formato del ID de tarea

`<PROJECT>-<AREA>-<NNN>`

Áreas: FEAT, SEC, BUG, TECH, INFRA, DOC, TEST

Verificar los IDs existentes en `{backlog_path}` antes de asignar nuevos.

### Formato de tarea

**CRÍTICO:** siempre leer el `sprint-current.md` existente antes de agregar tareas. Coincidir con el formato que ya existe — nunca imponer un formato diferente.

El formato estándar usa **tablas**, no encabezados markdown:

**Fila de tabla en Backlog:**
```
| TASK-ID | Task description | P | Type | Agent | Pts | Repo |
```

**Fila de encabezado de sección:**
```
| | **── Feature Name (PARENT-ID, date) ──** | | | | | |
```

**Fila de tabla In Progress:**
```
| TASK-ID | Task | P | Agent | Start date | Branch |
```

**Fila de tabla Done:**
```
| TASK-ID | Task | Type | Date | Notes |
```

### Dónde vive el backlog (patrón universal)

El backlog vive **siempre como archivos locales** en `.project-context/` o en el repo. El humano es el orquestador: si usa una herramienta externa, la skill **describe qué crear/mover en ella** — nunca ejecuta acciones externas.

**Antes de crear archivos o tareas:**

1. Leer el campo `task_tool` de `.project-context/project.md`.
2. **Crear/actualizar el backlog local** (`sprint-current.md`, `board.md`, `dashboard.md` y los `task.md`) en `.project-context/` o el repo.
3. **Si `task_tool` tiene valor** (Linear, Jira, Notion, GitHub Issues) → al finalizar, **indicar al humano** qué tareas debe crear/mover, en texto libre. No llamar MCP ni APIs externas.
4. **Si `task_tool` está vacío, es `ninguna`, o el campo no existe** → preguntar al humano en texto libre ("¿Usas alguna herramienta de gestión de tareas para registrar este backlog, o lo mantengo solo en archivos locales?") y persistir la respuesta si la da.

### Archivos del backlog local

| Archivo | Propósito |
|---|---|
| `sprint-current.md` | Tabla del sprint con secciones |
| `board.md` | Tablero Kanban — ítems de checkbox con wiki-link y etiquetas |
| `dashboard.md` | Queries del dashboard (opcional; útil con Dataview) |

Cada tarea en `board.md` es un ítem de checkbox con wiki-link y etiquetas:

```
- [ ] [[TASK-ID/task]] Titulo de la tarea #proyecto #tag
```

`board.md` y `dashboard.md` son opcionales para proyectos ligeros — `sprint-current.md` y los `task.md` son el mínimo.

### Actualizar archivos companion

- Cuando las tareas cambian de estado → actualizar `sprint-current.md` Y `board.md` (mover el checkbox a la columna correcta).
- Cuando se agregan tareas → agregar a la tabla de `sprint-current.md` Y a la columna Backlog de `board.md`.
- Actualizar el campo `status` del frontmatter de la tarea para que coincida.
- `dashboard.md` se actualiza solo via queries.

### Transiciones de estado — la regla de los 3 lugares (CRÍTICO)

Cada vez que una tarea cambia de estado, se DEBEN actualizar **exactamente 3 archivos** locales. Olvidar cualquiera causa deriva.

**Lista de verificación para CADA transición:**

1. **`sprint-current.md`** — mover la fila a la sección correcta (Backlog / TODO / In Progress / Blocked / In Review / Done). Las filas Done incluyen: `| ID | Title | Type | YYYY-MM-DD | Notes |`.
2. **`board.md`** — mover la línea de checkbox `- [ ]` / `- [x]` a la columna Kanban correcta. Al mover a Done, cambiar `- [ ]` a `- [x]`.
3. **Frontmatter del `task.md`** — actualizar el campo `status`. Si se mueve a Done, TAMBIÉN agregar `completed: YYYY-MM-DD`. **Este es el que se olvida.**

Si `task_tool` tiene valor, además **describir al humano** la transición equivalente a aplicar en su herramienta — sin ejecutarla.

#### Mapeo de estado → valor en frontmatter

| Columna Kanban | Frontmatter `status` | Campos extra |
|---|---|---|
| Backlog | `backlog` | — |
| TODO | `todo` | — |
| In Progress | `in-progress` | `started: YYYY-MM-DD` |
| Blocked | `blocked` | `blocked_by: <TASK-ID>` |
| In Review | `in-review` | `pr: <URL>` |
| Done | `done` | `completed: YYYY-MM-DD` |

#### Error común — Done con frontmatter desactualizado

**Síntoma:** el usuario dice "veo la tarea X todavía en el backlog" aunque `sprint-current.md` y `board.md` la muestren en Done.

**Causa:** el `task.md` todavía tiene `status: backlog`. La query del `dashboard.md` lee los frontmatters, no las columnas Kanban.

**Corrección:** buscar statuses desactualizados antes de cerrar un sprint:

```bash
grep -rl "status: backlog" <ruta-de-tasks>/
```

**Prevención:** al mover a Done, hacer las 3 ediciones en el mismo lote de tool calls.

### Formato del tablero de sprint

`{backlog_path}` (archivo local `sprint-current.md`):

```markdown
# Sprint Backlog

> Sprint #N | YYYY-MM-DD → ongoing | Goal: <sprint goal>

## Backlog
| ID | Tarea | P | Tipo | Agente | Pts | Repo |
|----|-------|---|------|--------|-----|------|
| | **── Feature Name (TASK-ID, date) ──** | | | | | |
| PROJ-FEAT-001 | Create password reset endpoint | P1 | feat | developer-backend | 5 | my-service |
| PROJ-FEAT-002 | Password reset UI | P1 | feat | developer-frontend | 5 | my-web |
| PROJ-TEST-001 | Tests for reset endpoint | P1 | test | tester | 3 | my-service |

## TODO
| ID | Tarea | P | Tipo | Agente | Pts | Repo |
|----|-------|---|------|--------|-----|------|

## In Progress
| ID | Tarea | P | Agente | Inicio | Branch |
|----|-------|---|--------|--------|--------|

## Blocked
| ID | Tarea | P | Agente | Bloqueado por |
|----|-------|---|--------|---------------|

## In Review
| ID | Tarea | Agente | Reviewer | PR |
|----|-------|--------|----------|-----|

## Done
| ID | Tarea | Tipo | Fecha | Notas |
|----|-------|------|-------|-------|
```

### Ciclo de vida de la tarea

```
PM crea PRD
  → Architect crea Architecture Views + ADRs (propaga milestone del PRD)
  → spec-writer produce spec.md
  → task-writer genera los archivos de task (esta skill)
  → Tareas van a la columna Backlog
  → Orquestador toma tarea, asigna a agente
  → Agente comienza → tarea se mueve a In Progress
  → Agente termina → tarea se mueve a Done con fecha
  → Todas las tareas completadas → feature está completo
```

### Reglas generales

- **Sin trabajo sin ticket** — si un agente necesita hacer algo, debe haber una tarea.
- **Sin ticket sin spec/PRD** — cada tarea referencia su fuente padre (excepto bugs con repro).
- **Dependencias explícitas** — `Depends on: <TASK-ID>`.
- **Criterios de aceptación vienen de los escenarios GCE** embebidos en los RFs del `requirements.md` (o del `## 3. Criterios de aceptación` del spec).
- **Puntos Fibonacci** — 1, 2, 3, 5, 8, 13. Si > 8, descomponer.
- **Actualizaciones de estado obligatorias** — agentes deben actualizar al comenzar y al terminar.
