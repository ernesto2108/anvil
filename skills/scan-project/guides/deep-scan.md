# Deep Scan Guide

Used when scanner is invoked with `mode: deep` (documentation pipelines).

## Stack detection (FIRST)

Check marker files to determine the stack:

| File | Stack | Recipe |
|---|---|---|
| `go.mod` | Go | Check `docs/scanner-recipe-go-gin.md` if gin import exists |
| `package.json` + `src/` with `.tsx` | React | Follow frontend patterns below |
| `pubspec.yaml` | Flutter | Follow mobile patterns below |
| `Cargo.toml` | Rust | Follow generic approach |
| `requirements.txt` / `pyproject.toml` | Python | Follow generic approach |

If a stack-specific recipe exists in `docs/`, load it. It defines extraction order, grep patterns, what to skip, and token budget.

If no recipe exists for the detected stack, follow the generic approach.

## Generic approach

Works for any stack:

1. Perform standard scan (project structure, deps, tests, CI)
2. Trace endpoint/route flows: handler → logic → data layer chain
3. **Prefer Grep over Read:** extract function signatures, bindings, queries, external calls. Only Read full functions when grep context isn't enough
4. **Write THREE segmented files** (not one monolith)
5. The goal: **no subsequent agent needs to re-read source code** — all flows are already traced

## Output files

### context-summary.md (~200 lines max)

```markdown
# {project-name} — Contexto tecnico
## 1. Identificacion (framework, runtime, module, responsabilidades)
## 2. Estructura de directorios (arbol, patron arquitectonico)
## 3. Dependencias principales (tabla: libreria, version, uso)
## 4. Endpoints/routes expuestos (tabla: metodo, ruta, handler, descripcion)
## 5. Dependencias externas (DB, cache, HTTP, gRPC, brokers — con config keys)
## 6. Configuracion (estructura resumida, variables de entorno)
## 7. Middleware/interceptors (cadena completa en orden)
## 8. Logica de negocio destacada (state machines, async patterns, rollbacks)
## 9. Schema de BD inferido (tablas principales con campos clave)
## 10. Notas tecnicas (deuda tecnica, issues conocidos)
```

Be concise: tables over prose. Omit config values — only structure and keys.

### context-detail.md (~400 lines max)

The "detail" file adapts to project type:
- **Backend (Go, Python, Rust):** endpoint flows (handler → service → repo)
- **Frontend (React):** module/feature flows (routes, components, API calls, state)
- **Mobile (Flutter):** screen flows (widgets, BLoC/providers, API calls)

```markdown
# {project-name} — Flujos detallados

### POST /endpoint-name (or Screen/Module name)
- Entry: {file}:{line} → description
- Logic: {file}:{line} → what it does
- Data: {file}:{line} → queries, API calls, cache
- External: {url/service} (timeout, protocol)
- Side effects: events, notifications
```

Keep each flow to 5-15 lines. Summarize similar ones.

### context-risks.md (~100 lines max)

```markdown
# {project-name} — Areas de riesgo para auditoria
## Archivos riesgosos (tabla: archivo, linea, razon)
## Patrones de concurrencia (threads, async, shared state — con archivos)
## Input sin validar (datos que llegan directo a DB/services sin sanitizar)
## Dependencias externas sin proteccion (sin timeout, sin TLS, sin auth)
## Codigo sospechoso (error ignorado, panic sin recover, SQL concat, eval, dangerouslySetInnerHTML)
```

Only facts relevant to security. No architecture, no business logic.

## Line budget

**CRITICAL:** The three files combined MUST NOT exceed 700 lines. If many endpoints/screens, summarize patterns (e.g., "CRUD endpoints all follow the same handler → service → repo chain").

This enables the orchestrator to inject ONLY the relevant file into each agent prompt, keeping each under 10k tokens.
