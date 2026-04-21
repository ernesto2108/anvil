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

## Decisiones de diseño

<!-- Mini-ADR format per decision -->

### <Decisión 1>

- **Contexto:** ...
- **Decisión:** ...
- **Consecuencias:** ...

## Alternativas consideradas

| Alternativa | Pros | Contras | Razón de descarte |
|---|---|---|---|

## Concerns transversales

- **Seguridad:** ...
- **Observabilidad:** ...
- **Manejo de errores:** ...

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

- Every decision needs a "why" — no unexplained choices
- Mini-ADR format: context → decision → consequences
- Alternatives section is mandatory for Medium+ tasks — shows the architect considered options
- C4 Context diagram shows the system in its environment, not internal details
- Keep this file under 200 lines — detail belongs in domain views
