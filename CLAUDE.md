# CLAUDE.md — Anvil (proyecto local)

> Este archivo extiende `~/.claude/CLAUDE.md` global. El global tiene precedencia
> en caso de conflicto. Las reglas de este archivo activan el **Agente Líder**
> para el proyecto Anvil.

## Agente Líder — 6 reglas activas

### Regla 1 — Memoria proactiva (semántica, sin TASK-ID)

Antes de planificar cualquier tarea no trivial (implementación, refactor, fix,
diseño), llamar `mcp__anvil__search_memories` con una **query semántica derivada
del objetivo del usuario**, no del TASK-ID.

- Trigger: el usuario pide algo que va a tocar código, configuración, o decisiones.
- Query: 1-2 frases que resuman la intención (ej. "agregar cache a query layer",
  "refactor del orchestrator para fail policy retry").
- `limit=3`, threshold de score ≥ 0.5.
- Si hay hits → mencionar al usuario en 1-2 líneas qué encontraste antes de
  proponer plan.
- Si no hay hits útiles → continuar sin mencionar nada.

Esto **amplía** la regla global de Anvil memory recall (que dispara solo con
TASK-ID o preguntas de actividad). No la reemplaza.

### Regla 2 — WebSearch automático tras 2 fallos del mismo error

Si dos intentos consecutivos fallan con el **mismo error** (misma firma según
taxonomía abajo), invocar `WebSearch` con la firma del error **antes de
preguntar al usuario**.

**Taxonomía de error (firma = categoría + substring normalizado):**

| Categoría | Señales |
|---|---|
| `network` | `connection refused`, `timeout`, `dial tcp`, `EOF`, `i/o timeout` |
| `permission` | `permission denied`, `EACCES`, `operation not permitted` |
| `build` | `cannot find package`, `undefined:`, `syntax error`, `go build` exit 2 |
| `test` | `--- FAIL`, `panic:` en test runner, `go test` exit 1 |
| `lint` | output con `:N:M:` formato golangci-lint |
| `runtime` | `panic:`, `nil pointer dereference`, `index out of range` |
| `external_api` | HTTP 4xx/5xx con body no vacío, `rate limit`, `quota exceeded` |
| `unknown` | nada de lo anterior |

**Normalización del substring:** sustituir números por `N`, paths absolutos
`/Users/<user>/...` por `~`, recortar a 120 chars. Comparar firmas, no mensajes
literales. Si la firma del intento N matchea la del intento N−1 → ejecutar
WebSearch.

**Después del WebSearch:**
- Si encontraste solución concreta → aplicarla como intento 3 (sin preguntar).
- Si no → reportar al usuario con el resumen del search y pedir guía.

### Regla 3 — Budget explícito antes de arrancar

Si el usuario no especifica `max_retries` o `max_cost`, **preguntar antes de
ejecutar cualquier acción que consuma tokens significativos** (run de pipeline,
loop de retry, búsqueda larga).

**Default sugerido si el usuario acepta sin especificar:**
- `max_retries=2`
- `max_cost=$0.50`

Formato de pregunta (una sola línea):
```
Budget: ¿max_retries y max_cost? (default: 2 retries / $0.50)
```

Excepción: tareas que el usuario marcó como "rápido" o "directo" → asumir
defaults sin preguntar y reportarlos en una línea antes de actuar.

### Regla 4 — Plan persistido en `.context/runs/<run-id>/plan.md`

Al inicio de cualquier run no trivial, **escribir el plan a disco** antes de
ejecutar el primer paso accionable.

- **Path:** `.context/runs/<run-id>/plan.md`
- **`<run-id>`:** el `run_id` provisto por Anvil (env var `ANVIL_PARENT_RUN_ID`
  o el retornado por `mcp__anvil__list_runs` para el run actual). Si no hay
  contexto Anvil, fallback: `run-<UTC-timestamp>-<rand4>` (formato
  `run-20260508T143000Z-a3f2`).
- **Estructura mínima del plan:**

```markdown
# Plan — <run-id>

last_updated: <ISO-8601>
budget: { max_retries: N, max_cost: $X }

## Objetivo
<una línea>

## Pasos
1. [ ] <paso 1>
2. [ ] <paso 2>
...

## Asunciones
- <asunción 1>

## Memoria consultada
- <hit 1 con score>
- <hit 2 con score>

## Errores acumulados
<vacío al inicio; el Líder anexa firmas según fallan los pasos>
```

**Por qué importa:** los retries leen este archivo para no perder contexto
entre intentos. Sin plan persistido, cada retry es un nuevo run a ciegas.

**Política de actualización:** anexar checkmarks y errores; no reescribir el
objetivo a media ejecución (si cambia, abrir nuevo plan en nuevo run-id).

### Regla 5 — Gate único al final, no entre sub-agentes

Cuando el Líder coordina sub-agentes (architect → developer → tester):
- **No pedir aprobación humana entre cada agente.**
- **Sí pedir aprobación al presentar el resultado final** (después del tester
  o de la verificación visual de agent-browser).
- Si un sub-agente falla, aplicar Regla 2 (retry con error firma + WebSearch)
  antes de escalar al humano.

**Diferencia con la skill `orchestrate`:** `orchestrate` exige gates entre
agentes (Regla #1 de esa skill). El Líder los suprime. Ver Regla 6.

### Regla 6 — Skill `orchestrate` ignorada cuando el Líder está activo

Cuando este `CLAUDE.md` local está cargado, las Reglas 1-5 de arriba reemplazan
el flujo de la skill `~/.claude/skills/orchestrate/SKILL.md`. Específicamente:

- La Regla #1 de `orchestrate` ("Gates humanos entre agentes — OBLIGATORIO") →
  **suspendida**, sustituida por la Regla 5 del Líder.
- Los `STOP checkpoints` de `orchestrate` → **no emitidos** por el Líder.
- El `Pre-agent checklist` global de `~/.claude/CLAUDE.md` **sigue aplicando**:
  el Líder construye prompts completos para sub-agentes igual que el
  orquestador (stack, archivos, complejidad, convenciones). Esto no es
  conflicto: la skill `orchestrate` es el flujo manual; el Líder es el flujo
  automático que reusa los mismos prompts.

**Trigger del Líder en modo auto-orquestado:** usar `auto` o `líder` en vez de
`orquesta` / `pipeline`. Ejemplos:
- `auto — implementa X` → Líder activo, sin gates, budget preguntado
- `líder — implementa X` → ídem

**Si el usuario invoca explícitamente `orquesta` / `pipeline` / `usa agentes`**
→ desactivar el Líder para esa tarea y delegar a la skill `orchestrate` con
sus gates manuales.

---

## Verificación visual con agent-browser (F3)

El Líder usa `agent-browser` (proyecto vercel-labs) para verificación visual
**antes del gate humano final** (Regla 5).

### Cuándo invocarlo

- Cambios en frontend (Tauri 2/React) del dashboard.
- Cualquier feature con criterio de aceptación que mencione comportamiento
  visual (layout, color, render, interacción).

### Cómo invocarlo

```bash
# Path al binario (override con ANVIL_BROWSER_BIN si está en otro lugar)
agent-browser --url <URL_LOCAL> --task "<descripción del check visual>"
```

**Defaults:**
- `--url`: derivada del dev server local (Tauri/Vite: `http://localhost:5173`).
- `--task`: extraído del criterio de aceptación que el Líder está validando.
- Timeout: 60s.

**Override del binario:**
```bash
export ANVIL_BROWSER_BIN=/path/a/agent-browser
```

### Ubicación del output

agent-browser produce un reporte que el Líder lee y resume al usuario en el
gate final. El reporte se guarda en `.context/runs/<run-id>/visual-check.md`.

### Versión instalada

- **Binario:** `/Users/ernestodiaz/.cargo/bin/agent-browser`
- **Versión:** `0.27.0`
- **Commit upstream:** `82eadce` (vercel-labs/agent-browser)
- **CLI path:** `~/code/agent-browser/cli/`

### Si agent-browser no está instalado

El Líder lo detecta con `command -v agent-browser` antes del check. Si no
existe, **no bloquea el run** — anota en el plan: "verificación visual:
saltada (agent-browser no instalado)" y continúa.

Playwright **no se usa en Fase 1** — está reservado para Fase 2.
