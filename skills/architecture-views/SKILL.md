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
| overview | `guides/overview.md` | siempre — visión general del sistema y secciones comunes embebidas |

## Estructura mínima de una vista (arc42 + C4)

Cada `arch-<dominio>.md` debe contener al menos estas tres secciones:

````markdown
# Architecture View — <dominio>

> Feature: <feature_id> | Milestone: <milestone>

## 1. Vista (C4 — contexto o contenedores)

<Diagrama Mermaid embebido — nivel C4 apropiado al dominio:>
- Para `arch-backend.md` / `arch-infra.md` → nivel **Contenedores** (servicios, DBs, brokers)
- Para `arch-frontend.md` / `arch-mobile.md` → nivel **Componentes** (jerarquía + estado)
- Para `arch-database.md` → `erDiagram` (entidades + relaciones)

```mermaid
<diagrama>
```

## 2. Componentes principales

<Tabla o lista — para cada componente del diagrama:>
- **Nombre / path** — responsabilidad en una línea
- **Depende de** — lista corta de componentes upstream
- **Expuesto a** — lista corta de componentes downstream

## 3. Atributos de calidad relevantes

<Solo los que aplican al dominio. Una línea por atributo con valor concreto:>
- **Latencia objetivo:** p99 < 300ms (si aplica)
- **Disponibilidad:** SLO 99.9% (si aplica)
- **Seguridad:** auth method, boundaries de confianza
- **Escalabilidad:** dimensión de crecimiento esperado
- **Observabilidad:** logs/metrics/traces que el dominio emite
````

## Reglas duras

1. **Una vista por dominio.** No combinar backend + frontend en el mismo archivo.
2. **Al menos un diagrama Mermaid embebido.** Sin diagrama, no es una vista — es prosa.
3. **Las vistas NO duplican ADRs.** Si la vista necesita justificar una decisión, referenciar el ADR (`Ver ADR-NNN`). El razonamiento vive en el ADR, no en la vista.
4. **Las vistas son ligeras.** Máx 2 páginas por archivo. Si crece más → partir en sub-vistas (`arch-backend-auth.md`, `arch-backend-events.md`).
5. **Diagramas válidos.** Validar la sintaxis Mermaid con la skill `generate-diagram` antes de cerrar el archivo.
6. **Componentes referenciados existen.** Si la vista menciona un path/módulo del repo, debe existir (o estar marcado como `NEW`).

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
- `task-decomposer` — lee `arch-<dominio>.md` para entender capas y componentes, y `adrs/` para entender restricciones de decisiones.
- `dba` / `dba-nosql` — leen `arch-database.md` + ADRs relevantes.
- `diagrammer` — recibe `arch-<dominio>.md` como fuente de vistas a expandir en `.drawio`, y ADRs como contexto de decisiones.
- Cualquier developer / reviewer — la vista es el mapa de orientación; los ADRs explican por qué.

## Checklist de validación (antes de cerrar un `arch-<dominio>.md`)

- [ ] Nombre de archivo: `arch-<dominio>.md` (no `ard-`)
- [ ] Las tres secciones presentes: Vista, Componentes principales, Atributos de calidad
- [ ] Al menos un diagrama Mermaid embebido y válido
- [ ] Componentes referenciados existen en el repo o están marcados `NEW`
- [ ] Atributos de calidad cuantificados cuando aplican (números, no adjetivos)
- [ ] Decisiones que justifican la estructura referencian ADRs por número, no las re-explican
- [ ] Tamaño ≤ 2 páginas
