# Prompting y Context Engineering — Referencia

Verificado 2026-07-18. Las técnicas de esta sección son **universales** (features del prompt, portan a cualquier proveedor). Las secciones marcadas como específicas de un proveedor viven en `providers-*.md`. Ítems ⚠️ no re-verificados literalmente.

## Técnicas de prompting universales (portan a cualquier proveedor)

| Técnica | Recomendación actual |
|---|---|
| **Claridad y directness** | La #1. "Golden rule": si un colega sin contexto se confundiría, el modelo también. Tratarlo como "empleado brillante pero nuevo". |
| **Contexto/motivación** | Explicar el *por qué* mejora resultados ("tu respuesta será leída por TTS, no uses elipsis" > "NUNCA uses elipsis"). El modelo generaliza desde la explicación. |
| **Few-shot / multishot** | Confiable para formato/tono, envuelto en delimitadores (`<example>`). Frontier: **3-5 ejemplos**; modelos pequeños: **1-3 de alta calidad** (más shots pueden empeorarlos). |
| **Delimitadores / XML tags** | Tags consistentes y descriptivos (`<instructions>`, `<context>`, `<input>`), anidados cuando hay jerarquía. Reduce malinterpretación. |
| **Roles / system prompt** | Una frase de rol enfoca comportamiento y tono. Regla portable: **un único system prompt al inicio** (muchos templates locales solo renderizan bien uno). |
| **Long context** | Documentos largos (20k+ tokens) **arriba** del prompt, query/instrucciones al final. Envolver en XML con metadata. Pedir extracción de quotes en `<quotes>` antes de la tarea ("grounding in quotes"). |
| **Formato de salida** | Decir qué hacer, no qué no hacer; el estilo del prompt contagia la respuesta (quitar markdown del prompt reduce markdown en la salida). |
| **CoT explícito** | "piensa paso a paso" escrito en el prompt porta a todos. El reasoning *nativo* NO porta (ver `providers-*.md`). |
| **Prompt chaining / self-correction** | draft → review contra criterios → refine, cada paso una llamada separada para loggear/evaluar/bifurcar. |

## Reglas generales de robustez

- **Prefer general instructions over prescriptive steps**: "think thoroughly" suele producir mejor razonamiento que un plan paso a paso a mano (en modelos con razonamiento capaz).
- **Self-check** ("Before you finish, verify your answer against [criteria]") atrapa errores confiablemente, especialmente en código y matemática.
- **Tool use explícito**: "Can you suggest changes" produce sugerencias, no acciones; para actuar, instrucciones imperativas.
- **Zero-shot cada vez más viable en frontier**: empezar mínimo; prompts sobre-prescriptivos de modelos anteriores pueden **reducir** calidad.
- **Modelos pequeños/locales son frágiles**: chat template correcto, instrucciones cortas, una tarea por llamada, schema estricto > prosa; evaluar por modelo. Detalle en `providers-ollama-local.md`.

> Las palancas específicas de reasoning nativo (adaptive thinking, `effort`, snippets oficiales anti-overengineering / anti-alucinación / `<investigate_before_answering>`) son de Anthropic → ver `providers-anthropic.md`.

## Context engineering

Los conceptos de esta sección son universales; los **flags de API concretos** citados (`compact-2026-01-12`, `memory_20250818`, etc.) son de Anthropic — otros proveedores ofrecen equivalentes distintos o ninguno (verificar por proveedor).

Definición: prompt engineering = escribir instrucciones óptimas (tarea discreta); **context engineering = curar el conjunto óptimo de tokens en cada inferencia** (iterativo, cada turno).

- **Context rot**: la memoria del modelo degrada al crecer el contexto (presupuesto de atención finito, relaciones n² del transformer). Tratar el contexto como recurso finito.
- **System prompts a la "altitud correcta" (Goldilocks)**: ni lógica hardcodeada frágil ni vaguedad. Organizar en secciones delimitadas (background, instructions, tool guidance, output) con XML/Markdown.
- **Tools**: contrato entre agente e información. Minimizar solapamiento, retornar info token-eficiente. Si un ingeniero no puede decidir qué tool usar, el agente tampoco.
- **Just-in-time context**: mantener identificadores ligeros (paths, queries, links) y cargar dinámicamente vía tools en runtime, en vez de pre-cargar todo.

### Estrategias para horizonte largo

1. **Compaction** — resumir la conversación al acercarse al límite y reiniciar con el resumen. Limpiar resultados de tool calls viejos es la forma más ligera. Features de API: compaction server-side (beta `compact-2026-01-12`, trigger default 150K), context editing (`clear_tool_uses_20250919`, `clear_thinking_20251015`).
2. **Structured note-taking** — el agente escribe notas persistentes fuera del context window (NOTES.md, TODO) y las relee. Memory tool oficial (`memory_20250818`).
3. **Sub-agentes** — agentes especializados exploran con context limpio y devuelven resúmenes condensados (~1.000-2.000 tokens ⚠️) al orquestador.

Elección: compaction para conversaciones largas continuas; note-taking para proyectos iterativos; sub-agentes para investigación/análisis paralelo. Los modelos 4.5+/Haiku 4.5 tienen context awareness (rastrean su presupuesto de tokens); conviene decirle al modelo si el harness compacta, para que no cierre tareas prematuramente.

## Fuentes

- https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/claude-prompting-best-practices
- https://platform.claude.com/docs/en/build-with-claude/adaptive-thinking
- https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents
