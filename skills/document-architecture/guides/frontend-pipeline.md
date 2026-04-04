# Frontend Pipeline — React

## Output pattern

```
<docs>/04-architecture/<project>/
├── context-summary.md      # Stack, deps, structure, state, APIs
├── context-modules.md      # Modules/features with routes, components, API calls
├── context-risks.md        # XSS, token storage, deps, PII (with code snippets)
├── overview.md             # C4, route tree, state diagram, auth flow
├── security-audit.md       # OWASP client-side with score
└── modules/
    └── <module-name>.md    # One per module with user flow diagram
```

## Skeleton

File: `<docs>/07-references/frontend-skeleton.md`

Covers: store setup, auth query/hooks, i18n config, auth guards, route helpers.

## Scanner instructions

"The skeleton describes patterns common to frontends of this type. Do NOT re-explore these patterns. Focus on what DIFFERS: modules/features, API endpoints, state slices beyond auth, route tree with guards, external integrations (Sentry, Stripe, PostHog), build tool specifics, TypeScript vs JS. In context-summary.md, reference the skeleton for common patterns and only detail differences. In context-risks.md, include CODE SNIPPETS (5-10 lines) for each finding so security does not need to re-read files."

Output: `context-summary.md`, `context-modules.md`, `context-risks.md`

## Architect — Overview

Inject: context-summary.md INLINE. Do NOT inject modules or risks.

Output `overview.md` with:
1. Descripcion del frontend (que hace, quien lo usa, stack)
2. Diagrama de Contexto (C4) — frontend <-> backends <-> auth <-> 3rd parties
3. Arbol de rutas completo (Mermaid graph TD) — con guards y layouts
4. Diagrama de estado global (Mermaid) — store slices y relaciones
5. Auth flow diagram (Mermaid sequence) — login -> token -> refresh -> logout
6. Dependencias externas (tabla)
7. Notas tecnicas

## Architect — Detail (modules)

Before launching, read `<docs>/07-references/template-module.md`.

Inject: template-module.md + context-modules.md INLINE. Do NOT inject summary or risks.

Instructions: "Use the template as the EXACT format reference. Do NOT read example files."

Output: `modules/*.md` (one per module/feature)

## Security instructions (client-side)

"CLIENT-SIDE audit of a React frontend. Focus on: (1) Token storage — localStorage vs HttpOnly cookies, XSS theft risk. (2) XSS — dangerouslySetInnerHTML, unsanitized input, URL injection. (3) Sensitive data in console.log, Sentry, APM, URL params. (4) Dependency CVEs in package.json. (5) CORS/CSP headers. (6) Auth bypass — route guards, client-only validation. (7) Hardcoded secrets. Context-risks.md has CODE SNIPPETS — do NOT re-read those files. Only Read to trace cross-file dependencies."

Output: `security-audit.md`, bug files in `<docs>/05-bugs/` (critical/high only)

## Injection cheat sheet

| Agent | Inject INLINE | Do NOT inject |
|---|---|---|
| Scanner | frontend-skeleton.md | — |
| Architect (overview) | context-summary.md | modules, risks |
| Architect (modules) | template-module.md + context-modules.md | summary, risks |
| Security | known-systemic-issues.md + context-risks.md + overview summary | full modules |
