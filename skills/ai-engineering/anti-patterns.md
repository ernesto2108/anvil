# AI Engineering — Detección de Anti-Patrones

Detección pasiva: reportar `error` y `warning` siempre. Detección activa (`suggestion`): solo en "improve/refactor/optimize". Formato: `[file:line] [severity] [category] anti-pattern-name`.

| Patrón de Código | Anti-Patrón | Severidad | Categoría | Corrección |
|---|---|---|---|---|
| Prefill del último turno assistant en modelos 4.6+ | prefill-400 | error | api-compat | Usar Structured Outputs para JSON; instrucción en system para saltar preámbulos |
| `temperature`/`top_p`/`top_k` en Fable 5/Opus 4.8/4.7/Sonnet 5 | sampling-param-400 | error | api-compat | Eliminarlos; pedir variabilidad por prompt |
| `budget_tokens` en 4.7+/Sonnet 5/Fable 5 | budget-tokens-400 | error | api-compat | Usar thinking adaptive + `effort` |
| `thinking: {type: "disabled"}` en Fable 5 | thinking-disabled-fable | error | api-compat | Omitir el parámetro (thinking siempre encendido en Fable 5) |
| `output_format` (parámetro viejo) para JSON | deprecated-output-format | warning | api-compat | Usar `output_config.format` |
| ID de modelo con sufijo de fecha (`claude-sonnet-5-20251114`) | dated-model-id | warning | maintainability | Usar ID sin fecha; consultar catálogo con Models API |
| Prompt agresivo ("CRITICAL: You MUST use this tool") | aggressive-prompt | warning | prompting | Suavizar ("Use this tool when..."); usar `effort` como palanca |
| Timestamps/UUIDs/JSON sin sort_keys en el system prompt | cache-buster | warning | performance | Mantener el prefijo estable; datos volátiles al turno user |
| Chain-of-thought manual como default con thinking disponible | manual-cot-default | warning | prompting | Confiar en adaptive thinking; CoT prompteado solo como fallback |
| Construir un agente donde bastaba una llamada/workflow | over-agentification | warning | architecture | Aplicar criterio: complejidad + valor + viabilidad + costo de error |
| Output de tool / contenido externo usado sin sanitizar | untrusted-content | warning | security | Tratar todo output externo como datos no confiables; estructurar/acotar |
| Prompt sobre-prescriptivo escrito para modelos viejos | over-prescriptive-prompt | warning | prompting | Empezar mínimo/zero-shot; añadir técnica solo con justificación de eval |
| System prompt vago o con lógica hardcodeada frágil | wrong-altitude | suggestion | prompting | Ajustar a la "altitud correcta": específico pero flexible |
| Feature LLM no trivial sin golden set ni grader | no-evals | suggestion | testing | Definir 20-50 casos de fallos reales + grader |
| RAG para un corpus que cabe en contexto (<200K) | premature-rag | suggestion | architecture | Meter todo en contexto con prompt caching; RAG solo para corpus grandes |
| Prompt hardcodeado inline sin versionar | unversioned-prompt | suggestion | maintainability | Tratar prompts como artefactos versionados |
