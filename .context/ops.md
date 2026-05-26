# Operaciones — Anvil

last_updated: 2026-05-08

## Desarrollo local

```bash
# Build e instalar CLI en $HOME/bin (target por defecto)
make install

# Build local sin instalar (produce ./anvil)
make build

# Levantar Vite dev server del dashboard frontend (HMR en http://localhost:5173)
make dashboard-dev
# Nota: la ventana nativa requiere build separado — ver sección Dashboard abajo
```

## Build

```bash
# Build CLI (sin dashboard) — instala en $HOME/bin/anvil
make install
# go build -tags fts5 -o $HOME/bin/anvil ./cmd/anvil/

# Build solo local sin instalar
make build
# go build -tags fts5 -o ./anvil ./cmd/anvil/

# Build binario completo con dashboard embebido (requiere CGO, solo macOS)
make dashboard-build
# Produce: ./anvil-full
# go build -tags "dashboard production fts5" -o anvil-full ./cmd/anvil/

# Build frontend React (embebido en anvil-full)
make dashboard-frontend
# cd frontend && npm install --ignore-scripts && npm run build
```

## Tests

```bash
# Todos los tests con race detector
make test
# go test -race ./...

# Tests de un dominio específico
go test -race ./internal/memory/...
go test -race ./internal/orchestrator/...
go test -race ./internal/instrumentation/...

# Tests de integración (usan docker-compose.test.yml)
# Ver docker-compose.test.yml para setup requerido
```

## Lint y formato

```bash
# Vet estático
make vet
# go vet ./...

# Format (convención Go estándar)
gofmt -w .
```

## Base de datos

```bash
# Las migraciones corren automáticamente al inicio de anvil
# Path de la DB: determinado por config.App en runtime (ver pkg/config/)

# Conectarse a la DB local (path real depende de la config)
sqlite3 ~/.local/share/anvil/anvil.db

# Para tests: SQLite in-memory — los tests crean su propia DB
```

## Docker

```bash
# Solo para tests de integración
docker-compose -f docker-compose.test.yml up -d
docker-compose -f docker-compose.test.yml down
```

## Dashboard (Wails)

```bash
# Instalar CLI de Wails (solo primera vez)
make dashboard-install-cli

# Dev: solo el frontend con HMR
make dashboard-dev          # Vite en http://localhost:5173

# Build completo con ventana nativa
make dashboard-build        # Produce ./anvil-full
./anvil-full dashboard      # Lanzar ventana nativa
```

## Makefile — targets disponibles

| Target | Qué hace |
|--------|----------|
| `make install` | Build con fts5 e instala en `$HOME/bin/anvil` (default) |
| `make build` | Build con fts5 en `./anvil` sin instalar |
| `make clean` | Elimina `./anvil` del directorio local |
| `make test` | `go test -race ./...` |
| `make vet` | `go vet ./...` |
| `make dashboard-install-cli` | Instala Wails CLI en `$GOBIN` |
| `make dashboard-frontend` | Build del frontend React en `frontend/dist` |
| `make dashboard-build` | Build completo de `./anvil-full` con dashboard embebido |
| `make dashboard-dev` | Vite dev server en `http://localhost:5173` |

## Variables de entorno requeridas

| Variable | Ejemplo | Para qué |
|----------|---------|----------|
| `ANTHROPIC_API_KEY` | `sk-ant-...` | Summarización con Claude Haiku (opcional si se usa Ollama) |
| `ANVIL_VAULT_PATH` | `/Users/user/vault` | Path al vault Obsidian (opcional — para herramientas MCP de contexto) |
| `CLAUDE_HOME` | `/Users/user/.claude` | Override del home de Claude Code (default: `~/.claude`) |
| `ANVIL_PARENT_RUN_ID` | `run-abc123` | Propagado automáticamente a subprocesos para agrupar telemetría |
| `ANVIL_AGENT_ID` | `developer` | ID del agente actual en pipeline orquestado |
| `ANVIL_SKIP_EMIT` | `1` | Deshabilitar emit de eventos (útil en tests) |
| `CLAUDE_SESSION_FILE` | `/tmp/session.json` | Path al archivo de sesión de Claude Code |
| `EDITOR` | `nvim` | Editor para edición interactiva (fallback `vi`) |

## Flujo de trabajo típico

```bash
# 1. Setup inicial (primera vez)
make install                    # Instala anvil en $HOME/bin

# 2. Desarrollo día a día
make build                      # Build rápido local
go test -race ./internal/...    # Tests del dominio que tocaste
make vet                        # Vet antes de commit

# 3. Antes de hacer commit
make test                       # Race detector completo
make vet

# 4. Actualizar instalación
make install                    # Reemplaza el binario en $HOME/bin
```
