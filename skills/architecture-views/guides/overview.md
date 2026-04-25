# Template: architecture.md (Overview)

Inspirado en: Google Design Docs + formato ADR.

**Siempre se genera.** Este es el punto de entrada — decisiones, límites, trade-offs.

## Template

```markdown
# Arquitectura — <TASK-ID>

## Contexto y alcance

<!-- Landscape del sistema. Qué existe hoy, qué cambia, por qué. -->

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
```

## Reglas

- Cada decisión necesita un "por qué" Y alternativas rechazadas — sin elecciones inexplicadas
- Formato ADR: contexto → opciones consideradas → decisión + fuerzas → consecuencias
- La sección de alternativas es obligatoria para tareas Medium+ — demuestra que el architect sopesó opciones
- El diagrama C4 Context muestra el sistema en su entorno, no detalles internos
- Mantener este archivo bajo 200 líneas — el detalle pertenece a las vistas de dominio
- "Concerns transversales" debe abordar idempotencia y observabilidad — no solo seguridad
