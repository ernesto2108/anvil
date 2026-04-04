# Service Pipeline — Go Backend

## Architecture detection

| Root contains | Pattern | Skeleton |
|---|---|---|
| `internal/` (NO `domain/`) | MVC | `service-skeleton-mvc.md` |
| `domain/` + `usecase/` | Clean | `service-skeleton-clean.md` |
| `internal/` with `business/` | Hex | `service-skeleton-clean.md` (note deviations) |

## Output pattern

```
<docs>/04-architecture/<project>/
├── context-summary.md      # Technical summary
├── context-endpoints.md    # Flows per endpoint
├── context-risks.md        # Risk areas with code snippets
├── overview.md             # C4, components, ERD, sequence diagrams
├── security-audit.md       # OWASP audit with score
└── endpoints/
    └── <endpoint-name>.md  # One per endpoint with sequence diagram
```

## Skeleton reference

### MVC (`service-skeleton-mvc.md`)
- Structure: `internal/{domain}/handlers|services|repositories`
- Middleware: MiddlewareError + MiddlewareTracking (separate)
- Config: JSON local -> SSM fallback
- SQL: raw queries in `queries/*.go`

### Clean (`service-skeleton-clean.md`)
- Structure: `domain/` + `usecase/` + `interface/` + `infrastructure/`
- Middleware: trackRequestAndResponse + authorizationMiddleware (fused)
- Config: YAML local -> SSM fallback
- SQL: squirrel + scany
- Multi-mode: HTTP, gRPC, SQS, Lambda

## Scanner instructions

"The skeleton describes patterns common to services of this type. Do NOT re-explore these patterns. Focus on what DIFFERS: domains, endpoints, business logic, schema, integrations, extra libraries. In context-summary.md, reference the skeleton and only detail differences. In context-risks.md, include CODE SNIPPETS (5-10 lines) for: (a) auth/authorization, (b) input validation, (c) SQL risks, (d) data exposure, (e) CONCURRENCY — Redis locks, errgroup, goroutines, os.Exit, (f) INTEGRATION — Kafka/SQS config, gRPC interceptors, (g) ERROR HANDLING — ignored errors, silenced rollbacks, fmt.Errorf with external strings. Reference inherited systemic issues by ID."

Output: `context-summary.md`, `context-endpoints.md`, `context-risks.md`

## Architect — Overview

Inject: context-summary.md INLINE. Do NOT inject endpoints or risks.

Output `overview.md` (+ optional `state-machine.md`) with:
1. Descripcion del servicio
2. Diagrama de Contexto (C4)
3. Componentes internos
4. ERD
5. Sequence diagrams de flujos criticos
6. Dependencias externas (tabla)
7. Notas tecnicas

## Architect — Detail (endpoints)

Before launching, read `<docs>/07-references/template-endpoint.md`.

Inject: template-endpoint.md + context-endpoints.md INLINE. Do NOT inject summary or risks.

Instructions: "Use the template as the EXACT format reference. Do NOT read example files."

Output: `endpoints/*.md` (one per endpoint)

## Security instructions (backend)

"Systemic issues documented in `known-systemic-issues.md` are PRE-VALIDATED. Include with their ID prefix and note 'inherited from ecosystem'. Do NOT spend tool calls confirming them. For service-specific risks, context-risks.md has COMPREHENSIVE CODE SNIPPETS covering auth, input, SQL, data exposure, concurrency, integrations, and error handling. Do NOT re-read infrastructure/ or usecase/ files. Only Read to trace cross-file call chains not covered by snippets."

Output: `security-audit.md`, bug files in `<docs>/05-bugs/`, backlog entries in sprint-current.md

## Systemic issues reference

File: `<docs>/07-references/known-systemic-issues.md`
- Systemic issues confirmed across audited services
- Frequent issues found in subset of services
- Score reference from ecosystem audit

## Injection cheat sheet

| Agent | Inject INLINE | Do NOT inject |
|---|---|---|
| Scanner | skeleton (resolved by pattern) | — |
| Architect (overview) | context-summary.md | endpoints, risks |
| Architect (endpoints) | template-endpoint.md + context-endpoints.md | summary, risks |
| Security | known-systemic-issues.md + context-risks.md + overview summary | full endpoints |
