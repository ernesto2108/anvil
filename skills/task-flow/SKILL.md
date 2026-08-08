---
name: task-flow
description: Ciclo de vida de una tarea de trabajo con seguimiento en Linear y documentación en Outline, en dos fases — abrir el issue al inicio y cerrarlo (PR + doc + estado Done) al final. Úsalo cuando el usuario diga "crea la tarea de X", "abre el issue", "nueva tarea en Linear", "cierra la tarea", "termina TECH-XXX", "documenta en Outline", o al terminar la implementación de una tarea que se abrió con este flujo. Solo aplica en proyectos cuyo backend de documentación, según la tabla de ruteo del registro global de proyectos, es Linear + Outline.
---

# Task Flow (Linear + Outline)

Flujo de dos fases para el ciclo de vida de una tarea en los repos de trabajo:

- **Fase 1 — Abrir:** redactar y crear el issue en Linear.
- **Fase 2 — Cerrar:** commit + PR (delegado), mover el issue a Done, documentar en Outline.

Las fases son independientes: la Fase 2 puede ejecutarse en otra sesión indicando el identifier del issue.

## Filosofía

1. **Escribir en sistemas externos es irreversible en la práctica** — un issue o doc mal creado queda visible para el equipo. Toda creación pasa por confirmación humana previa.
2. **El idioma sigue a la audiencia** — Linear y Outline los lee el equipo en español; git y el remoto se leen en inglés; la conversación con el humano es en español.
3. **No duplicar flujos existentes** — el commit y el PR ya tienen su procedimiento; aquí solo se referencia, nunca se reimplementa.

## Gate 0 — Resolver el backend de docs del proyecto (OBLIGATORIO)

Antes de cualquier paso, determinar el proyecto activo (nombre del directorio raíz del repo) y resolverlo contra la tabla de ruteo del registro global de proyectos (sección "Routing Rules" de `project-registry.md`, en el directorio de configuración de Claude del usuario). La regla de resolución la define el propio registro: **first match wins**.

| Resultado del ruteo | Acción |
|---|---|
| Backend "Linear + Outline" | Continuar con el Gate 1. |
| Vault local, `.workspace/` u otro backend de docs | **DETENER** y reportar al humano, citando la fila del registro que aplicó: ese proyecto usa otro flujo de documentación. No crear nada en Linear ni en Outline. |
| Sin match en la tabla, o más de una fila plausible | **PREGUNTAR** al humano qué backend usar antes de continuar. Este es el único caso donde se pregunta. |

El registro es la única fuente de verdad del ruteo proyecto → backend; esta skill no mantiene copia propia de esa regla ni de patrones de nombre de repo. Sumar un proyecto al flujo = editar una fila del registro, sin tocar esta skill.

## Gate 1 — Credenciales desde el registro (OBLIGATORIO)

Leer en runtime el registro global de proyectos (`project-registry.md` en el directorio de configuración de Claude del usuario), sección de docs backend "Linear + Outline".

De ahí se obtienen, **siempre en runtime y nunca hardcodeados en esta skill ni en ningún archivo del repo**:

- Base URL y token de Linear (GraphQL).
- Base URL y token de Outline (REST).

Reglas:

- La comunicación es por HTTP directo con `curl` (Bash). No usar servidores MCP de Linear/Outline aunque existan en la configuración.
- **Nunca** imprimir tokens en el output, ni escribirlos en archivos, commits, issues o docs.
- Si el registro no existe o no tiene la sección → DETENER y pedir al humano la ubicación de las credenciales.

Forma de las llamadas (los valores entre `<>` salen del registro):

```bash
# Linear — GraphQL
curl -s -X POST <linear_api_url> \
  -H "Authorization: <linear_token>" \
  -H "Content-Type: application/json" \
  -d '{"query": "<query o mutation>"}'

# Outline — REST
curl -s -X POST <outline_api_url>/<endpoint> \
  -H "Authorization: Bearer <outline_token>" \
  -H "Content-Type: application/json" \
  -d '<json>'
```

## Inputs requeridos

### Fase 1

| Campo | Requerido | Fallback si falta |
|---|---|---|
| Descripción de la tarea | siempre | Preguntar al humano. No inventar alcance. |
| Equipo de Linear | no | Usar el equipo principal del registro; resolver su `id` en runtime. |
| ¿Se empieza a trabajar ya? | no | Preguntar en el Paso 1.5 antes de mover el estado. |

### Fase 2

| Campo | Requerido | Fallback si falta |
|---|---|---|
| Identifier del issue (ej. `TECH-410`) | siempre | Intentar detectarlo del nombre de rama; si no aparece, preguntar. |
| Colección de Outline destino | no | Resolver por `collections.list` / `documents.search`; si hay ambigüedad, preguntar. |
| Resumen de lo implementado | siempre | Derivarlo del diff y del PR; confirmar con el humano en el gate del doc. |

## Flujo — Fase 1: abrir la tarea

Triggers: "crea la tarea de X", "abre el issue de X", "nueva tarea en Linear".

### Paso 1.1 — Redactar el issue (en español)

El contenido de Linear va **en español**, siguiendo la convención del equipo.

- **Título:** imperativo, corto, específico y accionable.
  Ejemplos reales del equipo:
  - `Corregir ISR en la tabla de Resumen de cortes: debe permanecer consolidado en Egresos`
  - `Reetiquetar "ISR" a "Retención ISR" en el balance general`
- **Descripción:** tres bloques en markdown.

```markdown
## Contexto
Por qué existe la tarea, qué se observó, dónde.

## Alcance
Qué entra y qué queda explícitamente fuera.

## Criterios de aceptación
- [ ] Condición verificable 1
- [ ] Condición verificable 2
```

Anti-patrones de título — NUNCA: "Arreglar bug", "Cambios varios", "WIP", títulos en inglés, títulos sin objeto ("Revisar").

### Paso 1.2 — Gate: confirmación humana

Mostrar el borrador completo (título + descripción) al humano y **DETENERSE** hasta obtener confirmación explícita. Es una escritura en un sistema externo del equipo.

- Confirma → continuar.
- Pide ajustes → reescribir y volver a mostrar.
- No responde o rechaza → no crear nada.

### Paso 1.3 — Resolver el `teamId`

```graphql
{ teams { nodes { id key name } } }
```

Seleccionar el nodo cuyo `key` coincide con el equipo del registro. Si no aparece → DETENER y reportar los equipos disponibles.

### Paso 1.4 — Crear el issue

Mutation `issueCreate` con `teamId`, `title` y `description`. Pedir de vuelta `{ success issue { id identifier url } }`.

- Éxito → guardar `identifier` (ej. `TECH-410`), `id` y `url`.
- `success: false` o error HTTP → ver "Manejo de errores".

### Paso 1.5 — Arranque del trabajo (opcional)

Preguntar al humano si el trabajo empieza de inmediato.

- Sí → resolver el estado "In Progress" del equipo (`workflowStates`) y aplicarlo con `issueUpdate`. Sugerir nombrar la rama incluyendo el identifier, en el formato `<tipo>/<IDENTIFIER>-<slug>` (ej. `feat/TECH-410-resumen-cortes`), para que el flujo de commit detecte el ticket y lo incluya en el footer.
- No → dejar el issue en el estado inicial.

### Paso 1.6 — Reportar

Máx 80 palabras: identifier, título, URL del issue, estado, y la rama sugerida si aplica.

## Flujo — Fase 2: cerrar la tarea

Triggers: "cierra la tarea", "termina TECH-XXX", o al completar la implementación de una tarea abierta con este flujo.

### Paso 2.1 — Commit y PR (delegado)

Cargar la skill `committer-flow` y seguir su flujo para commit, push y apertura del PR. No duplicar aquí su lógica.

- El contenido de git y del PR va **en inglés** (regla global).
- El footer `Refs <IDENTIFIER>` proviene de la detección de ticket en el nombre de rama que ese flujo ya realiza; por eso importa el nombrado del Paso 1.5.
- Capturar la **URL del PR** — es input de los pasos siguientes.

Si no se obtiene URL de PR → DETENER y preguntar al humano antes de tocar Linear u Outline.

### Paso 2.2 — Actualizar Linear

Sin gate adicional: son consecuencia directa del cierre ya aprobado por el humano.

1. `commentCreate` sobre el issue con el link al PR (en español, una línea, ej. `PR: <url>`).
2. Resolver el estado "Done" del equipo (`workflowStates`) y aplicarlo con `issueUpdate`.

### Paso 2.3 — Documentar en Outline

Un **doc nuevo por tarea**, en español.

1. Resolver la colección del proyecto con `collections.list`; usar `documents.search` para ubicar el doc padre del proyecto si el equipo anida por proyecto. Si hay más de una colección plausible → **preguntar al humano** cuál usar; no elegir por defecto.
2. Título: `<IDENTIFIER> — <título de la tarea>`.
3. Cuerpo:

```markdown
## Qué se hizo
Resumen funcional del cambio.

## Decisiones
Alternativas consideradas y por qué se eligió esta.

## Archivos tocados
- ruta/al/archivo — qué cambió

## Enlaces
- PR: <url del PR>
- Issue: <url del issue de Linear>
```

4. **Gate: mostrar el contenido del doc al humano y esperar confirmación** antes de llamar `documents.create`.
5. Crear con `documents.create` (`title`, `text`, `collectionId`, `parentDocumentId` si aplica, `publish: true`). Guardar la URL devuelta.

### Paso 2.4 — Reportar

Máx 80 palabras, con los tres links:

```
Tarea <IDENTIFIER> cerrada.
- PR: <url>
- Linear: <url> (Done)
- Outline: <url>
```

## Reglas transversales

| Ámbito | Regla |
|---|---|
| Alcance | Solo proyectos cuyo backend de docs en el registro es Linear + Outline (Gate 0). Otro backend: detener; sin match: preguntar. |
| Idioma | Linear y Outline en español; commits, PR y contenido de git en inglés; interacción con el humano en español. |
| Gates de escritura | Crear issue (Paso 1.2) y crear doc (Paso 2.3.4) exigen confirmación previa. Actualizar estado y comentar el PR, no. |
| Credenciales | Siempre leídas del registro en runtime. Nunca hardcodeadas, nunca impresas, nunca commiteadas. |
| Transporte | `curl` HTTP directo. No MCP de Linear/Outline. |
| Reintentos | Ninguno a ciegas ante fallo de red o auth. |

## Manejo de errores

Ante cualquier fallo de `curl` o respuesta con `errors` / `success: false`:

1. Capturar el status HTTP y el cuerpo de la respuesta **sin exponer el header de autorización**.
2. **DETENER** el flujo en ese punto.
3. Reportar al humano: endpoint u operación, status, mensaje de error y en qué paso ocurrió.
4. No reintentar automáticamente. Casos frecuentes:

| Síntoma | Causa probable | Acción |
|---|---|---|
| HTTP 401 / `authentication failed` | Token vencido o mal leído del registro | Reportar y pedir al humano que revise el registro. |
| HTTP 400 con `errors` de GraphQL | Query mal formada o `teamId` inválido | Mostrar el mensaje de GraphQL; corregir la query, no el token. |
| Timeout / DNS | Red | Reportar; el humano decide si reintentar. |
| `collections.list` devuelve varias candidatas | Ambigüedad legítima | Preguntar al humano cuál colección usar. |

Si la Fase 2 falla a mitad (ej. Linear actualizado pero Outline no), reportar **qué quedó aplicado y qué no**, para que el humano pueda retomar sin duplicar.

## Auto-QA antes de cerrar

### Fase 1
- [ ] Gate 0 verificado: el proyecto rutea a backend Linear + Outline según el registro
- [ ] Borrador mostrado y confirmado por el humano antes de crear
- [ ] Título en español, imperativo, específico
- [ ] Descripción con contexto, alcance y criterios de aceptación en checklist
- [ ] `identifier` y URL capturados

### Fase 2
- [ ] PR abierto y URL capturada
- [ ] Contenido de git y del PR en inglés
- [ ] Comentario con el link al PR creado en el issue
- [ ] Issue en estado Done
- [ ] Doc de Outline confirmado por el humano antes de crearse, y creado en la colección correcta
- [ ] Reporte final con los tres links
- [ ] Ningún token apareció en output, archivos ni contenido publicado
