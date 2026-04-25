---
name: task-complete
description: Marcar una tarea como completada — actualiza el estado del archivo de tarea, mueve la tarjeta en el tablero Kanban, limpia duplicados. Úsalo cuando el usuario diga "tarea lista", "completar tarea", "marcar como hecho", o "/task-complete TASK-XXX".
disable-model-invocation: true
---

# Task Complete

Automatiza los 5 pasos necesarios para cerrar una tarea en el vault de Obsidian.

## Uso

```
/task-complete TASK-006
/task-complete TASK-006 "Documentación completada + 2 bugs encontrados"
```

## Flujo de Trabajo

Cuando se invoca con un TASK-ID (y resumen opcional):

### Paso 1 — Encontrar el archivo de tarea

Buscar `<TASK-ID>.md` en `<docs>/03-tasks/` (verificar primero las carpetas de sprint, luego el backlog).

### Paso 2 — Actualizar el estado de la tarea

En el frontmatter del archivo de tarea, establecer:
```yaml
status: done
```

### Paso 3 — Actualizar el tablero Kanban

En `<docs>/02-backlog/board.md`:
1. **Eliminar** la línea de la tarea de cualquier columna en la que se encuentre (Backlog, To Do, In Progress, Blocked)
2. **Eliminar** cualquier duplicado en Backlog si la tarea también existe en una carpeta de sprint
3. **Agregar** a la columna Done:
   ```
   - [x] [[<sprint>/<TASK-ID>]] <resumen o título> #<service> #<labels>
   ```

### Paso 4 — Actualizar métricas del sprint (si aplica)

En `<docs>/02-backlog/sprint-current.md`, incrementar los SP completados por los `story_points` de la tarea.

### Paso 5 — Archivar la nota de handoff

Seguir la operación de Archivo de la skill `/handoff`: leer, agregar resumen al archivo de tarea, eliminar el archivo de handoff. Omitir si no existe handoff.

### Paso 6 — Confirmar

Emitir una confirmación de una línea:
```
✓ <TASK-ID> marcada como done (<story_points> SP)
```

## Reglas

- Resolver `<docs>` desde `~/.claude/project-registry.md`
- Si no se encuentra el archivo de tarea → reportar error, no crearlo
- Si la tarea ya tiene `status: done` → omitir, reportar "already done"
- NO modificar ningún archivo que no sea: el .md de la tarea, board.md, sprint-current.md
- Máximo 6 llamadas a herramientas: 1 Read (archivo de tarea) + 1 Edit (estado de tarea) + 1 Edit (board) + 1 Edit (métricas de sprint) + 1 Read (handoff) + 1 Delete (handoff)
