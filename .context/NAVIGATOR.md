# Context Navigator — Anvil

last_full_scan: 2026-05-08
last_updated: 2026-05-08T14:35
coverage: bootstrap

## Índice

- [Proyecto](project.md) — stack, arquitectura, restricciones, SOLID
- [Operaciones](ops.md) — comandos para levantar, buildear, testear y operar
- [Patrones](patterns.md) — patrones de diseño inferidos con referencias
- [Contratos](contracts.md) — MCP server, servicios externos (Ollama, Anthropic), event schema, interfaces internas
- [Riesgos](risks.md) — gotchas operativos, deuda técnica, restricciones conocidas

### Dominios activos

- [cli](domains/cli.md) — punto de entrada de todos los subcomandos; orquesta el flujo completo de un run
- [memory](domains/memory.md) — captura, summarización, embedding y búsqueda semántica de outputs de agentes
- [orchestrator](domains/orchestrator.md) — ejecución de pipelines multi-agente como DAGs con gates y políticas de fallo
- [instrumentation](domains/instrumentation.md) — captura y persistencia de eventos de telemetría (event sourcing ligero)
- [dashboard](domains/dashboard.md) — capa de query read-only sobre todos los datos de runs, tasks, agents, métricas
- [mcp](domains/mcp.md) — servidor MCP JSON-RPC 2.0 sobre stdio; expone internals de Anvil a clientes AI
- [deploy](domains/deploy.md) — despliegue de configuración de Anvil a directorios de AI clients (Claude, Codex, Gemini, etc.)

### Decisiones arquitectónicas

<!-- No hay evidencia explícita de decisiones con handoff/SPEC — se llenan on-demand -->

## Notas para agentes

- Leer `project.md` siempre — es el punto de entrada
- Cargar solo los dominios relevantes a la tarea
- Si la tarea toca APIs externas o interfaces entre dominios → leer `contracts.md`
- Si `coverage: bootstrap`, el contexto fue generado automáticamente — puede tener gaps en dominios `tui` y `runner` (no generados por ser < umbral o secundarios)
- `internal/tui/` y `internal/runner/` existen pero no tienen domain file — leer código directamente si la tarea los toca
- `pkg/` contiene utilidades compartidas: `config`, `errors`, `fileutil`, `frontmatter`, `gitutil`, `output`, `registry`, `state`, `storage`
- No modificar este archivo manualmente — actualizarlo vía skill `context-nav`
