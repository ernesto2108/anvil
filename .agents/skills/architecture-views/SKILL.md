---
name: architecture-views
description: Guía al `architect` a producir Architecture Views ligeras (formato arc42 + C4) por dominio. Las vistas capturan la estructura del sistema — el "qué" — y conviven con los ADRs (que capturan el "por qué"). Usar cuando el architect necesite documentar la estructura del sistema, contenedores, componentes o despliegue desde múltiples perspectivas (backend, frontend, mobile, database, infra).
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->


# Architecture Views — arc42 + C4 ligero

> **Nota de contexto:** Esta skill se carga en el Paso 5 del flujo del architect, solo si el usuario eligió el formato por defecto en el Paso 3. No se carga automáticamente. Las vistas NO se escriben hasta que el agente pasó por el Paso 3 (confirmar plan) y el Paso 4 (confirmar paths).

## Filosofía

Las **Architecture Views** responden a *¿cómo está estructurado el sistema?*. Los **ADRs** responden a *¿por qué está estructurado así?*. Ambos artefactos coexisten y se complementan:

| Artefacto | Pregunta | Formato | Ubicación |
|---|---|---|---|
| Architecture View | ¿Cómo? (estructura) | arc42 + C4 ligero | `arch-<dominio>.md` (raíz o `docs/arch/`) |
| ADR | ¿Por qué? (decisión) | Nygard | `adrs/ADR-NNN-<slug>.md` |

Una vista NO es un ADR agregado — es un mapa estructural por dominio. Un ADR NO es una vista — es el registro de una decisión puntual con contexto, alternativas y consecuencias.

## Cuándo producir vistas

Cuando se usa esta skill, el `architect` produce una Architecture View por cada dominio relevante al feature:

- `arch-backend.md` — servicios, módulos, capas, integraciones backend
- `arch-frontend.md` — jerarquía de componentes, routing, estado, contratos con backend
- `arch-mobile.md` — pantallas, navegación, estado, integraciones nativas
- `arch-database.md` — entidades, relaciones, particionamiento, estrategia de migración
- `arch-infra.md` — topología de despliegue, redes, observabilidad, secrets
- `arch-api.md` — *(opcional)* contratos de API como producto (SDK público, OpenAPI compartido entre múltiples consumidores). Solo si la API es dominio central; si es endpoint interno, documentar dentro de `arch-backend.md`.
- `arch-auth.md` — *(opcional)* identidad, autorización, tokens, sesiones. Solo si auth es dominio central; si es solo un guard sobre un endpoint, documentar dentro de `arch-backend.md`.
- `arch-<otro-dominio>.md` — cualquier dominio cohesivo del sistema

Una vista cubre **una perspectiva** del sistema. No mezclar dominios en un solo archivo.

## Nomenclatura — `arch-<dominio>.md` (NUNCA `ard-<dominio>.md`)

El nombre de los archivos es **siempre** `arch-<dominio>.md`. El prefijo `ard-` está deprecado y prohibido.

## Guides por dominio

Cargar el guide correspondiente al dominio antes de producir la vista. Los guides contienen el template detallado con secciones específicas, tabla de archivos involucrados, NFRs cuantificadas y contratos de interfaz.

| Dominio | Guide | Cuándo cargar |
|---|---|---|
| backend | `guides/backend.md` | al producir `arch-backend.md` |
| frontend | `guides/frontend.md` | al producir `arch-frontend.md` |
| mobile | `guides/mobile.md` | al producir `arch-mobile.md` |
| database | `guides/database.md` | al producir `arch-database.md` |
| infrastructure | `guides/infrastructure.md` | al producir `arch-infra.md` |
| api | `guides/api.md` | al producir `arch-api.md` |
| auth | `guides/auth.md` | al producir `arch-auth.md` |
| overview | `guides/overview.md` | referencia — secciones comunes y convenciones transversales; no produce `arch-overview.md` |

## Estructura mínima de una vista (arc42 + C4)

> **Las Architecture Views documentan la estructura estable del sistema, no el scope de un feature puntual.** Lo que entra y sale en un cambio concreto (in/out of scope, tabla de archivos modificados) pertenece a `spec.md`, no a una vista. Una vista refleja cómo está organizado el sistema hoy y por qué; un spec refleja qué se está cambiando en este run.

Cada `arch-<dominio>.md` debe contener al menos estas cuatro secciones. Los nombres son canónicos y deben aparecer literalmente como encabezados en cada vista — los guides pueden adaptar el contenido al dominio pero NO renombrar las secciones obligatorias.

### Mapeo arc42 ↔ C4 (referencia)

| arc42 | C4 | Sección en la vista |
|---|---|---|
| § 3 Context & Scope | C4 System Context (L1) | Vista (cuando el dominio es el sistema completo) |
| § 5 Building Block View L1 | C4 Container (L2) | Vista (backend, infra) — whitebox del sistema |
| § 5 Building Block View L2 | C4 Component (L3) | Vista (frontend, mobile) — whitebox de un container |
| § 6 Runtime View | C4 Dynamic | **Runtime View** (obligatoria) |
| § 7 Deployment View | C4 Deployment | Vista (infra) |

El mínimo obligatorio según C4 es: System Context + Container + Deployment + ADRs + al menos un Runtime scenario.

````markdown
# Architecture View — <dominio>

> Feature: <feature_id> | Milestone: <milestone>

## 1. Vista (C4 — contexto, contenedores o componentes)

<Diagrama Mermaid embebido — nivel C4 apropiado al dominio:>
- Para `arch-backend.md` / `arch-infra.md` → nivel **Contenedores** (servicios, DBs, brokers)
- Para `arch-frontend.md` / `arch-mobile.md` → nivel **Componentes** (jerarquía + estado)
- Para `arch-database.md` → `erDiagram` (entidades + relaciones)

Esta es la vista **whitebox** del dominio: muestra la estructura interna y justifica la descomposición.

```mermaid
<diagrama>
```

## 2. Componentes principales (blackbox)

<Tabla blackbox — descripción externa de cada bloque del diagrama. Cubrir: nombre/path, responsabilidad (una línea), depende de, expuesto a. Para dominios con endpoints/screens/módulos, la tabla puede listar endpoints (api), screens (mobile/frontend) o módulos (backend) según corresponda.>

| Componente / path | Responsabilidad | Depende de | Expuesto a |
|---|---|---|---|

## 3. Runtime View (escenario principal)

<Diagrama de secuencia o flujo — arc42 § 6 / C4 Dynamic. Mostrar el happy path del escenario más importante del dominio. Si hay un path de fallo crítico (timeout, retry, fallback) incluirlo también.>

```mermaid
sequenceDiagram
  ...
```

## 4. Atributos de calidad relevantes

<Equivale a la tabla "Restricciones no-funcionales" (NFRs) presente en todos los guides. Una línea por atributo con valor cuantificado:>
- **Latencia objetivo:** p99 < 300ms (si aplica)
- **Disponibilidad:** SLO 99.9% (si aplica)
- **Seguridad:** auth method, boundaries de confianza
- **Escalabilidad:** dimensión de crecimiento esperado
- **Observabilidad:** logs/metrics/traces que el dominio emite
````

> **Equivalencias entre el SKILL y los guides:**
> - "Componentes principales (blackbox)" ↔ la tabla de módulos / endpoints / screens / entidades que cada guide describe en su template.
> - "Atributos de calidad relevantes" ↔ la tabla **Restricciones no-funcionales** presente en todos los guides.
> - "Runtime View" ↔ secciones tipo "Comportamiento runtime" / "Flujo principal" / `sequenceDiagram` presentes en los guides.

## Reglas duras

1. **Una vista por dominio.** No combinar backend + frontend en el mismo archivo.
2. **Al menos un diagrama Mermaid embebido en la sección Vista** (whitebox del dominio) **y un `sequenceDiagram` en la sección Runtime View**. Sin diagrama estructural + runtime, no es una vista — es prosa.
3. **Las vistas NO duplican ADRs.** Si la vista necesita justificar una decisión, referenciar el ADR (`Ver ADR-NNN`). El razonamiento vive en el ADR, no en la vista.
4. **Las vistas son ligeras.** Máx 2 páginas por archivo. Si crece más → partir en sub-vistas (`arch-backend-auth.md`, `arch-backend-events.md`).
5. **Diagramas válidos.** Validar la sintaxis Mermaid con la skill `generate-diagram` antes de cerrar el archivo.
6. **Componentes referenciados existen.** Si la vista menciona un path/módulo del repo, debe existir (o estar marcado como `NEW`).
7. **Whitebox + blackbox.** La sección Vista muestra la estructura interna (whitebox). La tabla Componentes principales describe externamente cada bloque (blackbox: responsabilidad + dependencias + interfaces expuestas).

## Relación con los ADRs

- La vista lista los componentes; los ADRs justifican las decisiones que dieron forma a esos componentes.
- Si una decisión estructural cambia (ej. se introduce un broker entre dos servicios) → primero se escribe el ADR, luego se actualiza la vista.
- Si una vista contradice un ADR → la vista está desactualizada, corregir la vista, no el ADR.

## Output del architect

Cuando se usa esta skill (formato por defecto confirmado en el Paso 3), el `architect` produce **ambos** artefactos en su run:

1. **Architecture Views** (`arch-<dominio>.md`) — el mapa estructural del sistema desde las perspectivas relevantes al feature.
2. **ADRs** (`adrs/ADR-NNN-<slug>.md`) — los registros de decisión Nygard para cada decisión arquitectónica significativa.

Las vistas primero (estructura), los ADRs emergen mientras se toman las decisiones que dan forma a esa estructura.

## Consumidores aguas abajo

- `spec-writer` — lee **ambos** (vistas + ADRs) para producir `spec.md`.
- `task-writer` — lee `arch-<dominio>.md` para entender capas y componentes, y `adrs/` para entender restricciones de decisiones.
- `dba` / `dba-nosql` — leen `arch-database.md` + ADRs relevantes.
- `diagrammer` — recibe `arch-<dominio>.md` como fuente de vistas a expandir en `.drawio`, y ADRs como contexto de decisiones.
- Cualquier developer / reviewer — la vista es el mapa de orientación; los ADRs explican por qué.

## Diagramas embebidos en ADRs

Cuando un ADR se beneficie de una visualización (flujo, secuencia, estados, schema), incluir al menos un diagrama Mermaid embebido en `## Context` o `## Decision`. Cargar la skill `generate-diagram` para validar la sintaxis.

**Reglas duras:**

1. ADRs que documentan flujo de datos, comunicación entre componentes o ciclo de vida → incluir al menos un diagrama.
2. ADRs de persistencia / schema → incluir `erDiagram` con las entidades del cambio.
3. Cada diagrama debe pasar el checklist de validación de `generate-diagram`.
4. Si el diagrama excede el alcance de Mermaid → escalar al humano: `Pregunta abierta: el diagrama de [X] requiere drawio standalone — ¿quieres que lo produzca el agente diagrammer?`.

## Gate de verificación de estado real (pre-decisión)

Este gate corre **ANTES de tomar y escribir las decisiones** — el orden es verificar estado real → decidir, nunca decidir → verificar. Ninguna vista ni ADR debe afirmar estado del sistema sin haberlo verificado primero:

- Usar `Glob` para verificar que directorios/archivos referenciados existen.
- Usar `Grep` para confirmar que tipos/interfaces que referencias realmente existen.
- Si un path NO existe, marcarlo explícitamente como `NEW`.
- Verificar afirmaciones de estado — si la decisión dice "agregar X", `Grep` literal X primero para confirmar que NO existe.
- Si la decisión toca persistencia, leer el schema/migraciones directamente del repo (ver la regla de schema DB en `adr-writer`).

Estas lecturas dirigidas cuentan dentro del presupuesto global de tool calls del run y previenen re-invocaciones del developer — verificar antes es siempre más barato que iterar después. Al cerrar cada archivo, re-confirmar solo lo que haya cambiado durante la escritura.

### Reconocimiento obligatorio para archivos NEW (decisión de ubicación)

Para CADA archivo NEW que se introduzca:

1. **Listar el directorio destino** con `LS` (1 call). Si el directorio no existe, listar el directorio padre y justificar la creación.
2. **Leer 1 archivo vecino** (1 call) para identificar el patrón local (naming, organización por concern).
3. **Registrar la justificación** inline (en el ADR, dentro de `## Implementation notes`).

**Límite duro:** máximo 2 calls por archivo NEW. Si necesitas más exploración → escalar al humano con `Pregunta abierta: necesito que el explorer evalúe duplicados/utils existentes para [archivo NEW] en [áreas]`.

## Consistencia de contratos cross-ADR

Cuando múltiples ADRs tocan contratos relacionados (ej. backend define el schema, frontend define el tipo TS, mobile define el modelo Dart):

- Definir el contrato canónico UNA VEZ en el ADR primario (típicamente el backend).
- Los ADRs secundarios (frontend, mobile, infra) referencian el ADR primario por su número.
- Nunca duplicar la definición de contrato con formas diferentes entre ADRs.

## Checklist de validación (antes de cerrar un `arch-<dominio>.md`)

- [ ] Nombre de archivo: `arch-<dominio>.md` (no `ard-`)
- [ ] Las cuatro secciones obligatorias presentes: **Vista**, **Componentes principales (blackbox)**, **Runtime View**, **Atributos de calidad**
- [ ] Al menos un diagrama Mermaid estructural (Vista) **y** un `sequenceDiagram` (Runtime View), ambos válidos
- [ ] La tabla blackbox cubre nombre/path, responsabilidad, depende de, expuesto a
- [ ] Componentes referenciados existen en el repo o están marcados `NEW`
- [ ] Atributos de calidad cuantificados cuando aplican (números, no adjetivos)
- [ ] Decisiones que justifican la estructura referencian ADRs por número, no las re-explican
- [ ] Tamaño ≤ 2 páginas

## Gate de handoff al spec-writer (antes de cerrar el run completo)

Verificar antes de reportar al humano:

- [ ] Cada **Architecture View** pasa el checklist de validación de arriba
- [ ] Hay al menos una Architecture View por dominio relevante al feature (backend, frontend, mobile, database, infra según aplique)
- [ ] Cada ADR pasa el checklist de la skill `adr-writer`
- [ ] Las Views referencian ADRs por número cuando dependen de una decisión registrada (no re-explican la decisión)
- [ ] Cada archivo NEW referenciado tiene justificación de ubicación
- [ ] NFRs de `requirements.md` relevantes están cubiertos por al menos un ADR (latencia p99, SLO de disponibilidad, etc., con número concreto o `N/A` con justificación)
- [ ] Contratos cross-ADR consistentes (no duplicar definiciones con formas distintas)
- [ ] Para tareas Medium+ con componentes desplegables → existe al menos un ADR de infraestructura (topología, env vars, observabilidad, rollback). Excepción: tareas que no introducen ningún componente desplegable.
- [ ] Sección `## Preguntas abiertas` presente en el mensaje de cierre (con contenido o con "Ninguna")

Si algún ítem falta → completarlo antes de entregar al humano.
