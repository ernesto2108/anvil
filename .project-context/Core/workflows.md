# Workflows del Equipo — anvil

last_updated: 2026-08-13

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

anvil es un monolito CLI — no cargar `cross-service-dev`, no aplica análisis multi-servicio.

## Estrategia de ramas

- **Rama principal:** `master`
- **Ramas de desarrollo:** `develop` (rama activa detectada en este repo)
- **Convención de nombres:** sin convención formal documentada — no hay prefijos `feature/`/`fix/` verificados en el historial
- **Rama de release:** ninguna detectada

## Proceso de PR

<!-- Sin plantilla de PR ni política de reviewers formalizada en el repo -->
1. Crear cambios sobre `develop` o rama de trabajo
2. Abrir PR hacia `master` (o `develop`, según convención del equipo — no verificado explícitamente)
3. Revisión — sin número de reviewers documentado
4. Merge — estrategia (squash/merge/rebase) no documentada en el repo

## Ambientes

anvil es una CLI que se instala localmente (`make install` → `$HOME/bin/anvil`) — no tiene ambientes de deploy tipo dev/staging/production con URLs. El "ambiente" relevante es la máquina local del desarrollador y CI (GitHub Actions, si aplica).

| Ambiente | Rama | URL / Acceso | Deploy |
|---|---|---|---|
| Local (desarrollo) | cualquiera | binario en `$HOME/bin/anvil` | manual vía `make install` |
| CI (migraciones) | cualquiera | `docker-compose.test.yml` | automático en pipeline |

## Proceso de deploy

anvil no tiene proceso de deploy a servidores — es un binario CLI local. El "deploy" es la instalación del binario:

```bash
# Instalar (build con tags requeridos)
make install

# Solo build sin instalar
make build
```

## Comandos operativos

### Desarrollo local

```bash
# Instalar el binario CLI (incluye build tag fts5 obligatorio)
make install

# Dashboard (Wails) — modo dev con HMR del frontend
make dashboard-dev

# Build del dashboard completo (requiere CGO, macOS)
make dashboard-build
```

### Build

```bash
# Build directo (requiere -tags fts5 para SQLite FTS5/sqlite-vec)
go build -tags fts5 -o anvil ./cmd/anvil/
```

### Tests

```bash
# Todos los tests, con detección de race conditions
make test
# equivalente a: go test -race ./...
```

### Lint y formato

```bash
go vet ./...
```

<!-- No se detectó .golangci.yml en la raíz del repo — ver Core/coding-standards.md -->

### Base de datos

```bash
# Correr migraciones — via subcomando propio del CLI
anvil migrate   # implementado en internal/cli/cmd_migrate.go

# Migraciones validadas en CI con docker-compose
docker compose -f docker-compose.test.yml up --build
```

## Variables de entorno requeridas

| Variable | Ejemplo | Para qué |
|---|---|---|
| `ANTHROPIC_API_KEY` | `sk-ant-...` | Habilita llamadas a Claude API (`internal/cli/run.go`, `dream.go`, `emit_translate.go`, `capture.go`); si no está seteada, el flujo cae a Ollama local cuando esté disponible |
