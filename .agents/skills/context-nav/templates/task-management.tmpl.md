# Gestión de Tareas — <ProjectName>

<!-- Dónde viven las tareas, convenciones de tickets y definition of done. -->

last_updated: <YYYY-MM-DD>

## Herramienta de gestión

- **Herramienta:** `<Linear / Jira / GitHub Issues / Notion / ninguna>`
- **Workspace / Proyecto:** `<nombre o URL>`
- **Acceso:** `<link o instrucciones>`

## Convenciones de tickets

### Tipos de ticket

| Tipo | Prefijo / Label | Descripción |
|---|---|---|
| Feature | `feat` / `FEAT` | Nueva funcionalidad |
| Bug | `fix` / `BUG` | Corrección de error |
| Chore | `chore` | Mantenimiento técnico |
| Spike | `spike` | Investigación |

### Campos obligatorios por ticket

- **Título:** `<convención — ej: [TICKET-123] Descripción en imperativo>`
- **Descripción:** <qué incluir>
- **Criterios de aceptación:** <formato esperado>
- **Estimación:** <sistema — ej: story points / tallas>

## Definition of Done

Un ticket se considera terminado cuando:

- [ ] Código implementado y revisado
- [ ] Tests escritos y pasando
- [ ] Lint sin errores
- [ ] PR aprobado por al menos <N> reviewer(s)
- [ ] Deploy a staging exitoso
- [ ] `.project-context/` actualizado (via reporter)
- [ ] <criterio adicional del equipo>

## Flujo de estados

```
Backlog → In Progress → In Review → Done
```

## Estimación

- **Sistema:** <story points / tallas S/M/L/XL / horas>
- **Escala:** <1/2/3/5/8/13 / S=1d M=3d L=1w>
- **Quién estima:** <el equipo en refinamiento / el dev asignado>

## Ceremonias del equipo

| Ceremonia | Frecuencia | Duración | Propósito |
|---|---|---|---|
| Refinamiento | <semanal> | <1h> | Preparar y estimar tickets |
| Planning | <quincenal> | <2h> | Seleccionar tickets del sprint |
| Retro | <quincenal> | <1h> | Mejorar el proceso |
