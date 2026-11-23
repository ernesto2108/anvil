---
name: handoff
description: Continuidad de sesión para tareas Medium+. Crea, actualiza, lee y archiva notas de handoff para que los desarrolladores puedan retomar el trabajo entre sesiones sin tener que releer todo. Úsalo cuando un developer clasifique una tarea como Medium (5-8 pts) o Large (8-13 pts), al crear `.handoff/<TASK-ID>.md` o `.handoff/<slug>.md`, al actualizar el live document tras cada paso, o al archivar la tarea. Invocado por el desarrollador o el humano orquestador — no directamente por el usuario. El gate de calidad es verify-handoff.sh (invocado por el humano orquestador), no aprobación manual del usuario.
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->


# Notas de Handoff

Los archivos de handoff viven en `.handoff/` en la raíz del proyecto. Habilitan la continuidad de sesión — si a un desarrollador se le acaban los tokens, la siguiente sesión lee el handoff y retoma exactamente donde se quedó.

## Cuándo aplica el handoff

| Complejidad | Handoff | Por qué |
|---|---|---|
| **Small (1-5 pts)** | NO | Cabe en una sesión, no vale el overhead |
| **Medium (5-8 pts)** | SÍ | Puede abarcar sesiones, la pérdida de contexto es costosa |
| **Large (8-13 pts)** | SÍ | Casi con certeza abarcará varias sesiones |

Aplica tanto si la tarea viene del backlog (con TASK-ID) como si se invoca directamente sin uno.

## Nomenclatura de archivos

- **Con TASK-ID:** `.handoff/<TASK-ID>.md`
- **Sin TASK-ID:** `.handoff/<short-slug>.md` — derivar un slug de la descripción de la tarea (ej. `add-auth-middleware.md`, `fix-payment-flow.md`). El orquestador o el usuario pueden proporcionar un nombre; si no, generarlo.

## Operaciones

### Crear (al inicio de la tarea)

1. Crear el directorio `.handoff/` si no existe
2. Leer `template.md` de este skill y usarlo para escribir el archivo de handoff
3. Completar el plan de ejecución y la tabla de uso de tokens vacía
4. **Devolver el control al humano con el plan** — NO presentar el plan directamente al usuario ni continuar automáticamente. El humano lo incluirá en el output del modo Integración al final.

### Actualizar (continuo, no batch al final)

El handoff es un **live document**. La actualización ocurre **después de cada paso**, no al cerrar la tarea. Si la sesión crashea entre paso N y paso N+1, el handoff debe reflejar el estado real.

Actualizar incrementalmente — no reescribir todo el archivo, agregar o actualizar secciones:
- Marcar como completados los pasos en "Estado actual" (o "Fases" si cross-stack)
- Agregar nuevas entradas a "Archivos modificados" cada vez que tocas un archivo nuevo
- Registrar decisiones en "Decisiones tomadas" en el momento que las tomas (no al final)
- Actualizar "Siguiente paso" para reflejar dónde retomar si la sesión se corta

**Anti-patrón:** dejar el handoff vacío hasta el cierre y volcar todo en los últimos minutos. Ver el agente del stack correspondiente (`agents/developer-backend.md`, `agents/developer-frontend.md` o `agents/developer-mobile.md`) § "Output de cierre" para los momentos en que se actualiza el handoff.

El gate `scripts/verify-handoff.sh` (invocado por el orquestador después del developer) detecta handoffs incompletos y rebota la tarea — actualizar al final ya no es viable.

### Leer (al continuar una sesión)

El orquestador lee el handoff y lo pasa inline al desarrollador. El desarrollador reanuda desde "Siguiente paso" — NO relee PRD/design/context a menos que el handoff lo indique explícitamente.

### Archivar (al completar la tarea)

Llamado por `/task-complete` o manualmente por el orquestador:

1. Leer el contenido del handoff
2. **Agregar `## Resumen de completacion` al archivo de la tarea local** (`{task_path}/task.md` en `.project-context/` o el repo). Seguir el formato de tareas completadas existentes.
3. Mover el archivo de handoff a `.handoff/archive/<TASK-ID>.md`
4. Actualizar el tablero/backlog local si existe — mover la tarea a Done, actualizar frontmatter: `status: done`, `completed: <fecha>`.
5. **Si el proyecto tiene `task_tool` configurado** en `.project-context/project.md` (ej. Linear, Jira, Notion): **describir al humano** qué registrar en esa herramienta (ej. "Agrega el resumen de completación como comentario y mueve <TASK-ID> a Done en {task_tool}"). **Nunca ejecutar la acción en la herramienta externa** — solo describirla. Si `task_tool` está vacío o es `ninguna`, omitir este paso.

## Template

Ver `template.md` en el directorio de este skill para la estructura del archivo de handoff.

## Seguimiento de uso de tokens

Al final de cada sesión (completa o no), agregar una fila a la tabla de **Token usage**:

- **Session**: número secuencial
- **Tokens used**: tokens aproximados consumidos (de los metadatos del agente si están disponibles, de lo contrario estimar)
- **Tokens available**: presupuesto de tokens para el modelo utilizado
- **Tool calls**: número de invocaciones de herramientas
- **Files read**: número de archivos leídos
- **Files written**: número de archivos creados o modificados

## Tareas cross-stack

Cuando una tarea toca múltiples stacks (ej. backend Go + frontend React, o backend Go + mobile Flutter):

1. Usar `## Fases` en lugar de `## Estado actual` — una fase por stack, ordenadas por dependencia (backend primero, luego frontend/mobile)
2. Completar `## Puente de contratos` — el struct/DTO/interface exacto que conecta ambos lados, con JSON tags y tipos TypeScript/Dart lado a lado
3. Agrupar `### Tests requeridos — por stack` por stack — cada grupo con su propio file path, comando de ejecución y lista numerada de tests
4. Completar `#### Tests de automatización` — heredar los tipos marcados "Sí" de la sección "Automatización" del SPEC (E2E, API contract, visual regression, a11y). Solo incluir los que aplican; eliminar los que no

**Por qué importa:** los bugs cross-stack casi siempre ocurren en el límite del contrato. Si el struct Go tiene `json:"runId"` pero la interfaz TS espera `run_id`, un handoff plano no lo detectará. El puente de contratos hace visibles ambos lados en un solo lugar.

Para tareas single-stack, usar el checklist plano `## Estado actual` y omitir `## Fases`, `## Puente de contratos` y el agrupamiento por stack en los tests.

## Tareas cross-service

Cuando una tarea toca múltiples repos/servicios:

1. Completar `## Dependencias cross-service` — tabla con servicio, repo, qué cambia y orden de deploy
2. Documentar contratos compartidos (endpoints de API, schemas de eventos, tablas de DB que cruzan límites)
3. Señalar cambios breaking y el plan de migración

El orquestador DEBE verificar el orden de deploy antes de cerrar la tarea.

## Seguimiento de Input/Output

### Input recibido (al inicio de la tarea)

El desarrollador completa `## Input recibido` al crear el handoff. Es un acuse de recibo de lo que el orquestador proporcionó — si la siguiente sesión encuentra una brecha, sabe qué faltaba vs. qué se perdió.

### Completitud de `### Tests requeridos — por stack` (obligatorio para el developer)

La lista de tests que el developer entrega es la **lista cerrada** que define el alcance del tester. Una lista pobre produce cobertura incompleta, porque el tester no agrega tests en silencio. Por eso la lista DEBE incluir como mínimo:

- **(a)** un caso de éxito y un caso de error por cada interfaz pública listada en `### Public interfaces / contracts`
- **(b)** un caso por cada entrada de `### Edge cases descubiertos`

**Estos mínimos son CASOS, no funciones de test independientes.** Agrupar por función/método/componente: cada ítem de la lista es UN test parametrizado (table-driven en Go/Rust, `it.each`/`test.each` en TS/React, loop sobre casos en Flutter, `@pytest.mark.parametrize` en Python, `@Test(arguments:)` en Swift) que enumera sus casos en una sub-línea. Formato correcto: `1. Test_GetUser — casos: éxito, 404, error de DB` — NO tres ítems `Test_GetUser_exito` / `Test_GetUser_404` / `Test_GetUser_errorDB`. El tester implementa una función por unidad con N casos, nunca N funciones. Agrupar así no reduce la cobertura: cada caso listado sigue exigido.

Si un edge case declarado se deja **intencionalmente** sin caso, anotarlo en la lista con el motivo (ej. `— sin caso: cubierto por validación de tipos en compilación`). Nunca dejar un edge case declarado sin caso y sin explicación: el tester lo reportará como gap al humano.

### Output entregado (antes de terminar)

El desarrollador completa `## Output entregado` antes de reportar como terminado. Es el checklist de entrega que verifica el orquestador. Debe incluir: resultado del build, resultado del lint, resultado de tests existentes, conteo de archivos, verificación del puente de contratos (si es cross-stack) e impacto cross-service.

## Retro (al completar la tarea)

Después de que la tarea esté terminada (todos los agentes finalizaron, QA aprobado o saltado), completar `## Retro` antes de archivar. NO es opcional para tareas Medium+.

**Qué registrar:**
- **Qué funcionó** — patrones, decisiones, enfoques que vale la pena repetir
- **Qué no funcionó** — retrabajos, rebotes de QA, suposiciones incorrectas, lecturas desperdiciadas. Ser específico: "asumí que la columna nullable era NOT NULL, causó rebote de QA" no "debería haber verificado"
- **Métricas** — estimado vs real: story points, rebotes de QA, invocaciones del desarrollador, invocaciones del tester
- **Aprendizaje** — un aprendizaje concreto para tareas futuras (no genérico)

**Quién lo completa:**
- El desarrollador completa "qué funcionó" y "qué no funcionó" desde su perspectiva
- El orquestador completa "métricas" (tiene el panorama completo de invocaciones de agentes y rebotes)
- Cualquiera puede completar "aprendizaje"

**Cómo alimenta la mejora:**
- El orquestador lee las retros de `.handoff/archive/` al planificar tareas similares
- Los patrones que se repiten en 3+ retros deberían convertirse en entradas de memoria o actualizaciones del convention skill

## Reglas

- Un archivo por tarea
- `.handoff/` DEBE estar en el `.gitignore` del proyecto
- Los archivos de handoff son temporales — no pertenecen al control de versiones ni a la documentación
- NO crear handoff para tareas Small
