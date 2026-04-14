# Handoff Template

Use this template when creating a new handoff file at `.handoff/<TASK-ID or slug>.md`. Fill in the sections as you work — do not leave placeholder comments in the final file.

**Cross-stack rule:** if the task touches more than one stack (e.g., Go + React), use the `## Phases` section with one phase per stack and include the `## Contract bridge` section. If single-stack, collapse phases into the flat `## Estado actual` checklist and omit the contract bridge.

---

```markdown
# Handoff — <TASK-ID or slug>

## Input recibido
<!-- Qué proporcionó el orchestrator. Llenar al crear — es el recibo de lo que se recibió. -->

| Campo | Valor |
|---|---|
| Complejidad | <Small/Medium/Large> — <pts> pts |
| Stacks | <Go / React / Flutter / etc.> |
| Skills de convención cargados | <go-conventions, react-conventions, etc. o "reglas inline"> |
| PRD | <path o "N/A" o "inline"> |
| Design | <path o "N/A" o "inline"> |
| Context.md | <leído / inline / N/A> |
| Handoff existente | <continuación / nuevo> |
| Modo | <normal / maquetation / integration / qa-fix> |

## Estado actual

<!-- SINGLE-STACK tasks: flat checklist -->
- [ ] Step 1: description
- [ ] Step 2: description
- [ ] Step 3: description

## Fases

<!-- CROSS-STACK: una fase por stack, ordenadas por dependencia (backend primero, luego frontend).
     Eliminar esta sección para tareas single-stack. -->

### Fase 1 — Backend (<stack>)
- [ ] Paso 1: descripción
- [ ] Paso 2: descripción

### Fase 2 — Frontend (<stack>)
- [ ] Paso 1: descripción
- [ ] Paso 2: descripción

<!-- Agregar más fases si es necesario (e.g., Fase 3 — Mobile) -->

## Puente de contratos

<!-- Solo para tareas CROSS-STACK. Eliminar para single-stack.
     Documentar los contratos exactos que conectan los stacks — el DTO, API, binding o evento
     que un stack expone y el otro consume. Esta es la fuente #1 de bugs en tareas cross-stack. -->

### Backend expone
<!-- Struct/función Go exacta con JSON tags que el frontend va a consumir -->
```go
// Ejemplo:
type MetricsDTO struct {
    RunsCount int `json:"runsCount"`
}
func (a *App) GetMetrics() (*MetricsDTO, error)
```

### Frontend consume
<!-- Interface/call TypeScript exacto que mapea al contrato del backend -->
```typescript
// Ejemplo:
interface MetricsDTO {
  runsCount: number
}
async function getMetrics(): Promise<MetricsDTO>
```

### Validación del contrato
<!-- Cómo verificar que ambos lados coinciden — comandos exactos -->
- Backend compila: `<comando>`
- Frontend compila: `<comando>`
- Tipos coinciden: <verificación manual / codegen / schema compartido>

## Dependencias cross-service

<!-- Solo para tareas que tocan múltiples repos/servicios. Eliminar en caso contrario.
     Si está presente, el orchestrator DEBE verificar el orden de deploy antes de cerrar. -->

| Servicio | Repo | Qué cambia | Orden de deploy |
|---|---|---|---|
| <service-a> | <repo> | <qué> | 1 — deploy primero (provee API) |
| <service-b> | <repo> | <qué> | 2 — depende de service-a |

### Contratos compartidos
<!-- Endpoints de API, schemas de eventos, o tablas de DB que cruzan fronteras de servicio -->

### Cambios que rompen compatibilidad
<!-- ¿Esto rompe consumidores existentes? ¿Plan de migración? -->

## Archivos modificados
<!-- path — what was done and why -->

## Decisiones tomadas
<!-- decision — reasoning -->

## Siguiente paso
<!-- Exactly what to do next to resume work -->

## Notas
<!-- Non-obvious context: workarounds, bugs found, blockers -->

## Handoff for tester
<!--
MANDATORY for the developer to fill BEFORE finishing. The tester reads THIS section
instead of re-reading production files, so it must be complete and precise.
If this section is empty or incomplete when the developer reports done, the
orchestrator will bounce the task back to the developer.
-->

### Archivos de producción tocados
<!-- one line per file: path — role (store query / handler / DTO / custom component / etc.) -->

### Public interfaces / contracts
<!--
Exact signatures of what was added or modified. Copy-paste from the code.
- New types/structs with all fields and tags
- New functions/methods with full signatures (params, return types, error behavior)
- New DTOs with JSON tags
-->

### Patrones aplicados
<!-- which patterns from the convention skill the developer followed -->

### Edge cases descubiertos
<!-- NULL handling, empty states, error paths, race conditions considered, unusual inputs -->

### Build tags / constraints
<!-- //go:build tags, embed.FS layout, Wails bindings, Rust cfg, any quirk that affects how tests must be written -->

### Tests requeridos — por stack

<!-- CROSS-STACK: agrupar por stack. SINGLE-STACK: usar un solo grupo.
     Cada grupo incluye: path del archivo, comando de ejecución, y lista numerada.
     El tester implementa SOLO estos — sin extras. -->

#### Tests Go
- **Archivo:** `<path_test.go>`
- **Ejecutar:** `go test -tags <tag> ./<pkg>/...`
- Tests:
  1. <nombre del test> — qué valida
  2. <nombre del test> — qué valida

#### Tests React/TS
- **Archivo:** `<path.test.tsx>`
- **Ejecutar:** `npm test -- --run <scope>` o `vitest run <scope>`
- Tests:
  1. <nombre del test> — qué valida
  2. <nombre del test> — qué valida

<!-- Agregar más grupos según necesidad: Flutter, Python, Rust -->

### Validación ya ejecutada
<!-- go build, go vet, npm run build, lint — exact commands and their exit status. Tester does NOT repeat build checks. -->

## Output entregado
<!-- Llenar ANTES de terminar. Es el recibo de entrega — lo que el orchestrator y el siguiente agente pueden verificar. -->

| Verificación | Resultado |
|---|---|
| Build (<stack>) | PASS / FAIL — comando |
| Lint (<stack>) | PASS / FAIL — comando + cantidad de issues |
| Tests existentes | PASS / FAIL — comando |
| Archivos creados | <cantidad> |
| Archivos modificados | <cantidad> |
| Puente de contratos verificado | SÍ / N/A |
| Impacto cross-service | NINGUNO / <lista de servicios afectados> |

## Retro
<!-- Llenar DESPUÉS de completar la tarea (antes de archivar). Evaluación honesta — alimenta mejoras futuras. -->

### Qué funcionó
<!-- Patrones, decisiones o enfoques que funcionaron bien y deberían repetirse -->

### Qué no funcionó
<!-- Retrabajo, bounces de QA, suposiciones incorrectas, reads desperdiciados — ser específico -->

### Métricas
| Métrica | Estimado | Real |
|---|---|---|
| Story points | <pts> | <pts — igual salvo re-scope> |
| Bounces de QA | 0 | <real> |
| Invocaciones del developer | <esperado> | <real> |
| Invocaciones del tester | 1 | <real> |

### Aprendizaje para futuras tareas
<!-- Un takeaway concreto. No genérico ("planear mejor") — específico ("columnas FK nullable necesitan edge case explícito en handoff o el tester las pierde") -->

## Token usage
| Session | Tokens used | Tokens available | Tool calls | Files read | Files written |
|---------|------------|-----------------|------------|------------|---------------|
```
