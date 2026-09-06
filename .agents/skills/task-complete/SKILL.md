---
name: task-complete
description: Marcar una tarea como completada — actualiza el estado del archivo de tarea, mueve la tarjeta en el tablero Kanban, limpia duplicados. Úsalo cuando el usuario diga "tarea lista", "completar tarea", "marcar como hecho", o "/task-complete TASK-XXX".
disable-model-invocation: true
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->


# Task Complete

Automatiza los pasos necesarios para cerrar una tarea. El estado de la tarea vive siempre en `.project-context/` o en el repo. Si el proyecto tiene una herramienta de gestión externa (campo `task_tool` de `.project-context/project.md`), esta skill **describe al humano** qué marcar como done en esa herramienta — **nunca ejecuta acciones en herramientas externas**.

## Uso

```
/task-complete TASK-006
/task-complete TASK-006 "Documentación completada + 2 bugs encontrados"
```

## Flujo de Trabajo

Cuando se invoca con un TASK-ID (y resumen opcional):

### Paso 0 — Leer `task_tool`

Leer el campo `task_tool` del frontmatter de `.project-context/project.md`.
- Si tiene valor (ej. `Linear`, `Jira`, `Notion`) → al final describirás al humano qué hacer en esa herramienta.
- Si está vacío o es `ninguna` → todo el cierre se hace en archivos locales de `.project-context/` o del repo.

### Paso 1 — Encontrar el archivo de tarea

Buscar el archivo de tarea local: `.project-context/tasks/<TASK-ID>/task.md` (o la ubicación que el proyecto use en el repo). Si no existe localmente y `task_tool` tiene valor, la tarea vive en la herramienta externa — saltar a Paso 5.

### Paso 2 — Actualizar el estado de la tarea

En el frontmatter del archivo de tarea local, establecer:
```yaml
status: done
```

### Paso 3 — Actualizar el tablero local (si existe)

Si existe un tablero local (`board.md` / `sprint-current.md` en `.project-context/` o el repo):
1. **Eliminar** la línea de la tarea de cualquier columna en la que se encuentre
2. **Eliminar** cualquier duplicado en Backlog
3. **Agregar** a la columna Done:
   ```
   - [x] [[<sprint>/<TASK-ID>]] <resumen o título> #<service> #<labels>
   ```

Si no hay tablero local — omitir este paso.

### Paso 4 — Actualizar métricas del sprint local (si existen)

Si existe `sprint-current.md` local, incrementar los SP completados. Si no existe — omitir.

### Paso 5 — Archivar la nota de handoff

Seguir la operación de Archivo de la skill `/handoff`: leer, agregar resumen al archivo de tarea, eliminar el archivo de handoff. Omitir si no existe handoff.

### Paso 6 — Describir acción al humano (si hay `task_tool`)

Si `task_tool` tiene valor, **no ejecutar nada** en la herramienta. Emitir una instrucción al humano:
```
Marca <TASK-ID> como done en {task_tool}.
```

### Paso 7 — Confirmar

Emitir una confirmación de una línea:
```
✓ <TASK-ID> marcada como done (<story_points> SP)
```

## Reglas

- Leer `task_tool` desde `.project-context/project.md` — nunca ejecutar acciones en herramientas externas, solo describirlas al humano
- Si no se encuentra el archivo de tarea local y no hay `task_tool` → reportar error, no crearlo
- Si la tarea ya tiene `status: done` → omitir, reportar "already done"
- NO modificar ningún archivo que no sea: el .md de la tarea, board.md, sprint-current.md (todos locales)
- Máximo 6 llamadas a herramientas: 1 Read (project.md/archivo de tarea) + 1 Edit (estado de tarea) + 1 Edit (board) + 1 Edit (métricas de sprint) + 1 Read (handoff) + 1 Delete (handoff)
