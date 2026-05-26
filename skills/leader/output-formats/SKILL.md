---
name: leader-output-formats
description: Templates de output para el cierre de cada modo del Líder (Explorador, Planeación, Integración, Pruebas) y formatos de artefactos estructurados que produce el Líder al cerrar (vault `integration-summary.md`, `plan.md` del run). Cárgalo cuando el Líder esté por presentar el resultado al usuario al final de un modo, o cuando vaya a escribir al vault del proyecto, o cuando inicialice el `plan.md` del run.
user-invocable: false
---

# leader-output-formats

Templates exactos que el Líder debe seguir al cerrar cada modo y al producir sus dos artefactos estructurados (nota del vault de Integración, `plan.md` del run).

Cada template combina **en un solo bloque integrado**:
1. Header del modo completado
2. Árbol de agentes usados
3. Resumen ejecutivo en bullets
4. Hallazgos / archivos modificados / resultado según el modo
5. Próximos pasos

No hay dos bloques separados — el LLM debe escribir el output completo del modo siguiendo el template integrado correspondiente, de principio a fin.

---

## Reglas comunes a los 4 templates de modo

1. **Árbol de agentes:**
   - `┌─` para la raíz (`líder`), `├───` para hijos intermedios, `└───` para el último hijo.
   - 3 espacios después del conector antes del nombre del agente.
   - Si dos agentes corrieron en paralelo, ambos llevan `├───` y la anotación `(∥ con <otro>)` al final.
   - Agentes saltados con skip rule **no aparecen** en el árbol.
   - Agentes que fallaron y se re-invocaron aparecen **una sola vez** con el resultado final.

2. **Resumen ejecutivo:**
   - 3-5 bullets, en presente o pretérito, describiendo el comportamiento del sistema (no el código).
   - Correcto: "El query package ahora expone `GetRunsByProject` con filtro opcional por estado".
   - Incorrecto: "Se agregó una función al archivo `runs.go`".

3. **Archivos modificados:**
   - Paths absolutos cuando es posible. Si son relativos al repo, ser consistente en toda la lista.
   - Una línea por archivo con qué cambió. Sin diffs, sin snippets.
   - Si son más de 8, listar 8 + `(+N archivos menores)` y nombrar los menores al final.

4. **Próximos pasos:**
   - Si el modo encadena con otro → mencionarlo: `seguir a Integración con [TASK-ID]`.
   - Si el run cerró del todo → `ninguno — run cerrado`.
   - Si hay deuda o follow-ups → resumirlo en una línea.

5. **Lo que NO va en ningún template:**
   - Comandos bash, file reads, tool calls — eso vive en el log interno y en Anvil Dashboard.
   - Stack traces, errores de retry, iteraciones de self-critique.
   - Internal monologue de sub-agentes.
   - Outputs crudos de sub-agentes — el Líder los digiere.

---

## Explorador

```
✅ Explorador completó — [objetivo en una línea]

**Árbol de agentes invocados:**
┌─ líder
└─── explorer          → <qué produjo, una línea>

**Resumen ejecutivo:**
- [bullet 1 — máx 5 bullets]
- [bullet 2]

## Hallazgos
- [hallazgo 1 — viene del explorer]
- [hallazgo 2]

## Fuentes consultadas
- .project-context/domains/X.md (local)
- internal/foo/bar.go:123-150 (local)
- https://... (web) — accedido <fecha>

## Preguntas abiertas que quedaron sin responder
- [si las hay, si no, omitir la sección]

## Recomendación
[opcional — qué hacer con los hallazgos]

**Próximos pasos:** <una línea — ej. "seguir a Planeación con TASK-ID por definir" o "ninguno — run cerrado">

---
¿Continuamos a Planeación, o con esto es suficiente?
```

**Notas para Explorador:**
- Si el run no modificó archivos (caso típico de Explorador puro), no hay sección "Archivos modificados". Los hallazgos hacen las veces de output material.
- Si el fast-path resolvió la pregunta sin spawn de `explorer`, no se usa este template — es respuesta conversacional directa.

### Ejemplo compacto

```
✅ Explorador completó — Mapear cómo el frontend consume /v1/runs

**Árbol de agentes invocados:**
┌─ líder
└─── explorer          → 3 hallazgos del flujo HTTP cliente→server

**Resumen ejecutivo:**
- El frontend pollea `/v1/runs?status=running` cada 5s desde `dashboard.tsx`
- No hay manejo de error 5xx — la UI queda colgada en "Cargando..."
- El backend ya soporta el filtro pero el cliente nunca lo usa
[... resto del template ...]
```

---

## Planeación

```
✅ Planeación completó — [TASK-ID si existe]

**Árbol de agentes invocados:**
┌─ líder
├─── pm                → PRD con N criterios de aceptación
├─── designer          → Specs visuales (∥ con architect)
└─── architect         → SPEC + N ADRs (∥ con designer)

**Resumen ejecutivo:**
- [bullet 1 — qué cambió en términos del comportamiento del sistema]
- [bullet 2]
- [bullet 3 — máx 5 bullets]

## PRD — puntos clave
- [criterios de aceptación]
- [no-objetivos]

## Decisiones de arquitectura
- [decisión 1]
- [decisión 2]

## Riesgos identificados
- [si los hay, si no omitir]

## Archivos que se van a tocar (estimado)
- [lista]

**Archivos modificados:**
- `path/a.md` — <qué cambió>
- `path/b.md` — <qué cambió>

**Próximos pasos:** <ej. "seguir a Integración con [TASK-ID]" o "esperar confirmación del usuario">

---
¿Continuamos a Integración, o ajustamos el plan primero?
```

### Ejemplo compacto

```
✅ Planeación completó — TASK-042

**Árbol de agentes invocados:**
┌─ líder
├─── pm                → PRD con 4 criterios de aceptación
└─── architect         → SPEC + 1 ADR sobre estructura del template

**Resumen ejecutivo:**
- El feed de orquestación se renderiza con un template fijo en lugar de tool calls crudos
- Decisión: usar ASCII tree para compatibilidad con terminal monospace
[... resto del template ...]
```

---

## Integración

```
✅ Integración completó — [TASK-ID]

**Árbol de agentes invocados:**
┌─ líder
├─── developer         → <qué produjo>
├─── tester            → <qué produjo>
└─── reporter          → delta aplicado a .project-context/

**Resumen ejecutivo:**
- [bullet 1 — qué cambió en términos del comportamiento del sistema, no del código]
- [bullet 2]
- [bullet 3 — máx 5 bullets]

## Archivos modificados
- [lista con descripción de qué cambió, una línea por archivo]

## Validación
- Build: PASS
- Lint: 0 issues nuevos
- Tests: N passed / 0 failed
- Handoff verificado: ✅

## Nota escrita al vault
- [path absoluto del archivo creado]

**Próximos pasos:** <ej. "seguir a Pruebas" o "ninguno — run cerrado">

---
¿Continuamos a Pruebas, o revisas el código primero?
```

### Ejemplo compacto

```
✅ Integración completó — TASK-042

**Árbol de agentes invocados:**
┌─ líder
├─── developer         → handler + tests + handoff
├─── tester            → 6 tests nuevos, todos pasan
└─── reporter          → delta aplicado a domains/dashboard.md
[... resto del template ...]
```

---

## Pruebas

```
✅ Pruebas completó — [TASK-ID]

**Árbol de agentes invocados:**
┌─ líder
├─── tester            → <qué produjo>
├─── reviewer          → <hallazgos> (∥ con security)
├─── security          → <hallazgos> (∥ con reviewer)
└─── qa                → score X/10

**Resumen ejecutivo:**
- [bullet 1]
- [bullet 2 — máx 5 bullets]

## Resultado
- Tests: N passed / M failed
- Review: [limpio / N críticos / N mejoras] [si corrió Reviewer]
- QA score: X/10 [si corrió QA]
- Security: [limpio / hallazgos] [si corrió security]

## Issues encontrados
- [si los hay, con severidad]

## Estado final
[listo para merge / requiere fixes]

**Próximos pasos:** <ej. "volver a Integración para fixes" o "ninguno — run cerrado">

---
[Si hay issues] ¿Volvemos a Integración para los fixes, o los manejas directo?
[Si está limpio] ¿Cerramos el run?
```

### Ejemplo compacto

```
✅ Pruebas completó — TASK-042

**Árbol de agentes invocados:**
┌─ líder
├─── tester            → 6 tests nuevos
└─── reviewer          → 1 mejora menor, 0 críticos
[... resto del template ...]
```

---

## Vault — `integration-summary.md`

Formato no negociable de la nota que el Líder escribe al vault al cerrar Modo Integración (Regla inviolable #6).

```markdown
# <Título corto del cambio>

**Fecha:** <YYYY-MM-DD>
**Run ID:** <run-id de Anvil MCP>
**TASK-ID:** <TASK-ID si existe, si no "N/A">
**Modo:** Integración
**Estado:** <success | partial | failed>

## Qué se implementó

<2-4 líneas — describir el cambio en términos del comportamiento del sistema, no del código>

## Por qué (problema que resolvía)

<1-3 líneas — el síntoma o gap que motivó el cambio>

## Archivos clave tocados

- `<path>` — <qué cambió en una línea>
- `<path>` — <qué cambió en una línea>

## Validación

- Build: <PASS | FAIL>
- Lint: <0 issues nuevos | N issues>
- Tests: <N passed / M failed>
- Handoff verificado: <sí | no>

## Notas para el futuro

<si hay deuda, follow-ups, decisiones abiertas — si no, omitir>
```

**Resolución del path** (ver §Modo Integración → Cierre en `agents/leader.md` para reglas completas):
- Match `blt-*` → ir a Outline manualmente, no escribir local.
- Match a un proyecto del registry → `<vault>/tasks/<TASK-ID>/integration-summary.md`.
- Default → `<repo>/.workspace/03-tasks/<TASK-ID>/integration-summary.md`.

---

## `plan.md` del run

Scratchpad operativo del Líder durante el run. Vive en `.project-context/runs/<run-id>/plan.md`. Se inicializa en Paso 0.5 y se actualiza con `mcp__anvil__save_leader_log` después de cada sub-agente.

```markdown
# Plan — <run-id>

last_updated: <ISO-8601>
modo: <Explorador | Planeación | Integración | Pruebas>
budget: { max_retries: N, max_cost: $X }

## Objetivo
<una línea>

## Pipeline
[ ] paso 1 — sub-agente
[ ] paso 2 — sub-agente

## Asunciones
- <asunción>

## Memoria consultada
- <hit con score, o "ninguna relevante">

## Errores acumulados
<vacío al inicio>
```

**Notas:**
- `last_updated` en ISO-8601 (`2026-05-10T14:32:00Z`).
- Cada paso del pipeline se marca `[x]` cuando completa.
- "Errores acumulados" guarda firmas de error (categoría + substring) para detectar bucles de retry.
- Al cerrar el run en `success` → el directorio `.project-context/runs/<run-id>/` se limpia.
