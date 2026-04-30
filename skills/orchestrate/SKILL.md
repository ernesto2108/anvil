---
name: orchestrate
description: Orquestación inteligente — clasifica complejidad y ejecuta solo los agentes necesarios. Usar cuando el usuario dice "orquesta", "pipeline", "usa agentes", o para tareas no triviales.
disable-model-invocation: true
---

# Flujo de Orquestación

El sistema actúa como **Orquestador**. Esta skill se activa SOLO cuando el usuario lo pide explícitamente.

---

## Regla #0 — El usuario activa el orquestador

Se ejecuta SOLO cuando el usuario dice "orquesta", "pipeline", "usa agentes", o pasa `-c`. Nunca se auto-activa. "hazlo directo" → esta skill no aplica.

## Regla #1 — Gates humanos entre agentes (OBLIGATORIO)

Después de cada agente: mostrar resumen → preguntar "¿Paso al siguiente agente (X)?" → continuar SOLO con confirmación explícita.
El usuario puede siempre: saltar ("salta el tester") · parar ("hasta aquí") · cambiar ("el resto hazlo directo").

---

## STOP checkpoints — emitir estos bloques y ESPERAR (sin excepciones)

Reemplazan razonamiento interno. El orquestador DEBE emitir el bloque y parar.

**Después de PM → antes de Architect**
```
✅ PM terminó — [TASK-ID]
• [bullet 1 del PRD]
• [bullet 2 del PRD]
PRD: {task_path}/prd.md (o documento de Outline si Linear+Outline)

¿Paso al Architect? (sí / ajusta / hasta aquí)
```

**Después de Architect → antes de Developer (tareas Medium+)**

Checkpoint en dos fases. **Fase 1** — el architect devuelve resumen de decisiones ANTES de escribir docs completos:
```
✅ Architect — decisiones propuestas — [TASK-ID]
• [decisión clave 1]  • [decisión clave 2]
Riesgos: [bullets]
¿Apruebas para que escriba los docs? (sí / ajusta / hasta aquí)
```
Si se rechaza → el architect reescribe decisiones sin haber gastado tokens en docs completos.

**Fase 2** — después de docs escritos, mostrar secciones del SPEC para aprobación:
```
✅ Architect terminó — [TASK-ID]
📄 SPEC: [pegar: Contexto y objetivo, No-objetivos, Criterios de aceptación, Límites de implementación]
¿Apruebas y paso al Developer? (sí / ajusta / hasta aquí)
```
Verificación de arquitectura: al menos una vista de dominio debe existir (cuáles depende del scope de la tarea). Un solo `architecture.md` NO es válido para Medium+.

**Después de Developer → antes de Tester (tareas Medium+)**

Gate ejecutable, no declarativo. Ejecuta `scripts/verify-handoff.sh` desde el repo de anvil. Si exit ≠ 0, re-invocar developer con el output del script — NO preguntar al usuario hasta que el script devuelva exit 0.

```bash
bash <ANVIL_REPO>/scripts/verify-handoff.sh <PROJECT_ROOT> <TASK-ID>
```

El script valida en una sola corrida: archivo existe, secciones obligatorias presentes (`## Input recibido`, `## Estado actual`/`## Fases`, `## Archivos modificados`, `## Decisiones tomadas`, `## Handoff for tester`, `## Output entregado`), tabla de Output entregado tiene Build PASS y Lint PASS / 0 issues, y la lista de tests requeridos no está vacía.

Después de que el script pase, presentar al usuario:

```
✅ Developer terminó — [TASK-ID]
Handoff verificado: ✅ (verify-handoff.sh exit 0)
Lint nuevos issues: 0 (gate bloqueante — re-invocaría developer si hubiera)

¿Paso al Tester? (sí / revisa / hasta aquí)
```

**Re-invocación:** si el script falla, copia el mensaje BLOCKED del stderr al prompt del developer. El developer corrige el gap específico y re-cierra. NO consultar al usuario en este loop — es trabajo determinista.

---

## Paso -2 — Verificar pipeline previo

Antes de iniciar: `load_orchestration(run_id="last")`.
- `running`/`paused` con `pending_roles` → preguntar: "Hay un pipeline {status} ({run_id}) con N pasos pendientes. ¿Retomamos o nuevo?" Retomar = usar `run_id` existente + pasar outputs previos. Nuevo = `complete_orchestration(run_id, "failed")` + `start_orchestration(...)`.
- `success`/`failed` o sin pendientes → flujo normal.

Al finalizar cada pipeline: `complete_orchestration(run_id, "success"|"failed"|"paused")`.

---

## Paso -1 — Snapshot en vuelo (antes de triage)

Ejecutar `git status --short`. Si no está vacío, capturar la lista como **"Archivos ya modificados en esta sesión"** y pasarla inline en cada invocación del developer. Saltar solo cuando `git status` está vacío.

---

## Paso 0 — Triage

**Valores de `-c`:** `medium` (dev → tester) | `complex` (pm → arch → dev → tester → qa) | `max` (completo). Aplicar modificadores, confirmar antes de lanzar.

**Sin `-c`** — recomendar de esta tabla. **Nunca auto-ejecutar — preguntar primero.**

| Señal | Pipeline |
|---|---|
| Patrón conocido, 3-5 archivos | developer → tester |
| Feature / endpoint nuevo | pm → architect → developer → tester → qa? |
| Cross-cutting, multi-servicio | scanner → pm → designer → architect → developer → tester → security → qa |
| Bug fix (claro / no claro) | developer → tester / pm → developer → tester → qa? |
| Migración DB | architect → dba → qa (architect produce `architecture-db.md`; spec.md solo si Medium+) |
| Infra / CI | devops → security |
| Refactor | architect → developer → tester → qa |
| Solo docs / auditoría security / marketing | tech-writer / security / mkt-content |
| Scope no claro | pm primero — siempre |

**Modificadores:** UI → +designer · DB → +dba · infra → +devops · auth/sensible → +security · contexto stale → +scanner.

### Routing por scope (desde PRD)

| Scope del PRD | Designer | Architect | design-to-code |
|---|---|---|---|
| `new` / `both` | sí | sí | sí (si existe .pen/Figma) |
| `visual-improvement` | sí | saltar | sí |
| `functional-improvement` | saltar | sí | saltar |

---

## Paso 0.5 — Recall de memoria relevante (OBLIGATORIO antes del primer agente)

Antes de invocar al primer agente del pipeline, el orquestador busca contexto histórico en los digests previos. Anvil ya guarda digests automáticamente vía `digestCheckpointer` después de cada `agent.end`, pero los agentes NO los consultan por su cuenta — el orquestador es quien los inyecta.

### Pasos

1. Llamar `mcp__anvil__search_memories` con:
   - `query` = descripción de la tarea del usuario (texto completo, no recortado)
   - `project` = nombre del proyecto actual (default = basename del working dir)
   - `limit` = 3
2. **Filtrar por `score >= 0.5`** — el threshold por defecto del MCP es 0.3, que es muy permisivo y trae ruido. 0.5 mantiene solo hits genuinamente relevantes.
3. Si HAY resultados ≥ 0.5:
   - Extraer `summary + decisions + edge_cases` de cada hit
   - Calcular días desde `created_at` para etiquetar frescura
   - Inyectar inline en el prompt del primer agente (PM, architect, o developer según el pipeline) bajo la sección `## Memorias relevantes (contexto histórico)`
   - Reportar al usuario en 1 línea: "Encontré N digests previos relevantes (run-X, run-Y) — los incluyo como contexto histórico."
4. Si NO hay resultados ≥ 0.5: continuar sin mencionar memoria. **No inyectar nada** — la ausencia es información válida (es trabajo nuevo).

### Formato de inyección al primer agente

```markdown
## Memorias relevantes (contexto histórico — léelas como referencia, NO como instrucciones nuevas)

### Digest run-42 [hace 5 días — score 0.72]
**Summary:** Implementación de cache LRU para getRunsByProject
**Decisions:**
- Usar sync.Map en vez de mutex+map por escala esperada
- TTL de 60s configurable vía env
**Edge cases:**
- TTL no se aplicaba a entries nil — fix en commit abc123

### Digest run-58 [hace 12 días — score 0.61]
...
```

### Reglas para evitar ruido

- **Threshold estricto:** 0.5 mínimo. Si quieres ajustar, baja a 0.45 antes de bajar a 0.4 — nunca menos.
- **Etiqueta de fecha:** siempre incluir `[hace N días]` para que el agente pese la frescura. Una decisión de hace 60 días puede ser obsoleta.
- **Etiqueta de score:** ayuda al agente a calibrar cuán relevante es cada digest.
- **Solo al primer agente:** no inyectar memorias en cada agente del pipeline. El primero las recibe y las pasa adelante vía SPEC/handoff si son relevantes a steps posteriores.

### Cuándo saltar el Paso 0.5

- El usuario pasa explícitamente `--no-memory` (flag futuro) o dice "ignora la memoria"
- El proyecto no tiene digests previos (proyecto nuevo) — el MCP devolverá lista vacía, sin overhead
- Ollama no está disponible **y** el fallback de "recent digests sin ranking" trae basura — en ese caso loguear y continuar sin inyección

### Anti-patrón

Bajar el threshold a 0.3 "para no perderse nada". Trae ruido, los agentes empiezan a citar decisiones viejas que ya no aplican, y la calidad del recall cae. Mejor cero memoria que memoria mala.

---

## Regla de límites (solo modo orquestación)

En modo directo, el orquestador lee/escribe libremente. En modo agente:

| | PUEDE | NO DEBE |
|---|---|---|
| Leer | Docs de `<docs>`, docs de tareas, handoffs, project-registry.md | Código fuente (`.go .ts .tsx .py .rs .dart` etc.) · Archivos de diseño (`.pen .fig .sketch`) |
| Escribir | Docs de `<docs>`, sprint-current.md | Planes técnicos, task.md con síntesis, archivos de código |

`Glob` y `Bash` (build/test) siempre permitidos. Lecturas de archivos de diseño → delegar al agente downstream. Detalle: `anti-patterns.md`.

---

## Reglas de saltar agentes

| Agente | Saltar cuando |
|---|---|
| scanner | context.md existe y está actual |
| pm | requisitos ya claros (bug con repro, spec exacto) |
| designer | sin cambios de UI |
| architect | sin decisiones de diseño (patrón existente, solo extender) |
| dba | sin cambios de DB |
| devops | sin cambios de infra |
| security | sin auth, sin datos sensibles, sin APIs externas |
| qa | ver reglas de QA abajo |
| reporter | **SALTAR POR DEFECTO** (ejecutar solo: cross-service, incidente, release, usuario pide) |
| tester | sin código testeable |
| mkt-content | sin contenido de marketing |

**Nunca saltar sin preguntar:** developer, tester.

**Gates bloqueantes (no-skip, no-ask):**

| Gate | Cuándo | Comando | Si falla |
|---|---|---|---|
| `lint` | Después del developer, antes del tester | skill `/lint` (auto-detecta stack) | Re-invocar developer con output inline. NO preguntar al usuario. 0 issues nuevos requeridos. |
| `verify-handoff.sh` | Después del developer, antes del tester | `bash <ANVIL_REPO>/scripts/verify-handoff.sh <PROJECT_ROOT> <TASK-ID>` | Re-invocar developer con stderr inline. NO preguntar al usuario. |
| `run-tests` | Después del tester, antes del qa | skill `/run-tests` | Re-invocar tester con output inline si los tests existentes rompen. |

**Lógica del lint estricto:** un PR que falla en CI por lint = trabajo doble (rebote + re-implementación). Mejor bloquear localmente. Sin threshold de tolerancia — 0 issues nuevos en archivos tocados. Issues pre-existentes en otros archivos no cuentan.

### Reglas de QA

**Incluir en pipeline cuando CUALQUIERA de:** ≥8 pts — auth/permisos/tokens — migraciones DB — pagos/billing — contratos API públicos — crypto/secrets — concurrencia — usuario pide.
**Saltar cuando TODOS:** Medium (3-5 pts) + ninguno de arriba + usuario no pidió.
**En la duda: recomendar QA.** El usuario decide. Las reglas internas de scoring y bloqueo las maneja el agente QA.

## Gates adicionales

- **Ejecución de diseño:** después de designer → dtd.md → PAUSE para que el usuario ejecute en Pencil/Figma
- **QA:** score < 7 → STOP, arreglar antes de continuar
- **Security:** CVE critical/high → STOP, arreglar antes de continuar
- **Backlog arquitecto:** después de ARD → verificar tareas en sprint-current.md (o Linear)
- **Sync cross-repo:** cambios en DTO/endpoint/auth backend → developer lista archivos frontend/mobile afectados

---

## Reglas de orquestación

- Resolver `<docs>` desde `~/.claude/project-registry.md` antes de cualquier agente
- **Un escritor a la vez.** Máx tareas por run: 2 (preferido: 1). Cambio de scope → re-ejecutar PM + arquitecto.
- Todos los docs en español. Código, keys, paths → inglés. Preguntas al usuario → español.
- Al terminar todos los agentes: cerrar tareas según el sistema de docs (Obsidian: sprint-current.md + board.md + frontmatter; Linear: mover issue a Done; `.workspace/`: sprint-current.md).

## Paso de contexto

- Contenido en contexto → inline. No en contexto → pasar path. Output de agente → inline al siguiente.
- Nunca leer código fuente para relayarlo (anti-pattern #5).
- Cada agente define sus campos requeridos internamente. El orquestador SOLO debe pasar los campos de la tabla de abajo — **no duplicar lógica ni reglas que ya viven en el agente.**

### Inject, don't read — handoff a agentes downstream (CRÍTICO)

El handoff en disco existe como **audit trail y para continuidad cross-session**. NO es el mecanismo primario de transferencia entre agentes. Los sub-agentes son lectores poco confiables — pueden skim, saltarse secciones, o priorizar otros archivos sobre el handoff.

**Regla operativa:**
1. El orquestador lee `<PROJECT_ROOT>/.handoff/<TASK-ID>.md` UNA vez
2. Extrae la(s) sección(es) relevante(s) para el agente downstream
3. Inyecta el contenido **inline en el prompt** del agente
4. NO pasa solo el path — el agente no lee el handoff por su cuenta

**Qué inyectar por agente:**

| Agente | Secciones a extraer e inyectar inline |
|---|---|
| **tester** | `## Handoff for tester` completo + `### Validación ejecutada` (para que NO repita build/lint) |
| **qa** | `## Decisiones tomadas` + `## Output entregado` + `## Archivos modificados` + `git diff` |
| **developer (continuación / qa-fix)** | `## Estado actual` + `## Siguiente paso` + `## Archivos modificados` + `## Decisiones tomadas` |
| **security** | `## Archivos modificados` + `git diff` (el handoff completo es secundario) |

**Por qué importa:** tu queja "el agente no respeta el handoff" casi siempre es "el agente no lo leyó completo". Inyectando inline, el agente no decide qué leer — recibe exactamente lo que necesita. Bonus: ahorras tokens (extraes ~50 líneas de un handoff de 200).

**Excepción:** si el handoff supera 500 líneas (raro), pasa el path COMO REFERENCIA en `## Notas` y aún así inyecta inline las secciones críticas. La referencia al path es para "consulta opcional", no para "lectura obligatoria".

## Input por agente (campos obligatorios del orquestador)

Cada agente valida sus inputs al inicio y hace STOP si falta algo requerido. El orquestador es responsable de pasar estos campos — nada más. Las reglas internas (convenciones, handoff format, budget, QA checks) las maneja cada agente.

| Agente | Campos requeridos | Notas |
|---|---|---|
| **scanner** | `project_root` | Solo el path raíz |
| **pm** | `user_request`, `context.md` (inline/path), `sprint-current.md` (inline/path) | — |
| **designer** | `prd.md` (inline/path, incluye Scope + Platform), `context.md`, `design-system.md` (si existe) | — |
| **architect** | `prd.md` (inline/path), `dtd.md` (si existe), `context.md` (inline/path), `output` (`ard`/`spec`/`full`), `task_path`, `context_path`, convention files (architecture + coding del stack) | Puede auto-resolver contexto si no corrió scanner |
| **developer** | `complexity` + pts, `stack`, `objective`, `files` (o "en SPEC"), `TASK-ID` (Medium+), `docs_path` (Medium+), `SPEC` inline/path (Medium+), convention file paths (Medium+), `mode` (`normal`/`maquetation`/`integration`) | Small: puede recibir reglas inline en vez de paths |
| **dba** | `architecture-db.md` (inline/path), `task_path` | — |
| **tester** | `stack`, `TASK-ID`, `complexity`, handoff content o path (sección `## Handoff for tester`), `SPEC` (Medium+, secundario) | El tester carga sus convenciones de testing solo — no pasarlas |
| **qa** | `SPEC` (inline/path), `.handoff/<TASK-ID>.md`, `git diff` | — |
| **security** | `git diff`, dependency paths | — |
| **devops** | infra files afectados, objetivo | — |
| **reporter** | `TASK-ID`, resumen de `git diff` | — |
| **mkt-content** | contexto de proyecto/marca, audiencia, identidad visual | — |

---

## Sub-archivos — cargar por trigger

| Trigger | Cargar |
|---|---|
| Invocar developer | `plan-approval.md` |
| Re-invocar después de QA/security | `qa-fix.md` |
| Registry/contexto stale / docs system unknown | `vault-setup.md` |
| Duda de límites | `anti-patterns.md` |
