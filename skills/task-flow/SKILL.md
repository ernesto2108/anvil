---
name: task-flow
description: Compatibilidad para solicitudes de crear, abrir o cerrar tareas en Linear. Úsalo cuando el usuario diga "crea la tarea", "abre el issue", "cierra la tarea" o entregue un identificador Linear; aplica el ciclo canónico delivery-flow.
user-invocable: false
---

# Task Flow (compatibility)

`delivery-flow` es el ciclo canónico de planificación y entrega. Esta skill existe para que los triggers históricos de tareas continúen funcionando sin mantener un segundo flujo.

## Flujo de trabajo

1. Cargar `delivery-flow` y resolver `.project-context/` antes de realizar cualquier escritura externa.
2. Para “crear” o “abrir”, ejecutar las fases de clasificación y resolución de tracking; mostrar el borrador del issue y pedir confirmación antes de crearlo.
3. Para “cerrar” o “terminar”, reanudar desde `delivery-state.yaml` usando el identificador dado o detectado; completar reporter, validaciones, PR y sincronización antes de actualizar el estado de Linear.
4. Si el proyecto no declara Linear o el usuario indicó `no-tracking`, seguir la excepción registrada en `delivery-flow`; no crear issues por inferencia.

## Formato de salida

Usar exactamente `## Delivery status` definido por `delivery-flow`. Nunca anunciar una tarea como cerrada si falta PR, reporter, documentación requerida o sincronización.
