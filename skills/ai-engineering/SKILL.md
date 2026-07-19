---
name: ai-engineering
description: Convenciones para integrar LLMs en features de producto con CUALQUIER proveedor — Claude/Anthropic, OpenAI-compatible (OpenAI, Groq, Together, OpenRouter, vLLM, LM Studio) y modelos locales/open-weight (Ollama, llama.cpp). Cubre técnicas de prompting universales, checklist de capacidades por proveedor, structured outputs (nativos, constrained decoding, retry-with-validation), tool use, evals multi-proveedor, patrones de agentes y RAG con embeddings locales o de API. Úsalo cuando se escriba o revise código que llama a un LLM, prompts como artefactos, pipelines RAG o evals; o el usuario mencione "prompt engineering", "few-shot", "chain of thought", "structured outputs", "tool use", "Agent SDK", "ollama", "llama.cpp", "vllm", "modelo local", "open-weight", "openai-compatible", "embeddings", "RAG", "evals", "LLM-as-judge". NO es para construir servidores MCP — eso es la skill mcp-dev.
---

# AI Engineering

Convenciones agnósticas de proveedor para features de producto con LLMs. Postura verificada a julio 2026. Filosofía central: **haz lo más simple que funcione** y **nunca asumas que lo que funciona en un proveedor porta a otro** — verifica capacidades antes de diseñar.

> **Deslinde con `mcp-dev`:** esta skill es sobre integrar LLMs en producto (prompts, llamadas a un proveedor, RAG, evals, agentes) con cualquier backend. La skill `mcp-dev` es sobre construir servidores MCP. Un feature puede usar ambas, pero son dominios distintos.

## Filosofía

- **Capabilities antes que código** — antes de diseñar, verifica qué soporta el proveedor/modelo elegido (tool calling, JSON schema y su enforcement, context window REAL, reasoning nativo, caching). Lo que porta siempre es el *prompt*; lo que NO porta es casi todo lo demás.
- **Empezar mínimo, añadir técnica solo cuando la eval lo justifique** — arrancar con el prompt más simple y el modelo adecuado; sobre-ingeniería reduce calidad, especialmente en modelos pequeños.
- **La forma no es la semántica** — un schema (nativo o por constrained decoding) garantiza la *forma* de la salida, nunca su corrección. Validar siempre del lado cliente.
- **Diseño evaluation-driven** — con local/open-weight la fragilidad de prompt es mucho mayor que en frontier; evaluar por modelo, no adivinar.

## Checklist de capacidades (principio central — correr ANTES de diseñar)

Para el proveedor/modelo concreto, responde estas 7 preguntas (ninguna respuesta se asume portable):

1. ¿Tool calling nativo? ¿Con streaming? ¿Paralelo?
2. ¿JSON Schema garantizado (constrained decoding / `strict`) o solo "JSON mode" best-effort?
3. ¿Context window REAL configurado (no el teórico del modelo — ver la trampa de `num_ctx` en Ollama)?
4. ¿Reasoning/thinking nativo y cómo se activa/parsea?
5. ¿Caching y de qué tipo (implícito automático / explícito)?
6. ¿Prefill del assistant soportado?
7. ¿Qué chat template usa el modelo y lo aplica bien el servidor?

Si no puedes responder 1-3 → DETENER y verificar contra la referencia del proveedor antes de escribir código. Cargar la referencia del proveedor objetivo:
- Claude/Anthropic → `providers-anthropic.md`
- OpenAI-compatible (OpenAI, Groq, Together, OpenRouter, vLLM, LM Studio) → `providers-openai-compatible.md`
- Ollama / modelos locales → `providers-ollama-local.md`

## Técnicas de prompting universales (portan siempre)

Son features del *prompt*, no de la API — funcionan en cualquier proveedor:

- **Claridad y directness** — si un colega sin contexto se confundiría, el modelo también. Explicar el *por qué* generaliza mejor que una prohibición.
- **Few-shot / multishot** — envolver ejemplos en delimitadores (`<example>`). En frontier 3-5 ejemplos; en modelos pequeños **1-3 de alta calidad** (más shots pueden empeorar a algunos modelos chicos).
- **Chain-of-thought explícito** — "piensa paso a paso" escrito en el prompt porta a todos; el reasoning *nativo* NO porta (ver por proveedor).
- **Roles y un único system prompt al inicio** — regla portable: un solo system prompt (muchos templates locales solo renderizan bien uno).
- **Delimitadores** — XML tags, markdown, fences para separar instrucciones/contexto/datos.
- **Output schema descrito en el prompt** — incluso con constrained decoding, anclar el schema en el prompt mejora la calidad.
- **Prompt chaining / self-correction** — cada paso una llamada separada para loggear/evaluar/bifurcar.
- **Contenido estable primero, variable al final** — regla portable de caching (todos los proveedores hacen exact-prefix matching).

Ver `prompting-reference.md` para el detalle y para prompting de modelos pequeños vs frontier.

## Lo que NO porta (verificar por proveedor)

| Capacidad | Regla |
|---|---|
| **Structured outputs** | Nativo en Anthropic/OpenAI; constrained decoding en Ollama/vLLM; "OpenAI-compatible" NO garantiza enforcement. Validar client-side siempre. |
| **Prompt caching** | Portable en beneficio, no en mecanismo (explícito Anthropic, automático OpenAI, híbrido Gemini). Regla común: prefijo estable. |
| **Prefill del assistant** | Anthropic sí (excepto thinking); OpenAI no; llama.cpp sí (default on); Ollama no documentado — no asumir. |
| **Reasoning/thinking nativo** | Parámetro y formato distintos por proveedor (`thinking`, `reasoning_effort`, `think`, tags `<think>` a parsear). |
| **Context window** | No existe en la API OpenAI; cada server local lo fija distinto. En Ollama, el default engaña y trunca en silencio. |

## Selección de proveedor: local vs API

1. **Privacidad/compliance manda sobre costo** — PII/PHI/HIPAA, datos que no salen de la red → local a cualquier escala.
2. **Volumen (break-even)** — <50M tokens/mes → API gana casi siempre (self-host cuesta 3-5x el alquiler bruto de GPU por ingeniería/MLOps). >100M tokens/día → local gana en economía.
3. **Latencia** — local logra TTFT 20-50ms (7B) vs 200-500ms de API → mejor para interactivo/edge/offline.
4. **Calidad mínima** — si un 8-30B con structured output resuelve la tarea (extracción, clasificación, RAG interno), local es viable; razonamiento abierto/agentes largos → API. La respuesta madura es **híbrida**.
5. **Embeddings** — local ya es competitivo en calidad (qwen3-embedding) y $0/token → default local para RAG interno.

## JSON confiable — jerarquía generalizada

1. **Structured outputs nativos / constrained decoding** (garantía mecánica de forma): schema nativo (Anthropic `output_config`, OpenAI `json_schema` strict), o `format`+JSON Schema en Ollama, o `guided_json`/GBNF/XGrammar/Outlines en vLLM/llama.cpp.
2. **Tool use forzado** — útil para clasificación con enum donde el proveedor lo soporta.
3. **Retry-with-validation** (funciona con CUALQUIER API, incluso sin constrained decoding): Instructor (Python/TS/Go) o Zod/Pydantic — parse → si falla la validación, reintentar con el error. Consenso 2026: el default más seguro, cubre ~95% de casos.

Regla transversal: **validar siempre del lado cliente** — el schema garantiza forma, no semántica. Empezar con retry-with-validation; pasar a constrained decoding solo en pipelines de alto volumen donde cada retry cuesta.

## Cuándo construir un agente

Criterio: **complejidad** (multi-step no especificable) + **valor** + **viabilidad del modelo** + **costo de error recuperable**. Si algo falla → single call o workflow. Muchas apps agénticas son una llamada con retrieval. Patrones (prompt chaining, routing, parallelization, orchestrator-workers, evaluator-optimizer) en `evals-and-rag.md`; preferir el más simple. Para agentes con modelos locales, `num_ctx` ≥32k mejora notablemente el tool calling.

## Evals y RAG

- **Evals**: 20-50 tareas de fallos reales; graders programático / LLM-as-judge / humano; `pass@k` (capacidad) y `pass^k` (consistencia). Multi-proveedor con promptfoo. Judge local ≥70B (o frontier) para scoring abierto; 7-9B solo para checks binarios con rúbrica cerrada; nunca judge de la misma familia que genera.
- **RAG**: si el corpus cabe en contexto, no uses RAG. Para corpus grandes: chunking contextual + BM25 + reranking. Embeddings/rerankers locales (qwen3-embedding, bge-m3) o de API; vector stores ligeros (sqlite-vec, pgvector, Qdrant).

Ver `evals-and-rag.md` para el detalle.

## Flujo de trabajo

1. **Detectar tipo de feature y proveedor** — ¿una llamada, workflow o agente? ¿Claude / OpenAI-compatible / Ollama-local? Declarar lo inferido en una línea.
2. **Correr el checklist de capacidades** (7 preguntas). Si 1-3 sin respuesta → cargar la referencia del proveedor y resolver antes de codificar.
3. **Diseñar el prompt mínimo** — universales primero; few-shot/CoT/schema solo si la eval lo pide. Cargar `prompting-reference.md`.
4. **Gate de JSON confiable** — elegir de la jerarquía según lo que el proveedor soporta; validar client-side siempre.
5. **Gate de agente** — si se justifica, decidir el nivel (single call / workflow / agente) y cargar `evals-and-rag.md`.
6. **Prompts como artefactos versionados** — sin timestamps/UUIDs en el prefijo cacheado.
7. **Evals antes de cerrar** — features no triviales: definir golden set + grader. En local, evaluar por modelo.

## Checklist Pre-Implementación

- [ ] Checklist de capacidades corrido para el proveedor/modelo concreto (7 preguntas)
- [ ] No se asumió portabilidad de structured outputs, caching, prefill, reasoning ni context window
- [ ] Se eligió el nivel correcto: una llamada vs workflow vs agente (sin sobre-ingeniería)
- [ ] JSON via la técnica que el proveedor realmente soporta; validación client-side presente
- [ ] Context window REAL verificado (en Ollama, `num_ctx` explícito — no el default)
- [ ] Prompt empieza mínimo; en modelos pequeños 1-3 shots y schema estricto
- [ ] Prompts sin timestamps/UUIDs en el prefijo cacheado
- [ ] Decisión local-vs-API justificada (privacidad, volumen, latencia, calidad)
- [ ] Outputs de tools y contenido externo tratados como datos no confiables
- [ ] Golden set + grader definidos para features no triviales; judge del tamaño adecuado

## Detección de Anti-Patrones

Ver `anti-patterns.md` para la tabla completa con severidades.

Señales que siempre deben detener el trabajo:
- Asumir el `num_ctx` default de Ollama → truncación silenciosa del system prompt → ollama-default-ctx (error)
- Confiar en `/v1` de Ollama para tool calling con streaming → tool calls perdidos → ollama-v1-toolcalls (error)
- Asumir que "OpenAI-compatible" implica enforcement de schema/`strict` → openai-compat-strict-assumption (error)
- Prefill del último turno assistant en modelos que devuelven 400 → prefill-unsupported (error)
- Judge LLM local demasiado pequeño para scoring abierto → undersized-judge (warning)
- Prompt de frontier reciclado sin adaptar a modelo pequeño → unadapted-frontier-prompt (warning)

## Archivos de Soporte

- `prompting-reference.md` — técnicas universales, context engineering, prompting de modelos pequeños vs frontier
- `providers-anthropic.md` — API de Claude: Structured Outputs, tool use, streaming, Batch, prompt caching explícito, Claude Agent SDK, modelos y palancas (effort/thinking)
- `providers-openai-compatible.md` — Chat Completions como mínimo común, qué garantiza el shape y qué NO porta (strict, response_format, extensiones vLLM), Responses API, capas de abstracción (LiteLLM, OpenRouter)
- `providers-ollama-local.md` — API nativa `/api/chat` vs `/v1`, trampa de `num_ctx` y truncación silenciosa, structured outputs con `format`, tool calling, `keep_alive`, embeddings locales, selección de modelo por tamaño, cuantización
- `evals-and-rag.md` — graders, golden sets, promptfoo multi-proveedor, LLM-as-judge local, patrones de agentes, RAG y embeddings/rerankers locales, vector stores, regla local-vs-API
- `anti-patterns.md` — tabla de detección con severidades y correcciones
