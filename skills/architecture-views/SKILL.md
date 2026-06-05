---
name: architecture-views
description: Guía al `architect` a producir Architecture Views ligeras (formato arc42 + C4) por dominio. Las vistas capturan la estructura del sistema — el "qué" — y conviven con los ADRs (que capturan el "por qué"). Usar cuando el architect necesite documentar la estructura del sistema, contenedores, componentes o despliegue desde múltiples perspectivas (backend, frontend, mobile, database, infra).
---

# Architecture Views — arc42 + C4 ligero

## Filosofía

Las **Architecture Views** responden a *¿cómo está estructurado el sistema?*. Los **ADRs** responden a *¿por qué está estructurado así?*. Ambos artefactos coexisten y se complementan:

| Artefacto | Pregunta | Formato | Ubicación |
|---|---|---|---|
| Architecture View | ¿Cómo? (estructura) | arc42 + C4 ligero | `arch-<dominio>.md` (raíz o `docs/arch/`) |
| ADR | ¿Por qué? (decisión) | Nygard | `adrs/ADR-NNN-<slug>.md` |

Una vista NO es un ADR agregado — es un mapa estructural por dominio. Un ADR NO es una vista — es el registro de una decisión puntual con contexto, alternativas y consecuencias.

## Cuándo producir vistas

El `architect` produce una Architecture View por cada dominio relevante al feature:

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

El `architect` produce **ambos** artefactos en su run:

1. **Architecture Views** (`arch-<dominio>.md`) — el mapa estructural del sistema desde las perspectivas relevantes al feature.
2. **ADRs** (`adrs/ADR-NNN-<slug>.md`) — los registros de decisión Nygard para cada decisión arquitectónica significativa.

Las vistas primero (estructura), los ADRs emergen mientras se toman las decisiones que dan forma a esa estructura.

## Consumidores aguas abajo

- `spec-writer` — lee **ambos** (vistas + ADRs) para producir `spec.md`.
- `task-writer` — lee `arch-<dominio>.md` para entender capas y componentes, y `adrs/` para entender restricciones de decisiones.
- `dba` / `dba-nosql` — leen `arch-database.md` + ADRs relevantes.
- `diagrammer` — recibe `arch-<dominio>.md` como fuente de vistas a expandir en `.drawio`, y ADRs como contexto de decisiones.
- Cualquier developer / reviewer — la vista es el mapa de orientación; los ADRs explican por qué.

## Checklist de validación (antes de cerrar un `arch-<dominio>.md`)

- [ ] Nombre de archivo: `arch-<dominio>.md` (no `ard-`)
- [ ] Las cuatro secciones obligatorias presentes: **Vista**, **Componentes principales (blackbox)**, **Runtime View**, **Atributos de calidad**
- [ ] Al menos un diagrama Mermaid estructural (Vista) **y** un `sequenceDiagram` (Runtime View), ambos válidos
- [ ] La tabla blackbox cubre nombre/path, responsabilidad, depende de, expuesto a
- [ ] Componentes referenciados existen en el repo o están marcados `NEW`
- [ ] Atributos de calidad cuantificados cuando aplican (números, no adjetivos)
- [ ] Decisiones que justifican la estructura referencian ADRs por número, no las re-explican
- [ ] Tamaño ≤ 2 páginas
