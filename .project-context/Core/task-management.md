# Gestión de Tareas — anvil

<!-- Dónde viven las tareas, convenciones de tickets y definition of done. -->

last_updated: 2026-08-13

## Herramienta de gestión

- **Herramienta:** ninguna
- **Workspace / Proyecto:** N/A
- **Acceso:** N/A

Nota: `vault-template/` en este repo es un artefacto/template que anvil **genera** para otros proyectos (Obsidian vault con backlog, tasks, ADRs, etc.) — no es la herramienta de gestión usada para el propio desarrollo de anvil.

## Convenciones de tickets

### Tipos de ticket

| Tipo | Prefijo / Label | Descripción |
|---|---|---|
| Feature | `feat` / `FEAT` | Nueva funcionalidad |
| Bug | `fix` / `BUG` | Corrección de error |
| Chore | `chore` | Mantenimiento técnico |
| Spike | `spike` | Investigación |

### Campos obligatorios por ticket

- **Título:** <sin convención documentada — no se usa herramienta externa>
- **Descripción:** <sin convención documentada>
- **Criterios de aceptación:** <sin convención documentada>
- **Estimación:** <sin convención documentada>

## Definition of Done

Un ticket se considera terminado cuando:

- [ ] Código implementado y revisado
- [ ] Tests escritos y pasando (`go test -race ./...`)
- [ ] Lint sin errores (`go vet ./...`)
- [ ] PR aprobado por al menos 1 reviewer (según convención de GitHub, sin herramienta formal)
- [ ] `.project-context/` actualizado (via reporter)

## Flujo de estados

```
Backlog → In Progress → In Review → Done
```

<!-- Sin herramienta de gestión formal — el flujo de estados no está instrumentado externamente. -->

## Estimación

- **Sistema:** <sin convención documentada>
- **Escala:** <sin convención documentada>
- **Quién estima:** <sin convención documentada>

## Ceremonias del equipo

| Ceremonia | Frecuencia | Duración | Propósito |
|---|---|---|---|
| <sin ceremonias documentadas> | — | — | — |
