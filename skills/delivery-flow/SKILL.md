---
name: delivery-flow
description: 'Orquesta la trazabilidad completa de plan, feat, fix, hotfix, refactor y chore: tarea, estado persistente, documentación, validación, PR y sincronización con Linear. Úsalo cuando se inicie o cierre trabajo entregable, se cree una tarea, se prepare un PR, o se necesite enlazar Linear.'
user-invocable: false
---

# Delivery Flow

Ciclo de entrega idempotente para trabajo que llega a un remoto. Un cambio no está terminado hasta que su estado, evidencia, documentación, PR y tarea de seguimiento coinciden.

## Filosofía

1. **Una entrega tiene estado verificable.** El identificador de tarea, los enlaces y las validaciones viven en un archivo de estado, no en la memoria de una conversación.
2. **La configuración no es un secreto.** El proyecto decide si usa seguimiento y qué perfil usa; las credenciales se leen en runtime desde el registro privado del usuario y nunca entran al repositorio.
3. **No hay cierre parcial silencioso.** Cada operación externa es idempotente y un fallo deja un estado retomable con el siguiente paso explícito.

## Configuración y límites

La fuente de verdad de routing es `.project-context/`. Debe declarar, en su documento de gestión de tareas o configuración equivalente:

```yaml
task_tool: Linear | none
tracking_profile: <non-secret-profile>
documentation_backend: reporter | outline
pr_target_branch: <optional-branch>
```

`tracking_profile` solo selecciona una entrada no secreta del registro privado. Las credenciales de Linear y, si aplica, de Outline se leen automáticamente en runtime de `$HOME/.claude/project-registry.md`; nunca se piden repetidamente, se imprimen, se copian al estado ni se escriben en git. Si falta el archivo, perfil o credencial, detener y pedir al humano que lo configure allí, no el secreto.

`task_tool: none` permite un flujo sin tracking externo. `documentation_backend: reporter` exige el delta local de `reporter`; `outline` exige además el documento externo configurado. No inferir herramientas desde el nombre del repositorio.

## Estado persistente

Crear o reutilizar `.project-context/runs/<run-id>/delivery-state.yaml`. Si no hay `run-id`, usar `ad-hoc`. Nunca crear un segundo issue, documento o PR si ya hay URL/ID en este archivo.

```yaml
kind: plan | feat | fix | hotfix | refactor | chore
tracking: required | no-tracking
exception_reason: ""
task_id: LIN-123
task_url: https://...
milestone: payments-v2
parent_branch: feature/payments-v2
branch: feat/LIN-123-short-slug
reporter_status: pending | complete | skipped
validations: []
commit: ""
pr_url: ""
parent_branch_pr_url: ""
documentation_url: ""
linear_sync: pending | in_progress | in_review | done | skipped
status: initialized | implementing | ready_for_review | delivered | blocked
next_step: ""
```

`milestone` se omite y `parent_branch` vale `null` cuando el trabajo no pertenece a ningún milestone. Cuando existen, `parent_branch` siempre es `feature/<milestone-slug>` y es la base del PR.

## Flujo de trabajo

### 1. Clasificar y resolver tracking

1. Inferir `kind` del prompt: aceptar `plan`, `feat`, `fix`, `hotfix`, `refactor` y `chore`. Si no es seguro, preguntar una vez.
2. Leer `.project-context/` para obtener `task_tool`, perfil y backend de documentación.
3. Reutilizar `task_id` aportado por el usuario, detectado en la rama o guardado en el estado. Si Linear es requerido y no existe, redactar título y descripción en español con `Contexto`, `Alcance` y `Criterios de aceptación`; pedir confirmación solo antes de crear el issue.
4. Resolver el milestone antes de crear ramas:
   - Intentar derivarlo del proyecto/milestone del issue en Linear.
   - **La validación humana es obligatoria siempre.** Si Linear trae el dato, confirmarlo con el humano; si no lo trae, preguntarle directamente si el trabajo pertenece a un milestone y cuál.
   - Persistir el resultado en `delivery-state.yaml` (`milestone` y `parent_branch: feature/<milestone-slug>`, o `parent_branch: null` si no hay milestone) para no volver a preguntar dentro del mismo run.
5. Crear la tarea, persistir `task_id`/URL, moverla a **In Progress** y nombrar la rama `<kind>/<TASK-ID>-<slug>`. El origen de esa rama depende de `parent_branch`:
   - Con `parent_branch`: hacer `git fetch` primero; si la rama padre no existe (ni local ni remota), crearla desde `develop` actualizado, pushearla al remoto y aplicar el Paso 1.5.1 (PR draft de tracking) antes de continuar. Luego crear la rama de trabajo **desde la rama padre**, nunca desde `develop`.
   - Sin `parent_branch`: comportamiento actual — crear la rama de trabajo desde la base habitual del proyecto.

#### 1.5.1 — PR draft de tracking del `parent_branch` (solo al crearlo por primera vez)

Cuando este paso crea un `parent_branch` que no existía, abrir además un PR **draft** de tracking `<parent_branch> → develop` (o la base habitual del repo), reutilizando la mecánica de creación de PR de `committer-flow` Fase 3 (`gh pr create`, redacción en español):

- **Título:** `[NO MERGEAR] integración <milestone>: <descripción corta>` (Conventional Commits en el tipo cuando aplique, ej. `feat`).
- **Cuerpo:** advertencia de que es solo-destino-de-merge (los PRs de las tareas hijas se mergean aquí, nunca commits directos), que nunca se mergea a la base mientras el milestone esté incompleto, referencia a `task_url` de la tarea padre del milestone en Linear si existe, y nota de que se mergea al cerrar el milestone completo.
- Marcar el PR como `draft` (`gh pr create --draft`).
- **Nota práctica:** si `parent_branch` es idéntico a `develop` (aún sin commits propios porque no se mergeó ningún PR hijo), `gh pr create` falla con "No commits between X and Y". En ese caso, antes de crear el PR, ejecutar:
  ```bash
  git commit --allow-empty -m "chore: branch de integración milestone <slug> [<TASK-ID>]"
  git push origin <parent_branch>
  ```
- Guardar la URL resultante en `delivery-state.yaml` bajo `parent_branch_pr_url`.
6. Para trabajo local o experimental, aceptar `tracking: no-tracking` únicamente cuando el usuario lo pide explícitamente; guardar `exception_reason` y marcar sincronización como `skipped`. Un hotfix puede empezar de inmediato, pero debe crear/enlazar la tarea antes del cierre.

**Puerta:** si tracking requerido no tiene `task_id`, detener antes de modificar código.

### 2. Entregar contexto de implementación

Antes de implementar, pasar al ejecutor el path del estado, `kind`, `task_id`, criterios de aceptación y branch. Tras cada hito, actualizar `status` y `next_step`.

El estado no reemplaza handoffs: los handoffs conservan detalles técnicos; este archivo conserva trazabilidad de entrega.

### 3. Documentar y validar

1. Ejecutar `reporter` en modo delta-only después de que el run modifique archivos. Es obligatorio para cualquier cambio no cosmético; registrar los paths actualizados y `reporter_status: complete`.
2. Registrar en `validations` los comandos y resultados de lint, tests, smoke tests y gates requeridos por el modo. `plan` puede no tener validaciones de código, pero debe registrar sus artefactos.
3. Si `documentation_backend: outline`, generar un documento en español desde el diff, estado y evidencia. Incluir qué se hizo, decisiones, archivos, validación, riesgos/rollback, issue y PR. Pedir confirmación antes de crearlo y guardar su URL.

**Puerta:** no continuar a commit si reporter está pendiente, si una validación requerida falló o si la documentación externa configurada no tiene URL.

### 4. Commit, push y PR

Seguir `committer-flow` en sus tres fases, pasándole el path del estado. Si el estado tiene `parent_branch`, esa es la rama base del PR — no `develop` ni `pr_target_branch`. La Fase 3 crea o reutiliza el PR con `gh pr create`/`gh pr view` y usa esta descripción mínima en español:

```markdown
## Contexto
<Por qué se necesita este cambio.>

## Cambios
- <Cambio observable 1>
- <Cambio observable 2>

## Validación
- `<comando>` — <resultado>

## Riesgo y rollback
<Nivel de riesgo y acción concreta de rollback.>

## Trazabilidad
- Linear: <URL de la tarea o motivo de no-tracking>
- Documentación: <URL o reporter delta-only>
```

El título sigue Conventional Commits y contiene el `TASK-ID` cuando exista. No aceptar cuerpos de una línea ni placeholders. Guardar commit, PR URL y rama en el estado.

**Puerta:** si `gh` no está autenticado, si no hay remoto o si falla la creación/reutilización del PR, detener sin mover la tarea a Done.

### 5. Sincronizar y cerrar

Con PR URL persistida:

1. Comentar el PR en Linear una sola vez (buscar comentario existente antes de crearlo).
2. Cambiar Linear a **In Review** cuando el PR queda abierto. Cambiar a **Done** solo después de merge o una confirmación explícita de que la política del proyecto permite cerrar al abrir PR.
3. Marcar `status: delivered` y `next_step: ""` únicamente si todos los gates pasaron. Si no, `status: blocked` y describir el siguiente paso.

## Matriz de gates

| Tipo | Linear | Reporter | Validación | PR | Cierre |
|---|---|---|---|---|---|
| plan | crear/reutilizar salvo no-tracking | sí, si escribe artefactos | artefactos revisados | opcional si no hay cambio versionado | tarea lista para implementar |
| feat / fix / refactor / chore | requerido salvo excepción | requerido | según workflow | requerido | In Review / Done según merge |
| hotfix | crear o enlazar antes de cerrar | requerido | smoke + tests aplicables | requerido | no omitir trazabilidad |

## Reintentos y excepciones

- Reanudar desde `delivery-state.yaml`; comprobar primero recursos existentes por ID/URL antes de crear otros.
- No reintentar a ciegas solicitudes fallidas ni repetir mutaciones de Linear, Outline o GitHub.
- `no-tracking` y una documentación omitida solo son válidos con motivo persistido; no aplican a cambios destinados a merge salvo autorización explícita del humano.
- Nunca escribir secretos en archivos, issue, comentario, PR, salida o handoff.

## Formato de salida

```markdown
## Delivery status
- Kind: <kind>
- Task: <ID + URL | no-tracking: reason>
- Documentation: <URL | reporter delta-only | pending>
- Validation: <passed / pending / failed>
- PR: <URL | pending>
- State: <delivered / blocked>
- Next step: <action or none>
```
