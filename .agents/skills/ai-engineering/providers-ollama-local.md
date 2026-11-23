# Proveedor: Ollama / modelos locales — Referencia

Ollama es el stack local principal. Verificado 2026-07-18 (v0.32.1, 16 jul 2026). Ítems ⚠️ no verificados en fuente primaria.

## Índice
- [API nativa `/api/chat` vs `/v1`](#api-nativa-apichat-vs-v1)
- [La trampa de `num_ctx`](#la-trampa-de-num_ctx)
- [Structured outputs](#structured-outputs)
- [Tool calling](#tool-calling)
- [keep_alive](#keep_alive)
- [Embeddings locales](#embeddings-locales)
- [Selección de modelo y cuantización](#selección-de-modelo-y-cuantización)

## API nativa `/api/chat` vs `/v1`

**Usar la API nativa `/api/chat` para cualquier cosa seria.** El `/v1` (OpenAI-compat) es solo para reusar SDKs OpenAI sin tocar código, y tiene trampas:

| Aspecto | `/api/chat` (nativa) | `/v1/chat/completions` (compat) |
|---|---|---|
| Streaming + tool calls | Soportado desde v0.8.0 (may 2025), delta real | Tool calls **silenciosamente perdidos** al streamear (respuesta vacía con `finish_reason: "stop"`) |
| `num_ctx` por request | Sí (`options.num_ctx`) | **No** — hay que usar Modelfile o env var |
| `think`, `keep_alive`, `options.*` | Todos | `think` no funciona; `reasoning_effort` parcial/no documentado |
| Structured outputs | `format` con JSON Schema completo | `response_format` con menos control |

## La trampa de `num_ctx`

**Es la trampa #1 de Ollama en agentes.**

- El default **escala con la VRAM** (docs jul 2026): <24 GiB → 4k; 24-48 GiB → 32k; ≥48 GiB → 256k. En máquinas modestas el default efectivo sigue siendo **4k**.
- **Truncación silenciosa**: al exceder `num_ctx`, Ollama NO lanza error — descarta los mensajes más viejos (la cabeza). El system prompt puede desaparecer sin aviso ("el agente olvidó su prompt").
- Precedencia: `options.num_ctx` (por request) > `OLLAMA_CONTEXT_LENGTH` (env) > `PARAMETER num_ctx` (Modelfile) > default. Por `/v1` no se puede fijar por request.
- Recomendación: **fijar `num_ctx` explícitamente**; ≥32k mejora el tool calling; ≥64k para agentes/coding/web search. Más contexto = más VRAM (evitar offload a CPU).

## Structured outputs

- Parámetro `format`: `"json"` (JSON libre) o un **objeto JSON Schema completo** → constrained decoding: JSON malformado mecánicamente imposible. Fiabilidad de *forma* ~100%; la **calidad del contenido** depende del modelo (un 7B no iguala a un frontier).
- Recomendaciones oficiales: `temperature: 0` y **incluir el schema también en el prompt**.
- ⚠️ **Ollama Cloud NO soporta structured outputs** (solo local). Cloud sí configura contexto al máximo.
- v0.31.2 arregló structured output en modelos con thinking.

```python
from ollama import chat
from pydantic import BaseModel

class Persona(BaseModel):
    nombre: str
    edad: int
    ciudad: str

resp = chat(
    model="qwen3",
    messages=[{"role": "user", "content": "Ana tiene 23 años y vive en Lima"}],
    format=Persona.model_json_schema(),
    options={"temperature": 0, "num_ctx": 32768},
)
persona = Persona.model_validate_json(resp.message.content)  # validar SIEMPRE aparte
```

## Tool calling

- Array `tools` (schema estilo OpenAI) en `/api/chat`; el modelo responde con `message.tool_calls`; resultados como mensajes `role: "tool"`. Loop: repetir hasta que no haya más `tool_calls`.
- **Streaming de tool calls** en la API nativa desde v0.8.0: acumular `thinking` / `content` / `tool_calls` de los chunks.
- Modelos con capability **tools** (ollama.com/search?c=tools); docs usan **qwen3** de facto. Generación vigente 2026: Qwen3 / Gemma 4 / gpt-oss / DeepSeek. Contexto ≥32k mejora el tool calling; con 4k los agentes se degradan.

```python
msgs = [{"role": "user", "content": "¿Clima en Lima?"}]
tools = [get_weather]  # el SDK Python convierte funciones a schema
while True:
    r = chat("qwen3", messages=msgs, tools=tools, options={"num_ctx": 32768})
    msgs.append(r.message)
    if not r.message.tool_calls:
        break
    for tc in r.message.tool_calls:
        out = get_weather(**tc.function.arguments)
        msgs.append({"role": "tool", "content": str(out), "tool_name": tc.function.name})
```

## keep_alive

- Default: el modelo se descarga de VRAM tras **5 min** de inactividad.
- Control: env `OLLAMA_KEEP_ALIVE`, campo `keep_alive` por request (gana), CLI `--keepalive`. Valores: duración (`"24h"`), `-1` (residente), `0` (descarga inmediata).
- Multi-modelo en una GPU justa: keep_alive corto deja que Ollama swapee (costo: latencia de recarga).

## Embeddings locales

- Endpoint: **`POST /api/embed`** (batch como array). ⚠️ `/api/embeddings` (singular) es legacy → 404/confusión.
- Modelos vigentes: **qwen3-embedding** (0.6b/4b/8b; 8B = 70.58 MTEB multilingüe, compite con APIs; 0.6b la mejor sub-1GB), **nomic-embed-text** (274MB, CPU-only), **mxbai-embed-large** (supera a text-embedding-3-large), **bge-m3** (MIT, 100+ idiomas, 8k ctx), **snowflake-arctic-embed2**.
- La brecha vs APIs comerciales esencialmente se cerró en 2025-2026 → default local para RAG interno.

## Selección de modelo y cuantización

| Tamaño | Modelos (⚠️ nombres 2026 de agregadores) | Para qué |
|---|---|---|
| 1-4B | Qwen3 1.7B/4B, Gemma 4 chicas, Phi-4-mini | Clasificación, extracción con schema, routing. Sin razonamiento multi-paso. |
| 7-14B | Qwen3 8B, Gemma 4 12B, Phi-4 14B | Sweet spot laptop/GPU 8-16GB. Tool calling de una función fiable; judges binarios domain-specific. |
| 24-32B | Gemma 4 26B/31B (256K ctx), Qwen3 30B-A3B | GPU alta de consumo (Q4). Agentes locales serios, coding. |
| 70B+ / MoE | Qwen3 235B-A22B, DeepSeek V4/R1, Llama 4 Scout, gpt-oss-120b | Razonamiento profundo, judges fiables, largo horizonte. |

- Licencias: **Qwen3 y Gemma 4 son Apache 2.0** → los más seguros para comercial. Ollama v0.32 marca deprecados-para-agentes CodeLlama, Qwen2.5, Llama 3.x, Mistral clásico.
- **Cuantización**: **Q4_K_M** default universal (pérdida ~1-3% en 7-13B). Q8_0 casi lossless pero 2x memoria y ~29% más lento. **Por debajo de Q4 se rompen instruction following y tool calling; Q4_K_M ≥ es seguro.** Un modelo más grande en Q4 gana a uno más chico en Q8.

## Prompting para modelos pequeños

- **Chat template correcto = crítico** — usar la API de chat (no `/api/generate` con prompt crudo); el server aplica el template.
- Few-shot pesa más: **1-3 ejemplos de alta calidad** (muchos shots empeoran a algunos modelos chicos).
- Instrucciones simples y cortas, una tarea por llamada.
- **Schema estricto > prosa** — confiar en constrained decoding (`format`), no en "responde en JSON" por prosa.
- Evaluar por modelo: no hay prompt que porte consistente entre modelos chicos.

## Fuentes

- https://docs.ollama.com/api/openai-compatibility · /capabilities/structured-outputs · /capabilities/tool-calling · /context-length · /capabilities/embeddings
- https://github.com/ollama/ollama/releases · https://ollama.com/blog/streaming-tool · https://ollama.com/pricing
