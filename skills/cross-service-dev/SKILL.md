---
name: cross-service-dev
description: Orquesta agentes en múltiples repos de microservicios en una sola sesión. Usar cuando el usuario diga "implement across services", "this touches X and Y services", "cross-service feature", "work on multiple repos", "remove this endpoint from all services", "deprecate this across services", "refactor cross-service", o describa cualquier cambio (crear, actualizar, eliminar, deprecar) que requiera trabajo coordinado en 2+ servicios. Extiende el workflow de orchestrate para escenarios multi-repo. Requiere service-map.yaml para resolver rutas de repos.
disable-model-invocation: true
---

# Cross-Service Dev — Orquestación Multi-Repo

## Propósito

Extender el workflow de `orchestrate` para coordinar agentes en múltiples repos de microservicios en una sesión. Los mismos agentes, los mismos gates — el orquestador resuelve rutas, descubre dependencias y enruta agentes a los repos correctos.

```
orchestrate          = pipeline para 1 repo
cross-service-dev    = orchestrate × N repos (coordinado)
```

## Prerequisitos

- `service-map.yaml` existe en `<vault>/04-architecture/`
- Los repos de servicios afectados están en disco (local_path debe resolver)
- Si existe `service-map.local.yaml`, usarlo para overrides de rutas locales

---

## Workflow

### Fase 1 — Descubrimiento PM (agente pm)

Invocar al agente pm con la solicitud del usuario + ruta del vault. El PM maneja el descubrimiento en español.

Adicionalmente, el orquestador debe:

1. **Clasificar el tipo de operación:** Create | Update | Delete | Deprecate

2. **Resolver rutas de servicio desde service-map.yaml:**
   ```
   full_path = projects_root + "/" + service.local_path
   Verificar que cada ruta existe en disco. Si falta → advertir al usuario, pedir la ruta.
   ```

3. **Descubrir dependencias transitivas (OBLIGATORIO):**
   Para cada servicio que se está cambiando, verificar service-map.yaml:
   - ¿Quién consume este endpoint? (consumed_by)
   - ¿Quién se suscribe a este evento? (consumed_by en publishes)
   - ¿Quién lee esta tabla? (readers en shared_database)
   - ¿Quién depende de este servicio? (depends_on)

   **Si se encuentran servicios adicionales → DETENERSE y reportar al usuario antes de continuar.**
   Reglas:
   - NUNCA omitir silenciosamente servicios afectados
   - DELETE/DEPRECATE → la verificación transitiva es CRÍTICA
   - UPDATE con cambios de contrato → todos los consumidores están afectados

4. El PM escribe **un** `prd.md` en `<vault>/03-tasks/<TASK-ID>/prd.md`
   - Debe listar TODOS los servicios en scope bajo Dependencias
   - Debe notar los servicios omitidos como pendientes
   - Debe especificar el tipo de operación

### Fase 2 — Arquitectura (1 agente architect)

Un arquitecto recibe:
- `<vault>/03-tasks/<TASK-ID>/prd.md`
- `<vault>/01-project/context.md` de **cada** servicio en scope

Produce un solo `<vault>/03-tasks/<TASK-ID>/design.md` con:
- Una sección por servicio
- Definiciones de contrato (compartidas)
- Orden de ejecución
- Propiedad de migración

**GATE: veto del architect → DETENER**

### Fase 3 — Implementación (N agentes developer)

Un agente developer por servicio. Cada uno recibe:
- `<vault>/03-tasks/<TASK-ID>/prd.md`
- `<vault>/03-tasks/<TASK-ID>/design.md`
- Skill de convención a cargar
- La ruta de su servicio específico como directorio de trabajo

**Paralelismo:** servicios independientes → en paralelo. Si B depende de la salida de A → secuencial.
**DBA:** 0-1 agente, ejecuta antes que los developers si se necesita migración.
**Operaciones DELETE:** orden inverso — consumidores primero, productor último.

### Fase 4 — Testing (N agentes tester, en paralelo)

Un tester por servicio modificado. Todos ejecutan en paralelo.

### Fase 5 — QA (1 agente QA)

Un agente QA ve el diff combinado de todos los servicios. Foco en:
- Consistencia de contrato entre productor y consumidores
- Coincidencia de payloads de eventos
- Alineación de tipos de API
- Consistencia de DB en tablas compartidas

**GATE: puntuación < 7 → DETENER**

### Fase 6 — Documentar + Reportar

**6a.** Agregar a `<vault>/06-reports/cross-service-changes.md`:
- Fecha, operación, scope, cambios por servicio, contratos, trabajo pendiente, orden de deploy

**6b.** Actualizar `<vault>/04-architecture/service-map.yaml` para reflejar el nuevo estado

**6c.** Agente reporter → `<vault>/06-reports/last-run.md`

---

## Resumen de enrutamiento de agentes

| Fase | Agente | Cantidad | ¿En paralelo? |
|-------|-------|-------|-----------|
| PM | pm | 1 | — |
| Arquitectura | architect | 1 | — |
| Migración DB | dba | 0-1 | — |
| Implementación | developer | N | Sí (cuando son independientes) |
| Testing | tester | N | Sí |
| QA | qa | 1 | — |
| Seguridad | security | 0-1 | — |
| Reporte | reporter | 1 | — |

## Reglas Clave

1. Los mismos agentes, los mismos gates que orchestrate
2. Architect y QA ven TODOS los servicios — contexto cross-service completo
3. Developer y Tester son por servicio — guiados por design.md consolidado
4. NUNCA omitir silenciosamente servicios afectados
5. El orden de eliminación es inverso al de creación — consumidores primero, productor último
6. Todos los docs centralizados en el vault — sin duplicación entre repos
