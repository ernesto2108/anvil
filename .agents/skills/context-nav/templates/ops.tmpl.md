# Operaciones — <ProjectName>

last_updated: <YYYY-MM-DD>

<!-- Comandos reales para levantar, buildear, testear y operar el proyecto.
     Claude los usa directamente — sin preguntar al usuario cómo se hace algo. -->

## Desarrollo local

```bash
# Levantar entorno completo
<comando — ej: make dev / docker-compose up / pnpm dev>

# Solo backend
<comando>

# Solo frontend
<comando>

# Con hot-reload
<comando>
```

## Build

```bash
# Build de producción
<comando — ej: make build / go build ./... / pnpm build>

# Build para plataforma específica
<comando>

# Con optimizaciones / flags de release
<comando>
```

## Tests

```bash
# Correr todos los tests
<comando — ej: make test / go test ./... / pnpm test>

# Tests de un paquete/módulo específico
<comando — ej: go test ./internal/memory/...>

# Con coverage
<comando>

# Con race detector (Go)
<comando — ej: go test -race ./...>

# Tests de integración (si son separados)
<comando>
```

## Lint y formato

```bash
# Lint
<comando — ej: make lint / golangci-lint run / pnpm lint>

# Formatear código
<comando — ej: gofmt -w . / pnpm format>
```

## Base de datos

```bash
# Correr migraciones
<comando — ej: make migrate / goose up>

# Revertir migración
<comando>

# Resetear DB local
<comando>

# Conectarse a la DB local
<comando — ej: sqlite3 ./data/anvil.db>
```

## Docker / docker-compose

```bash
# Levantar todos los servicios
<comando — ej: docker-compose up -d>

# Levantar servicio específico
<comando — ej: docker-compose up -d postgres redis>

# Ver logs
<comando — ej: docker-compose logs -f <service>>

# Detener todo
<comando — ej: docker-compose down>

# Rebuild imagen
<comando — ej: docker-compose build <service> / docker build -t <name> .>

# Ejecutar comando dentro del contenedor
<comando — ej: docker-compose exec <service> sh>
```

## Makefile — targets disponibles

<!-- Listar todos los targets con su descripción real -->

| Target | Qué hace |
|--------|----------|
| `make <target>` | <descripción> |
| `make <target>` | <descripción> |

## Scripts disponibles

<!-- package.json scripts, scripts/ directory, etc. -->

| Script | Comando completo | Qué hace |
|--------|-----------------|----------|
| `<nombre>` | `<comando>` | <descripción> |

## Variables de entorno requeridas

<!-- Solo las que necesitan valor real para que el proyecto corra localmente -->

| Variable | Ejemplo | Para qué |
|----------|---------|----------|
| `<VAR>` | `<valor de ejemplo>` | <descripción> |

## Archivo de env

```bash
# Copiar template
cp .env.example .env

# Ubicación del .env usado por el proyecto
<path — ej: .env / config/.env>
```

## Flujo de trabajo típico

```bash
# 1. Setup inicial (primera vez)
<comandos en orden>

# 2. Desarrollo día a día
<comandos>

# 3. Antes de hacer commit
<comandos — lint, test, etc.>
```
