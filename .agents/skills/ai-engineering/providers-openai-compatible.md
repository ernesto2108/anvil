# Proveedor: OpenAI-compatible — Referencia

Cubre el estándar de facto que implementan OpenAI, Groq, Together, OpenRouter, vLLM, LM Studio, llama.cpp-server y el endpoint `/v1` de Ollama. Verificado 2026-07-18. Ítems ⚠️ no verificados en fuente primaria.

## Regla central

**"OpenAI-compatible" garantiza el *shape* del request, no el comportamiento.** Chat Completions (`/v1/chat/completions`) es el mínimo común denominador portable: messages con roles, `tools` + `tool_calls`, `response_format`, streaming SSE, `temperature`/`max_tokens`. Todo lo demás se verifica por proveedor.

## Chat Completions vs Responses API

- **Chat Completions** sigue siendo el estándar portable de facto en 2026 — para multi-proveedor es el objetivo.
- **Responses API** (OpenAI): recomendada por OpenAI para proyectos nuevos desde 2025 (mejor caching, tools nativas). Señales de transición: Codex eliminó chat/completions (feb 2026); Assistants API se deprecia en agosto 2026; vLLM implementa ambos. Pero **fuera del ecosistema OpenAI no es aún portable** — no adoptarla como base multi-proveedor.

## Lo que NO porta entre implementaciones "compatibles"

| Aspecto | Realidad |
|---|---|
| `response_format: json_schema` con `strict: true` | Groq lo garantiza solo en gpt-oss-20b/120b (en otros modelos `strict` se **ignora en silencio**); exige todos los campos en `required` + `additionalProperties: false`; **no soporta streaming ni tools junto con structured outputs**. TGI usa `json_object` + campo `value` (formato distinto). |
| Extensiones de schema | vLLM tiene `guided_json` / `guided_regex` / `guided_grammar` (no-OpenAI); bugs conocidos con combinaciones (`pattern` + `maxLength` puede violarse en silencio). |
| Context window | No existe parámetro en la API OpenAI → cada server local lo fija distinto (vLLM: flag de arranque; Ollama: Modelfile/env). |
| Reasoning/thinking | Nombres y semántica distintos (`reasoning_effort`, `think`, `enable_thinking`). |
| Concurrencia | LM Studio es mono-usuario, sin colas/rate-limit/auth — no apto para producción concurrente (4.2x tail latency con 2 clientes). |

**Regla para diseñar:** verificar por proveedor el enforcement real de schema, si `strict` se respeta, y si tools + streaming + structured outputs coexisten.

## Caching

Portable en beneficio (~90% de descuento del input cacheado), no en mecanismo: OpenAI = automático ≥1024 tokens de prefijo (writes gratis, sin TTL garantizado); todo es exact-prefix matching. Regla portable: **contenido estable primero, variable al final**. vLLM tiene prefix caching.

## Capas de abstracción multi-proveedor (2026)

- **LiteLLM** (open source, self-hosted proxy/SDK): traduce ~todos los proveedores al formato OpenAI; datos no salen de tu red; tú operas routing/budgets.
- **OpenRouter** (hosted): una API para cientos de modelos; routing/fallbacks incluidos; los requests pasan por su capa (revisar ZDR si hay data residency).
- Otros: Portkey (gateway con governance), Vercel AI SDK (capa de app TS). No emergió un estándar que desplace al formato OpenAI como lingua franca; **promptfoo + LiteLLM/OpenRouter** es el stack multi-proveedor típico.

## Fuentes

- https://platform.openai.com/docs/guides/migrate-to-responses · https://developers.openai.com/api/docs/deprecations
- https://docs.vllm.ai/en/latest/serving/online_serving/ · https://console.groq.com/docs/structured-outputs · https://lmstudio.ai/docs/developer/openai-compat
- https://openrouter.ai/blog/insights/openrouter-vs-litellm/
