# Template: spec.md

**Generado por:** agente `spec-writer` a partir del contexto disponible (brief, requirements, ADRs, Architecture Views, resumen del explorer, o cualquier combinación).
**Consumido por:** los developers de stack (`developer-backend` / `developer-frontend` / `developer-mobile`), `tester`, `QA`, `task-writer`.

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
| Tests por criterio de aceptación | Siempre |
| Pre-condiciones | Si el cambio tiene dependencias de estado previo |
| Decisiones tomadas (ADR) | Si hay ADRs o decisiones en el brief |
| Mapa de contratos (cross-stack) | Si hay contratos entre componentes (cross-stack o explícitos en ADRs) |
| Requerimientos de observabilidad | Si hay NFRs de observabilidad o el cambio lo amerita |
| Variables de entorno nuevas | Si el cambio introduce env vars |
| Coordinación externa | Si hay dependencias de equipos externos |
| Design references | Si la tarea toca UI |

---

## Template

```markdown
# SPEC: <Nombre del feature> — <TASK-ID>

> Milestone: <milestone> (si existe)

## Fuentes consumidas

| Fuente | Tipo | Origen |
|---|---|---|
| {nombre descriptivo} | brief \| requirements \| arch-view \| adr \| explorer \| design \| repo | {path local, URL, repo, o "inline en el prompt"} |

<!--
Columna **Tipo**: `brief`, `requirements`, `arch-view`, `adr`, `explorer`, `design`, `repo`.
Columna **Origen**: path absoluto local, URL (Linear, GitHub, Figma, Jira), nombre del repo (`github.com/org/repo`), o `inline en el prompt`.
Una fila por fuente consumida. Si solo hubo brief inline, una sola fila con `brief` / `inline en el prompt`.
-->

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

## Criterios de aceptación

<!-- Testeables. Formato: GIVEN / WHEN / THEN. Uno por comportamiento observable. -->
<!-- Obligatorio: cada AC debe incluir una línea "→ Ejemplo:" con input concreto y output esperado. -->
<!-- El ejemplo debe ser verificable por el humano sin leer código. -->
<!-- Agrupar por feature si hay múltiples features en scope. -->

### <Feature 1>

1. GIVEN ... WHEN ... THEN ...
   → Ejemplo: [dato concreto de input] → [resultado observable esperado]

## Señales de alerta

<!-- Comportamientos que NO deben ocurrir. Verificables sin leer código. -->
<!-- Obligatoria para features Medium+. Si no aplica, escribir "Ninguna." -->

- [descripción de lo que no debe pasar]

## Tests por criterio de aceptación

<!-- Una fila por AC declarado. Sin paths de archivos ni comandos exactos — eso es del tester. -->
<!-- El spec declara qué debe verificarse y con qué tipo de test. -->

| AC | Tipo | Qué verifica |
|---|---|---|
| AC-1: <descripción corta> | unit \| api \| e2e \| visual \| manual | descripción del comportamiento que verifica, en lenguaje de dominio |

**Tipos:**
- `unit` → lógica interna aislada
- `api` → contrato HTTP o MCP observable externamente
- `e2e` → flujo completo de usuario
- `visual` → verificación visual de UI
- `manual` → no automatizable; describe los pasos

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

## Design references

<!-- Incluir SOLO si la tarea toca UI (frontend/mobile/fullstack con UI nueva o cambio visual). -->
<!-- Si es backend pura o frontend sin UI nueva, omitir esta sección. -->

- **Type:** figma | pen | screenshots | none
- **Location:** <link de Figma, path al `.pen`, path a screens, o `none`>
- **Notes:** <observaciones opcionales — si vacío: "—">

```

---

## Reglas

- spec.md referencia las vistas de arquitectura cuando existen — no las duplica. Si no hay vistas, trabaja con el contexto disponible (brief, ADRs, resumen del explorer).
- Cada AC debe ser testeable tal cual — "el sistema funciona correctamente" no es un AC
- Decisiones de ubicación (en qué paquete/directorio va un archivo nuevo) son **decisión arquitectónica**, no detalle de implementación. El developer NO decide ubicación — solo verifica que el SPEC tenga justificación y que el path exista en disco
- La sección `## Pre-condiciones` se incluye **solo si el cambio tiene dependencias de estado previo**. Si se incluye y no hay nada concreto que listar, escribir "Ninguna" explícitamente.
- La sección `## Coordinación externa` se incluye **solo si hay dependencias de equipos externos** que bloquean el feature. Si no hay dependencias externas, omitir la sección.
- La sección `## Mapa de contratos (cross-stack)` se incluye **solo si hay contratos entre componentes** (cross-stack o explícitos en ADRs). Para features de un solo stack sin contratos explícitos, omitir la sección.
- La sección `## Requerimientos de observabilidad` se incluye **cuando hay NFRs de observabilidad o el cambio lo amerita**. No es obligatoria por tamaño de tarea — se incluye si el feature introduce logs, métricas o traces nuevos.
- La sección `## Variables de entorno nuevas` se incluye **solo si el cambio introduce env vars nuevas**. Si no hay env vars nuevas, omitir la sección. Usar nombres estándar de la tabla en `backend.md` (ej. `REDIS_URL`, no `CACHE_ADDR`)
- "Tests por criterio de aceptación" es la lista cerrada que el tester sigue — el architect define el scope, no el tester. Una fila por AC, sin excepción.
- Para tareas Medium+: E2E aplica a flujos de usuario nuevos, API contract a endpoints nuevos, a11y a páginas públicas, visual regression a cambios de UI. Justificar "N/A" cuando no aplica.
- Para features de una sola capa: target 100–150 líneas. Si supera 150, revisar si se están duplicando contratos de Architecture Views existentes.
- Para features multi-capa con specs separados: cada archivo individual apunta a 100–150 líneas. El límite aplica por archivo, no al conjunto total de specs del feature.
- **Cada AC debe tener una línea "→ Ejemplo:"** con input concreto y output observable por el humano. Sin ejemplo → AC incompleto.
- **"Señales de alerta" es obligatoria en features Medium+.** Lista lo que NO debe ocurrir — más fácil de detectar en code review que lo que sí debe ocurrir.
- El spec no incluye paths de archivos, nombres de métodos ni comandos de implementación — eso es responsabilidad del developer-agent y el task-writer.
