# Template: spec.md

**Generado por:** agente `spec-writer`, después de recibir el ARD del `architect`.
**Consumido por:** los developers de stack (`developer-backend` / `developer-frontend` / `developer-mobile`), `tester`, `QA`, `task-decomposer`.

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

## Secciones del spec.md

| Sección | Condición de inclusión |
|---|---|
| Contexto y objetivo | Siempre |
| No-objetivos | Siempre |
| Criterios de aceptación | Siempre |
| Testing strategy | Siempre |
| Pre-condiciones | Si el cambio tiene dependencias de estado previo |
| Decisiones / alternativas consideradas | Si hay ADRs o decisiones en el brief |
| Mapa de contratos | Si hay contratos entre componentes (cross-stack o explícitos en ADRs) |
| Mapa de implementación | Si hay Architecture Views, ADRs, o el brief es suficientemente detallado |
| Observabilidad | Si hay NFRs de observabilidad o el cambio lo amerita |
| Variables de entorno nuevas | Si el cambio introduce env vars |
| Coordinación externa | Si hay dependencias de equipos externos |
| Design references | Si la tarea toca UI |

---

## Template

```markdown
# SPEC: <Nombre del feature> — <TASK-ID>

> Inputs consumidos: {lista de lo que existía — requirements.md / ADRs / Architecture Views / brief inline}
> Milestone: <milestone> (si existe)

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

## Coordinación externa

<!-- Solo si hay dependencias de equipos externos que bloquean esta tarea. Si no hay, escribir "Ninguna". -->

| Qué | Responsable | Deadline | Estado |
|---|---|---|---|
| Migración `add_notifications_table` | @equipo-db | YYYY-MM-DD | pendiente |

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

## Tests por criterio de aceptación (OBLIGATORIO)

<!-- Una fila por AC declarado arriba. Sin filas vacías. -->
<!-- Si no es automatizable: tool=manual. Todos los ACs deben tener row → SPEC sin row por AC es rechazado. -->

| AC | Tipo | Tool | Test ID / archivo | Comando | Resultado esperado |
|---|---|---|---|---|---|
| AC-1: <descripción corta> | unit \| api \| e2e \| visual \| manual | go test \| hurl \| playwright \| agent-browser \| manual | `path/file_test.go::TestName` o `tests/api/resource.hurl` | comando exacto o pasos numerados | qué evidencia confirma el pass |

**Tipos:**
- `unit` → `go test ./...` (o stack equivalente), incluir package específico.
- `api` → contract test Hurl sobre endpoint MCP o HTTP.
- `e2e` → flujo completo (Playwright Fase 2+; no usar en Fase 1).
- `visual` → `agent-browser` para verificación visual antes del gate humano.
- `manual` → no automatizable; se promueve a `features/manual-checks` en Fase 3 de LEADER-001.

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
- "Coordinación externa" es obligatoria cuando la tarea tiene dependencias de equipos externos (migraciones manuales, aprobaciones, configuraciones de infra) — si no hay, escribir "Ninguna"
- El bridge de contratos cross-stack es obligatorio para cualquier tarea que toque 2+ stacks
- La sección de observabilidad es obligatoria para tareas Medium+ — "N/A" requiere justificación explícita
- La sección de variables de entorno es obligatoria — si la tarea no introduce env vars nuevas, escribir "Ninguna" explícitamente. Usar nombres estándar de la tabla en `backend.md` (ej. `REDIS_URL`, no `CACHE_ADDR`)
- "Tests por criterio de aceptación" es la lista cerrada que el tester sigue — el architect define el scope, no el tester. Una fila por AC, sin excepción.
- Para tareas Medium+: E2E aplica a flujos de usuario nuevos, API contract a endpoints nuevos, a11y a páginas públicas, visual regression a cambios de UI. Justificar "N/A" cuando no aplica.
- Mantener spec.md bajo 150 líneas — si es más largo, se están duplicando contratos de archivos de arquitectura
