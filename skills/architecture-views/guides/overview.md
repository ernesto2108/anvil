# Template: architecture.md (Overview)

Inspired by: Google Design Docs + ADR format.

**Always generated.** This is the entry point — decisions, boundaries, trade-offs.

## Template

```markdown
# Arquitectura — <TASK-ID>

## Contexto y alcance

<!-- System landscape. What exists today, what changes, why. -->

## Objetivos

- ...

## No-objetivos

- ...

## Convenciones aplicadas

<!-- 3-5 convention rules that shaped this architecture -->
- ...

## Decisiones de diseño (ADR)

<!-- MADR format per decision. One block per non-obvious choice. -->

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

<!-- Repeat block for each non-obvious decision -->

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

## Rules

- Every decision needs a "why" AND rejected alternatives — no unexplained choices
- ADR format: context → options considered → decision + forces → consequences
- Alternatives section is mandatory for Medium+ tasks — shows the architect weighed options
- C4 Context diagram shows the system in its environment, not internal details
- Keep this file under 200 lines — detail belongs in domain views
- "Concerns transversales" must address idempotency and observability — not just security
