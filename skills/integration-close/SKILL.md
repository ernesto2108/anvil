---
name: integration-close
description: Cierre completo del modo Integración del Líder — resuelve vault path desde project-registry, escribe nota al vault según tipo de cambio, spawnea reporter para aplicar delta a `.context/`, ejecuta `/task-complete`, cierra la orquestación en Anvil MCP, actualiza `last_updated` en NAVIGATOR, llama `digest_from_handoff` cuando no corrió reporter, y limpia `.context/runs/`. Cárgalo cuando el Líder esté por cerrar Modo Integración. Reemplaza las secciones de cierre de Integración y de Persistencia de runs del leader.md.
user-invocable: false
---

# integration-close

Secuencia de cierre del Modo Integración que el Líder ejecuta antes de presentar el output final al usuario. Cubre las dos responsabilidades acopladas: (a) escritura al vault del proyecto (Regla inviolable #6 del `leader.md`) y (b) persistencia/limpieza del run (Anvil MCP + `.context/runs/`).

El orden es **obligatorio** — saltarse o reordenar pasos rompe la trazabilidad del run y la consistencia de `.context/`.

## Cuándo se ejecuta

Inmediatamente antes del output final al usuario en Modo Integración, después de que todos los gates internos (`lint`, `verify-handoff.sh`, `run-tests`) hayan pasado y los sub-agentes productivos (`developer`, `tester`) hayan cerrado.

## Flujo (orden estricto)

### 1 — Resolver vault path desde `project-registry.md`

1. Leer `~/.claude/project-registry.md`
2. Identificar el directorio raíz del proyecto activo
3. Aplicar routing rules del registry **en orden** — primer match gana
4. Obtener path absoluto del vault (ej: `~/projects/notes/02-projects/anvil/`)
5. **Si matchea `blt-*`** → la doc va a Outline vía HTTP, no al vault local. Saltar el paso 2 y dejar nota en el output final: "Proyecto Boletia — la nota debe ir a Outline manualmente o vía pipeline aparte."
6. **Si cae al `default` (`.workspace/`)** → escribir en `<repo>/.workspace/03-tasks/<TASK-ID>/integration-summary.md`

### 2 — Escribir nota al vault

Decidir el destino dentro del vault según el tipo de cambio:

| Tipo de cambio | Destino |
|---|---|
| Implementación nueva con TASK-ID | `tasks/<TASK-ID>/integration-summary.md` (crear el directorio si no existe) |
| Decisión arquitectónica explícita o nuevo subsistema | `decisions/<NNN>-<slug>.md` (numerar tras el último ADR) |
| Bug fix sin TASK-ID | apéndice al final de `context.md` bajo `## Cambios recientes` con fecha |
| Fix urgente sin TASK-ID con impacto cross-domain | nuevo `decisions/` + apéndice en `context.md` |

**Formato de la nota:** cargar la skill `leader/output-formats` sección `## Vault — integration-summary.md` para el formato no negociable. Sin la nota escrita, el cierre del modo Integración no es válido (Regla inviolable #6).

**Si el vault no es accesible:**

- Path no existe → crearlo (`mkdir -p` al directorio padre, luego `Write`)
- Permiso denegado → escalar con formato de §Protocolo de debate del `leader.md`
- `project-registry.md` no existe → escalar: "No encontré `~/.claude/project-registry.md`. ¿Dónde escribo el resumen?"

### 3 — Spawnear `reporter` con archivos modificados

Después de escribir al vault y antes del output final, spawnear `reporter` con la lista de archivos modificados durante el run. El `reporter` aplica el delta a `.context/domains/`, `.context/patterns.md`, `.context/contracts.md`, `.context/ops.md`, `.context/risks.md` según corresponda.

**El Líder NO escribe en esos paths** — están en `denied_tools` del frontmatter del `leader.md`. El mapeo de archivos tocados → secciones de `.context/` vive en `skills/context-nav/update.md` — el `reporter` lo carga al ejecutar el delta. El Líder no necesita conocer este mapeo.

**Saltar `reporter` solo si el run NO modificó archivos del proyecto** (caso atípico en Integración, pero posible si todo el cambio fue revertido). Si hubo cualquier modificación → invocar siempre.

**Instrucción adicional al `reporter`:** incluir explícitamente en el prompt "Actualiza también `last_updated` en `.context/NAVIGATOR.md` a la fecha de hoy" — esto delega el paso 6 al `reporter` y evita que el Líder lo haga después por separado.

### 4 — Ejecutar `/task-complete <TASK-ID>`

Después del spawn del `reporter` y de la escritura al vault, ejecutar `/task-complete <TASK-ID>` para marcar la tarea como `done` en el backlog, archivar el handoff y actualizar las métricas del sprint.

Este paso es **responsabilidad exclusiva del Líder** — el `developer` solo reporta que la implementación está lista. Si no hay TASK-ID (invocación directa sin backlog), **omitir este paso** sin escalar.

### 5 — Cerrar la orquestación en Anvil MCP

`mcp__anvil__complete_orchestration(run_id, status)` con el `run_id` activo del Paso 0.5 (skill `run-init`).

- `status="success"` si todos los gates pasaron y los sub-agentes cerraron correctamente.
- `status="partial"` si hubo gates rechazados pero el usuario decidió cerrar igual.
- `status="failed"` si se violó alguna Regla inviolable durante el run.

### 6 — Actualizar `last_updated` en `.context/NAVIGATOR.md`

Criterio binario:

- **Si el `reporter` fue spawneado en el paso 3** → ya se delegó como instrucción explícita en su prompt. Omitir aquí.
- **Si el `reporter` NO fue spawneado** (caso atípico en Integración) → el Líder lo hace directamente con `Edit` sobre `.context/NAVIGATOR.md`. Esta es la única escritura del Líder en `.context/` fuera de `runs/`.

### 7 — Cerrar ciclo con `digest_from_handoff` (cuando no corrió reporter)

Si el run produjo un `.handoff/<TASK-ID>.md` pero el `reporter` NO fue invocado en este run: el Líder mismo DEBE llamar `mcp__anvil__digest_from_handoff(path=<path al handoff>)` antes de presentar el resultado final al usuario.

Esto cierra el ciclo y persiste lo aprendido en memoria para runs futuros. Si no hay handoff producido en el run, omitir este paso.

**Nota:** cuando el `reporter` corre (caso normal en Integración), él mismo llama `mcp__anvil__digest_from_handoff` al final de su flujo si recibió el path del handoff. Por eso este paso solo aplica cuando el `reporter` fue saltado.

### 8 — Limpiar `.context/runs/<run-id>/`

Si el run cerró en `status="success"` → limpiar el directorio `.context/runs/<run-id>/` (es scratchpad temporal — el historial vive en Anvil MCP).

Si cerró en `partial` o `failed` → **NO limpiar**. Dejar disponible para diagnóstico.

## Fuentes de verdad — separación estricta

| Qué | Dónde | Propósito |
|---|---|---|
| Estado del run, decisiones, digests | **Anvil MCP** (`start_orchestration`, `save_step`, `complete_orchestration`) | Persistencia cross-service, searchable, sobrevive `/clear` |
| Plan de trabajo activo | `.context/runs/<run-id>/plan.md` | Scratchpad local — temporal |
| Outputs intermedios, visual check | `.context/runs/<run-id>/` | Solo mientras el run está activo |
| Conocimiento del repo | `.context/` (project.md, patterns.md, domains/, contracts.md) | Fuente de verdad — siempre actualizar al cierre |
| Trazabilidad cross-proyecto | Vault del proyecto (resuelto vía `project-registry.md`) | Nota humana del cambio |

`.context/runs/` no es historial — es un workspace temporal. El historial vive en Anvil MCP.

## Microservicios

En proyectos cross-service: el run vive en Anvil MCP con referencias a todos los repos tocados. Cada repo actualiza su propio `.context/` al cierre vía spawn de `reporter`. El Líder coordina que todos los `reporter` apliquen el delta antes de marcar `success`.

## Reglas

- El orden de los 8 pasos NO se altera. Si un paso falla, NO continuar al siguiente sin resolver.
- Sin la nota al vault, el cierre del modo Integración **no es válido** (Regla inviolable #6).
- El Líder NUNCA escribe en `.context/domains/`, `.context/patterns.md`, `.context/contracts.md`, `.context/ops.md`, `.context/risks.md` ni `.context/decisions/` — esos paths están en `denied_tools`. Toda escritura ahí pasa por `reporter`.
- `/task-complete` se omite **solo** cuando no hay TASK-ID — no escalar, no preguntar.
- `digest_from_handoff` se llama **una sola vez** por handoff: o lo hace el `reporter` al final de su flujo (caso normal), o lo hace el Líder en el paso 7 (cuando no corrió reporter). Nunca ambos.
- Si `status != "success"` → NO limpiar `.context/runs/<run-id>/`.
