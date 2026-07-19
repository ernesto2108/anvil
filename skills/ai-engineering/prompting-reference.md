# Prompting y Context Engineering — Referencia

Verificado 2026-07-18 contra guías oficiales de Anthropic. Ítems marcados ⚠️ no re-verificados literalmente.

## Técnicas de prompting vigentes

| Técnica | Recomendación actual |
|---|---|
| **Claridad y directness** | La #1. "Golden rule": si un colega sin contexto se confundiría, Claude también. Tratar a Claude como "empleado brillante pero nuevo". |
| **Contexto/motivación** | Explicar el *por qué* mejora resultados ("tu respuesta será leída por TTS, no uses elipsis" > "NUNCA uses elipsis"). Claude generaliza desde la explicación. |
| **Few-shot / multishot** | Vigente y confiable para formato/tono. **3-5 ejemplos** relevantes y diversos, envueltos en `<example>`/`<examples>`. |
| **XML tags** | Vigente. Tags consistentes y descriptivos (`<instructions>`, `<context>`, `<input>`), anidados cuando hay jerarquía. Reduce malinterpretación. |
| **Roles / system prompt** | Vigente; incluso una frase de rol en `system` enfoca comportamiento y tono. |
| **Long context** | Documentos largos (20k+ tokens) **arriba** del prompt, query/instrucciones al final (hasta +30% de calidad). Envolver en XML con metadata. Pedir extracción de quotes en `<quotes>` antes de la tarea ("grounding in quotes"). |
| **Formato de salida** | Decir qué hacer, no qué no hacer; el estilo del prompt contagia el estilo de la respuesta (quitar markdown del prompt reduce markdown en la salida). |
| **Prompt chaining** | Degradado en importancia (el modelo maneja multistep internamente). Útil para inspeccionar salidas intermedias o forzar pipeline. Patrón más común: **self-correction** (draft → review contra criterios → refine), cada paso una llamada API separada. |

## Reglas específicas de modelos con razonamiento

- **Prefer general instructions over prescriptive steps**: "think thoroughly" suele producir mejor razonamiento que un plan paso a paso escrito a mano.
- Los ejemplos multishot con `<thinking>` **sí** funcionan con thinking nativo: Claude generaliza el patrón de razonamiento.
- **Self-check** ("Before you finish, verify your answer against [criteria]") atrapa errores confiablemente, especialmente en código y matemática.
- **Tool use explícito**: "Can you suggest changes" produce sugerencias, no acciones; para actuar, instrucciones imperativas. Paralelización steerable por prompt.
- **Zero-shot cada vez más viable**: empezar mínimo; para Fable 5, prompts sobre-prescriptivos de modelos anteriores **reducen** calidad.
- Snippets oficiales nuevos: anti-overengineering, anti-hardcodeo de tests, anti-alucinación (`<investigate_before_answering>`), estética frontend anti-"AI slop", confirmar acciones destructivas.

## Context engineering

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
