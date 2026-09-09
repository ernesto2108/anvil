# Exploración — task-workflows

**Run:** adhoc
**Fecha:** 2026-08-19 00:00

## Objetivo
Diagnosticar por qué Linear, documentación de cambios y descripciones de PR no se ejecutan automáticamente en Anvil.

## Fuentes consultadas

| Fuente | Tipo | Origen |
|---|---|---|
| Context Navigator | local | /Users/ernestodiaz/projects/anvil/.project-context/NAVIGATOR.md |
| Gestión y workflow del proyecto | local | /Users/ernestodiaz/projects/anvil/.project-context/Core/task-management.md; /Users/ernestodiaz/projects/anvil/.project-context/Core/workflows.md |
| Flujos de tareas, commit y reporte | local | /Users/ernestodiaz/projects/anvil/skills/task-flow/SKILL.md; /Users/ernestodiaz/projects/anvil/skills/committer-flow/SKILL.md; /Users/ernestodiaz/projects/anvil/skills/reporter/SKILL.md |
| Agente backend y cierre local | local | /Users/ernestodiaz/projects/anvil/agents/developer-backend.md; /Users/ernestodiaz/projects/anvil/skills/task-complete/SKILL.md |
| Registro global de rutas | local | /Users/ernestodiaz/.claude/project-registry.md |
| Historial y PRs remotos | github | origin/develop (fetch bloqueado por sandbox); gh pr list (red no disponible) |

## Hallazgos clave
- La selección de Linear+Outline depende exclusivamente de `project-registry.md`, no de `.project-context/`; para `anvil` el registro enruta al vault local y su contexto declara `task_tool: ""`/"ninguna". `/Users/ernestodiaz/projects/anvil/skills/task-flow/SKILL.md:21-31`; `/Users/ernestodiaz/projects/anvil/.project-context/Technical domain/project.md:3-4`; `/Users/ernestodiaz/projects/anvil/.project-context/Core/task-management.md:7-13`.
- `task-flow` solo se activa por frases explícitas, no está cargada por los developers ni tiene invocador automático. `/Users/ernestodiaz/projects/anvil/skills/task-flow/SKILL.md:1-3`; `/Users/ernestodiaz/projects/anvil/agents/developer-backend.md:1-21,81`.
- `task-flow` promete que `committer-flow` abre un PR, pero este último define únicamente commit y push; no contiene un paso `gh pr create`. Por tanto no obtiene URL de PR y el cierre Linear/Outline se detiene. `/Users/ernestodiaz/projects/anvil/skills/task-flow/SKILL.md:148-156`; `/Users/ernestodiaz/projects/anvil/skills/committer-flow/SKILL.md:3,8,213-259`.
- No existe plantilla, esquema ni comando de creación de PR; el contexto dice explícitamente que no hay plantilla. Esto deja título/cuerpo al comportamiento ad hoc. `/Users/ernestodiaz/projects/anvil/.project-context/Core/workflows.md:42-48`.
- El reporter exige actualización de `.project-context/` tras cambios, pero su trigger depende de que el agente lo invoque; developers declaran reporter como paso final, aunque separan commits/PRs al humano. `/Users/ernestodiaz/projects/anvil/skills/reporter/SKILL.md:10-25`; `/Users/ernestodiaz/projects/anvil/agents/developer-backend.md:81,132-134`.
- `task-complete` y `handoff` prohíben actualizar sistemas externos: solo instruyen al humano, contradiciendo el objetivo de automatizar Linear. `/Users/ernestodiaz/projects/anvil/skills/task-complete/SKILL.md:9,22-26,59-75`; `/Users/ernestodiaz/projects/anvil/skills/handoff/SKILL.md:55-60`.

## Gaps y preguntas abiertas
- No se pudo actualizar `origin` (restricción sobre `.git/FETCH_HEAD`) ni listar PRs (sin red); el diagnóstico se basa en fuentes locales.
- El run usa los fallbacks `run-id=adhoc` y `topic=task-workflows` porque no se proporcionaron.
