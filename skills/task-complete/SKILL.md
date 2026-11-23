---
name: task-complete
description: Cierra una tarea local o externa después de verificar reporter, validaciones, PR y sincronización. Úsalo cuando el usuario diga "tarea lista", "completar tarea", "marcar como hecho", o "/task-complete TASK-XXX".
disable-model-invocation: true
---

# Task Complete

El cierre de una tarea es una verificación de entrega, no un cambio aislado de estado. `delivery-flow` es la fuente canónica para Linear, documentación y PR.

## Flujo de trabajo

1. Cargar `delivery-flow` y localizar el `delivery-state.yaml` por `TASK-ID` o run activo.
2. Verificar que `reporter_status` esté `complete` (o `skipped` con excepción registrada), que todas las validaciones requeridas pasaron y que no haya `next_step` bloqueante.
3. Si el proyecto usa tracking externo, verificar que existe `pr_url`, que el comentario del PR está enlazado una sola vez y que el estado de Linear es **In Review** o **Done** según el merge/política configurada. Ejecutar las acciones pendientes idempotentes mediante `delivery-flow`; no pedirle al humano que las haga manualmente.
4. Si hay una tarea local, actualizar su estado/tablero solo después de los gates externos. Archivar el handoff según `handoff` sin borrar el estado de entrega.
5. Actualizar `status: delivered` únicamente cuando todos los gates estén satisfechos; si uno falta, dejar `blocked` y reportar el siguiente paso.

## Formato de salida

```markdown
✓ <TASK-ID> delivery verified
- PR: <url | not required>
- Tracking: <In Review | Done | no-tracking: reason>
- Documentation: <url | reporter delta-only>
- Next step: <none | action>
```
