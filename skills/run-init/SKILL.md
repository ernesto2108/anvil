---
name: run-init
description: Arranque del run de orquestación — verifica runs previos, captura snapshot git, carga Context Navigator, hace recall de memoria e inicia persistencia en Anvil MCP. Cargar al inicio de cada sesión de orquestación multi-agente. Cárgalo al inicio de cada run de orquestación, antes del primer sub-agente productivo. Reemplaza el Paso 0 inline del leader.md.
user-invocable: false
---

# run-init

Secuencia obligatoria que ejecuta el Líder al inicio de cada run, antes del primer sub-agente productivo. NO se salta — ni siquiera para "tareas triviales". Reemplaza la sección `## Paso 0 — Arranque` del `leader.md` con un único punto de carga.

## Cuándo se ejecuta

Al inicio de cada run de orquestación, cuando el humano invoca al Líder. Es el primer trabajo del Líder en cada turno conversacional que abra un run nuevo.

## Flujo (orden estricto)

### 0.1 — Verificar run previo

`mcp__anvil__load_orchestration(run_id="last")`.

- Estado `running`/`paused` con `pending_roles` → preguntar al usuario: "Hay un pipeline {status} ({run_id}) con N pendientes. ¿Retomamos o nuevo?"
  - Retomar → usar `run_id` existente + outputs previos inline
  - Nuevo → `complete_orchestration(run_id, "failed")` + continuar
- Estado `success`/`failed`/sin pendientes → continuar

### 0.2 — Snapshot git

`git status --short`. Si no vacío → capturar como **"Archivos ya modificados en esta sesión"** y pasarlo al developer cuando llegue su turno.

Complementar con `mcp__anvil__get_recent_changes(days=1)` para incluir contexto de runs de pipeline recientes que git no muestra (commits + runs cerrados en el último día). Si hay output relevante (cambios de hoy), incluirlo en el contexto del run actual bajo `## Cambios recientes` para inyectar inline al primer sub-agente.

### 0.3 — Cargar Context Navigator (vía `explorer`)

**Este chequeo es el primer paso operativo del run** — se ejecuta ANTES de spawnear cualquier sub-agente productivo. **El Líder NO lee `.context/` directamente** (ver Reglas inviolables #9 del `leader.md`): la carga del Context Navigator se delega SIEMPRE al `explorer`.

**Existencia de `.context/NAVIGATOR.md`** — el Líder no puede verificarla con `Read` (`.context/` está fuera de su whitelist). Para detectarla usa `Bash[ls .context/NAVIGATOR.md]` (cubierto por `Bash[ls *]`) o delega la verificación al `explorer` en el mismo spawn.

**Flujo:**

1. Spawnear `explorer` con prompt:

   ```
   ## Objetivo
   Lee `.context/NAVIGATOR.md`, `.context/project.md`, `.context/patterns.md` y los dominios relevantes a [objetivo del run]. Devuelve el contenido condensado más el valor de `last_updated`.

   ## Fuentes a consultar
   - `.context/` del proyecto (NAVIGATOR + project + patterns + dominios relevantes)
   - Memoria — `mcp__anvil__search_memories` con el objetivo del run
   - GitHub: rama de referencia `develop` (usar `skip_github: true` para omitir si no aplica)

   ## Restricciones
   Solo lectura — no modificar ningún archivo. No hacer checkout ni reset.

   ## Done-when
   Navigator cargado, memoria consultada y estado de GitHub disponibles.
   ```

   **Contrato estricto de respuesta (sin excepciones):** DEBES devolver exactamente uno de estos tres tokens — sin texto adicional, sin disculpas, sin markdown, sin envoltorios — cuando aplique la condición correspondiente:

   - `CONTEXT_MISSING` si:
     - `.context/NAVIGATOR.md` no existe o no se puede abrir (archivo no encontrado, permisos, path inválido).
     - Tuviste un error de tool, timeout, o cualquier otra falla que te impida garantizar que el contexto cargado es válido y completo.
   - `CONTEXT_STALE` si:
     - `.context/NAVIGATOR.md` existe pero está vacío, solo contiene whitespace, o tiene menos de 10 líneas de contenido real.
     - El contenido se devolvió truncado, corrupto, o no se pudo parsear de manera confiable.
     - El contenido no es un NAVIGATOR válido: faltan los campos obligatorios (`last_updated`, índice de dominios, referencias a `project.md`/`patterns.md`), o el header no identifica al archivo como Context Navigator.
   - `CONTEXT_INSUFFICIENT: <razón concreta>` si:
     - `.context/NAVIGATOR.md` existe y es válido pero no cubre el dominio/área investigado en este run. Incluir una razón concreta tras el token (ej: `CONTEXT_INSUFFICIENT: no hay info sobre orkestapay`).

   En cualquiera de esos casos: una sola línea, sin formato, sin más texto que el token (más la razón cuando aplique). Si dudas entre devolver contenido parcial o un token de parada, devuelve el token.
2. Recibir el output del `explorer`:
   - **Devolvió contenido válido + `last_updated`:** calcular días desde esa fecha.
     - `>3 días` → etiquetar "⚠️ puede estar stale" pero continuar.
     - `>7 días` → recomendar correr `scanner` antes (no auto-spawnear; gate al usuario).
     - Inyectar el contenido devuelto inline en el primer agente productivo bajo `## Contexto del sistema`. NO releer los archivos — el contenido ya está inline.
   - **Devolvió `CONTEXT_MISSING` (token exacto, único contenido material):** agregar `context-bootstrap` + `scanner` (modo deep) al inicio del pipeline (ver §Manejo de `CONTEXT_MISSING` en `leader.md`). Excepción solo si el usuario dijo "sin bootstrap".
   - **Devolvió `CONTEXT_STALE` (token exacto):** el `.context/` existe pero está desactualizado — agregar solo `scanner` (modo deep) al inicio del pipeline, sin `context-bootstrap`. Excepción solo si el usuario dijo "sin scanner".
   - **Devolvió `CONTEXT_INSUFFICIENT: <razón>`:** el `.context/` existe pero no cubre el dominio/área señalado — agregar `scanner` al inicio del pipeline con scope acotado al dominio o área que el explorer marcó como insuficiente (parsear la razón devuelta tras el token). Excepción solo si el usuario dijo "sin scanner".
   - **Devolvió cualquier otra cosa (texto ambiguo, error, output parcial, formato inesperado, token desconocido):** tratarlo como `CONTEXT_MISSING` (fallback seguro) y seguir la rama correspondiente. El Líder NO intenta interpretar outputs no conformes — un output que no es uno de los tres tokens válidos se trata como contexto faltante.

**Sin excepción de complejidad:** una tarea Small sin `.context/` igual arranca con `context-bootstrap` + `scanner`. El Líder nunca abre `.context/` por su cuenta para "ahorrar un spawn" — la regla #9 no admite atajos.

**Por qué este paso pasa por el `explorer` y no es opcional:** el chequeo sigue siendo el primer paso del run, pero ahora se ejecuta vía spawn (no vía `Read` directo). El `explorer` carga el contexto, el Líder lo recibe y lo inyecta hacia los siguientes sub-agentes. Sin este paso, los agentes posteriores pueden reportar `CONTEXT_MISSING` mid-run y forzar reintentos costosos.

### 0.4 — Recall de memoria

`mcp__anvil__search_memories(query=<descripción>, limit=3)`.

Hits con `score >= 0.5` → inyectar inline en primer agente bajo `## Memorias relevantes` + reportar 1 línea al usuario. Sin hits → continuar en silencio.

### 0.5 — Iniciar persistencia

1. `mcp__anvil__start_orchestration(objetivo, pipeline)` → obtener `run-id`
2. Escribir `.context/runs/<run-id>/plan.md` (formato en la skill `leader/output-formats`, sección `## plan.md del run`)
3. Escribir `.context/runs/<run-id>/log.md` con encabezado mínimo `# Log del run <run-id>` (archivo donde el Líder irá apendando entradas del run en orden cronológico)
4. `mcp__anvil__save_leader_log(run_id, content)` con plan inicial completo

## Output disponible para el Líder al terminar

Al completar los 5 sub-pasos, el Líder tiene listo:

- **`run_id` activo** — identificador devuelto por `start_orchestration`, requerido por todas las llamadas MCP posteriores (`save_step`, `save_leader_log`, `complete_orchestration`)
- **Contenido del Context Navigator** — texto condensado de `.context/NAVIGATOR.md`, `project.md`, `patterns.md` y dominios relevantes (vía `explorer`), o flag `CONTEXT_MISSING` que dispara `context-bootstrap` + `scanner`
- **Archivos ya modificados en la sesión** — output de `git status --short` capturado en 0.2, listo para inyectar al `developer` bajo `## Archivos ya modificados`
- **Memorias relevantes** — hits con `score >= 0.5` del recall de 0.4, listos para inyectar al primer sub-agente bajo `## Memorias relevantes` (vacío si no hubo hits)

Estos cuatro artefactos son los inputs base de cualquier sub-agente productivo que el Líder spawnee a continuación.

## Reglas

- Sin excepciones de complejidad — corre completo aun para tareas Small.
- El Líder NO lee `.context/` directamente en ningún sub-paso — la lectura siempre pasa por `explorer` (Regla inviolable #9 del `leader.md`).
- Si el `explorer` devuelve `CONTEXT_MISSING`, la secuencia obligatoria es `context-bootstrap` + `scanner` (deep) ANTES de re-invocar al `explorer` con los mismos inputs y continuar el pipeline original.
- El orden de los 5 sub-pasos NO se altera — 0.1 → 0.2 → 0.3 → 0.4 → 0.5.
- Sin este paso ejecutado completo, el gate de visibilidad del Líder (definido en `~/.claude/CLAUDE.md`) no puede iniciarse — no mostrar árbol de agentes ni spawnear nada hasta que `run-init` termine.
