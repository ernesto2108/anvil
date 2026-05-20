# Guía: Convenciones transversales (Overview)

Inspirado en: Google Design Docs + formato ADR.

**Esta guía NO produce un archivo de salida.** `architecture.md` genérico ya no es un output válido — todo el ARD vive en vistas de dominio nombradas (`architecture-backend.md`, `architecture-db.md`, etc.).

Esta guía existe para dos usos:

1. **Plantilla de secciones comunes** que cada vista de dominio incorpora al inicio del archivo (Contexto y alcance, Objetivos, No-objetivos, Convenciones aplicadas, Decisiones de diseño, Concerns transversales, Diagramas). Cuando la tarea es multi-dominio, estas secciones viven en la vista del dominio que las origina, con referencias cruzadas desde las demás.
2. **Formato canónico MADR de ADRs** — el bloque `### ADR-NN` de abajo es el formato que el architect usa tanto en archivos ADR standalone (`adrs/ADR-NNN-<slug>.md`) como en decisiones inline dentro de una vista de dominio.

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

## No-objetivos

- ...

## Convenciones aplicadas

<!-- 3-5 reglas de convención que moldearon esta arquitectura -->
- ...

## Decisiones de diseño (ADR)

<!-- Formato MADR por decisión. Un bloque por cada elección no obvia. -->

### ADR-01: <Título de la decisión>

- **Estado:** accepted | superseded-by ADR-XX | deprecated
- **Contexto:** ¿Qué problema o restricción motivó esta decisión?
- **Opciones consideradas:**
  - Opción A — [pro / con]
  - Opción B — [pro / con]
  - Opción C — [pro / con]
- **Decisión:** Elegimos [opción] porque [fuerza principal que pesó].
- **Consecuencias positivas:** ...
- **Consecuencias negativas / tradeoffs aceptados:** ...

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

| # | Pregunta | Owner | Fecha límite |
|---|---|---|---|
| 1 | | | |

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
