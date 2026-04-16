# Preflight y Prompts Estructurados

## El problema

Los agentes de Anvil (developer, tester, architect, etc.) arrancan con contexto limpio. Si reciben un prompt vago como "agrega un campo al entity", no saben que stack usar, que archivos tocar, ni que convenciones seguir. El resultado es trabajo generico o fallido.

Esto pasa en **dos escenarios**:

1. **Claude Code orquestando** — el usuario dice "orquesta" y Claude invoca agentes via Agent tool
2. **Anvil CLI** — el usuario corre `anvil feat --task "descripcion"`

En ambos casos, la solucion es la misma: validar que el prompt tenga la informacion minima antes de lanzar agentes.

## Que cambio

### En Claude Code (CLAUDE.md)

Nueva seccion "Pre-agent checklist" que obliga a Claude a validar estos campos antes de invocar cualquier agente:

| Campo | Requerido | Ejemplo |
|-------|-----------|---------|
| Stack | siempre | Go, React, Flutter |
| Objetivo | siempre | "Agregar metodo GetRunsByProject" |
| Archivos afectados | siempre | `internal/dashboard/query/runs.go` o "por descubrir" |
| Complejidad | siempre | Small (3 pts), Medium (5 pts), Large (8 pts) |
| Convention files | Medium+ | rutas absolutas a archivos de convenciones |
| Servicios afectados | solo cross-service | `blt-bookings`, `blt-payments` |

Si falta algo, Claude pregunta en espanol antes de proceder. Muestra el prompt final y espera aprobacion.

**No aplica en modo directo** — si el usuario dice "hazlo directo", Claude trabaja sin agentes y no necesita el checklist.

### En Anvil CLI

Tres cambios en el runner de pipelines:

#### 1. Preflight interactivo (`internal/cli/preflight.go`)

Antes de correr cualquier pipeline, Anvil pregunta lo que falta:

```
$ anvil feat --task "agregar campo ProjectName al entity Run"

[anvil] Preflight — necesito algunos datos para armar los prompts de los agentes.

  ¿Stack? (Go, React, Flutter, Python, Rust, Astro, otro): Go
  ¿Archivos afectados? (rutas separadas por coma, o 'no sé'): internal/dashboard/entity/run.go
  ¿Complejidad? (small, medium, large — default: medium): small

[anvil] Prompt que recibiran los agentes:

  Objetivo:    agregar campo ProjectName al entity Run
  Stack:       Go
  Archivos:    internal/dashboard/entity/run.go
  Complejidad: Small (3 pts)

  ¿Dale? (si/no): si
```

Con `-y` (auto-approve) se salta las preguntas y usa defaults (stack: auto, complexity: medium).

#### 2. Runner mejorado (`internal/runner/runner.go`)

Antes:
```
claude --print --bare -p "You are the developer agent. Task: agregar campo..."
```

Ahora:
```
claude --print --agent developer -p "Complexity: Small (3 pts)\nStack: Go\nMode: normal\nObjective: agregar campo...\nFiles to change: internal/dashboard/entity/run.go"
```

Cambios clave:
- `--bare` eliminado — ahora carga CLAUDE.md, OAuth, hooks, y skills
- `--agent <role>` agregado — carga la definicion del agente con sus skills precargados
- Prompt estructurado con metadata completa

#### 3. Auth fix (`internal/cli/run.go`)

`checkClaudeAuth` ya no usa `--bare`, asi que OAuth funciona correctamente.

## Como probar

### Prerequisito: deploy

Los cambios en agentes y config necesitan desplegarse primero:

```bash
cd ~/projects/anvil
go build -o ~/bin/anvil ./cmd/anvil
anvil deploy   # o el comando que uses para desplegar
```

### Test 1: verificar que auth funciona

```bash
anvil feat --task "test" 
```

Debe mostrar el preflight interactivo, NO el error "Claude CLI not authenticated".

Si ves el error de auth, corre `claude /login` primero.

### Test 2: preflight interactivo

```bash
anvil quick --task "agregar campo ProjectName al entity Run"
```

Debe preguntar stack, archivos, complejidad, mostrar resumen, pedir confirmacion.

### Test 3: preflight con auto-approve

```bash
anvil quick --task "agregar campo ProjectName" -y
```

Debe saltar las preguntas y usar defaults (stack: auto, complexity: medium).

### Test 4: verificar que --agent se pasa

Crear un fake claude para inspeccionar los argumentos:

```bash
# En una terminal temporal
mkdir /tmp/fake-claude
cat > /tmp/fake-claude/claude << 'EOF'
#!/bin/sh
echo "ARGS: $@" >> /tmp/claude-args.log
echo "pong"
EOF
chmod +x /tmp/fake-claude/claude

# Correr con el fake
PATH=/tmp/fake-claude:$PATH anvil quick --task "test" -y

# Verificar
cat /tmp/claude-args.log
# Debe contener: --print --agent developer -p "Complexity: ..."
# NO debe contener: --bare
```

### Test 5: verificar prompt estructurado

Mismo test que el anterior, pero revisar el contenido del `-p`:

```bash
cat /tmp/claude-args.log | grep -o '\-p .*'
```

Debe contener:
- `Complexity: Medium (5 pts)`
- `Stack: auto`
- `Mode: normal`
- `Objective: test`

### Test 6: pipeline completo (end-to-end)

```bash
anvil quick --task "agregar campo ProjectName al entity Run"
```

Responder al preflight:
- Stack: Go
- Archivos: internal/dashboard/entity/run.go
- Complejidad: small

El pipeline debe:
1. Correr el developer agent con el prompt estructurado
2. Pasar el output del developer al tester como upstream context
3. Completar sin errores de auth o prompts vacios

### Tests unitarios

```bash
go test ./internal/runner/... -v
```

Verifica:
- `TestNew` — construye runner con TaskContext
- `TestBuildPrompt_NoUpstream` — prompt tiene objective, stack, complexity
- `TestBuildPrompt_WithUpstream` — incluye context de agentes previos
- `TestRunAgent_WithModel` — pasa `--agent` y `--model` al CLI

## Flujo completo: antes vs ahora

### Antes

```
usuario: anvil feat --task "agregar campo"
anvil: claude --print --bare -p "You are the developer agent. Task: agregar campo"
claude: (sin CLAUDE.md, sin agent definition, sin skills, sin conventions)
        → resultado generico o fallido
```

### Ahora

```
usuario: anvil feat --task "agregar campo"
anvil: ¿Stack? Go. ¿Archivos? entity/run.go. ¿Complejidad? small. ¿Dale? si.
anvil: claude --print --agent developer -p "Complexity: Small (3 pts)
        Stack: Go
        Objective: agregar campo
        Files: entity/run.go"
claude: (carga CLAUDE.md + developer.md + skills lint/run-tests)
        → resultado con convenciones, lint, tests
```

## Archivos modificados

| Archivo | Que cambio |
|---------|-----------|
| `~/.claude/CLAUDE.md` | Nueva seccion "Pre-agent checklist" |
| `internal/cli/preflight.go` | Nuevo — preflight interactivo |
| `internal/cli/run.go` | Integra preflight, auth sin --bare |
| `internal/runner/runner.go` | TaskContext, --agent, sin --bare, prompt estructurado |
| `internal/runner/runner_test.go` | Tests adaptados a nueva API |
