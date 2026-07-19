# Evals, Patrones de Agentes y RAG — Referencia (multi-proveedor)

Verificado 2026-07-18. Ítems ⚠️ no verificados en fuente primaria.

## Evals

Tres tipos de graders (complementarios):
1. **Programáticos/código**: string checks, análisis estático, validación de estado. Rápidos y objetivos; frágiles ante variaciones válidas.
2. **LLM-as-judge**: rúbricas flexibles, comparaciones pareadas, consenso multi-juez. **Debe calibrarse contra expertos humanos**. Best practice: pedir al juez que **razone primero y luego dé el score**.
3. **Humanos**: gold standard; caros y lentos — por muestreo y para calibrar jueces LLM.

- **Golden sets**: **20-50 tareas extraídas de fallos reales** es un gran comienzo. Dos expertos independientes deben llegar al mismo veredicto; cada tarea con solución de referencia comprobada.
- **Métricas para agentes**: `pass@k` (al menos 1 de k correcto — capacidad) y `pass^k` (los k correctos — consistencia, clave en producto).
- **Pitfalls**: specs ambiguas → falsos negativos; graders demasiado rígidos que penalizan soluciones válidas; ambientes compartidos → correlaciones espurias; no revisar transcripciones.
- **En local la eval es obligatoria por modelo** — la fragilidad de prompt es mucho mayor que en frontier; un prompt que funciona en un modelo chico no porta a otro.

### Herramientas multi-proveedor

- **promptfoo** es la referencia: config YAML declarativa, **50+ providers en un mismo eval** (`openai:...`, `anthropic:...`, `ollama:chat:qwen3`, Bedrock, Azure, cualquier endpoint OpenAI-compatible), asserts (igualdad, regex, is-json, LLM-rubric), matriz web UI, CI/CD con umbrales. Core MIT.
  - ⚠️ Fuentes secundarias 2026 reportan adquisición de promptfoo por OpenAI (mar 2026) manteniendo el core open source y los providers Anthropic/Ollama — no verificado en fuente primaria.
- Console de Anthropic (Evaluation tool) para el ecosistema Claude; Braintrust/LangSmith/Langfuse/Arize para anotación humana y regresiones.

### LLM-as-judge con modelos locales

- Solo modelos grandes (GPT-4-class, **Llama-3.x 70B+**) alcanzan alineación alta con humanos como judges generales — y aun así divergen hasta 5 puntos en escalas absolutas.
- Los **8B instruction-tuned** (Llama-3.1-8B, Qwen-2.5-7B) sirven para **juicios binarios/rubrics simples en dominio acotado** (hasta 99.75% en clasificación binaria domain-specific), pero **colapsan out-of-domain**.
- **Regla**: judge local ≥70B (o frontier API) para scoring abierto; 7-9B solo para checks binarios con rúbrica cerrada; **nunca judge del mismo modelo/familia que genera** (self-preference bias).

## Patrones de arquitectura de agentes

Bloque base: **augmented LLM** (retrieval + tools + memoria). Patrones: **prompt chaining** (secuencia con checks), **routing** (clasificar y derivar), **parallelization** (sectioning y voting), **orchestrator-workers** (descomposición dinámica), **evaluator-optimizer** (generador + evaluador en loop), **agents autónomos** (loop con feedback del entorno).

Principios: **simplicidad** (muchas apps agénticas son una llamada con retrieval), **transparencia** (mostrar el plan), **ACI** (diseñar tools con el cuidado de una API humana). Usar agentes solo cuando la tarea es abierta/impredecible. En local, `num_ctx` ≥32k para que el agente no pierda el system prompt ni degrade el tool calling.

## RAG y embeddings

- **Regla de decisión**: si el corpus cabe en el context window real, **no uses RAG** — mételo en contexto con caching. RAG para corpus mayores.
- **Chunking contextual**: generar 50-100 tokens de contexto por chunk y prependearlo antes de embeddear e indexar en BM25 (contextual embeddings + BM25 + reranking reduce fallos de retrieval drásticamente). Recuperar top-20; siempre correr evals sobre tu caso.
- **Embeddings**:
  - API: Voyage AI (`voyage-4-*`, `voyage-context-4`), OpenAI, etc.
  - **Locales** (calidad ya competitiva, $0/token → default para RAG interno): **qwen3-embedding** (0.6b/4b/8b; 8B a 70.58 MTEB multilingüe), **nomic-embed-text** (CPU), **mxbai-embed-large** (supera a text-embedding-3-large), **bge-m3** (MIT, multilingüe), **snowflake-arctic-embed2**. En Ollama vía `POST /api/embed`.
  - Usar `input_type="query"`/`"document"`; vectores normalizados; cuantización int8/binary para storage.
- **Rerankers locales**: **bge-reranker-v2-m3** (default histórico), **qwen3-reranker** 0.6b/4b/8b (el 0.6b ≈ bge-v2-m3 con 1/10 del tamaño, corre en CPU), **mxbai-rerank-large-v2**. ⚠️ Ollama no sirve rerankers vía endpoint dedicado — usar sentence-transformers/TEI/Xinference.
- **Vector stores ligeros**: **sqlite-vec** (local-first, single-user, edge, tests sin contenedores); **pgvector** (si ya hay Postgres, hasta pocos millones de vectores); **Qdrant** (filtrado pesado, data residency, escala — el self-hosted "serio").

## Regla local vs API (resumen)

Privacidad/compliance manda sobre costo → local a cualquier escala. Volumen: <50M tokens/mes → API gana; >100M/día → local gana. Latencia interactiva/edge → local. Calidad frontier (razonamiento abierto) → API. Respuesta madura: **híbrida**. Embeddings: default local. (Detalle en `SKILL.md` § Selección de proveedor.)

## Fuentes

- https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents · https://www.anthropic.com/research/building-effective-agents
- https://www.promptfoo.dev/docs/providers/ · https://arxiv.org/html/2510.09738v1
- https://www.morphllm.com/ollama-embedding-models · https://futureagi.com/blog/best-rerankers-for-rag-2026/ · https://www.firecrawl.dev/blog/best-vector-databases
