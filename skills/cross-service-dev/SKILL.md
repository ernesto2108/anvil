---
name: cross-service-dev
description: Protocolo procedimental para coordinar trabajo cross-service en múltiples repos de microservicios en una sola sesión. Úsalo cuando el usuario diga "implement across services", "this touches X and Y services", "cross-service feature", "work on multiple repos", "remove this endpoint from all services", "deprecate this across services", "refactor cross-service", o describa cualquier cambio (crear, actualizar, eliminar, deprecar) que requiera trabajo coordinado en 2+ servicios. Define fases, gates y formato de salida. Requiere service-map.yaml para resolver rutas de repos.
disable-model-invocation: true
---

# Cross-Service Dev — Protocolo Multi-Repo

## Filosofía

1. **Mismos gates que un run de un solo repo** — multi-repo no relaja el pipeline; lo replica coordinado.
2. **Cero omisión silenciosa** — todo servicio afectado debe aparecer en el scope o ser declarado explícitamente como excluido con justificación.
3. **Contexto cross-service consolidado** — las decisiones de diseño y QA se toman sobre el conjunto completo, no por repo aislado.

## Prerequisitos

- `service-map.yaml` existe en `{service_map_path}` (en `.project-context/service-map.yaml` o el repo)
- Los repos de servicios afectados están en disco (`local_path` resuelve)
- Si existe `service-map.local.yaml`, se usa para overrides de rutas locales

## Flujo de Trabajo

### Fase 1 — Descubrimiento y scope

1. **Clasificar la operación:** Create | Update | Delete | Deprecate.

2. **Resolver rutas de servicio desde service-map.yaml:**
   ```
   full_path = projects_root + "/" + service.local_path
   ```
   Verificar que cada ruta existe en disco y aplicar:
   - **Repo existe localmente** → continuar con esa ruta.
   - **Repo no existe localmente pero hay URL en `service-map.yaml`** → clonar a directorio temporal (`/tmp/<nombre-servicio>-<timestamp>`); la Fase 1.5 lo explora desde ahí; al terminar, eliminar el clon temporal.
   - **Repo no existe localmente y no hay URL** → DETENER. Pedir al usuario URL del repo o ruta local. Persistir la respuesta en `service-map.local.yaml` bajo `local_path`.
   - **Repo no accesible por ninguna vía** → marcar como `sin contexto de código` en el documento de diseño. Documentar explícitamente qué decisiones quedan sin verificar por falta de acceso.

3. **Descubrir dependencias transitivas (OBLIGATORIO):**
   Por cada servicio en scope, consultar service-map.yaml:
   - ¿Quién consume este endpoint? (`consumed_by`)
   - ¿Quién se suscribe a este evento? (`consumed_by` en `publishes`)
   - ¿Quién lee esta tabla? (`readers` en `shared_database`)
   - ¿Quién depende de este servicio? (`depends_on`)

   **Si aparecen servicios adicionales → DETENER y reportar al usuario antes de continuar.**
   Reglas:
   - NUNCA omitir silenciosamente servicios afectados
   - DELETE/DEPRECATE → la verificación transitiva es CRÍTICA
   - UPDATE con cambios de contrato → todos los consumidores están afectados

4. **Producto de la fase:** un PRD local en `{task_path}/prd.md` que:
   - Lista TODOS los servicios en scope bajo Dependencias
   - Anota los servicios excluidos como pendientes con justificación
   - Especifica el tipo de operación
   - Si el proyecto tiene `task_tool` configurado en `.project-context/project.md`, indicar al humano que vincule el PRD en esa herramienta — nunca ejecutar acciones en ella

5. **Coordinar `parent_branch` cross-repo cuando el trabajo pertenece a un milestone.**

   Mismo criterio de resolución que `delivery-flow`: intentar derivar el milestone de Linear; **la validación humana es obligatoria siempre** — confirmar con el humano si Linear trae el dato, preguntarle directamente si no lo trae.

   - Resolver `parent_branch: feature/<milestone-slug>` **una sola vez para todo el run**, no por repo. Persistir el resultado (milestone + parent_branch) para no volver a preguntar dentro del mismo run.
   - Por cada repo en scope: si `parent_branch` no existe (ni local ni remoto) en ese repo, crearlo desde `develop` actualizado, pushearlo, y aplicar el mecanismo de PR draft de tracking de `delivery-flow` (Paso 1.5.1) en ESE repo — título `[NO MERGEAR] integración <milestone>: <descripción corta>`, cuerpo con advertencia de solo-merge y referencia a la tarea padre, guardando la URL resultante.
   - Si `parent_branch` ya existe en algunos repos del scope pero no en otros (milestone con trabajo previo en un repo y uno nuevo que se suma ahora), crearlo únicamente donde falte — **nunca fallar el run completo por esto**.
   - Si el trabajo NO pertenece a ningún milestone, omitir esta sub-sección y continuar con el comportamiento habitual (cada repo usa su base por defecto).
   - **NUNCA omitir silenciosamente** la resolución de `parent_branch` cuando hay milestone: si no se puede resolver o crear en algún repo, DETENER y reportar al humano antes de avanzar a la Fase 1.5.

### Fase 1.5 — Exploración de código por repo (OBLIGATORIA)

Esta fase no se puede saltar. Garantiza contexto técnico real de cada servicio antes de decisiones de diseño.

Por cada repo resuelto en la Fase 1:

1. **Explorar el repo** leyendo firmas de función, contratos de API, schemas, estructuras de evento, tipos e interfaces relevantes al scope. Cuando hay 2+ repos → ejecutar en paralelo.

2. **Inputs de la exploración:**
   - Ruta local del repo (`full_path` o el clon temporal cuando aplique)
   - Paths específicos dentro del repo que la tarea va a tocar (tomados del PRD)

3. **Output por repo:** un bloque `## Contexto técnico — <nombre-servicio>` con firmas, contratos, schemas y tipos relevantes.

4. **Consolidación:** concatenar los bloques de cada repo en un único documento, que resuelve `{context_path}` como input de la Fase 2.

**Servicios marcados como `sin contexto de código` se omiten en esta fase** — anotarlos explícitamente como faltantes en el bloque consolidado.

### Fase 2 — Arquitectura

Inputs:
- PRD (inline o desde `{task_path}/prd.md`)
- `{context_path}` consolidado con bloques de cada servicio en scope

Producto: `{task_path}/design.md` con:
- Una sección por servicio
- Definiciones de contrato (compartidas)
- Orden de ejecución
- Propiedad de migración

**GATE: veto arquitectónico → DETENER.**

### Fase 3 — Implementación

Por cada servicio: una unidad de implementación independiente que recibe:
- PRD (inline o desde `{task_path}/prd.md`)
- `{task_path}/design.md`
- Skill de convención a cargar
- Ruta del servicio como directorio de trabajo

Reglas de orden:
- **Paralelismo:** servicios independientes → en paralelo. Si B depende de la salida de A → secuencial.
- **Migraciones de BD:** se ejecutan antes que el resto de la implementación del servicio afectado.
- **Operaciones DELETE:** orden inverso — consumidores primero, productor último.

**Milestone:** si la Fase 1 resolvió `parent_branch` para el run, pasarlo a cada delegación de `committer-flow`/`delivery-flow` por servicio, para que el PR de cada tarea hija apunte a ese `parent_branch` y no a `develop`.

### Fase 4 — Testing

Por cada servicio modificado: un ciclo de tests propio. Todos en paralelo cuando sea posible.

### Fase 5 — QA cross-service

QA sobre el diff combinado de todos los servicios. Foco en:
- Consistencia de contrato entre productor y consumidores
- Coincidencia de payloads de eventos
- Alineación de tipos de API
- Consistencia de BD en tablas compartidas

**GATE: puntuación QA < 7 → DETENER.**

### Fase 5.5 — Actualizar service-map.yaml (CONDICIONAL)

**Activación:** el diff combinado del run toca al menos uno de:
- Handlers HTTP (rutas, endpoints, controllers)
- Archivos `.proto` o `.graphql`
- Definiciones de eventos (publishers/subscribers, payloads)
- Schemas de BD compartidos (tablas con readers cross-service)

**Si el diff NO toca contratos → omitir esta fase.** Reportar "No hay cambios de contrato en este run" y continuar a Fase 6.

**Si el diff sí toca contratos:**
1. Leer el diff consolidado de todos los servicios del run.
2. Actualizar `{service_map_path}`:
   - Agregar entradas nuevas (endpoints, eventos, dependencias descubiertas)
   - Modificar las que cambiaron (firmas, payloads, consumers)
   - **Proponer** eliminar las obsoletas — la eliminación requiere **confirmación humana explícita** antes de aplicarse
3. Output: diff aplicado a `service-map.yaml` + lista de eliminaciones propuestas pendientes de confirmación.

> Si esta fase se ejecuta, ya deja `service-map.yaml` consistente con el nuevo estado y la Fase 6 omite su actualización manual.

### Fase 6 — Documentar y reportar

**6a.** Agregar a `{reports_path}/cross-service-changes.md`:
- Fecha, operación, scope, cambios por servicio, contratos tocados, trabajo pendiente, orden de deploy

**6b.** Actualizar `{service_map_path}` para reflejar el nuevo estado — **omitir si la Fase 5.5 ya lo actualizó**.

**6c.** Generar reporte de cierre del run en `{reports_path}/last-run.md`.

## Reglas Clave

1. Los mismos gates que un pipeline single-repo
2. La arquitectura y el QA ven TODOS los servicios — contexto cross-service completo
3. Implementación y testing son por servicio — guiados por el `design.md` consolidado
4. NUNCA omitir silenciosamente servicios afectados
5. Orden de eliminación inverso al de creación — consumidores primero, productor último
6. Todos los docs centralizados en `.project-context/` o el repo — sin duplicación entre repos

## Formato de Salida

Al cerrar el run, producir:

```
## Cross-Service Run — <operación>
- Servicios en scope: <lista>
- Servicios excluidos: <lista + razón>
- PRD: {task_path}/prd.md
- Design: {task_path}/design.md
- QA score: <n>/10
- service-map.yaml actualizado: sí | no | propuestas pendientes
- Cambios documentados en: {reports_path}/cross-service-changes.md
- Reporte de run: {reports_path}/last-run.md
- Orden de deploy: <lista ordenada>
- Trabajo pendiente: <lista>
```
