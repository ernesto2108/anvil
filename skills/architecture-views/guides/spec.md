# Template: spec.md

**Generado por:** agente `spec-writer`, después de recibir el ARD del `architect`.
**Consumido por:** `developer`, `tester`, `QA`, `task-decomposer`.

## Qué ES y qué NO ES el spec.md

| ES | NO ES |
|---|---|
| Decisiones + razones | Duplicado de contratos de arquitectura |
| Criterios de aceptación (testeables) | Instrucciones de implementación |
| Límites (hacer/preguntar/nunca) | Código o SQL |
| Bridge de contratos cross-stack | Detalle específico de stack |
| Pre-condiciones | Narrativa de diseño o arquitectura |

**Regla:** Si una sección ya existe en una vista de arquitectura, referenciarla — NO copiarla. Los archivos de arquitectura son la fuente de verdad para contratos.

---

## Template

```markdown
# SPEC: <Nombre del feature> — <TASK-ID>

## Contexto y objetivo

<!-- Un párrafo. Qué problema resuelve y por qué ahora. -->

## No-objetivos

<!-- Exclusiones explícitas de scope. Qué esta tarea NO hace. -->
- ...

## Pre-condiciones

<!-- Qué debe ser verdad / existir antes de que esta tarea pueda empezar. -->
- [ ] Migración XXX aplicada
- [ ] Feature YYY desplegado
- [ ] ...

## Decisiones tomadas (ADR)

<!-- Un bloque por decisión no obvia. Formato MADR resumido. -->

### ADR-01: <Título>

- **Opciones consideradas:** A (pro/con) · B (pro/con) · C (pro/con)
- **Decisión:** Elegimos [opción] porque [fuerza principal].
- **Tradeoff aceptado:** ...

<!-- Repetir por decisión -->

## Mapa de contratos (cross-stack)

<!-- Solo para tareas que tocan múltiples stacks. Mapea el límite entre ellos. -->
<!-- El detalle vive en archivos de arquitectura — esta tabla muestra la conexión. -->

| Productor | Contrato | Consumidor | Ver en |
|---|---|---|---|
| Go `EventWriter` | `ToolUsePayload.InputSizeBytes int` | Rust store `input_size_bytes` | architecture-backend.md §handler |

<!-- Para async: documentar también el canal/topic que los conecta -->
| Productor | Topic / Queue | Contrato (evento) | Consumidor |
|---|---|---|---|

## Mapa de implementación

<!-- Plan a nivel de archivo. Acción = CREATE / MODIFY / DELETE. -->
<!-- Esto es lo que el developer sigue — ser específico sobre qué cambia, no cómo. -->
<!-- Para acción = CREATE: la columna "Ubicación: por qué aquí" es OBLIGATORIA y debe anclar -->
<!-- la decisión en un archivo vecino existente o en el patrón del módulo. -->

| Archivo | Acción | Qué cambia | Ubicación: por qué aquí | Referencia | Fase |
|---|---|---|---|---|---|
| `internal/dashboard/store/runs.go` | MODIFY | Agregar método `UpdateToolUseDuration` | — (existente) | architecture-backend.md §writer | 1 |
| `internal/dashboard/store/cache.go` | CREATE | Cache LRU de runs por proyecto | Sigue el patrón de `internal/dashboard/store/runs.go` (mismo bounded context — persistencia). NO va en `internal/cache/` porque es específico de runs, no util genérico. | architecture-backend.md §cache | 1 |

### Utils a reutilizar (verificación previa OBLIGATORIA)

<!-- Antes de proponer un util/helper nuevo, el architect ejecuta Grep en directorios -->
<!-- de utilidades comunes (`internal/util/`, `pkg/util/`, `src/lib/`, `src/utils/`) -->
<!-- y reporta el resultado. NO crear un util nuevo sin descartar primero los existentes. -->

| Necesidad | Util existente reutilizable | Ubicación |
|---|---|---|
| Parsear duración ISO-8601 | `ParseDuration` | `internal/util/timefmt.go` |
| Hash SHA-256 de payload | (ninguno encontrado — proponer `internal/util/hash.go`) | NEW |

## Criterios de aceptación

<!-- Testeables. Formato: GIVEN / WHEN / THEN. Uno por comportamiento observable. -->
<!-- Agrupar por feature si hay múltiples features en scope. -->

### <Feature 1>

1. GIVEN ... WHEN ... THEN ...
2. GIVEN ... WHEN ... THEN ...

### <Feature 2>

3. GIVEN ... WHEN ... THEN ...

## Testing Strategy (OBLIGATORIO)

Tabla con un row por **criterio de aceptación** declarado arriba. Sin filas vacías.
Si un criterio no es testeable automáticamente, marcar `tool=manual`.

| Criterio de aceptación | Tipo | Tool | Comando/pasos | Resultado esperado |
|---|---|---|---|---|
| <ID + descripción corta> | unit \| api \| e2e \| visual \| manual | go test \| hurl \| playwright \| agent-browser \| manual | comando exacto o pasos numerados | qué evidencia confirma el pass |

**Tipos:**
- `unit` → `go test ./...` (o stack equivalente), incluir package específico.
- `api` → contract test Hurl sobre endpoint MCP o HTTP.
- `e2e` → flujo completo (Playwright Fase 2+; no usar en Fase 1).
- `visual` → `agent-browser` para verificación visual antes del gate humano.
- `manual` → no automatizable; se promueve a `features/manual-checks` en Fase 3 de LEADER-001.

**Cobertura mínima:** todos los ACs deben tener row. SPEC sin esta sección o con ACs sin row → rechazado.

## Requerimientos de observabilidad

<!-- Qué debe emitir esta feature. No es opcional para tareas Medium+. -->
- **Logs:** campos obligatorios en operaciones críticas (ej. `run_id`, `tool_name`, `duration_ms`)
- **Métricas:** counters/gauges que esta feature expone (o "N/A — feature puramente UI")
- **Spans / traces:** si hay operaciones distribuidas, qué spans se crean

## Variables de entorno nuevas

<!-- Listar SOLO las env vars que esta tarea introduce. Si no hay nuevas, escribir "Ninguna". -->
<!-- El developer agrega estas al .env.example con valores placeholder. -->

| Variable | Ejemplo | Secreto | Notas |
|---|---|---|---|
| `VAR_NAME` | `valor-placeholder` | Sí / No | Para qué se usa |

## Límites de implementación

### Siempre hacer
- ...

### Preguntar antes de hacer
- ...

### Nunca hacer
- ...

## Tests esperados

<!-- Lista cerrada. El tester implementa exactamente estos — ni más, ni menos. -->
<!-- Sección 1: unit/integration por stack. Sección 2: automatización (E2E, API, visual, a11y). -->

### Unit / Integration — por stack

#### Tests Go
- `path/to/file_test.go` — `TestFunctionName`: valida que ...

#### Tests Rust
- `src-tauri/tests/file_test.rs` — `test_function_name`: valida que ...

#### Tests React/TS
- `src/features/.../file.test.tsx` — `"description"`: valida que ...

### Automatización

<!-- Evaluar cuáles aplican según la tabla. Si no aplica, escribir "N/A" con justificación. -->

| Tipo | ¿Aplica? | Qué validar |
|---|---|---|
| **E2E web** (Playwright) | Sí / N/A | Flujos: login → ..., checkout → ... |
| **E2E mobile** (Maestro) | Sí / N/A | Flows: ... |
| **API contract** (Hurl) | Sí / N/A | Endpoints: POST /api/..., GET /api/... |
| **Visual regression** | Sí / N/A | Páginas: landing, dashboard |
| **Accesibilidad** (axe) | Sí / N/A | Páginas públicas: ... |

<!-- Detalle de cada tipo que aplica: -->

#### E2E web (si aplica)
- `tests/e2e/feature.spec.ts` — flujo: ...

#### E2E mobile (si aplica)
- `.maestro/feature.yaml` — flow: ...

#### API contract (si aplica)
- `tests/api/resource/crud-flow.hurl` — valida: ...

#### Visual regression (si aplica)
- En test E2E correspondiente con `toHaveScreenshot()`

#### Accesibilidad (si aplica)
- En test E2E correspondiente con axe-core
```

---

## Reglas

- spec.md se genera DESPUÉS de todas las vistas de arquitectura — las referencia, no al revés
- Cada AC debe ser testeable tal cual — "el sistema funciona correctamente" no es un AC
- El mapa de implementación debe listar cada archivo que el developer tocará — sin sorpresas a mitad de tarea
- **Cada archivo con acción CREATE debe tener "Ubicación: por qué aquí"** anclado en un archivo vecino existente o en el patrón del módulo. Sin esa columna llena → SPEC incompleto, el developer rebota la tarea ("SPEC sin justificación de ubicación para `X` — reinvocar architect")
- **La sección "Utils a reutilizar" es obligatoria si el SPEC propone cualquier helper, parser, formatter, validator o util nuevo.** El architect debe ejecutar `Grep` en `internal/util/`, `pkg/util/`, `src/lib/`, `src/utils/` (o equivalente del stack) y reportar lo encontrado. Si existe un util equivalente → reusar (poner en la tabla); si no existe → marcar `NEW` y justificar
- Decisiones de ubicación (en qué paquete/directorio va un archivo nuevo) son **decisión arquitectónica**, no detalle de implementación. El developer NO decide ubicación — solo verifica que el SPEC tenga justificación y que el path exista en disco
- La sección de pre-condiciones es obligatoria — si está vacía, escribir "Ninguna" explícitamente
- El bridge de contratos cross-stack es obligatorio para cualquier tarea que toque 2+ stacks
- La sección de observabilidad es obligatoria para tareas Medium+ — "N/A" requiere justificación explícita
- La sección de variables de entorno es obligatoria — si la tarea no introduce env vars nuevas, escribir "Ninguna" explícitamente. Usar nombres estándar de la tabla en `backend.md` (ej. `REDIS_URL`, no `CACHE_ADDR`)
- "Tests esperados" es la lista cerrada que el tester sigue — el architect define el scope, no el tester
- La sección "Automatización" es obligatoria para tareas Medium+ — evaluar cada tipo y escribir "N/A" con justificación si no aplica. Criterios: nuevo endpoint → API contract, flujo de usuario nuevo → E2E, página pública → a11y, cambio visual → visual regression
- Mantener spec.md bajo 150 líneas — si es más largo, se están duplicando contratos de archivos de arquitectura
