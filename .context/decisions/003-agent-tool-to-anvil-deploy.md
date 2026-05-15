# ADR 003 — Rechazo del swap `Agent` → `mcp__anvil__deploy_agents`; plan para habilitar spawn de sub-agentes desde contexto sub-agente

**Fecha:** 2026-05-10
**Run ID:** run-20260510T-adr003-agent-spawn
**Estado:** Aceptado (con redirección del enunciado original)

## Contexto

Se propuso "reemplazar el uso de la tool `Agent` (no disponible en contexto sub-agente) por `mcp__anvil__deploy_agents` para spawnear sub-agentes desde el Líder".

Al investigar las dos tools encontramos que **la premisa del swap no es ejecutable**: las dos herramientas tienen propósitos distintos e incompatibles. Sin embargo, el problema subyacente que motiva la propuesta sí existe y es real: cuando el Líder corre **como sub-agente** (spawneado por Claude vía `Agent`), el harness de Claude Code **no le otorga la tool `Agent`** — por lo tanto el Líder no puede spawnear sub-agentes anidados, rompiendo toda la cadena de delegación documentada en `agents/leader.md` (Reglas inviolables #1, #2, #9 y todos los modos).

Este ADR (1) documenta por qué el swap propuesto no aplica, (2) describe la causa raíz real del problema, y (3) propone un plan ordenado para resolverlo sin romper el contrato del Líder.

### Evidencia recolectada

- `internal/mcp/tools.go:197-202` — `deploy_agents` está registrado con la descripción "Deploy Anvil agents, skills, and commands to all enabled targets (Claude Code, OpenCode, Gemini, Codex)". Único argumento: `target` ∈ {`all`, `agents`, `skills`}.
- `internal/mcp/utilities.go:167-229` — handler `deployAgents`. Su comportamiento es 100% I/O de filesystem: invoca `deploy.Claude(...)`, `deploy.OpenCode(...)`, etc., que sincronizan archivos del repo a directorios del target. Devuelve `{ target, deployed_agents: [...], deployed_skills: [...] }`. **No spawnea ningún proceso de modelo, no invoca un LLM, no produce output de razonamiento.**
- `internal/deploy/claude.go:22-67` — la función `Claude(cfg, paths)` deploya agentes copiando/adaptando los `agents/*.md` al directorio del Claude Code harness; `adaptClaude` reescribe frontmatter (resuelve `tier` → modelo, `permission` → tools concretas vía `cfg.ResolvePermission`). No tiene ningún path de ejecución que dispare un sub-agente.
- `anvil.config.yaml:47` — `claude.execute: Glob, Grep, LS, Read, Write, Edit, Bash, Skill, Agent`. La tool `Agent` se otorga a todos los agentes con `permission: execute` en deploy time.
- `agents/leader.md:11-34` — `allowed_tools` del Líder incluye explícitamente `Agent` como "única forma de hacer trabajo concreto" y excluye `mcp__anvil__deploy_agents` (solo lista `start_orchestration`, `save_step`, `save_leader_log`, `complete_orchestration`, `load_orchestration`, `search_memories`).
- **Comportamiento observado del harness:** un sub-agente spawneado vía `Agent` **no recibe la tool `Agent` en su contexto**, aun cuando su spec declara `permission: execute` en `allowed_tools`. Esto bloquea la delegación anidada. (Evidencia indirecta: este mismo run del Líder como sub-agente no tiene `Agent` disponible — solo `Read`, `Write`, `Edit`, `Bash`, `Skill`.)

## Decisión

### 1) Rechazar el swap `Agent` → `mcp__anvil__deploy_agents`

No reemplazar `Agent` por `mcp__anvil__deploy_agents` en `agents/leader.md`. **`mcp__anvil__deploy_agents` no es un spawner de sub-agentes**; es una utilidad de deployment que copia archivos de specs del repo Anvil a los directorios target de los harnesses (Claude Code, OpenCode, Gemini, Codex). Tratarla como spawner produciría:

- Llamadas que no devuelven output del sub-agente (solo lista de archivos deployados) → el Líder se quedaría esperando un output que nunca llega.
- Re-deploy completo de TODOS los agentes y skills en cada turno de orquestación → I/O masivo, posible race con deploys legítimos, ensuciamiento del filesystem de targets.
- Ruptura silenciosa de las Reglas inviolables #1, #2, #9 y del Protocolo de debate: sin spawn real, el Líder terminaría haciendo el trabajo de los sub-agentes él mismo o devolviendo "no puedo continuar" en cada turno.

### 2) Reconocer la causa raíz real

El problema real es que **el harness de Claude Code no expone la tool `Agent` cuando un agente es spawneado vía `Agent`**, aunque su spec declare `permission: execute`. El sistema actual asume implícitamente que el Líder se invoca como agente top-level (Claude → Leader directo), no como sub-agente anidado. Cuando Claude (siguiendo `~/.claude/CLAUDE.md`) hace `Agent(subagent_type="leader", ...)`, el Líder recibe un toolset reducido y pierde la capacidad de spawnear.

Esto **es** una limitación arquitectónica genuina, pero la solución no vive en `agents/leader.md`. Vive en una de tres capas:

| Capa | Cambio | Quién lo aplica |
|---|---|---|
| **Harness de Claude Code** | Permitir `Agent` recursivo en sub-agentes con `permission: execute` | Anthropic (fuera de nuestro control) |
| **Modelo de invocación** | El Líder se invoca como agente top-level, no anidado; Claude solo hace prompt-pass-through, no spawn | `~/.claude/CLAUDE.md` + cambio en cómo Claude llama al Líder |
| **MCP propio de Anvil** | Crear un tool `mcp__anvil__spawn_agent` que invoque un sub-modelo y devuelva su output | `internal/mcp/` (nuevo handler) + nuevo agent runner Go |

### 3) Camino recomendado (ordenado por costo/beneficio)

**Opción A — Cambiar el modelo de invocación (recomendado, costo bajo, sin código Go nuevo).**

En lugar de que Claude haga `Agent(subagent_type="leader", ...)`, Claude inyecta el spec del Líder como `system prompt` adicional en su propia conversación y opera como Líder él mismo. Concretamente:

- Reescribir `~/.claude/CLAUDE.md` para que en lugar de spawnear al Líder vía `Agent`, **adopte su rol** cargando inline `agents/leader.md` (vía Read sobre el repo activo en la sesión top-level, donde Claude sí tiene Read).
- Claude sigue siendo el único punto de contacto con el usuario; pero asume el rol del Líder (frontmatter, reglas inviolables, Paso 0, detección de modo, pipelines).
- El spawn de sub-agentes concretos (`explorer`, `pm`, `developer`, `tester`, etc.) sigue funcionando porque ocurre desde la sesión top-level donde `Agent` sí está disponible.
- Las restricciones de toolset del Líder (whitelist de `Read`, denied `Grep`/`Glob`/`WebFetch`/`WebSearch`, etc.) se enforcean por auto-corrección textual (igual que hoy, según ADR 002).

Trade-off: Claude top-level y el Líder se fusionan conceptualmente. Esto contradice la frontera "Claude no orquesta, el Líder sí" que `~/.claude/CLAUDE.md` afirma. Hay que decidir si esa frontera vale más que la capacidad técnica de spawn anidado.

**Opción B — Implementar `mcp__anvil__spawn_agent` (costo medio-alto, requiere código Go nuevo).**

Agregar un tool MCP en `internal/mcp/tools.go` y handler en `internal/mcp/execution.go` o un archivo nuevo que:

1. Reciba `agent_name` (ej. `"explorer"`, `"developer"`), `prompt` (markdown completo), opcionalmente `run_id`.
2. Localice `agents/<agent_name>.md`, resuelva su frontmatter (tier → modelo, permission → tools).
3. Invoque el provider activo (claude, gemini, openai) vía API directa con el modelo resuelto, pasando el spec + prompt.
4. Devuelva el output del sub-agente como string al caller.
5. Persista el step vía `save_step` automáticamente para mantener consistencia con la orquestación.

Trade-off: replica funcionalidad del harness dentro de Anvil. Costo de mantenimiento: drivers por provider, manejo de tools dentro del sub-agente (Read/Write/Bash), passthrough de filesystem, gestión de errores. Beneficio: el Líder puede spawnear desde cualquier nivel de anidamiento, y la persistencia queda 100% controlada por Anvil.

**Opción C — Status quo + documentación (costo nulo, no resuelve).**

Aceptar que el Líder solo funciona como agente top-level. Documentar en `~/.claude/CLAUDE.md` que Claude **no debe** usar `Agent` para invocar al Líder; en su lugar, debe re-prompt al usuario instruyendo a invocar al Líder como sesión top-level separada (ej. via slash command `/leader` o similar). No resuelve el problema; lo desplaza al usuario.

**Recomendación: Opción A** primero (lift inmediato), seguida de Opción B si después se decide que Claude y el Líder deben mantenerse como entidades separadas (ej. para auditoría granular, métricas por agente, costos separados, o sub-agentes que no sean Claude).

### 4) Diferencias entre `Agent` y `mcp__anvil__deploy_agents`

| Dimensión | `Agent` (Claude Code harness) | `mcp__anvil__deploy_agents` (Anvil MCP) |
|---|---|---|
| **Naturaleza** | Tool del harness — spawnea un sub-modelo con spec + prompt | Tool MCP de Anvil — copia/sincroniza archivos de specs a directorios target |
| **Input** | `subagent_type` (nombre del agente), `prompt` (markdown) | `target` opcional ∈ {`all`, `agents`, `skills`} |
| **Output** | Texto generado por el sub-modelo (razonamiento + artefactos producidos) | JSON con lista de archivos deployados: `{ target, deployed_agents: [...], deployed_skills: [...] }` |
| **Efecto sobre el sistema** | Crea una sesión efímera de modelo, devuelve su output al caller | Reescribe `~/.claude/agents/`, `~/.config/opencode/agent/`, etc. con copias adaptadas de `agents/*.md` del repo |
| **Idempotencia** | No (cada call cuesta tokens y puede producir output distinto) | Sí (re-ejecutarlo produce el mismo estado de filesystem) |
| **Costo** | Tokens del modelo del sub-agente | Solo I/O de filesystem |
| **Latencia típica** | Segundos a minutos (depende del modelo y prompt) | Milisegundos |
| **Disponibilidad en sub-agente** | NO (limitación del harness — causa raíz del problema) | SÍ (es una tool MCP, no del harness) |
| **Uso correcto** | Spawn de sub-agentes en pipelines | Setup inicial / re-deploy tras cambiar provider o agregar agentes |

### 5) Cambios necesarios en `agents/leader.md` (si se adopta Opción A)

Ninguno relacionado con `Agent` → `deploy_agents`. El swap propuesto no aplica.

Cambios mínimos si se adopta Opción A:

- Sección "Rol" — clarificar que el Líder puede ser asumido por Claude top-level (no solo invocado vía `Agent`). Quitar la implicación "el Líder siempre se spawnea".
- `allowed_tools` — sin cambios; `Agent` sigue siendo válido porque la sesión que asume el rol del Líder lo tiene.
- Reglas inviolables — sin cambios (las invariantes #1, #2, #8, #9 son independientes del mecanismo de invocación).
- Eliminar (si existe) cualquier asunción de "el Líder es un sub-agente". No la encontramos al revisar el spec actual — el spec no afirma top-level ni sub-agente explícitamente, solo asume que `Agent` está disponible.

Cambios si se adopta Opción B (nuevo MCP tool):

- `allowed_tools` — agregar `mcp__anvil__spawn_agent`.
- Reescribir las referencias a "spawnear sub-agente vía `Agent`" para que digan "spawnear sub-agente vía `Agent` o `mcp__anvil__spawn_agent` (cuando el Líder corre anidado)".
- Documentar el orden de preferencia: usar `Agent` si está disponible (top-level), sino `mcp__anvil__spawn_agent`.
- Las invariantes #1, #2, #8, #9 quedan iguales — solo cambia el verbo, no la semántica de delegación.

### 6) Cambios necesarios en el harness Go (si se adopta Opción B)

Archivos a crear/modificar:

| Archivo | Cambio |
|---|---|
| `internal/mcp/tools.go` | Registrar `spawn_agent` con `add("spawn_agent", "Spawn a sub-agent...", schema(prop("agent_name", ...), prop("prompt", ...), optProp("run_id", ...)), s.spawnAgent)` |
| `internal/mcp/execution.go` (nuevo handler) | Implementar `func (s *Server) spawnAgent(ctx, args) (string, error)`: resolver spec del agente, llamar al provider, persistir step, devolver output |
| `internal/provider/` (nuevo paquete) | Drivers por provider (`claude.go`, `openai.go`, `gemini.go`) con interface `Invoke(model string, system string, prompt string) (string, error)` |
| `internal/deploy/claude.go::adaptClaude` | Procesar `allowed_tools`/`denied_tools` del frontmatter para producir whitelist real cuando el sub-agente corra vía `spawn_agent` (follow-up de ADR 002) |
| `anvil.config.yaml` | Posiblemente agregar `mcp__anvil__spawn_agent` a la lista `execute` del Líder, si se quiere uniforme |

Tests:

- `internal/mcp/execution_test.go` — `TestSpawnAgent_HappyPath`, `TestSpawnAgent_UnknownAgent`, `TestSpawnAgent_ProviderError`, `TestSpawnAgent_PersistsStep`.
- Mock provider para tests sin tokens reales.

Si se adopta Opción A: ningún cambio en el harness Go. Solo `~/.claude/CLAUDE.md` y posiblemente `agents/leader.md` (texto, no frontmatter).

### 7) Pasos de implementación ordenados

**Si se adopta Opción A (recomendada):**

1. Confirmar con el usuario que la fusión conceptual Claude ↔ Líder es aceptable. Si no → ir a Opción B.
2. `agent-designer` reescribe `~/.claude/CLAUDE.md`: en lugar de "Claude spawnea al Líder con `Agent`", queda "Claude carga `agents/leader.md` del repo activo y opera bajo su contrato". Cambia el bloque "Cómo spawnear al Líder" y las 7 condiciones de spawn obligatorio (siguen aplicando, pero como triggers para Claude-asumiendo-rol-de-Líder, no para `Agent(subagent_type="leader")`).
3. `agent-designer` ajusta `agents/leader.md` con los cambios menores de la sección (5) Opción A — sin tocar `allowed_tools`, solo aclarando el modelo de invocación.
4. Test manual: usuario hace una tarea típica (ej. "investiga X"). Verificar que Claude detecta el rol del Líder, ejecuta Paso 0, spawnea `explorer` vía `Agent` (funciona porque Claude está top-level), recibe el output, y presenta el formato del modo Explorador.
5. ADR de seguimiento (si necesario) documentando el cambio de contrato Claude ↔ Líder.

**Si se adopta Opción B:**

1. Diseño detallado del MCP tool `spawn_agent` (PRD del `pm` + ARD del `architect` — modo Planeación completo).
2. Implementación incremental:
   - 2a. Paquete `internal/provider/` con interface + driver Claude solo.
   - 2b. Handler `spawnAgent` en `internal/mcp/execution.go` con resolución de spec y llamada al provider.
   - 2c. Persistencia automática vía `save_step`.
   - 2d. Tests unitarios e integración.
3. `developer` extiende `internal/deploy/claude.go::adaptClaude` para emitir tools resueltas a partir de `allowed_tools`/`denied_tools` (follow-up de ADR 002 que se desbloquea aquí).
4. `agent-designer` actualiza `agents/leader.md` con `allowed_tools` extendido (`mcp__anvil__spawn_agent`) y reglas de preferencia (`Agent` cuando disponible, `spawn_agent` sino).
5. Test de aceptación: invocar al Líder como sub-agente (vía Claude → `Agent(leader)`), verificar que detecta ausencia de `Agent`, usa `spawn_agent` para invocar `explorer`, recibe el output, y completa el pipeline.
6. ADR de seguimiento documentando la nueva tool MCP y su uso.

## Alternativas consideradas

- **Aplicar el swap propuesto literalmente** (`Agent` → `mcp__anvil__deploy_agents` en `allowed_tools`). Rechazado por incompatibilidad de propósito documentada arriba — rompería el sistema en silencio.

- **Eliminar la cadena de delegación del Líder** y permitir que cualquier agente top-level (`developer`, `tester`, etc.) sea invocado directamente por Claude sin pasar por el Líder. Rechazado: viola ADR 001 (Regla #8 — sub-agentes no hablan con el usuario), elimina el self-critique gate (Regla #2), elimina la persistencia coherente vía `mcp__anvil__start_orchestration` (un run por pipeline), y devuelve el sistema a un estado pre-Anvil donde el usuario coordina manualmente.

- **Restringir el Líder a operar siempre como top-level** (documentar que `Agent(leader)` nunca debe llamarse). Rechazado como solución completa: empuja el problema al usuario y no resuelve casos legítimos donde el Líder puede invocarse anidado (ej. una skill que requiere coordinar varios agentes y debe llamarse desde otro agente).

- **Esperar a que Anthropic habilite `Agent` recursivo en el harness.** Rechazado como única solución: no tenemos control sobre el roadmap del harness; bloquea funcionalidad indefinidamente.

## Consecuencias

- **El swap propuesto queda explícitamente rechazado en el repo.** Futuras propuestas similares pueden citar este ADR para evitar re-debatirlo.
- **El problema real (sub-agente sin `Agent`) queda documentado con causa raíz identificada y dos opciones de solución viables.** La elección entre Opción A y B se delega al usuario en función de su valoración de la frontera Claude ↔ Líder.
- **`mcp__anvil__deploy_agents` mantiene su rol actual** como utilidad de deployment de specs — nada cambia en su comportamiento ni en su signature.
- **Si se elige Opción A:** Claude top-level absorbe el rol del Líder; las reglas del Líder (#1, #2, #8, #9, modos, gates) siguen vigentes pero aplicadas por Claude. Hay riesgo de erosión textual si Claude no respeta el spec; mitigarlo con un hook que valide al inicio que `agents/leader.md` está cargado.
- **Si se elige Opción B:** se introduce un nuevo paquete `internal/provider/` y un nuevo tool MCP. Aumenta la superficie de mantenimiento (drivers por provider, manejo de tools del sub-agente, costos cross-provider) a cambio de spawn anidado sin depender del harness.
- **Sin acción adicional, el sistema sigue funcionando** mientras Claude invoque al Líder como top-level (camino actual de `~/.claude/CLAUDE.md`). El gap solo se manifiesta si algún día se intenta llamar al Líder desde otro agente vía `Agent`.

## Archivos modificados

- `.context/decisions/003-agent-tool-to-anvil-deploy.md` — nuevo ADR (este archivo).

Sin cambios en código ni en specs de agentes en este ADR. Los cambios concretos se aplicarán en runs separados según la opción que el usuario elija (A o B).

## Verificación

- `internal/mcp/tools.go:197-202` y `internal/mcp/utilities.go:167-229` confirman que `deploy_agents` es deployment, no spawn.
- `internal/deploy/claude.go` no contiene ninguna ruta de ejecución que dispare un sub-modelo.
- `anvil.config.yaml:47` confirma que `Agent` se otorga vía `permission: execute`.
- `agents/leader.md` `allowed_tools` actual incluye `Agent` y NO incluye `mcp__anvil__deploy_agents`.
- Este ADR no modifica ningún archivo fuera de `.context/decisions/`.
