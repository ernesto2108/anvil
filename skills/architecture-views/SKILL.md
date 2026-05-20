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

El criterio primario es **cuántos dominios toca la tarea**, no los puntos de historia. El tamaño influye en la profundidad de cada vista (cuánto detalle, cuántos diagramas, cuántos specs ejecutables), pero no en si el archivo es por dominio o genérico — siempre es por dominio.

| Alcance de la tarea | Output |
|---|---|
| Single-dominio Small (1-5 pts) | Una sola vista de dominio: `ard-<dominio>.md` — narrativa pura, sin specs ejecutables ni diagramas extensos |
| Single-dominio Medium (5-8 pts) | Una sola vista de dominio: `ard-<dominio>.md` con specs ejecutables del dominio (OpenAPI, DBML, etc.) |
| Multi-dominio (2+ dominios, cualquier tamaño) | Una vista por dominio: `ard-backend.md` + `ard-database.md` + … — cada archivo cubre solo su dominio, sin consolidación |
| Large (8+ pts), multi-servicio | Todas las vistas aplicables, specs SDD completos, bridge de contratos entre vistas |

**Regla dura:** `architecture.md` genérico **NO es un output válido en ningún caso**. La narrativa de contexto, objetivos, no-objetivos y concerns transversales vive **dentro** de la vista de dominio correspondiente (sección `## Contexto y alcance` y siguientes). Para multi-dominio, las preocupaciones transversales se documentan en la vista del dominio que las origina, con referencia cruzada desde las demás vistas.

## Nombrado de archivos (OBLIGATORIO)

Cada dominio tiene su propio archivo. Usar exactamente estos nombres.

| Archivo | Cuándo crearlo |
|---|---|
| `ard-backend.md` | Servicios backend, APIs internas, lógica de dominio server-side (Go, Rust, Python, etc.) |
| `ard-database.md` | Schema, migraciones, índices, patrones de acceso a datos |
| `ard-frontend.md` | UI web, jerarquía de componentes (React, Astro, etc.), rutas, estado cliente |
| `ard-mobile.md` | iOS/Android/Flutter — navegación, offline/sync, push, platform channels |
| `ard-infrastructure.md` | Topología de despliegue, IaC, brokers/colas, observabilidad, CI/CD |
| `ard-api.md` | Contrato de API cross-stack cuando la API es el dominio central (SDK público, OpenAPI compartido) |
| `ard-auth.md` | Cuando auth (identidad, autorización, tokens, sesiones) es el dominio central |

❌ `architecture.md` genérico no es un output válido — usar siempre vistas de dominio nombradas.

El orquestador verifica que las vistas de dominio relevantes existan antes de invocar al `spec-writer`. Archivos faltantes → se re-invoca al architect.

## Guías — cargar por vista

Cada guía contiene el template + reglas de formato para una vista. Cargar SOLO las guías relevantes a la tarea.

| Vista | Guía | Cuándo cargar |
|---|---|---|
| Backend (`ard-backend.md`) | `guides/backend.md` | Trabajo de backend |
| Frontend web (`ard-frontend.md`) | `guides/frontend.md` | Trabajo de frontend web |
| Mobile (`ard-mobile.md`) | `guides/mobile.md` | Trabajo de mobile (Flutter, RN, nativo) |
| Base de datos (`ard-database.md`) | `guides/database.md` | Cambios de DB |
| Infraestructura (`ard-infrastructure.md`) | `guides/infrastructure.md` | Cambios de infra |
| API cross-stack (`ard-api.md`) | `guides/api.md` | Contrato de API es el dominio central |
| Auth (`ard-auth.md`) | `guides/auth.md` | Auth es el dominio central |
| Convenciones transversales (MADR, etc.) | `guides/overview.md` | Solo para consultar formato MADR de ADRs y convenciones — **no produce archivo overview** |

**Orden de generación (obligatorio):** vistas de dominio (en el orden en que el dominio aparece en la cadena de impacto: datos → backend → contratos → consumidores) → `adrs/`. No existe paso de "overview" separado — cada vista de dominio se autocontiene. El `spec.md` lo produce el `spec-writer` en una invocación separada después del cierre del architect — NO cargar `guides/spec.md` desde el architect.

## Consistencia de contratos cross-vista

Cuando el architect genera múltiples vistas, los contratos DEBEN ser consistentes:

1. **Schema OpenAPI backend ↔ Interface TypeScript frontend** — mismos nombres de campo, mismos tipos, mismo required/optional
2. **Schema OpenAPI/gRPC backend ↔ Modelos Dart/Kotlin/Swift mobile** — mismos nombres de campo, mismos tipos
3. **Tipos de persistencia backend ↔ Schema intent DB** — mismas columnas, mismos tipos, mismas constraints
4. **Env vars de infra ↔ Referencias de config backend** — mismos nombres de variables
5. **Push notification payloads infra/backend ↔ Handlers mobile** — misma estructura de payload

**Regla:** Definir el contrato UNA VEZ en la vista primaria (usualmente backend), luego referenciar o derivar en vistas secundarias. Nunca duplicar con formas diferentes.

## Checklist de validación (auto-check del architect antes de cerrar)

- [ ] Cada decisión en las vistas de dominio (o en su ADR correspondiente) tiene una razón ("por qué")
- [ ] Contratos cross-vista son consistentes (mismas formas)
- [ ] Todos los paths referenciados verificados con Glob/Grep
- [ ] Archivos/paths nuevos marcados como `NEW`
- [ ] Spec OpenAPI es YAML válido (si aplica)
- [ ] DBML/DDL es sintácticamente correcto (si aplica)
- [ ] Las reglas de convención no contradicen la arquitectura
- [ ] Diagramas legibles (no más de 15 nodos por diagrama)
