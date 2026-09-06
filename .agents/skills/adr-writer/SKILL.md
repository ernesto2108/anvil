---
name: adr-writer
description: Guía al `architect` a producir ADRs (Architecture Decision Records) en formato Nygard estándar. Define el formato exacto, qué decisiones ameritan un ADR, la regla de contratos-no-código, el orden de generación, la regla de schema DB y el presupuesto de tokens. Cargar antes de producir cualquier ADR. Úsalo cuando el usuario diga 'escribir ADR', 'documentar decisión técnica', 'Architecture Decision Record', 'registrar decisión', o cuando el architect tome una decisión de diseño relevante.
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->


# ADR Writer — formato Nygard estándar

> **Nota de contexto:** Esta skill se carga en el Paso 5 del flujo del architect, solo si el usuario eligió el formato por defecto en el Paso 3. No se carga automáticamente.

## Filosofía

Un **ADR** responde a *¿por qué está estructurado así?*. Captura el razonamiento detrás de una decisión arquitectónica significativa: contexto + decisión + alternativas + consecuencias. Convive con las Architecture Views (que capturan el *qué* / estructura).

Una decisión por ADR, nunca agregadas. Los ADRs son conversacionales con el developer futuro: 1-2 páginas, concisos, honestos sobre trade-offs.

## Mentalidad — orden de diseño

Siempre seguir este orden al razonar antes de escribir un ADR:

1. Diseño de sistema (alto nivel)
2. Límites y dominios
3. Contratos (specs ejecutables cuando aplique)
4. Comportamiento runtime
5. Infraestructura y operaciones
6. Solo entonces → hints de implementación

Nunca empezar desde la estructura de código.

## Cuándo producir un ADR

**Sí — un ADR (archivo) por decisión cuando:**

- La decisión es **arquitectónica relevante** al feature: contratos de comunicación (REST / eventos / gRPC), estrategia de persistencia y schema, modelo de auth, topología de despliegue, jerarquía de componentes frontend/mobile, ubicación de archivos NEW que no es obvia, taxonomía de errores, patrón de integración (sync/async).
- La decisión **se desvía de convenciones del proyecto** (stack, patrones, naming, manejo de errores).
- La decisión **afecta otros equipos/servicios** o introduce un patrón nuevo.

**No — no requiere ADR cuando:**

- La decisión es trivial y se alinea con convenciones existentes.

**Atomicidad:** una decisión por ADR. Si tienes dos decisiones independientes (ej. "usamos Postgres" y "indexamos por `tenant_id`"), son dos ADRs. Si son dependientes (ej. "usamos event sourcing → necesitamos outbox pattern"), pueden vivir en el mismo ADR si la segunda es consecuencia directa de la primera.

## Ruta y nomenclatura

- **Ruta de output:** `{task_path}/adrs/`
- **Archivo:** `ADR-NNN-<slug>.md` (ej. `ADR-001-estrategia-cache.md`)
- **Numeración:** secuencial (001, 002, 003, …)
- **Slug:** kebab-case en español

## Formato Nygard estándar (obligatorio)

```markdown
# ADR-NNN — <Título en español>

> Milestone: <milestone> | Motivado por: FR-XX, NFR-YY (si aplica)

## Status
Accepted

## Context
<2-5 párrafos: fuerzas en juego, restricciones técnicas y de negocio,
supuestos, estado actual del sistema. Aquí pueden vivir diagramas Mermaid,
firmas de tipos como evidencia, y referencias a paths concretos del repo
verificados con Glob/Grep.>

## Decision
<La decisión tomada en una o dos oraciones claras. Sin ambigüedad.
Si hay alternativas consideradas, listarlas brevemente con el motivo
de descarte.>

## Consequences
**Positivas:**
- <consecuencia 1>

**Negativas / Trade-offs aceptados:**
- <trade-off 1>

<Opcional: ## Implementation notes con invariantes, archivos NEW con
justificación de ubicación, o referencias cruzadas a otros ADRs.>
```

### Reglas duras

- **Una decisión por ADR.** 1-2 páginas máx.
- **Idioma:** contenido en español. Solo quedan en inglés los nombres canónicos Nygard (`## Status`, `## Context`, `## Decision`, `## Consequences`), identificadores de operación (`CREATE`, `MODIFY`, `DELETE`), paths e identificadores técnicos (FR-XX, NFR-YY, IDs de tipo).
- **Trazabilidad a requirements:** si el ADR está motivado por uno o más FR/NFR de `requirements.md`, incluirlos en el encabezado como `Motivado por: FR-XX, NFR-YY`. Si es puramente técnica, omitir el campo.
- **Milestone obligatorio** en el encabezado.
- **Excepciones a convenciones:** si una decisión contradice una convención del proyecto, el ADR debe argumentar en `## Decision` por qué la excepción se justifica.
- **Archivos NEW:** registrar path + justificación de ubicación en `## Implementation notes`. Ejemplo: *"`internal/dashboard/cache/store.go` (NEW) — sigue patrón de `internal/dashboard/store/X.go` (mismo bounded context — persistencia)"*. No se acepta justificación vacía ni genérica.

## Contratos, no código (REGLA DURA)

El ADR son **decisiones y contratos** — no un borrador de código. Código que el developer copie verbatim está fuera de scope.

**PUEDE incluir:**

- Firmas de tipos y contratos de interfaces (Go structs, TS interfaces, listas de columnas SQL) — **solo declaraciones, sin cuerpos** — como evidencia del `## Context` o ilustración del `## Decision`.
- **Firmas** de funciones/métodos (nombre, params, tipos de retorno, invariantes) — no implementaciones.
- Fragmentos OpenAPI / AsyncAPI (YAML) para contratos de API/eventos — specs ejecutables, no prosa.
- **Intención** SQL en pseudo-código, DBML, o formato anotado — la query exacta es trabajo del developer.
- Diagramas Mermaid (C4, secuencia, flowchart, estado, ERD) embebidos en `## Context` o `## Decision`.
- Tablas de invariantes y taxonomía de errores (enum/códigos, no strings).

**NO DEBE escribir:**

- **Cuerpos** de funciones/métodos — nada de `{ ... return dto }`.
- Nombres de helpers que prescriban implementación (ej. `calcDeltaPct`) — el developer elige.
- Queries SQL completas con sintaxis de driver (`?`, `$1`, `:named`).
- Import paths — el developer verifica qué existe.
- Strings de error o mensajes de log — las convenciones controlan eso.
- Casos de test completos.

**Si sientes la tentación de escribir un detalle de implementación**, regístralo como **invariante** dentro de `## Decision` o `## Consequences`.

- Mal: `func calcDeltaPct(cur, prev int) *float64 { if prev == 0 { return nil }; ... }`
- Bien: Invariante: *"El delta porcentual es nil cuando la línea base es cero o no existe. No-nil en otros casos, calculado como `(current - baseline) / baseline`."*

## Orden de generación

Los ADRs se generan en el orden en que la decisión aparece en la cadena de impacto: **datos → backend → contratos → consumidores**. Numeración secuencial.

El `spec.md` NO entra en este orden — lo produce el `spec-writer` aparte.

## Regla de schema DB (CRÍTICA)

**NUNCA proponer una tabla nueva sin verificar primero el schema real.** La memoria del humano NO es fuente de verdad del schema — el repo sí.

Antes de diseñar cualquier cambio de DB:

1. **Leer el schema/migraciones directamente:** localizar con Glob/Grep dirigidos dónde vive el schema (directorios de migraciones, `schema.sql`, DBML, modelos ORM) y leer las tablas relacionadas al cambio. Si el schema no vive en el repo o la lectura dirigida no basta → pedir el artefacto de hallazgos del explorer con el schema relevante; nunca sustituirlo por lo que el humano recuerde.
2. Evaluar si `ALTER TABLE` (agregar columnas) sobre una tabla existente resuelve el problema.
3. Solo proponer tabla nueva si hay justificación técnica clara Y el humano confirma la propuesta.

### Qué confirmar con el humano (solo lo que el repo no responde)

1. Si la evaluación apunta a tabla nueva: "Verifiqué el schema — existen <tablas leídas>. ¿Extendemos <tabla candidata> o hay justificación para una tabla separada (ej. relación N:M, entidad con ciclo de vida propio)?"
2. Nunca preguntar "¿qué tablas existen?" — eso se lee del repo, no se pregunta. Presentar siempre lo leído como evidencia y pedir confirmación de la decisión, no del estado.

## Estrategia de migración — verificar, luego confirmar (CRÍTICA)

No todos los proyectos usan archivos de migración en el repo. Antes de diseñar la estrategia de persistencia:

1. **Verificar en el repo primero:** ¿existe un directorio o herramienta de migraciones (Glob dirigido)? ¿Qué formato y convención siguen las migraciones existentes?
2. **Confirmar con el humano solo lo que el repo no puede responder:** el estado de la DB y, si no se detectó sistema de migraciones, cómo se gestionan los cambios de schema.

### Preguntas a hacer al humano (solo las que el repo no responde)

1. Solo si no se detectó sistema de migraciones en el repo: "No encontré migraciones en el repo — ¿cómo se gestionan los cambios de schema? (SQL manual / herramienta de sync / la DB ya existe sin sistema de migraciones)"
2. "¿En qué estado está la DB? (nueva sin datos / existente solo en dev / existente en producción con datos reales)"

Si la DB ya existe en producción, incluir en el ADR de persistencia:

- **Riesgos de deploy:** bloqueos de tabla, downtime, incompatibilidad con código actual.
- **Orden de ejecución:** ¿migración antes o después del deploy de código?
- **Plan de rollback.**
- **Backfill** si hay datos existentes que deben transformarse.

## Presupuesto de tokens

- **Objetivo:** 25K tokens | **Máximo:** 40K tokens
- **Llamadas a tools de lectura/verificación:** comparten el presupuesto global del run (máx 15 tool calls — ver `docs/token-optimization.md`), incluyendo el gate de verificación pre-decisión (schema, paths, tipos) y el reconocimiento por archivo NEW (LS + 1 vecino). La verificación pre-decisión tiene prioridad sobre cualquier otra lectura: nunca recortarla para ahorrar llamadas — si el presupuesto no alcanza, escalar al humano en vez de decidir sin verificar.
- **Máx ADRs por run:** 12.

## Checklist de validación (antes de cerrar un ADR)

- [ ] Formato Nygard completo (`## Status`, `## Context`, `## Decision`, `## Consequences`)
- [ ] Una sola decisión por archivo
- [ ] Milestone en el encabezado
- [ ] `Motivado por: FR-XX, NFR-YY` cuando aplica
- [ ] Sin cuerpos de funciones ni queries con sintaxis de driver
- [ ] Archivos NEW con justificación de ubicación en `## Implementation notes`
- [ ] Contratos consistentes con otros ADRs del mismo run (contrato canónico definido UNA VEZ, los demás lo referencian por número)
- [ ] Diagramas Mermaid embebidos cuando aplica (ver `architecture-views` para reglas de diagramas)
- [ ] Tamaño ≤ 2 páginas
