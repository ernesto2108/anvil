---
name: backlog-management
description: Creación de tareas, gestión del backlog y formato del tablero de sprint. Define cómo descomponer PRDs en tickets, asignar agentes y hacer seguimiento del progreso. Usado por el agente PM después de escribir un PRD.
---

# Gestión del Backlog

## Cuándo usar

Después de escribir un PRD, el PM DEBE descomponerlo en tareas antes de que cualquier agente comience a trabajar. Sin PRD no hay tareas. Sin tareas no hay referencia a un PRD.

## Jerarquía de trabajo

```
PROJECT (el repo/producto — ej. Anvil, Dashboard)
  └── MILESTONE (milestone de entrega — ej. MVP, v1.0, v2.0)
       └── US (User Stories / Features — cada PRD es una US)
            └── TASKS (tareas técnicas — backend, frontend, DB, tests)
                 └── SUB TASKS (pasos dentro de una tarea — rastreados en handoff, no en backlog)
```

- **PROJECT** — implícito desde el repo. No se rastrea en el backlog
- **MILESTONE** — grupos de features relacionadas con un objetivo de entrega compartido. Rastreado como campo en el PRD Scope y en el frontmatter de la tarea
- **US** — cada PRD representa una User Story o Feature. El PRD es la US
- **TASKS** — la descomposición de una US en ítems de trabajo técnico. Estas son las filas en sprint-current.md
- **SUB TASKS** — pasos de implementación dentro de una tarea. Rastreados en archivos `.handoff/`, no en el backlog

### Gestión de milestones

Los milestones se definen en la sección `## Scope` del PRD (campo `Milestone`) y se propagan a cada tarea.

**En sprint-current.md:** agrupar tareas por milestone usando encabezados de sección:
```
| | **── 🎯 MVP ──** | | | | | |
| PROJ-FEAT-001 | Create auth flow | P0 | feat | developer | 5 | my-service |
| PROJ-FEAT-002 | Auth UI | P0 | feat | developer | 5 | my-web |
| | **── 🎯 v1.0 ──** | | | | | |
| PROJ-FEAT-010 | Add analytics | P1 | feat | developer | 3 | my-service |
```

**En el frontmatter de la tarea:** incluir `milestone: <name>` para que Dataview pueda agrupar/filtrar por milestone.

**Query del dashboard:** `dashboard.md` puede usar `GROUP BY milestone` para mostrar el progreso por milestone.

## Formato del ID de tarea

`<PROJECT>-<AREA>-<NNN>`

Áreas: FEAT, SEC, BUG, TECH, INFRA, DOC, TEST

Verificar los IDs existentes en `<docs>/02-backlog/sprint-current.md` antes de asignar nuevos.

## Descomponer un PRD en tareas

Leer los requisitos funcionales y criterios de aceptación del PRD. Crear una tarea por:

1. **Cada requisito P0** → al menos una tarea
2. **Cada componente que requiere trabajo separado** (backend, frontend, DB, infra)
3. **Tests** → tarea separada por componente (el desarrollador escribe el código, el tester escribe los tests)
4. **Migraciones** → tarea separada si se necesitan cambios en DB
5. **Documentación** → tarea separada si se necesitan docs orientadas al usuario

### Reglas de descomposición

- Una preocupación por tarea — si una tarea toca backend Y frontend, separarla
- Las tareas deben ser completables en 1-8 puntos (si > 8, descomponer más)
- Cada tarea debe referenciar su PRD: `PRD: <TASK-ID>`
- Cada tarea debe tener un tipo de agente asignado
- Los tests son SIEMPRE una tarea separada de la implementación

### Ejemplo de descomposición

PRD: `PROJ-FEAT-042` — Add password reset flow

| Task ID | Título | Agente | Puntos | Depende de |
|---|---|---|---|---|
| PROJ-FEAT-042-01 | Create password reset endpoint | developer | 5 | — |
| PROJ-FEAT-042-02 | Add email sending service | developer | 3 | — |
| PROJ-FEAT-042-03 | Create password reset UI | developer | 5 | 01 |
| PROJ-FEAT-042-04 | Add migration for reset tokens table | dba | 2 | — |
| PROJ-FEAT-042-05 | Tests for reset endpoint | tester | 3 | 01 |
| PROJ-FEAT-042-06 | Tests for email service | tester | 2 | 02 |
| PROJ-FEAT-042-07 | Tests for reset UI | tester | 3 | 03 |
| PROJ-FEAT-042-08 | Security review | security | 2 | 01, 02 |

## Formato de tarea

**CRÍTICO:** Siempre leer el `sprint-current.md` existente antes de agregar tareas. Coincidir con el formato que ya existe — nunca imponer un formato diferente.

El formato estándar usa **tablas**, no encabezados markdown:

### Fila de tabla en Backlog
```
| TASK-ID | Task description | P | Type | Agent | Pts | Repo |
```

### Fila de encabezado de sección (para agrupar tareas relacionadas)
```
| | **── Feature Name (PARENT-ID, date) ──** | | | | | |
```

### Fila de tabla In Progress
```
| TASK-ID | Task | P | Agent | Start date | Branch |
```

### Fila de tabla Done
```
| TASK-ID | Task | Type | Date | Notes |
```

## Integración con Obsidian (OBLIGATORIO)

El backlog vive en un vault de Obsidian. Cada tarea y archivo de sprint debe ser compatible con los plugins de Obsidian: **Dataview** (queries) y **Kanban** (tablero visual).

**Todos los templates viven en `vault-template/` en la raíz del repo Anvil.** Leerlos directamente — nunca hardcodear templates inline en skills o agentes.

### Frontmatter de archivo de tarea (OBLIGATORIO para cada task.md)

Cada `<docs>/03-tasks/<TASK-ID>/task.md` DEBE incluir frontmatter para Dataview. Leer `vault-template/03-tasks/task-template.md` para el formato y campos exactos.

**Sin este frontmatter, las queries de Dataview y el tablero Kanban no funcionarán.** Esto no es opcional.

### Archivos companion del sprint (crear junto con sprint-current.md)

Al crear un nuevo sprint, el PM DEBE crear 3 archivos basados en los templates en `vault-template/02-backlog/`:

| Archivo | Fuente del template | Propósito |
|---|---|---|
| `<docs>/02-backlog/sprint-current.md` | `vault-template/02-backlog/sprint-current.md` | Tabla del sprint con secciones |
| `<docs>/02-backlog/board.md` | `vault-template/02-backlog/board.md` | Tablero Kanban (plugin Obsidian) |
| `<docs>/02-backlog/dashboard.md` | `vault-template/02-backlog/dashboard.md` | Queries del dashboard Dataview |

**Los tres archivos deben existir juntos.** Nunca crear uno sin los otros.

Cada tarea en board.md es un ítem de checkbox con un wiki-link y etiquetas relevantes:
```
- [ ] [[TASK-ID/task]] Titulo de la tarea #proyecto #tag
```

### Actualizar archivos companion

- Cuando las tareas cambian de estado → actualizar tanto `sprint-current.md` COMO `board.md` (mover el checkbox a la columna correcta)
- Cuando se agregan tareas → agregar a la tabla de `sprint-current.md` Y a la columna Backlog de `board.md`
- Actualizar el campo `status` del frontmatter de la tarea para que coincida
- El `dashboard.md` se actualiza solo via queries Dataview — no se necesitan actualizaciones manuales

### Transiciones de estado — la regla de los 3 lugares (CRÍTICO)

Cada vez que una tarea cambia de estado, se DEBEN actualizar **exactamente 3 archivos**. Olvidar cualquiera de ellos causa deriva: `sprint-current.md` y `board.md` son vistas manuales, pero `dashboard.md` se basa en queries Dataview que leen los frontmatters de las tareas — si solo se actualizan los primeros dos, el dashboard muestra datos desactualizados y la próxima sesión del orquestador pensará que la tarea aún está pendiente.

**Lista de verificación para CADA transición de estado:**

1. **`<docs>/02-backlog/sprint-current.md`** — mover la fila a la sección correcta (Backlog / TODO / In Progress / Blocked / In Review / Done). Las filas Done incluyen: `| ID | Title | Type | YYYY-MM-DD | Notes |`.

2. **`<docs>/02-backlog/board.md`** — mover la línea de checkbox `- [ ]` / `- [x]` a la columna Kanban correcta. Al mover a Done, cambiar `- [ ]` a `- [x]`.

3. **Frontmatter de `<docs>/03-tasks/<TASK-ID>/task.md`** — actualizar el campo `status`. Si se mueve a Done, TAMBIÉN agregar el campo `completed: YYYY-MM-DD`. **Este es el que se olvida.**

### Mapeo de estado → valor en frontmatter

| Columna Kanban | Frontmatter `status` | Campos extra |
|---|---|---|
| Backlog | `backlog` | — |
| TODO | `todo` | — |
| In Progress | `in-progress` | `started: YYYY-MM-DD` |
| Blocked | `blocked` | `blocked_by: <TASK-ID>` |
| In Review | `in-review` | `pr: <URL>` |
| Done | `done` | `completed: YYYY-MM-DD` |

### Error común — Done con frontmatter desactualizado

**Síntoma:** el usuario dice "veo la tarea X todavía en el backlog" aunque `sprint-current.md` y `board.md` la muestren en Done.

**Causa:** `03-tasks/<ID>/task.md` todavía tiene `status: backlog`. La query Dataview en `dashboard.md` lee los frontmatters, no las columnas Kanban, por lo que muestra la tarea como pendiente.

**Corrección:** buscar statuses desactualizados antes de cerrar un sprint:
```bash
# Cualquier archivo de tarea cuyo status no coincida con la sección de sprint-current en la que está
grep -r "status: backlog" <docs>/03-tasks/ | grep -v "$(awk '/## Backlog/,/## TODO/' <docs>/02-backlog/sprint-current.md)"
```

**Prevención:** al mover a Done, hacer las 3 ediciones de archivo en el mismo lote de llamadas a herramientas. Nunca dividirlas entre mensajes — ahí es donde se olvida el tercer archivo.

## Formato del tablero de sprint

`<docs>/02-backlog/sprint-current.md`:

```markdown
# Sprint Backlog

> Sprint #N | YYYY-MM-DD → ongoing | Goal: <sprint goal>

## Backlog
| ID | Tarea | P | Tipo | Agente | Pts | Repo |
|----|-------|---|------|--------|-----|------|
| | **── Feature Name (TASK-ID, date) ──** | | | | | |
| PROJ-FEAT-001 | Create password reset endpoint | P1 | feat | developer | 5 | my-service |
| PROJ-FEAT-002 | Password reset UI | P1 | feat | developer | 5 | my-web |
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

## Ciclo de vida de la tarea

```
PM crea PRD
  → PM descompone en tareas (este skill)
  → Tareas van a la columna Backlog
  → Orquestador toma tarea, asigna a agente
  → Agente comienza → tarea se mueve a In Progress
  → Agente termina → tarea se mueve a Done con fecha
  → Todas las tareas completadas → PRD está completo
```

## Reglas

- **Sin trabajo sin ticket** — si un agente necesita hacer algo, debe haber una tarea para ello
- **Sin ticket sin PRD** — cada tarea referencia su PRD padre (excepto bugs con pasos de reproducción)
- **Las dependencias deben ser explícitas** — si la tarea B necesita que la tarea A esté terminada primero, escribir "Depends on: A"
- **Los criterios de aceptación vienen del PRD** — los criterios de cada tarea provienen de los escenarios Given/When/Then del PRD
- **Los puntos son fibonacci** — 1, 2, 3, 5, 8, 13. Si > 8, descomponer
- **Las actualizaciones de estado son obligatorias** — los agentes deben actualizar el estado de la tarea al comenzar y al terminar
