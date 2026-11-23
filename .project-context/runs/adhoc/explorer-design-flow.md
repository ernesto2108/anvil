# Exploración — design-flow

**Run:** adhoc
**Fecha:** 2026-09-09

## Objetivo
Diagnosticar dónde está (o dónde falta) la cobertura de flujos/navegación en el pipeline de diseño (design-spec → construcción visual en Pencil), que hoy produce pantallas aisladas sin pantallas destino.

## Fuentes consultadas

| Fuente | Tipo | Origen |
|---|---|---|
| Agente designer-spec | local | /Users/ernestodiaz/projects/anvil/agents/designer-spec.md |
| Agente designer-visual | local | /Users/ernestodiaz/projects/anvil/agents/designer-visual.md |
| Skill design-recipes | local | /Users/ernestodiaz/projects/anvil/skills/design-recipes/SKILL.md |
| Referencia Pencil de recipes | local | /Users/ernestodiaz/projects/anvil/skills/design-recipes/reference/pencil.md |
| Skill design-system | local | /Users/ernestodiaz/projects/anvil/skills/design-system/SKILL.md |
| Workflow Pencil de design-system | local | /Users/ernestodiaz/projects/anvil/skills/design-system/reference/pencil-workflow.md |
| Skill design-review | local | /Users/ernestodiaz/projects/anvil/skills/design-review/SKILL.md |
| Skill design-project | local | /Users/ernestodiaz/projects/anvil/skills/design-project/SKILL.md |
| Skill design-to-code | local | /Users/ernestodiaz/projects/anvil/skills/design-to-code/SKILL.md |
| Skill visual-fidelity-qa (grep dirigido) | local | /Users/ernestodiaz/projects/anvil/skills/visual-fidelity-qa/SKILL.md |
| Agente pm (reglas #3/#4) | local | /Users/ernestodiaz/projects/anvil/agents/pm.md |
| Skill prd-template (estructura) | local | /Users/ernestodiaz/projects/anvil/skills/prd-template/SKILL.md |
| NAVIGATOR del proyecto | local | /Users/ernestodiaz/projects/anvil/.project-context/NAVIGATOR.md |
| git log + PRs abiertos | github | rama origin/develop (sin PRs abiertos relevantes) |

## Hallazgos clave

### Mapa del pipeline actual
1. `pm` → PRD (skill `prd-template`). Regla #3 "Cada CTA necesita un destino" (pm.md:44-47), pero el template de PRD NO tiene sección de user flows/journeys/mapa de pantallas (prd-template/SKILL.md:75-147: secciones 1-10, ninguna de flujos).
2. `designer-spec` → `design-spec.md` + `DESIGN.md`. AQUÍ vive casi toda la cobertura de flujo: Paso 2.5 "Validación del Inventario de Pantallas" con auditoría de navegación (designer-spec.md:202-228, punto 1 en :206: "Cada botón, enlace o CTA → ¿tiene una pantalla de destino diseñada?"), sección 3.2 Flujos de Usuario con Mermaid (:263-267), 3.3 Arquitectura de Información con "estructura de navegación" (:269-271), y en DESIGN.md las secciones `User Flow` (Mermaid) y `Screen Inventory` (:321-326) e `Interaction Flows` (:340-344).
3. `designer-visual` → construye el `.pen`. Ejecuta "exactamente lo que el Design Spec ya especificó" (designer-visual.md:21) y prohíbe inventar spec (:142). NO tiene auditoría de navegación propia; su Paso 4 (:108-123) construye variables→componentes→pantallas→estados sin verificar CTA→frame destino. Su output de cierre legitima dejar pantallas sin construir "por presupuesto" (:165).
4. `design-review` (opcional, post-construcción): único check de flujo es un checkbox en Completitud: "Cada CTA tiene una pantalla de destino" (design-review/SKILL.md:68) — sin procedimiento de verificación cross-frame ni carácter bloqueante.
5. `design-to-code` y `visual-fidelity-qa`: operan frame a frame (design-to-code Paso 2 inventaría secciones/estados/tokens de UN frame, SKILL.md:40-53; visual-fidelity-qa compara por `frame_id`, SKILL.md:30-35). Sin noción de flujo.

### Causas del comportamiento reportado (pantalla aislada, sin destinos)
- El único gate de flujo (designer-spec Paso 2.5) solo corre cuando el pipeline entra por `designer-spec`. La ruta directa/conversacional ("diseña un login") entra por `design-project` → `design-system`/`design-recipes`, donde NO existe ninguna noción de navegación: design-recipes Paso 1 dice solo "Identifica qué tipos de pantalla vas a construir" (design-recipes/SKILL.md:41-44) y sus recetas son de pantallas individuales; design-system va variables→componentes→pantallas sin inventario (design-system/SKILL.md:266-270) y su tabla de anti-patrones (:305-332) no incluye "CTA sin destino".
- `designer-visual` construye solo lo especificado y no puede inventar (designer-visual.md:21, :142); si el spec no llega o no cubre destinos, no se generan.
- Presupuestos de eficiencia refuerzan el corte: "Máximo 25 ops por lote" (design-recipes/SKILL.md:318), pendientes "por presupuesto" (designer-visual.md:165).
- No hay representación de flujo en el canvas: cero menciones de conectores/prototipado/flechas en todos los agents/skills de diseño (grep sin hits); las tools Pencil MCP declaradas no incluyen conectores (designer-visual.md:6); la organización del canvas es cronológica por iteración, no por flujo (pencil-workflow.md:86-96; design-project/SKILL.md:95-101).

### Gaps rankeados
1. **[Crítico] Ruta directa sin gate de flujo** — design-system, design-recipes y design-project no exigen inventario de navegación antes/durante la construcción.
2. **[Crítico] designer-visual sin auditoría de navegación** — no verifica CTA→frame destino ni reporta huérfanos; puede omitir pantallas por presupuesto sin gate.
3. **[Alto] Sin convención de representación de flujo en el canvas .pen** — ni numeración por flujo, ni agrupación por journey, ni conectores/flechas.
4. **[Alto] design-review con checkbox sin procedimiento** — "Cada CTA tiene una pantalla de destino" (línea 68) no dice cómo enumerar CTAs por frame ni cruzar contra frames existentes; no es bloqueante.
5. **[Medio] PRD sin sección de flujos** — la regla #3 del pm (pm.md:44-47) no tiene sección estructurada donde aterrizar en prd-template.
6. **[Bajo] visual-fidelity-qa por frame** — coherente con su alcance, pero nadie valida flujo aguas abajo.

### Puntos de inserción recomendados (para agent-designer)
1. `skills/design-recipes/SKILL.md` — nuevo gate previo al Paso 1: derivar inventario de pantallas desde los elementos interactivos de la pantalla pedida (CTA→destino: diseñar, o inventariar como pendiente explícito).
2. `agents/designer-visual.md` — Paso 4: sub-paso final de auditoría de navegación (cada CTA del .pen → frame destino existe o va a lista "CTAs huérfanos"); Output de cierre: sección obligatoria de huérfanos/pantallas destino no construidas.
3. `skills/design-system/SKILL.md` — Paso 6 (Ensamblar Pantallas): gate de inventario de navegación; agregar fila a Detección de Anti-Patrones ("pantalla con CTAs sin destino diseñado ni inventariado" → error).
4. `skills/design-review/SKILL.md` — Paso 2 Completitud: convertir el checkbox de línea 68 en procedimiento (enumerar elementos interactivos por frame vía batch_get → cruzar contra frames top-level → reportar huérfanos) con severidad alta.
5. `skills/design-system/reference/pencil-workflow.md` §Organización del Canvas — convención de flujo: prefijo de nombre de frame por journey + orden (ej. "Auth / 1-Login"), agrupación por filas de flujo. Verificar con `get_guidelines` si Pencil soporta conectores antes de prescribir flechas.
6. `skills/prd-template/SKILL.md` — sección nueva en el template (User journeys / mapa de pantallas) que materialice la regla #3 del pm.
7. `agents/designer-spec.md` — refuerzo menor: exigir que el "Plan de ejecución Pencil" (Paso 3, punto 5, :253) liste TODAS las pantallas del Screen Inventory en orden de flujo, para que designer-visual no pueda recortar sin dejarlo explícito.

## Gaps y preguntas abiertas
- No se pudo consultar memoria (`mcp__anvil__search_memories` no disponible en esta sesión).
- Capacidad real de Pencil MCP sobre conectores/prototyping no verificable desde los archivos del repo — las skills no la documentan; confirmar con `get_editor_state(include_schema:true)`/`get_guidelines` antes de diseñar la convención de representación de flujo.
- `run-id`/`topic` no provistos con formato estándar completo — se usó `adhoc`/`design-flow` según lo pedido por el prompt.
