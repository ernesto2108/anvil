# Integración con la API de Claude — Referencia

Verificado 2026-07-18 (referencia oficial de la API, snapshot caché 2026-06-24). Ítems ⚠️ no re-verificados literalmente.

## Structured Outputs

- Canónico: `output_config: {format: {type: "json_schema", schema: {...}}}` en `messages.create()`. El viejo `output_format` está deprecado.
- SDKs ofrecen `client.messages.parse()` que valida contra el schema (Pydantic/Zod).
- **Strict tool use**: `strict: true` top-level en la definición del tool (requiere `additionalProperties: false` + `required`) garantiza que `tool_use.input` valide.
- Soportado en Fable 5, Opus 4.8, Sonnet 5, Haiku 4.5. Limitaciones: sin recursión, sin constraints numéricos/de longitud (validados client-side por los SDKs), `additionalProperties` solo `false`. Primer request con schema nuevo paga compilación (caché 24h). Incompatible con citations y prefill.

## Tool use

- Todo pasa por `POST /v1/messages`. `tool_choice`: `auto` / `any` / `tool` / `none`, con `disable_parallel_tool_use` opcional.
- **Parallel tool use activado por default**: devolver **todos** los `tool_result` en un solo mensaje user.
- **Tool Runner** (beta de los SDKs: `client.beta.messages.tool_runner` / `@beta_tool` / `betaZodTool`): maneja el loop request→execute→loop automáticamente, con hooks por turno (approval gates, interceptar errores, retries, streaming, compaction). Recomendado sobre el loop manual. Es parte del SDK de API normal, no del Agent SDK.
- **Server-side tools** (corren en infra de Anthropic): web search / web fetch (`web_search_20260209` / `web_fetch_20260209`), code execution (`code_execution_20260521`), tool search (para catálogos grandes con `defer_loading`). Client-side Anthropic-defined: bash, text editor, memory, computer use.
- **Programmatic tool calling**: Claude compone múltiples tool calls en un script que corre en el contenedor de code execution; los resultados intermedios no entran al contexto — solo la salida final. Para cadenas largas o resultados intermedios grandes.

## Streaming, Batch, Caching

- **Streaming**: obligatorio en la práctica para `max_tokens` > ~16K (timeouts HTTP). Helpers: `.stream()` + `.get_final_message()`/`.finalMessage()`.
- **Batch API**: `POST /v1/messages/batches` — asíncrono con **50% de descuento**; resultados en cualquier orden (clavear por `custom_id`); polling de `processing_status` hasta `ended`.
- **Prompt caching**: prefix-match estricto (orden de render: `tools` → `system` → `messages`); cualquier byte cambiado invalida lo posterior. `cache_control: {type: "ephemeral"}` (TTL 5 min, escritura 1.25×; TTL 1h, 2×; lecturas ~0.1×). Máx 4 breakpoints; mínimo cacheable 2048 tokens (Fable 5/Sonnet 4.6) o 4096 (Opus 4.8/4.7/4.6/Haiku 4.5). **Anti-patrones**: timestamps/UUIDs en el system prompt, JSON sin `sort_keys`, tool sets variables. Verificar con `usage.cache_read_input_tokens`. Novedad Opus 4.8: mensajes `{"role": "system"}` mid-conversation sin invalidar el prefijo cacheado.

## Claude Agent SDK

- **Qué es**: Claude Code empaquetado como librería — mismo agent loop, tools y context management, programable. Renombrado desde "Claude Code SDK" (sep-2025).
- **Lenguajes**: Python (`claude-agent-sdk`, ≥3.10) y TypeScript (`@anthropic-ai/claude-agent-sdk`, incluye el binario nativo de Claude Code). Otros lenguajes: CLI headless (`claude -p --output-format json`).
- **API central**: `query(prompt, options)` — iterador async de mensajes. Opciones: `allowed_tools`/`disallowed_tools`, `permission_mode`, `hooks`, `agents`, `mcp_servers`, `resume`, `setting_sources`.
- **Capacidades**: tools integradas (Read, Write, Edit, Bash, Glob, Grep, WebSearch, WebFetch, AskUserQuestion), **subagents** (via `agents={...}` con `AgentDefinition`), **hooks** (`PreToolUse`, `PostToolUse`, `Stop`, `SessionStart`, `SessionEnd`, `UserPromptSubmit`...), **MCP** (`mcp_servers` stdio/HTTP), permisos granulares, sesiones (resume/fork, JSONL en tu filesystem).
- **Cuándo qué**: Client SDK (una llamada o workflow con código); Agent SDK (agente batteries-included en tu infra); Managed Agents (beta, REST hosteada, loop y sandbox por Anthropic). Patrón: prototipar con Agent SDK local → producción en Managed Agents.
- Auth: `ANTHROPIC_API_KEY` (también Bedrock/Vertex/Foundry via env vars).

## Fuentes

- https://platform.claude.com/docs/en/build-with-claude/structured-outputs · /streaming · /batch-processing · /prompt-caching
- https://platform.claude.com/docs/en/agents-and-tools/tool-use/overview
- https://code.claude.com/docs/en/agent-sdk/overview
