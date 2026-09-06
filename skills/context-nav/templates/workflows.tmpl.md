<!-- SECCIONES FIJAS (preservar literalmente): Modos de trabajo, Reglas por modo, Para agentes
     SECCIONES A RELLENAR (sustituir placeholders <...>): Estrategia de ramas, Proceso de PR, Ambientes, Comandos operativos, Variables de entorno -->

# Workflows del Equipo — <ProjectName>

<!-- Cómo trabaja el equipo: ramas, PRs, ambientes y proceso de deploy.
     También incluye comandos operativos para levantar, buildear, testear y operar. -->

last_updated: <YYYY-MM-DD>

## Modos de trabajo

El equipo opera bajo cinco modos según el tipo de cambio. Cada modo determina qué pasos del workflow son obligatorios y cuáles se omiten.

| Modo | Cuándo usarlo | ¿Actualiza business-rules? | ¿Actualiza contracts? | ¿Crea ADR? | ¿Requiere PR review? |
|---|---|---|---|---|---|
| `feature` | Nueva funcionalidad visible para el usuario o nuevo capability del sistema | sí | sí | solo si hay decisión arquitectónica | sí |
| `bug` | Corrección de comportamiento incorrecto observado en producción o staging | no (si la cambia, escalar a feature) | solo si hay cambio de contrato | solo si hay decisión arquitectónica | sí |
| `fix` | Corrección técnica menor (typo, refactor puntual, ajuste de config) | no | no | no | depende del equipo |
| `chore` | Mantenimiento técnico (upgrade de dependencia, linting masivo, reorganización de carpetas) | no | no | no | depende del equipo |
| `spike` | Investigación o prototipo descartable; no va a producción | no | no | no | no |

### Reglas por modo

**feature** — nueva funcionalidad. Puede cambiar reglas de negocio, contratos, patrones y dominio. Requiere tests, lint, validación con el humano y `reporter` (con diff completo para que actualice `.project-context/`).

**bug** — corrección de comportamiento incorrecto. NO debe cambiar reglas de negocio; si las cambia, escalar a `feature`. Solo actualiza `risks.md` si revela un gotcha nuevo. Requiere tests que reproduzcan el bug, lint y validación con el humano. `reporter` obligatorio.

**fix** — corrección técnica menor. No cambia reglas ni contratos. `reporter` obligatorio.

**chore** — mantenimiento técnico. No cambia comportamiento observable. `reporter` obligatorio.

**spike** — investigación o prototipo. No va a producción. No requiere tests. Solo documentar hallazgos en `runs/` vía `reporter`.

### Para agentes

Al inicio de cualquier run, determinar el modo de trabajo (`feature`, `bug`, `fix`, `chore`, `spike`) antes de implementar: inferirlo del prompt cuando sea inequívoco y declararlo en el bloque de arranque del agente; preguntar explícitamente solo si no es inferible. El modo determina qué pasos del workflow son obligatorios y cuáles se omiten.

Si el cambio toca más de un servicio, cargar `cross-service-dev` antes de implementar — no continuar en modo single-repo.

## Estrategia de ramas

- **Rama principal:** `<main / master>`
- **Ramas de desarrollo:** `<develop / feature branches>`
- **Convención de nombres:** `<feature/TICKET-descripcion / fix/descripcion>`
- **Rama de release:** `<release/vX.X / ninguna>`

## Proceso de PR

1. <paso 1 — ej: crear rama desde develop>
2. <paso 2 — ej: abrir PR con template>
3. <paso 3 — ej: revisión requerida de N personas>
4. <paso 4 — ej: merge con squash / rebase>

## Ambientes

| Ambiente | Rama | URL / Acceso | Deploy |
|---|---|---|---|
| Development | `develop` | `<url o local>` | <automático / manual> |
| Staging | `staging` | `<url>` | <automático / manual> |
| Production | `main` | `<url>` | <manual con aprobación> |

## Proceso de deploy

```bash
# Deploy a staging
<comando>

# Deploy a producción
<comando>
```

## Comandos operativos

### Desarrollo local

```bash
# Levantar entorno completo
<comando>

# Solo backend / Solo frontend
<comando>
```

### Build

```bash
<comando de build>
```

### Tests

```bash
# Todos los tests
<comando>

# Con coverage
<comando>
```

### Lint y formato

```bash
<comando de lint>
```

### Base de datos

```bash
# Correr migraciones
<comando>
```

## Variables de entorno requeridas

| Variable | Ejemplo | Para qué |
|---|---|---|
| `<VAR>` | `<valor>` | <descripción> |
