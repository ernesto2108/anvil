# ADR 002 — Restricción del toolset del Líder y creación del agente `explorer`

**Fecha:** 2026-05-10
**Run ID:** run-20260510T184539Z-leader-explorer
**Estado:** Aceptado

## Contexto

El Líder (`agents/leader.md`) declara como Rol único "orquestar — no ejecuta", y la Regla inviolable #1 prohíbe que escriba código, tests, configs del proyecto, specs del sistema de IA y otros archivos sensibles. ADR-001 reforzó esa invariante a nivel declarativo (delegación de specs al `agent-designer` y aislamiento de sub-agentes).

Sin embargo, dos gaps persistían:

1. **Toolset implícito.** El frontmatter del Líder declaraba `permission: execute`, lo que en el adapter del provider se traduce a "todas las herramientas permitidas" (Edit, Write, Bash arbitrario, Grep, Glob, WebFetch, WebSearch). El cumplimiento de la Regla #1 dependía exclusivamente de la disciplina textual del spec — nada en el harness impedía técnicamente que el Líder editara código o leyera el repo.

2. **Modo Explorador ejecutado por el Líder.** La sección "Modo Explorador" del Líder decía explícitamente "Líder investiga directamente. No delega salvo casos específicos". Eso contradecía el Rol ("el Líder orquesta, no ejecuta") y obligaba al Líder a hacer Read sobre código del repo, WebFetch, WebSearch — exactamente las herramientas que la Regla #1 quería restringir. La contradicción daba cobertura para que el Líder leyera archivos del proyecto bajo el pretexto "estoy investigando", erosionando la frontera.

## Decisión

- **Toolset del Líder restringido técnicamente vía `allowed_tools` y `denied_tools` en el frontmatter.** El Líder pasa de toolset implícito (heredado de `permission: execute`) a una whitelist explícita: `Read` solo sobre `.context/`, `~/.claude/project-registry.md`, `~/.claude/CLAUDE.md`, `.handoff/**`, `CLAUDE.md` del proyecto; `Write`/`Edit` solo sobre `.context/runs/**`, `.context/` (Navigator, domains, patterns, contracts, ops, risks, decisions) y vault del proyecto; `Bash` solo en una whitelist de comandos exactos (`git status`, `git diff`, `ls`, `mkdir -p`, `verify-handoff.sh`, `date`); `Agent` como única vía para spawnear trabajo concreto; tools MCP de Anvil para persistencia de orquestación. Todo lo demás cae en `denied_tools`: `Edit`/`Write` sobre código (`.go`, `.ts`, `.tsx`, `.py`, `.dart`, `.rs`, `Makefile`, `Dockerfile`, `package.json`), specs del sistema de IA (`agents/`, `skills/`, `commands/`, `pipelines/`, `settings.json`, `~/.claude/CLAUDE.md`), y exploración de código (`Grep`, `Glob`, `WebFetch`, `WebSearch`, `Bash[*]`).

- **Crear `agents/explorer.md`** para cubrir el trabajo de investigación que el Líder hacía directamente. Toolset: `Read[**]` (cualquier path), `Grep`, `Glob`, `WebFetch`, `WebSearch`, `Bash` con whitelist read-only (`ls`, `find`, `file`, `wc`, `head`, `tail`, `cat`, `git log/show/blame/diff`, `gh view/api`, `curl -sI`). Sin `Edit`, `Write`, `Agent`, sin MCP de modificación. Modelo `medium` (sonnet-class) — síntesis cross-fuente sin requerir diseño arquitectónico. Nunca habla con el usuario directamente; devuelve hallazgos estructurados al Líder.

- **Reescribir el Modo Explorador del Líder** para delegar siempre al `explorer`. Excepción única: si la pregunta es 100% resuelta por `.context/` cargado en Paso 0.3, el Líder responde directo sin spawn.

- **Agregar Regla inviolable #9** ("Investigación se delega al `explorer`") en `agents/leader.md`. Refuerza textualmente la frontera que ya está en el frontmatter — el Líder NO usa `Read`/`Grep`/`Glob`/`WebFetch`/`WebSearch` sobre archivos que no sean `.context/` o configuración.

- **Documentar el flujo de escalación** en una sección dedicada de `agents/leader.md`. Sub-agente con problema → escala al Líder → Líder escala al usuario (después de Paso 1 del Protocolo de debate). Tabla explícita de qué escala el sub-agente vs qué resuelve el Líder.

**Premisa de enforcement:** las listas `allowed_tools`/`denied_tools` son declarativas en el spec. La aplicación efectiva depende del adapter del provider (`internal/deploy/claude.go::adaptClaude`). Si el adapter aún no procesa estas listas, el guardrail se aplica por auto-corrección del Líder mediante la subsección "Self-check antes de tool call" (Regla inviolable #1). Extender el adapter para enforcement real queda como follow-up.

## Alternativas consideradas

- **Mantener el toolset implícito y reforzar solo el spec textual.** Rechazado. Es lo que ya existía pre-ADR-002 y produjo la contradicción del Modo Explorador. Un guardrail que vive solo en prosa se erosiona con cada interacción ambigua.

- **Permitir al Líder usar `Read`/`Grep` sobre el repo "solo en Modo Explorador".** Rechazado. Genera dos modos de operación distintos para el mismo agente, duplica responsabilidades con el sub-agente que de todos modos haría falta para investigaciones más profundas, y reintroduce el patrón "el Líder ejecuta cuando le conviene" que la Regla #1 quiere prohibir.

- **Reutilizar el `scanner` para investigación puntual.** Rechazado. El `scanner` corre una vez al inicio del proyecto para bootstrap del Context Navigator (`.context/`); su skill `scan-project` está optimizada para producir un mapa estructural completo, no para responder preguntas puntuales. Forzarlo a doble-rol diluye su responsabilidad y rompe el patrón "una skill por agente".

- **Extender primero el adapter (`internal/deploy/claude.go`) y después actualizar los specs.** Rechazado como bloqueante. El cambio declarativo no requiere el adapter para entregar valor (la Regla #9 + el Self-check del Líder son enforceable como auto-corrección). Extender el adapter es un follow-up independiente que puede correr en paralelo.

## Consecuencias

- **Frontera técnica reforzada.** Cuando el adapter procese `allowed_tools`/`denied_tools`, el Líder no podrá ejecutar `Edit`/`Write` sobre código ni `Grep`/`Glob`/`WebFetch`/`WebSearch` sobre el repo — el harness lo bloqueará. Mientras tanto, el Self-check del Líder produce el mismo efecto vía auto-corrección.

- **Nuevo agente `explorer`** disponible en el catálogo del meta-system. Aparece en las tablas "Sub-agentes disponibles", "Routing por complejidad", "Skip rules" e "Input por sub-agente" del Líder.

- **Modo Explorador reorganizado.** El Líder pasa de ejecutor a orquestador puro también en este modo. El output al usuario queda igual visualmente; cambia upstream — el contenido lo produce el `explorer`.

- **Contrato de escalación explícito.** El flujo sub-agente → Líder → usuario queda documentado con tabla de qué escala cada nivel. Reduce ambigüedad sobre cuándo el Líder debe resolver internamente vs cuándo debe consultar al usuario.

- **Patrón replicable.** El esquema `allowed_tools`/`denied_tools` introducido en el frontmatter del Líder y del `explorer` puede aplicarse a otros agentes (`developer`, `tester`, `reviewer`, `qa`, `security`, `dba`) en iteraciones futuras para reforzar sus fronteras técnicas. Fuera de scope de este ADR.

- **Follow-up identificado:** extender `internal/deploy/claude.go::adaptClaude` (y los demás adapters) para que procesen `allowed_tools`/`denied_tools` y emitan whitelists/blacklists granulares al provider. Sin esto, el guardrail efectivo se reduce a la disciplina del Líder durante el run.

## Archivos modificados

- `agents/leader.md` — frontmatter con `allowed_tools`/`denied_tools`; refuerzo de Rol con delegación de investigación; subsección "Self-check antes de tool call" en Regla #1; nueva Regla inviolable #9; nueva sección "Flujo de escalación documentado"; reescritura completa de Modo Explorador; filas de `explorer` en tablas "Sub-agentes disponibles", "Routing por complejidad", "Skip rules", "Input por sub-agente".
- `agents/explorer.md` — nuevo agente con frontmatter restringido, rol, inputs esperados, flujo, output al Líder, presupuesto, reglas y no-objetivos.

## Verificación

- `agents/leader.md` declara `allowed_tools` con `Read[.context/**]`, `Read[~/.claude/project-registry.md]`, `Read[.handoff/**]`, `Agent`, MCPs de Anvil y whitelist de bash; declara `denied_tools` con `Grep`, `Glob`, `WebFetch`, `WebSearch`, `Edit`/`Write` sobre código y specs del sistema de IA, y `Bash[*]` para todo lo no listado.
- `agents/leader.md` Sección "Modo Explorador" pipeline = `explorer`. Ningún paso del modo deja al Líder ejecutando `Grep`/`Glob`/`WebFetch`/`WebSearch`.
- `agents/leader.md` Regla inviolable #9 existe y dice explícitamente que toda investigación se delega al `explorer`.
- `agents/leader.md` sección "Flujo de escalación documentado" incluye camino feliz, camino con sub-agente con problema, y tabla "qué escala el sub-agente vs qué escala el Líder".
- `agents/explorer.md` existe con `permission: read`, `model: medium`, skills `[read-files]`, `denied_tools` con `Edit`, `Write`, `Agent`, `Bash[*]` y MCPs de modificación.
- `agents/explorer.md` declara explícitamente "NO te invocan agentes que no sean el Líder" y el formato de output al Líder.
- ADR-001 sigue vigente — Reglas #1 y #8 no se reescriben; Regla #9 las complementa.
