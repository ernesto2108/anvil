---
name: task-complete
description: Marcar una tarea como completada — actualiza el estado del archivo de tarea, mueve la tarjeta en el tablero Kanban, limpia duplicados. Úsalo cuando el usuario diga "tarea lista", "completar tarea", "marcar como hecho", o "/task-complete TASK-XXX".
disable-model-invocation: true
---

# Task Complete

Automatiza los pasos necesarios para cerrar una tarea. El comportamiento se adapta al sistema de docs del proyecto (Obsidian vault, Linear+Outline, o `.workspace/`). Resolver `<docs>` desde `~/.claude/project-registry.md` para detectar el sistema.

## Uso

```
/task-complete TASK-006
/task-complete TASK-006 "Documentación completada + 2 bugs encontrados"
```

## Flujo de Trabajo

Cuando se invoca con un TASK-ID (y resumen opcional):

### Paso 1 — Encontrar el archivo de tarea

Buscar el archivo de tarea según el sistema de docs:
- **Obsidian vault:** `<docs>/03-tasks/<TASK-ID>/task.md`
- **`.workspace/`:** `.workspace/tasks/<TASK-ID>/task.md`
- **Linear+Outline:** buscar el issue en Linear por TASK-ID

### Paso 2 — Actualizar el estado de la tarea

En el frontmatter del archivo de tarea, establecer:
```yaml
status: done
```

### Paso 3 — Actualizar el tablero (si aplica)

**Obsidian vault:**
En `{board_path}` (solo Obsidian vault):
1. **Eliminar** la línea de la tarea de cualquier columna en la que se encuentre
2. **Eliminar** cualquier duplicado en Backlog
3. **Agregar** a la columna Done:
   ```
   - [x] [[<sprint>/<TASK-ID>]] <resumen o título> #<service> #<labels>
   ```

**Linear+Outline:** mover el issue a Done en Linear. No hay archivos locales.

**`.workspace/`:** no hay board.md — omitir este paso.

### Paso 4 — Actualizar métricas del sprint (si aplica)

**Obsidian vault:** En `<docs>/02-backlog/sprint-current.md`, incrementar los SP completados.
**`.workspace/`:** En `.workspace/sprint-current.md`, incrementar los SP completados.

**Linear+Outline:** las métricas viven en Linear — omitir.

### Paso 5 — Archivar la nota de handoff

Seguir la operación de Archivo de la skill `/handoff`: leer, agregar resumen al archivo de tarea, eliminar el archivo de handoff. Omitir si no existe handoff.

### Paso 6 — Confirmar

Emitir una confirmación de una línea:
```
✓ <TASK-ID> marcada como done (<story_points> SP)
```

## Reglas

- Resolver `<docs>` y el tipo de sistema de docs desde `~/.claude/project-registry.md`
- Si no se encuentra el archivo de tarea → reportar error, no crearlo
- Si la tarea ya tiene `status: done` → omitir, reportar "already done"
- NO modificar ningún archivo que no sea: el .md de la tarea, board.md, sprint-current.md
- Máximo 6 llamadas a herramientas: 1 Read (archivo de tarea) + 1 Edit (estado de tarea) + 1 Edit (board) + 1 Edit (métricas de sprint) + 1 Read (handoff) + 1 Delete (handoff)
