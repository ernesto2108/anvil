---
name: ai-architecture
description: Decisiones arquitectónicas de nivel sistema para features que usan LLMs — NO implementación. Cubre agente vs workflow vs llamada simple, RAG vs fine-tuning vs prompting, selección de proveedor a nivel arquitectónico (API frontier vs hosted OpenAI-compatible vs local/open-weight), estrategia de evals a nivel producto, contratos de dominio para features IA (structured outputs, límites de tools), consideraciones de seguridad de diseño y qué decisiones IA ameritan ADR. Úsalo cuando se tomen decisiones de arquitectura y el feature mencione "LLM", "IA", "RAG", "embeddings", "agente", "workflow de IA", "MCP", "prompts", "evals", "fine-tuning", "structured outputs", o proveedores como "Claude", "OpenAI", "Ollama", "vLLM". NO cubre cómo escribir el código que llama al LLM — eso es la skill ai-engineering.
---

# AI Architecture

Guía de **decisiones** para features de producto que dependen de LLMs. Esta skill se ejecuta durante el diseño arquitectónico: produce trade-offs justificados, contratos de dominio y candidatos a ADR — no código.

> **Deslinde con `ai-engineering`:** esta skill **decide** a nivel sistema (qué construir, qué proveedor, qué se evalúa, qué amerita ADR). La skill `ai-engineering` **implementa** (cómo llamar al proveedor, prompting, structured outputs client-side, checklist de 7 capacidades). No dupliques su contenido: cuando una decisión dependa de un detalle de capacidad de proveedor, **referencia** el checklist de 7 capacidades de `ai-engineering` en lugar de reproducirlo. La auditoría formal de seguridad (OAuth, prompt injection, token handling como gate de calidad) tampoco vive aquí: aquí solo se enuncian las consideraciones que condicionan el diseño.

## Filosofía

- **Lo más simple que resuelva el problema gana** — la mayoría de features IA no necesitan un agente ni RAG ni fine-tuning; necesitan una buena llamada con un buen prompt. La complejidad se justifica con evidencia (una eval, un límite de contexto real), no con expectativa.
- **La decisión de proveedor es arquitectónica, la de capacidad es de implementación** — a nivel diseño se decide la *clase* de proveedor por trade-offs de costo/latencia/privacidad/calidad; la verificación de qué soporta el modelo concreto es de `ai-engineering`. No decidas arquitectura sobre capacidades asumidas.
- **Sin eval no hay contrato** — un feature IA sin criterio de "está bien" no es diseñable ni verificable. Definir qué se evalúa es parte del diseño, no un extra posterior.
- **La salida del modelo es un contrato de dominio, no texto libre** — si otro componente consume la salida, su forma y sus límites son parte de la arquitectura.

## Flujo de trabajo

Estas decisiones se toman durante el diseño arquitectónico, antes de escribir Architecture Views y ADRs.

1. **Clasificar la forma de la solución** — ¿una llamada simple, un workflow determinista, o un agente? Aplicar la puerta de "Agente vs workflow vs llamada" (abajo). Empezar por lo más simple; escalar solo con justificación.
2. **Decidir la estrategia de conocimiento** — ¿el modelo base basta (prompting), necesita contexto externo en runtime (RAG), o requiere especialización de pesos (fine-tuning)? Aplicar la puerta correspondiente.
3. **Decidir la clase de proveedor** — API frontier vs hosted OpenAI-compatible vs local/open-weight, por trade-offs de costo/latencia/privacidad/calidad. Si la decisión depende de si el modelo soporta X → DETENER y delegar la verificación de capacidad a `ai-engineering`; no asumir.
4. **Definir el contrato de dominio** — forma de la salida (structured output como contrato), conjunto y límites de tools disponibles, qué entra y sale del sistema IA.
5. **Definir la estrategia de evals a nivel producto** — qué se evalúa, golden set, tipo de grader (programático / LLM-as-judge / humano). Si el feature no tiene criterio de éxito evaluable → DETENER y resolver con el humano antes de cerrar el diseño.
6. **Enumerar consideraciones de seguridad de diseño** — prompt injection, datos no confiables, token passthrough. Marcarlas como restricciones de diseño; el gate formal es de `security`.
7. **Decidir qué amerita ADR** — aplicar el criterio de "Qué amerita ADR" y listar los candidatos con su razón en una línea.

## Puertas de decisión

### Agente vs workflow vs llamada simple

Criterio combinado: **complejidad** (¿los pasos son especificables de antemano?) + **valor** + **viabilidad del modelo** + **costo de un error** (¿recuperable?).

- **Llamada simple** — la tarea es un paso (extracción, clasificación, generación acotada). Default. La mayoría de "features de IA" son esto, a veces con retrieval.
- **Workflow determinista** — los pasos se conocen de antemano y se pueden encadenar (prompt chaining, routing, parallelization, orchestrator-workers). Preferible al agente siempre que el flujo sea especificable.
- **Agente** — solo cuando los pasos NO son especificables por adelantado, el valor lo justifica, el modelo es viable para decidir su propia trayectoria y el costo de error es recuperable.

Si cualquiera de las cuatro condiciones del agente falla → DETENER en workflow o llamada simple. "Necesitamos un agente" sin las cuatro condiciones es sobre-ingeniería.

### RAG vs fine-tuning vs prompting

- **Prompting simple** — el conocimiento cabe en el contexto o el modelo base ya lo tiene. Default. Si el corpus entra en la ventana de contexto, **no uses RAG**.
- **RAG** — el conocimiento es grande, cambia con frecuencia, o requiere trazabilidad a la fuente. Elegir cuando el corpus no cabe en contexto y la respuesta debe fundamentarse en documentos específicos.
- **Fine-tuning** — solo cuando se necesita un *comportamiento/forma/estilo* consistente que el prompting no logra, hay dataset suficiente y de calidad, y el costo de mantener el ciclo de entrenamiento se justifica. Rara vez es la primera respuesta; casi nunca sustituye a RAG para "saber datos frescos".

Regla de precedencia: prompting → RAG → fine-tuning. No saltar de largo a fine-tuning sin haber agotado las dos anteriores.

### Clase de proveedor (decisión arquitectónica)

Decidir la **clase**, no el modelo exacto:

- **API frontier** (Claude/Anthropic, OpenAI) — razonamiento abierto, agentes largos, máxima calidad; se acepta que los datos salen de la red y el costo por token es mayor. Default para tareas de alto razonamiento con datos no sensibles.
- **Hosted OpenAI-compatible** (Groq, Together, OpenRouter, vLLM gestionado) — buen punto medio de costo/latencia con superficie de API estándar; útil para volumen medio y para no acoplarse a un solo proveedor.
- **Local / open-weight** (Ollama, llama.cpp, vLLM propio) — cuando privacidad/compliance manda (PII/PHI/HIPAA, datos que no salen de la red), a alto volumen sostenido, o para latencia interactiva/edge/offline. Viable si un modelo 8-30B con structured output resuelve la tarea.

Trade-offs a explicitar en el ADR: **costo** (break-even de volumen: <50M tok/mes suele favorecer API; >100M tok/día suele favorecer local), **latencia**, **privacidad/compliance**, **calidad mínima requerida**. La respuesta madura suele ser **híbrida** (p. ej. embeddings/RAG interno local + razonamiento frontier). Para embeddings de RAG interno, local suele ser el default por calidad competitiva y costo cero por token.

Puerta de deslinde: si elegir la clase depende de si el modelo concreto soporta tool calling nativo, JSON schema garantizado o el context window real → **DETENER** y remitir esa verificación al checklist de 7 capacidades de `ai-engineering`. La arquitectura fija la clase y los trade-offs; la capacidad exacta se verifica en implementación.

### Contrato de dominio de un feature IA

- **Structured output = contrato** — si otro componente consume la salida, define su schema como contrato de dominio (campos, tipos, enums). La *garantía de forma* (nativa / constrained decoding / retry-with-validation) es decisión de implementación; a nivel diseño defines la forma esperada y que la validación client-side es obligatoria.
- **Límites de tools** — para agentes/workflows, define qué tools existen, qué efectos tienen (lectura vs escritura vs acción irreversible) y qué queda fuera del alcance del modelo. Un agente sin límites de tools explícitos no tiene contrato.
- **Fronteras de entrada/salida** — qué datos entran al sistema IA y qué sale; los datos no confiables se marcan en la frontera.

### Estrategia de evals a nivel producto

Decidir, no implementar:

- **Qué se evalúa** — la propiedad de calidad que define "está bien" para este feature (exactitud de extracción, fidelidad a la fuente en RAG, seguridad de la acción del agente).
- **Golden set** — que exista un conjunto de casos reales (incluyendo fallos) como criterio; su tamaño y construcción concreta es de implementación.
- **Tipo de grader** — programático (determinista, preferible cuando aplica), LLM-as-judge (para scoring abierto; el judge no debe ser de la misma familia que genera), o humano. La elección del *tipo* es de diseño.

Si el feature no admite un criterio evaluable → DETENER: no es diseñable como contrato. Resolver con el humano antes de cerrar.

### Consideraciones de seguridad de diseño

Enunciar como restricciones que condicionan la arquitectura (el gate formal es de `security`):

- **Prompt injection** — todo contenido externo y toda salida de tool es dato no confiable; el diseño debe separar datos de instrucciones.
- **Datos no confiables cruzando fronteras** — marcar qué entradas no son de confianza y qué componentes las procesan.
- **Token passthrough** — un diseño que reenvía tokens/credenciales a través del modelo o de tools es una violación; marcarlo como no negociable en el diseño.

## Qué decisiones de IA ameritan ADR

Amerita ADR cuando la decisión es costosa de revertir o condiciona a otros componentes:

- Elección de forma de solución cuando se opta por **workflow o agente** (una llamada simple rara vez amerita ADR propio).
- Estrategia de conocimiento cuando se elige **RAG o fine-tuning** (prompting simple no suele ameritarlo).
- **Clase de proveedor** cuando implica compromisos de costo/privacidad/latencia o acoplamiento a un vendor.
- **Contrato de dominio** de la salida IA cuando otros servicios lo consumen.
- **Estrategia de evals** cuando define el criterio de aceptación del feature.
- Decisiones de seguridad estructurales (aislamiento de datos no confiables, política de tools).

No ameritan ADR: elección de una técnica de prompting, ajuste de un few-shot, un valor de temperatura — son detalles de implementación de `ai-engineering`.

## Checklist Pre-Decisión

- [ ] Forma de solución elegida con las cuatro condiciones del agente verificadas (o justificado por qué basta workflow/llamada)
- [ ] Estrategia de conocimiento (prompting/RAG/fine-tuning) elegida por precedencia, sin saltar pasos
- [ ] Clase de proveedor decidida con trade-offs explícitos (costo, latencia, privacidad, calidad)
- [ ] Ninguna decisión arquitectónica se apoya en una capacidad de proveedor asumida (las capacidades se verifican en `ai-engineering`)
- [ ] Contrato de dominio definido: forma de salida + límites de tools + fronteras de datos
- [ ] Estrategia de evals definida: qué se evalúa + golden set + tipo de grader
- [ ] Consideraciones de seguridad de diseño enumeradas como restricciones
- [ ] Candidatos a ADR listados con su razón en una línea
- [ ] Sin duplicar contenido de implementación de `ai-engineering` — solo referencias

## Detección de Anti-Patrones

| Patrón en el diseño | Anti-Patrón | Severidad | Corrección |
|---|---|---|---|
| Diseño elige "agente" sin las cuatro condiciones | agent-without-justification | warning | Bajar a workflow o llamada simple hasta que las cuatro condiciones se cumplan |
| Fine-tuning propuesto sin agotar prompting ni RAG | premature-fine-tuning | warning | Aplicar precedencia prompting → RAG → fine-tuning |
| RAG con corpus que cabe en contexto | unnecessary-rag | warning | Usar prompting con el corpus en contexto |
| Clase de proveedor decidida sobre una capacidad asumida | assumed-capability-decision | error | Fijar la clase por trade-offs; verificar la capacidad en `ai-engineering` antes de codificar |
| Feature IA sin criterio de eval definido | no-eval-criterion | error | Definir qué se evalúa y el grader antes de cerrar el diseño |
| Salida IA consumida por otro componente sin schema de contrato | untyped-ai-contract | warning | Definir el structured output como contrato de dominio |
| Detalle de implementación (temperatura, few-shot) elevado a ADR | over-specified-adr | suggestion | Dejarlo a `ai-engineering`; no es decisión arquitectónica |
