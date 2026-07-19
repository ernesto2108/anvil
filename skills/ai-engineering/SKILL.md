---
name: ai-engineering
description: Convenciones para integrar LLMs (familia Claude) en features de producto — técnicas de prompting (zero/few-shot, adaptive thinking), context engineering, integración con la API de Claude, structured outputs, tool use, Claude Agent SDK, evals, patrones de agentes y RAG. Úsalo cuando se escriba o revise código que llama a la API de Claude, prompts como artefactos, pipelines RAG, o evals; o el usuario mencione "prompt engineering", "few-shot", "chain of thought", "structured outputs", "tool use", "Agent SDK", "claude-agent-sdk", "messages.create", "output_config", "evals", "LLM-as-judge", "RAG", "embeddings", "context engineering", "prompt caching". NO es para construir servidores MCP — eso es la skill mcp-dev.
---

# AI Engineering

Convenciones para features de producto con LLMs de la familia Claude. Postura por defecto verificada a julio 2026. Filosofía central de Anthropic 2025-2026: **haz lo más simple que funcione** — con modelos más capaces se necesita menos ingeniería prescriptiva humana.

> **Deslinde con `mcp-dev`:** esta skill es sobre integrar LLMs en producto (llamadas a la API, prompts, RAG, evals, agentes). La skill `mcp-dev` es sobre construir servidores MCP. Un feature puede usar ambas, pero son dominios distintos.

## Filosofía

- **Empezar mínimo, añadir técnica solo cuando la eval lo justifique** — arrancar con prompts mínimos y el mejor modelo; los prompts sobre-prescriptivos escritos para modelos viejos **reducen** la calidad en los actuales.
- **Prompt engineering evolucionó a context engineering** — no basta redactar el prompt; hay que curar el conjunto óptimo de tokens en cada inferencia (system, tools, historial, datos). El contexto es un recurso finito y precioso ("context rot").
- **Claridad ante todo** — si un colega sin contexto se confundiría con tu prompt, Claude también. Explicar el *por qué* de una instrucción generaliza mejor que una prohibición.
- **Diseño evaluation-driven** — 20-50 casos extraídos de fallos reales antes que pulir a ciegas; medir, no adivinar.

## Cambios críticos con los modelos con razonamiento (2026)

Estos rompen patrones de años anteriores — verificar en cada integración:

1. **Prefill eliminado** — prefill del último turno assistant devuelve **400** en Claude 4.6+/Sonnet 5/Fable 5. Para forzar JSON usar Structured Outputs; para saltar preámbulos, instrucción en el system prompt.
2. **Chain-of-thought manual es fallback, no default** — los modelos actuales usan **adaptive thinking** (`thinking: {type: "adaptive"}`); Claude decide cuándo/cuánto pensar. `budget_tokens` deprecado en 4.6, **400** en 4.7+/Sonnet 5/Fable 5. El CoT prompteado ("piensa paso a paso") es solo fallback con thinking apagado.
3. **Los prompts agresivos sobre-disparan** — "CRITICAL: You MUST..." causa sobre-triggering; suavizar ("Use this tool when..."). Usar `effort` como palanca antes que más prosa.
4. **Sampling params removidos** — `temperature`/`top_p`/`top_k` devuelven 400 en Fable 5/Opus 4.8/4.7/Sonnet 5. La variabilidad se pide por prompt.
5. **IDs de modelo sin sufijo de fecha** — usar `claude-sonnet-5`, nunca `claude-sonnet-5-20251114`. Consultar el catálogo vivo con la Models API.

Ver `prompting-reference.md` para técnicas vigentes (few-shot, XML, long context, self-correction) y context engineering (compaction, note-taking, sub-agentes).

## Selección de modelo (verificado, caché docs 2026-06-24)

| Modelo | ID | Cuándo elegirlo |
|---|---|---|
| Claude Fable 5 | `claude-fable-5` | El más capaz; razonamiento extremo, agentic long-horizon. Solo si el costo lo justifica. |
| Claude Opus 4.8 | `claude-opus-4-8` | Default para trabajo exigente/agéntico. |
| Claude Sonnet 5 | `claude-sonnet-5` | Caballo de batalla de features de producto — mejor velocidad/inteligencia. |
| Claude Haiku 4.5 | `claude-haiku-4-5` | Clasificación, extracción simple, subagentes baratos, latencia. |

`effort` (`low`/`medium`/`high`/`xhigh`/`max`, default `high`) es la palanca de profundidad/costo; `xhigh` para coding/agentic en los modelos top. En Fable 5 el thinking está siempre encendido (omitir el parámetro). Verificar precios/IDs contra platform.claude.com antes de fijar (pricing intro de Sonnet 5 vence 2026-08-31).

## JSON confiable — jerarquía de técnicas

1. **Structured Outputs nativos** (recomendado): `output_config: {format: {type: "json_schema", schema: {...}}}` en `messages.create()`. Los SDKs ofrecen `client.messages.parse()` que valida contra Pydantic/Zod. El parámetro viejo `output_format` está deprecado.
2. **Tool use forzado**: `tool_choice: {type: "tool", name}` — útil para clasificación con enum. `strict: true` en la definición del tool garantiza que `tool_use.input` valide.
3. **Prompt + retries** — último recurso. El **prefill ya NO es opción** (400).

Limitaciones de schema: sin recursión, sin constraints numéricos/de longitud (los SDKs los validan client-side), `additionalProperties` solo `false`. Incompatible con citations y prefill.

Ver `api-integration.md` para tool use (parallel, Tool Runner, server-side tools), streaming, Batch API (50% descuento), prompt caching (breakpoints, anti-patrones) y el Claude Agent SDK.

## Cuándo construir un agente

Criterio oficial: **complejidad** (multi-step no especificable de antemano) + **valor** + **viabilidad del modelo** + **costo de error recuperable**. Si algo falla → quedarse en single call o workflow. Muchas "apps agénticas" son una sola llamada con retrieval.

- **Client SDK (Messages API directa)**: una llamada (clasificar, resumir, extraer), o workflow con lógica controlada por código.
- **Agent SDK** (`claude-agent-sdk` Python / `@anthropic-ai/claude-agent-sdk` TS): agente "batteries-included" con filesystem/bash/web en tu infra (CI/CD, automatización).
- **Managed Agents** (beta): API REST hosteada; Anthropic corre el loop y el sandbox.

Patrones de workflow (Building Effective Agents): prompt chaining, routing, parallelization, orchestrator-workers, evaluator-optimizer. Preferir el más simple. Ver `evals-and-rag.md`.

## Evals y RAG

- **Evals**: 20-50 tareas de fallos reales para empezar; tres graders (programático, LLM-as-judge calibrado contra humanos, humano por muestreo); métricas `pass@k` (capacidad) y `pass^k` (consistencia). Pedir al juez LLM que razone antes de puntuar.
- **RAG**: si el corpus cabe en <200K tokens (~500 páginas), NO uses RAG — mételo todo en contexto con prompt caching. Para corpus grandes: Contextual Retrieval (contexto por chunk antes de embeddear) + BM25 + reranking. Anthropic no ofrece embeddings propios (documenta Voyage AI).

Ver `evals-and-rag.md` para graders, golden sets, pitfalls, Contextual Retrieval con cifras, embeddings Voyage 4 y memoria.

## Flujo de trabajo

1. **Detectar el tipo de feature** — ¿una llamada (clasificar/extraer/resumir), un workflow con código, o un agente? Aplicar el criterio "cuándo construir un agente". Declarar lo inferido en una línea.
2. **Elegir modelo y palancas** — modelo por tabla de selección; `effort`/`thinking` según complejidad; no fijar sampling params (400 en modelos actuales).
3. **Diseñar el prompt mínimo** — empezar zero-shot; añadir few-shot (3-5 ejemplos en `<example>`), XML tags o self-correction solo si la eval lo pide. Cargar `prompting-reference.md`.
4. **Gate de JSON confiable** — si el feature necesita salida estructurada → Structured Outputs (`output_config.format`), NO prefill. Cargar `api-integration.md`.
5. **Gate de agente** — si se justifica un agente → decidir Client SDK vs Agent SDK vs Managed Agents; cargar `api-integration.md` (Agent SDK) y `evals-and-rag.md` (patrones).
6. **Prompts como artefactos versionados** — tratarlos como código: versionar, no hardcodear timestamps/UUIDs en el system prompt (rompen prompt caching).
7. **Evals antes de cerrar** — si el feature es no trivial, definir el golden set y el grader. Cargar `evals-and-rag.md`.

## Checklist Pre-Implementación

- [ ] Se eligió el nivel correcto: una llamada vs workflow vs agente (no sobre-ingeniería)
- [ ] ID de modelo sin sufijo de fecha; modelo apropiado al costo/latencia
- [ ] Sin `temperature`/`top_p`/`top_k` (400 en modelos actuales)
- [ ] Sin prefill; JSON via Structured Outputs o tool use forzado
- [ ] Sin `budget_tokens`; thinking adaptive o effort como palanca
- [ ] Prompt empieza mínimo; técnica añadida solo con justificación de eval
- [ ] Prompts sin timestamps/UUIDs en el prefijo cacheado; caching aprovechado
- [ ] System prompt a la "altitud correcta" (ni hardcodeo frágil ni vaguedad)
- [ ] Outputs de tools y contenido externo tratados como datos no confiables
- [ ] Golden set + grader definidos para features no triviales

## Detección de Anti-Patrones

Ver `anti-patterns.md` para la tabla completa con severidades.

Señales que siempre deben detener el trabajo:
- Prefill del último turno assistant en modelos 4.6+ → prefill-400 (error)
- `temperature`/`top_p`/`top_k` en Fable 5/Opus 4.8/4.7/Sonnet 5 → sampling-param-400 (error)
- `budget_tokens` en 4.7+/Sonnet 5/Fable 5 → budget-tokens-400 (error)
- ID de modelo con sufijo de fecha hardcodeado → dated-model-id (warning)
- Prompt agresivo ("CRITICAL: You MUST") en modelos 4.5+ → aggressive-prompt (warning)
- Construir un agente donde bastaba una llamada → over-agentification (warning)

## Archivos de Soporte

- `prompting-reference.md` — técnicas vigentes (few-shot, XML, long context, self-correction, tool use explícito), context engineering (context rot, altitud, just-in-time, compaction, note-taking, sub-agentes)
- `api-integration.md` — Structured Outputs, tool use (parallel, Tool Runner, server-side tools, programmatic tool calling), streaming, Batch API, prompt caching, Claude Agent SDK (query, hooks, subagents, sesiones)
- `evals-and-rag.md` — tipos de graders, golden sets, pass@k/pass^k, pitfalls, herramientas de evals, patrones de agentes, Contextual Retrieval, embeddings Voyage, memoria
- `anti-patterns.md` — tabla de detección con severidades y correcciones
