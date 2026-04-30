# Plantilla de Handoff

Usar esta plantilla al crear un nuevo archivo de handoff en `.handoff/<TASK-ID o slug>.md`. Llenar las secciones mientras trabajas — no dejar comentarios de placeholder en el archivo final.

**Regla cross-stack:** si la tarea toca más de un stack (ej., Go + React), usar la sección `## Fases` con una fase por stack e incluir la sección `## Puente de contratos`. Si es single-stack, colapsar las fases en el checklist plano `## Estado actual` y omitir el puente de contratos.

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

<!-- Tareas SINGLE-STACK: checklist plano -->
- [ ] Paso 1: descripción
- [ ] Paso 2: descripción
- [ ] Paso 3: descripción

## Fases

<!-- CROSS-STACK: una fase por stack, ordenadas por dependencia (backend primero, luego frontend/mobile).
     Eliminar esta sección para tareas single-stack. -->

### Fase 1 — Backend (<stack>)
- [ ] Paso 1: descripción
- [ ] Paso 2: descripción

### Fase 2 — Frontend (<stack>)
- [ ] Paso 1: descripción
- [ ] Paso 2: descripción

<!-- Agregar más fases si es necesario (ej., Fase 3 — Mobile) -->

## Puente de contratos

<!-- Solo para tareas CROSS-STACK. Eliminar para single-stack.
     Documentar los contratos exactos que conectan los stacks — el DTO, API, binding o evento
     que un stack expone y el otro consume. Esta es la fuente #1 de bugs en tareas cross-stack. -->

### Backend expone
<!-- Struct/función Go exacta con JSON tags que el frontend/mobile va a consumir -->
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

### Mobile consume — incluir si aplica
<!-- Modelo Dart/Kotlin/Swift exacto que mapea al contrato del backend -->
```dart
// Ejemplo:
class MetricsDTO {
  final int runsCount;
  MetricsDTO({required this.runsCount});
  factory MetricsDTO.fromJson(Map<String, dynamic> json) =>
    MetricsDTO(runsCount: json['runsCount'] as int);
}
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
<!-- path — qué se hizo y por qué -->

## Decisiones tomadas
<!-- decisión — razonamiento -->

## Verificación de ubicación
<!--
OBLIGATORIO. El developer NO decide ubicación de archivos NEW — eso es responsabilidad del architect en el SPEC.
Esta sección documenta que el developer verificó que el SPEC trae justificación de ubicación para cada archivo CREATE
y que el directorio destino existe. Si el SPEC no la traía, el developer escala al orquestador (no decide solo).
Una línea por archivo NEW. No listar MODIFY/DELETE aquí.
-->

- `<path/al/archivo.ext>` — SPEC justificó: "<copiar la columna 'Ubicación: por qué aquí' del SPEC>". Confirmado en disco. ✓

<!-- Si el SPEC tenía gap y se reinvocó al architect:
- `<path>` — SPEC original sin justificación, architect reinvocado (run X). Ubicación final: <path> porque <razón del SPEC actualizado>.
-->

## Siguiente paso
<!-- Exactamente qué hacer a continuación para retomar el trabajo -->

## Notas
<!-- Contexto no obvio: workarounds, bugs encontrados, blockers -->

## Handoff for tester
<!--
OBLIGATORIO que el desarrollador llene ANTES de terminar. El tester lee ESTA sección
en lugar de re-leer los archivos de producción, por lo que debe ser completa y precisa.
Si esta sección está vacía o incompleta cuando el desarrollador reporta done, el
orchestrator devolverá la tarea al desarrollador.
-->

### Archivos de producción tocados
<!-- una línea por archivo: path — rol (store query / handler / DTO / custom component / etc.) -->

### Public interfaces / contracts
<!--
Firmas exactas de lo que se agregó o modificó. Copiar-pegar del código.
- Nuevos tipos/structs con todos los campos y tags
- Nuevas funciones/métodos con firmas completas (params, tipos de retorno, comportamiento de error)
- Nuevos DTOs con JSON tags
-->

### Patrones aplicados
<!-- qué patrones del convention skill siguió el desarrollador -->

### Edge cases descubiertos
<!-- manejo de NULL, estados vacíos, caminos de error, condiciones de carrera consideradas, entradas inusuales -->

### Build tags / constraints
<!-- //go:build tags, layout de embed.FS, Wails bindings, Rust cfg, cualquier peculiaridad que afecte cómo deben escribirse los tests -->

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

#### Tests de automatización
<!-- Heredar de la sección "Automatización" del SPEC. Solo incluir los tipos marcados "Sí". -->

- **E2E web** (Playwright): `tests/e2e/<feature>.spec.ts`
  - Ejecutar: `npx playwright test tests/e2e/<feature>.spec.ts`
  - Flujos:
    1. <flujo> — qué valida
- **E2E mobile** (Maestro): `.maestro/<feature>.yaml`
  - Ejecutar: `maestro test .maestro/<feature>.yaml`
  - Flows:
    1. <flow> — qué valida
- **API contract** (Hurl): `tests/api/<resource>/<scenario>.hurl`
  - Ejecutar: `hurl --test tests/api/<resource>/`
  - Scenarios:
    1. <scenario> — qué valida
- **Visual regression**: en test E2E con `toHaveScreenshot()`
- **Accesibilidad**: en test E2E con axe-core

<!-- Eliminar los tipos que no aplican según el SPEC. Si el SPEC no tiene sección de automatización, omitir este bloque. -->

### Validación ya ejecutada
<!-- go build, go vet, npm run build, lint — comandos exactos y su estado de salida. El tester NO repite los checks de build. -->

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
