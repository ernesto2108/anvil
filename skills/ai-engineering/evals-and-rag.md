# Evals, Patrones de Agentes y RAG — Referencia

Verificado 2026-07-18. Ítems ⚠️ no verificados en fuente primaria.

## Evals

Tres tipos de graders (complementarios):
1. **Programáticos/código**: string checks, análisis estático, validación de estado. Rápidos y objetivos; frágiles ante variaciones válidas.
2. **LLM-as-judge**: rúbricas flexibles, comparaciones pareadas, consenso multi-juez. Escalable; **debe calibrarse contra expertos humanos**. Best practice: pedir al juez que **razone primero y luego dé el score** (y descartar el razonamiento).
3. **Humanos**: gold standard; caros y lentos — usarlos por muestreo y para calibrar jueces LLM.

- **Golden sets**: "**20-50 tareas simples extraídas de fallos reales** es un gran comienzo". Convertir verificaciones manuales en casos de prueba; priorizar por impacto. Dos expertos independientes deben llegar al mismo veredicto; cada tarea con solución de referencia comprobada.
- **Métricas para agentes**: `pass@k` (al menos 1 de k correcto — capacidad) y `pass^k` (los k correctos — consistencia, clave en producto).
- **Pitfalls**: specs ambiguas → falsos negativos; graders demasiado rígidos que penalizan soluciones válidas; ambientes compartidos entre tests → correlaciones espurias; no revisar transcripciones. Ejemplo: instrucciones "solo reporta high-severity" bajan el recall medido — reportar todo con confianza+severidad y filtrar downstream.
- **Herramientas**: Console de Anthropic (Workbench + Evaluation tool con variables `{{var}}`, CSV, autogeneración, side-by-side, versionado); terceros: Harbor, Braintrust, LangSmith, Langfuse, Arize Phoenix; **promptfoo** (CLI/librería open-source, YAML versionable, CI/CD gating, red-teaming). ⚠️ Adquisición de promptfoo por OpenAI (mar-2026) reportada solo en artículos de competidores; no confirmada en fuente primaria.

## Patrones de arquitectura de agentes

Bloque base: **augmented LLM** (retrieval + tools + memoria, ej. vía MCP). Patrones (Building Effective Agents):
- **Prompt chaining** — secuencia con checks programáticos.
- **Routing** — clasificar y derivar a prompts especializados.
- **Parallelization** — sectioning y voting.
- **Orchestrator-workers** — orquestador descompone dinámicamente y delega (subtareas no predecibles).
- **Evaluator-optimizer** — generador + evaluador en loop (criterios claros + refinamiento iterativo).
- **Agents autónomos** — loop con feedback del entorno, para problemas abiertos.

Principios: **simplicidad** (muchas apps agénticas son una llamada con retrieval), **transparencia** (mostrar el plan), **ACI** (diseñar tools con el cuidado de una API humana). Usar agentes solo cuando la tarea es abierta/impredecible y el costo extra se justifica.

Evolución 2025-2026: los patrones migran de "código que escribes" a features de plataforma (subagentes nativos, compaction/context editing server-side, memory tool, tool search, evaluator-optimizer como servicio). Menos scaffolding prescriptivo: declarar el objetivo up-front y controlar con `effort`.

## RAG y memoria

- **Regla de decisión**: si el corpus cabe en <200K tokens (~500 páginas), **no uses RAG** — mételo todo en contexto con prompt caching. RAG para corpus mayores (con ventanas de 1M el umbral práctico sube, pero el trade-off costo/latencia/context-rot mantiene vigente RAG para corpus grandes).
- **Contextual Retrieval** (técnica oficial): generar **50-100 tokens de contexto por chunk** ("este chunk viene del 10-K de ACME, Q2 2023...") y prependearlo antes de embeddear e indexar en BM25. Cifras oficiales: embeddings contextuales solos **−35%** de fallos; + BM25 contextual **−49%**; + reranking **−67%**. Con prompt caching la contextualización cuesta ~$1.02 por millón de tokens. Recuperar **top-20 chunks**; siempre correr evals sobre tu caso.
- **Embeddings**: Anthropic NO ofrece modelo propio; documenta **Voyage AI**. Generación Voyage 4 (ene-2026): `voyage-4-large`/`voyage-4`/`voyage-4-lite`/`voyage-4-nano` (32K contexto, 1024 dims default, Matryoshka 256/512/2048); dominio: `voyage-code-3`, `voyage-finance-2`, `voyage-law-2`; contextualized chunk: `voyage-context-4`; rerankers `rerank-2.5`. Usar `input_type="query"`/`"document"` en retrieval; vectores normalizados; cuantización int8/binary para reducir storage.
- **Memoria de producto**: Memory tool client-side (`memory_20250818`, directorio `/memories` que persistes); Managed Agents memory stores (beta, workspace-scoped, versionado inmutable, redact PII). Patrón Fable 5: dar siempre una "superficie de memoria" (aunque sea un `.md`) con formato prescrito.

## Fuentes

- https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents
- https://www.anthropic.com/research/building-effective-agents
- https://www.anthropic.com/news/contextual-retrieval
- https://platform.claude.com/docs/en/build-with-claude/embeddings
