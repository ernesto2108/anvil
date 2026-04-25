# Template: spec.md

**Generado por:** agente architect, después de producir todas las vistas de arquitectura.
**Aprobado por:** usuario (gate de aprobación del SPEC — obligatorio antes de que el developer empiece).
**Consumido por:** developer (input principal), tester (sección de AC), QA (AC + límites).

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

| Archivo | Acción | Qué cambia | Referencia | Fase |
|---|---|---|---|---|
| `path/to/file.go` | MODIFY | Agregar método `UpdateToolUseDuration` | architecture-backend.md §writer | 1 |

## Criterios de aceptación

<!-- Testeables. Formato: GIVEN / WHEN / THEN. Uno por comportamiento observable. -->
<!-- Agrupar por feature si hay múltiples features en scope. -->

### <Feature 1>

1. GIVEN ... WHEN ... THEN ...
2. GIVEN ... WHEN ... THEN ...

### <Feature 2>

3. GIVEN ... WHEN ... THEN ...

## Requerimientos de observabilidad

<!-- Qué debe emitir esta feature. No es opcional para tareas Medium+. -->
- **Logs:** campos obligatorios en operaciones críticas (ej. `run_id`, `tool_name`, `duration_ms`)
- **Métricas:** counters/gauges que esta feature expone (o "N/A — feature puramente UI")
- **Spans / traces:** si hay operaciones distribuidas, qué spans se crean

## Límites de implementación

### Siempre hacer
- ...

### Preguntar antes de hacer
- ...

### Nunca hacer
- ...

## Tests esperados — por stack

<!-- Lista cerrada. El tester implementa exactamente estos — ni más, ni menos. -->
<!-- Agrupar por stack. Cada uno con path de archivo y qué valida. -->

#### Tests Go
- `path/to/file_test.go` — `TestFunctionName`: valida que ...

#### Tests Rust
- `src-tauri/tests/file_test.rs` — `test_function_name`: valida que ...

#### Tests React/TS
- `src/features/.../file.test.tsx` — `"description"`: valida que ...
```

---

## Reglas

- spec.md se genera DESPUÉS de todas las vistas de arquitectura — las referencia, no al revés
- Cada AC debe ser testeable tal cual — "el sistema funciona correctamente" no es un AC
- El mapa de implementación debe listar cada archivo que el developer tocará — sin sorpresas a mitad de tarea
- La sección de pre-condiciones es obligatoria — si está vacía, escribir "Ninguna" explícitamente
- El bridge de contratos cross-stack es obligatorio para cualquier tarea que toque 2+ stacks
- La sección de observabilidad es obligatoria para tareas Medium+ — "N/A" requiere justificación explícita
- "Tests esperados" es la lista cerrada que el tester sigue — el architect define el scope, no el tester
- Mantener spec.md bajo 150 líneas — si es más largo, se están duplicando contratos de archivos de arquitectura
