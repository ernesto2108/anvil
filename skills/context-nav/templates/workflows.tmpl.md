# Workflows del Equipo — <ProjectName>

<!-- Cómo trabaja el equipo: ramas, PRs, ambientes y proceso de deploy.
     También incluye comandos operativos para levantar, buildear, testear y operar. -->

last_updated: <YYYY-MM-DD>

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
