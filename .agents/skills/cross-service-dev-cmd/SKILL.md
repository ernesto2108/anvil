---
name: cross-service-dev-cmd
description: Entry point explícito para runs multi-repo/cross-service. Define la asignación de agentes a cada fase del pipeline cross-service y los gates de cierre.
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->

# Cross-Service Dev — Entry Point Multi-Repo

Este command es el orquestador delgado del pipeline cross-service. La skill `cross-service-dev` define el **cómo** (procedimiento, fases, gates, formato de salida). Este command define el **quién** (qué agente ejecuta cada fase) y el orden de invocación.

## Brief del usuario

$ARGUMENTS

## Pre-flight

1. Tratar la tarea como cross-service desde el Paso L0 (snapshot git de cada repo involucrado, Navigator por repo cuando exista).
2. Cargar la skill `cross-service-dev` y seguir su workflow extremo a extremo.
3. Resolver `service-map.yaml` desde el vault del proyecto; si no existe, pedirlo al usuario antes de continuar.
4. Si el brief del usuario está vacío o ambiguo respecto a servicios involucrados, tipo de cambio o restricciones, ejecutar el Paso 0 de clarificación antes de spawnear cualquier agente.

## Asignación de agentes por fase

El humano spawnea los agentes en este orden, respetando los gates definidos en la skill.

| Fase de la skill | Agente | Cantidad | Paralelismo |
|---|---|---|---|
| Fase 1 — Descubrimiento y scope | `pm` | 1 | — |
| Fase 1.5 — Exploración por repo | `explorer` | N (uno por repo) | Sí, cuando hay 2+ repos |
| Fase 2 — Arquitectura | `architect` | 1 | — |
| Fase 3 — Migración de BD (si aplica) | `dba` | 0-1 | — (antes de los developers del servicio afectado) |
| Fase 3 — Implementación | `developer-backend` / `developer-frontend` / `developer-mobile` (según stack del servicio) | N | Sí cuando son independientes; secuencial cuando hay dependencias |
| Fase 4 — Testing | `tester` | N (uno por servicio modificado) | Sí |
| Fase 5 — QA cross-service | `qa` | 1 | — |
| Fase 5 — Seguridad (si aplica) | `security` | 0-1 | — |
| Fase 5.5 — Actualizar service-map | `service-map-updater` | 0-1 (condicional según la skill) | — |
| Fase 6 — Reporte de cierre | `reporter` | 1 | — |

## Gates

Aplicar los gates definidos en la skill al cerrar cada fase:

- **Fase 1:** dependencias transitivas adicionales descubiertas → DETENER y reportar al usuario.
- **Fase 2:** veto del `architect` → DETENER.
- **Fase 5:** QA < 7 → DETENER.
- **Fase 5.5:** eliminaciones propuestas en `service-map.yaml` requieren confirmación humana explícita antes de aplicarse.

## Reglas de orden cross-service

- **DELETE/DEPRECATE:** orden inverso al de creación — consumidores primero, productor último.
- **Migración de BD:** el `dba` ejecuta antes del developer del servicio que toca el schema.
- **Servicios sin contexto de código:** marcar en el `design.md`; el `architect` documenta qué decisiones quedan sin verificar.

## Cierre

Al finalizar, el output del run sigue el formato definido en la sección "Formato de Salida" de la skill.
