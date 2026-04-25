---
name: architecture-views
description: Templates y guías de formato para vistas de arquitectura. Lo carga el agente architect para producir archivos de arquitectura por dominio con specs ejecutables (SDD). Usar cuando el architect necesita templates de vistas.
disable-model-invocation: true
---

# Vistas de Arquitectura — Templates y Guía de Formato

Skill de referencia para el agente architect. Provee templates para cada vista de arquitectura con formatos de Spec Driven Development (SDD).

## Filosofía

Los documentos de arquitectura sirven a dos audiencias:
1. **Humanos** — developers, revisores, futuros maintainers necesitan contexto, razones, trade-offs
2. **Máquinas** — agentes, CI, generadores de código necesitan specs ejecutables que puedan consumir y validar

Cada vista balancea ambas. Las secciones narrativas explican "por qué"; las secciones de spec definen "qué" en formato legible por máquinas.

## Cuándo usar cada formato

| Complejidad | Output |
|---|---|
| Small (2 pts), single-stack, sin contratos | Solo `architecture.md` — narrativa con diagramas |
| Medium (5 pts), single-stack con DB o API | `architecture.md` + vista de dominio relevante |
| Medium+ con contratos cross-stack | **Archivos separados por concern** (ver Nombrado abajo) + specs ejecutables |
| Large (8+ pts), multi-servicio | Todas las vistas aplicables, specs SDD completos, bridge de contratos |

## Nombrado de archivos — tareas Medium+ (OBLIGATORIO)

Cada concern tiene su propio archivo.

| Archivo | Cuándo crearlo |
|---|---|
| `architecture-backend.md` | Cualquier cambio de backend (Go, Rust, Python, etc.) |
| `architecture-db.md` | Cualquier cambio de DB/schema |
| `architecture-frontend.md` | Cualquier cambio de frontend web (React, Astro, etc.) |
| `architecture-mobile.md` | Cualquier cambio de mobile (Flutter, React Native, Swift, Kotlin) |
| `architecture-infra.md` | Cualquier cambio de infra/CI |
| `architecture.md` | **Small:** único output. **Medium+:** suplemento overview (contexto, decisiones, concerns transversales) junto a las vistas de dominio |

El orquestador verifica que estos archivos existan antes de invocar al Developer. Archivos faltantes → se re-invoca al architect.

## Guías — cargar por vista

Cada guía contiene el template + reglas de formato para una vista. Cargar SOLO las guías relevantes a la tarea.

| Vista | Guía | Cuándo cargar |
|---|---|---|
| Overview | `guides/overview.md` | Siempre |
| Backend | `guides/backend.md` | Trabajo de backend |
| Frontend web | `guides/frontend.md` | Trabajo de frontend web |
| Mobile | `guides/mobile.md` | Trabajo de mobile (Flutter, RN, nativo) |
| Base de datos | `guides/database.md` | Cambios de DB |
| Infraestructura | `guides/infrastructure.md` | Cambios de infra |
| **SPEC** | `guides/spec.md` | **Siempre — se genera ÚLTIMO, después de todas las vistas** |

**Orden de generación (obligatorio):** overview → vistas de dominio (backend/db/frontend/mobile/infra) → spec.md. El spec referencia archivos de arquitectura — no puede escribirse antes.

## Consistencia de contratos cross-vista

Cuando el architect genera múltiples vistas, los contratos DEBEN ser consistentes:

1. **Schema OpenAPI backend ↔ Interface TypeScript frontend** — mismos nombres de campo, mismos tipos, mismo required/optional
2. **Schema OpenAPI/gRPC backend ↔ Modelos Dart/Kotlin/Swift mobile** — mismos nombres de campo, mismos tipos
3. **Tipos de persistencia backend ↔ Schema intent DB** — mismas columnas, mismos tipos, mismas constraints
4. **Env vars de infra ↔ Referencias de config backend** — mismos nombres de variables
5. **Push notification payloads infra/backend ↔ Handlers mobile** — misma estructura de payload

**Regla:** Definir el contrato UNA VEZ en la vista primaria (usualmente backend), luego referenciar o derivar en vistas secundarias. Nunca duplicar con formas diferentes.

## Checklist de validación (auto-check del architect antes de cerrar)

- [ ] Cada decisión en `architecture.md` tiene una razón ("por qué")
- [ ] Contratos cross-vista son consistentes (mismas formas)
- [ ] Todos los paths referenciados verificados con Glob/Grep
- [ ] Archivos/paths nuevos marcados como `NEW`
- [ ] Spec OpenAPI es YAML válido (si aplica)
- [ ] DBML/DDL es sintácticamente correcto (si aplica)
- [ ] Las reglas de convención no contradicen la arquitectura
- [ ] Diagramas legibles (no más de 15 nodos por diagrama)
