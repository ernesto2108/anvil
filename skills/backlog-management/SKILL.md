---
name: backlog-management
description: Creación de tareas, gestión del backlog y formato del tablero de sprint. Define cómo descomponer PRDs en tickets, asignar agentes y hacer seguimiento del progreso. Usado por el agente `task-decomposer` para crear tasks atómicas y actualizar el backlog.
---

# Gestión del Backlog

## Cuándo usar

Después de escribir el ARD, el `task-decomposer` DEBE descomponer en tareas antes de que cualquier agente comience a trabajar. Sin PRD no hay ARD. Sin ARD no hay tareas.

## Jerarquía de trabajo

```
PROJECT (el repo/producto — ej. Anvil, Dashboard)
  └── MILESTONE (milestone de entrega — ej. MVP, v1.0, v2.0)
       └── US (User Stories / Features — cada PRD es una US)
            └── TASKS (tareas técnicas — backend, frontend, DB, tests)
                 └── SUB TASKS (pasos dentro de una tarea — rastreados en handoff, no en backlog)
```

- **PROJECT** — implícito desde el repo. No se rastrea en el backlog
- **MILESTONE** — grupos de features relacionadas con un objetivo de entrega compartido. Rastreado como campo en la sección `## 9. Milestones y Timeline` del PRD y en el frontmatter de la tarea
- **US** — cada PRD representa una User Story o Feature. El PRD es la US
- **TASKS** — la descomposición de una US en ítems de trabajo técnico. Estas son las filas en sprint-current.md
- **SUB TASKS** — pasos de implementación dentro de una tarea. Rastreados en archivos `.handoff/`, no en el backlog

### Gestión de milestones

Los milestones se definen en la sección `## 9. Milestones y Timeline` del PRD y se propagan a cada tarea.

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

Verificar los IDs existentes en `{backlog_path}` (archivo local en `.project-context/` o el repo) antes de asignar nuevos.

## Descomponer un PRD en tareas

Leer los requisitos funcionales y los escenarios GCE (Dado/Cuando/Entonces) embebidos en los RFs del `requirements.md`. Crear una tarea por:

1. **Cada requisito P0** → al menos una tarea
2. **Cada componente que requiere trabajo separado** (backend, frontend, DB, infra)
3. **Tests** → tarea separada por componente (el desarrollador escribe el código, el tester escribe los tests)
4. **Migraciones** → tarea separada si se necesitan cambios en DB
5. **Documentación** → tarea separada si se necesitan docs orientadas al usuario

### Reglas de descomposición

- Una preocupación por tarea — si una tarea toca backend Y frontend/mobile, separarla
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

## Dónde vive el backlog (patrón universal)

El backlog vive **siempre como archivos locales** en `.project-context/` o en el repo. El humano es el orquestador: si usa una herramienta de gestión externa, la skill **describe qué crear/mover en ella** — nunca ejecuta acciones en herramientas externas.

**Antes de crear archivos o tareas:**

1. Leer el campo `task_tool` de `.project-context/project.md`.
2. **Crear/actualizar el backlog local** (`sprint-current.md`, `board.md`, `dashboard.md` y los `task.md` por tarea) en `.project-context/` o el repo según corresponda.
3. **Si `task_tool` tiene valor** (ej. Linear, Jira, Notion, GitHub Issues) → al finalizar, **indicar al humano** qué tareas debe crear/mover en esa herramienta, en texto libre. No llamar MCP ni APIs externas.
4. **Si `task_tool` está vacío, es `ninguna`, o el campo no existe** → preguntar al humano en texto libre ("¿Usas alguna herramienta de gestión de tareas para registrar este backlog, o lo mantengo solo en archivos locales?") y persistir la respuesta en `task_tool` si la da. No ofrecer un enum cerrado de opciones.

### Frontmatter y archivos del backlog local

Cada `task.md` usa frontmatter YAML simple. Los archivos del sprint viven juntos en `.project-context/` o el repo:

| Archivo | Propósito |
|---|---|
| `sprint-current.md` | Tabla del sprint con secciones (formato abajo) |
| `board.md` | Tablero Kanban — ítems de checkbox con wiki-link y etiquetas |
| `dashboard.md` | Queries del dashboard (opcional; útil si el proyecto usa un visor tipo Dataview) |

Cada tarea en board.md es un ítem de checkbox con un wiki-link y etiquetas relevantes:
```
- [ ] [[TASK-ID/task]] Titulo de la tarea #proyecto #tag
```

`board.md` y `dashboard.md` son opcionales para proyectos ligeros — `sprint-current.md` y los `task.md` son el mínimo. Si el proyecto usa un visor tipo Dataview/Kanban, leer las plantillas de referencia que el proyecto tenga y mantener su frontmatter.

#### Actualizar archivos companion

- Cuando las tareas cambian de estado → actualizar tanto `sprint-current.md` COMO `board.md` (mover el checkbox a la columna correcta)
- Cuando se agregan tareas → agregar a la tabla de `sprint-current.md` Y a la columna Backlog de `board.md`
- Actualizar el campo `status` del frontmatter de la tarea para que coincida
- El `dashboard.md` se actualiza solo via queries — no se necesitan actualizaciones manuales

#### Transiciones de estado — la regla de los 3 lugares (CRÍTICO)

Cada vez que una tarea cambia de estado, se DEBEN actualizar **exactamente 3 archivos** locales. Olvidar cualquiera de ellos causa deriva.

**Lista de verificación para CADA transición de estado:**

1. **`sprint-current.md`** — mover la fila a la sección correcta (Backlog / TODO / In Progress / Blocked / In Review / Done). Las filas Done incluyen: `| ID | Title | Type | YYYY-MM-DD | Notes |`.

2. **`board.md`** — mover la línea de checkbox `- [ ]` / `- [x]` a la columna Kanban correcta. Al mover a Done, cambiar `- [ ]` a `- [x]`.

3. **Frontmatter del `task.md`** — actualizar el campo `status`. Si se mueve a Done, TAMBIÉN agregar el campo `completed: YYYY-MM-DD`. **Este es el que se olvida.**

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

**Prevención:** al mover a Done, hacer las 3 ediciones de archivo en el mismo lote de llamadas a herramientas.

## Formato del tablero de sprint

`{backlog_path}` (archivo local `sprint-current.md` en `.project-context/` o el repo):

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
  → Arquitecto crea ARD (propaga milestone del PRD)
  → task-decomposer descompone en tareas (este skill)
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
- **Los criterios de aceptación vienen de los escenarios GCE (Dado/Cuando/Entonces) embebidos en los RFs del `requirements.md`** — los criterios de cada tarea provienen de esos escenarios
- **Los puntos son fibonacci** — 1, 2, 3, 5, 8, 13. Si > 8, descomponer
- **Las actualizaciones de estado son obligatorias** — los agentes deben actualizar el estado de la tarea al comenzar y al terminar
