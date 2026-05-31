# Guía: Convenciones transversales (Overview)

Inspirado en: Google Design Docs + formato ADR.

**Esta guía NO produce un archivo de salida.** `architecture.md` genérico ya no es un output válido — toda la arquitectura vive en vistas de dominio nombradas (`arch-backend.md`, `arch-database.md`, etc.).

Esta guía existe para dos usos:

1. **Plantilla de secciones comunes** que cada vista de dominio incorpora al inicio del archivo (Contexto y alcance, Objetivos, Alcance del cambio con `### Out of scope`, Convenciones aplicadas, Decisiones de diseño, Concerns transversales, Diagramas). Cuando la tarea es multi-dominio, estas secciones viven en la vista del dominio que las origina, con referencias cruzadas desde las demás.
2. **Formato canónico Nygard de ADRs** — el bloque `### ADR-NN` de abajo es el formato que el architect usa tanto en archivos ADR standalone (`adrs/ADR-NNN-<slug>.md`) como en decisiones inline dentro de una vista de dominio.

## Template — secciones a embeber en cada vista de dominio

```markdown
# Arquitectura — <DOMINIO> — <TASK-ID>

## Contexto y alcance

<!-- Landscape del sistema. Qué existe hoy, qué cambia, por qué. -->

## Restricciones no-funcionales

| Atributo | Target | Condición de medición |
|---|---|---|
| Latencia p99 | < Xms | bajo carga nominal (Y rps) |
| Throughput máximo | X rps | sin degradación |
| Disponibilidad (SLO) | 99.X% | ventana de 30 días |
| Error budget | X% | basado en SLO anterior |
| Tiempo máx de recuperación (RTO) | Xmin | tras fallo de componente crítico |
| Constraints de seguridad | — | e.g. datos PII, cifrado en tránsito |
| Constraints de compliance | — | e.g. GDPR, SOC2 |

<!--
Completar con los NFRs trazados en `requirements.md`. Si un campo no aplica, escribir `N/A` con justificación.
Para tareas Small, mínimo latencia p99 y disponibilidad.
-->

## Objetivos

- ...

## Alcance del cambio

### In scope
- <qué sistemas, módulos, archivos y comportamientos ESTÁN incluidos en este cambio>

### Out of scope
- <qué NO está incluido — explícito, no asumido>

## Convenciones aplicadas

<!-- 3-5 reglas de convención que moldearon esta arquitectura -->
- ...

## Decisiones de diseño (ADR)

<!-- Formato Nygard por decisión. Un bloque por cada elección no obvia. -->

### ADR-001 — <título en español>

## Status
[proposed | accepted | deprecated | superseded by ADR-NNN]

## Context
[Descripción del problema y las fuerzas en juego]

## Decision
[La decisión tomada — voz activa: "We will..."]

## Consequences
[Resultados esperados: positivos, negativos y neutros]

<!-- Repetir bloque por cada decisión no obvia -->

## Concerns transversales

- **Seguridad:** ...
- **Observabilidad:** qué spans/métricas/logs emite esta feature
- **Idempotencia y reintentos:** ¿esta operación es idempotente? ¿cómo se maneja un reintento?
- **Manejo de errores:** errores retryable vs fatales, propagación

## Diagramas

### C4 Context

```mermaid
C4Context
  ...
```

### Flujo principal

```mermaid
sequenceDiagram
  ...
```

## Preguntas abiertas

| # | Pregunta | Impacto si no se resuelve | Responsable | Deadline |
|---|----------|--------------------------|-------------|----------|
| 1 | [pregunta concreta] | [qué se bloquea] | [persona/rol] | [fecha o "antes de implementación"] |

> Si no hay preguntas abiertas, escribir explícitamente: "Ninguna — todas las ambigüedades fueron resueltas en el diseño."
> No omitir la sección; su ausencia indica que no se revisaron las ambigüedades, no que no existen.
```

## Reglas

- Cada decisión necesita un "por qué" Y alternativas rechazadas — sin elecciones inexplicadas
- Formato ADR: contexto → opciones consideradas → decisión + fuerzas → consecuencias
- La sección de alternativas es obligatoria para tareas Medium+ — demuestra que el architect sopesó opciones
- El diagrama C4 Context muestra el sistema en su entorno, no detalles internos
- Mantener las secciones de overview embebidas bajo 200 líneas dentro de cada vista de dominio — el detalle de implementación de cada dominio va en sus propias secciones
- "Concerns transversales" debe abordar idempotencia y observabilidad — no solo seguridad
