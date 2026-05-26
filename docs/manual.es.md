[Read in English](manual.md)

# Manual de Uso — Anvil

Guia practica para usar Anvil en tu dia a dia. Desde instalar hasta orquestar agentes en un proyecto real.

---

## 1. Instalacion

```bash
# Clonar el repo
git clone https://github.com/ernesto2108/anvil.git ~/projects/anvil
cd ~/projects/anvil

# Elegir que herramientas de IA usas
anvil targets claude              # Solo Claude Code
anvil targets claude opencode     # Claude + OpenCode
anvil targets all                 # Todas

# Elegir proveedor de modelos
anvil provider claude             # Anthropic (Claude)
anvil provider gemini             # Google (Gemini)
anvil provider local              # Ollama/local

# Desplegar
anvil deploy

# Verificar
anvil status
```

Despues de `deploy`, tus herramientas de IA ya tienen acceso a todos los agentes y skills.

---

## 2. Como invocar skills (lo mas comun)

Las skills se invocan con `/nombre-de-skill` en tu chat con la IA. Son el mecanismo principal para activar conocimiento especializado.

### Ejemplos del dia a dia

```
# Antes de escribir codigo Go
/go-conventions

# Antes de escribir codigo React
/react-conventions

# Quieres crear un Dockerfile
/devops-conventions

# Quieres correr linters
/lint

# Quieres correr tests
/run-tests

# Quieres crear un PRD para una nueva feature
/prd-template

# Quieres revisar un diseno en Pencil o Figma
/design-review

# Quieres ver el esquema de la base de datos
/db-schema-scan

# Quieres generar un diagrama de arquitectura
/generate-diagram

# Quieres auditar accesibilidad
/a11y-check

# Quieres verificar dependencias
/dependency-check
```

### Como saber que skills existen

Preguntale a tu IA: **"que skills tengo disponibles?"** — el sistema las lista automaticamente porque estan registradas.

---

## 3. Como usar agentes

Los agentes son roles especializados que la IA puede asumir. Se invocan de dos formas:

### Forma 1: Guiada (via `/orchestrate`)

Dile a tu IA que quieres hacer y usa `/orchestrate`:

```
Quiero agregar un endpoint de notificaciones al backend.
/orchestrate
```

La skill `/orchestrate` clasifica la tarea por complejidad y propone los agentes en orden. Tu confirmas antes de que se ejecute algo:

| Complejidad | Que pasa |
|-------------|----------|
| **Trivial** (1-2 archivos) | La IA lo hace directo, sin agentes |
| **Media** (3-8 archivos) | PM (si falta PRD) → Developer → Tester → QA |
| **Alta** (8+ archivos) | PM → Architect → Developer → Tester → QA → Reporter |

### Forma 2: Manual (invocas un agente especifico)

Tu eres el orquestador — no existe un agente lider automatico. Elige el agente que corresponde a tu tarea e invocalo directamente:

```
"Usa el agente developer-backend para implementar el login"
"Necesito al agente dba para crear la migracion"
"Lanza el agente security para auditar el codigo"
```

### Que agente usar para que?

| Necesitas... | Agente | Ejemplo |
|---|---|---|
| Escribir requisitos | **pm** | "Escribe el PRD para la feature de invitaciones" |
| Disenar arquitectura | **architect** | "Disena el bounded context de notificaciones" |
| Disenar UI/UX | **designer** | "Disena la pantalla de configuracion en Pencil" |
| Escribir codigo Go/backend | **developer-backend** | "Implementa el endpoint GET /users/:id" |
| Escribir codigo React/Astro/TS | **developer-frontend** | "Implementa la pantalla de Workflows" |
| Escribir codigo Flutter/Dart | **developer-mobile** | "Implementa la pantalla de notificaciones" |
| Escribir tests | **tester** | "Escribe tests para el servicio de autenticacion" |
| Crear migraciones | **dba** | "Crea la migracion para agregar la tabla invitations" |
| CI/CD e infra | **devops** | "Crea el Dockerfile y el workflow de CI" |
| Revisar calidad | **qa** | "Revisa el codigo del ultimo PR" |
| Auditar seguridad | **security** | "Audita el manejo de tokens JWT" |
| Inicializar o refrescar contexto del proyecto | **context-init** | "Inicializa el contexto del proyecto" |
| Escribir docs | **tech-writer** | "Actualiza el README con los nuevos endpoints" |
| Generar reporte + actualizar contexto | **reporter** | "Genera el reporte de esta sesion" |

---

## 4. Flujo de trabajo tipico

### Tarea trivial (fix rapido)

```
> "Corrige el typo en el archivo routes.go linea 42"
```

La IA lo hace directo. No necesitas agentes ni skills.

### Tarea media (nueva pantalla frontend)

```
> "Necesito implementar la pantalla de Workflows segun el diseno de Pencil"
> /orchestrate
```

La IA:
1. Clasifica como media (~5 pts)
2. Carga `/react-conventions`
3. Lanza **developer** en modo implementacion
4. Lanza **tester** para tests
5. Corre `/lint` y `/run-tests`

### Tarea grande (nuevo bounded context)

```
> "Quiero agregar gestion de equipo: invitar usuarios, asignar roles, listar miembros"
> /orchestrate
```

Tu (como orquestador):
1. Clasificas como alta (~13 pts)
2. Invocas **pm** → genera PRD
3. Invocas **architect** → disena contratos y bounded context
4. Invocas **designer** → disena pantallas
5. Invocas **developer-backend** y/o **developer-frontend** / **developer-mobile** → implementa los stacks necesarios
6. Invocas **tester** → escribe tests
7. Invocas **qa** → revisa calidad (bloquea si score < 7)
8. Invocas **reporter** → genera reporte de sesion y actualiza `.project-context/`

### Tarea de infraestructura

```
> "Necesito dockerizar el backend y crear el CI con GitHub Actions"
> /devops-conventions
```

La IA carga las convenciones de DevOps y tiene acceso a:
- Best practices de Docker (multi-stage, non-root, layer caching)
- Templates de GitHub Actions (CI con lint+test+build, CD con Cloud Run)
- Patrones de Terraform
- Guias de AWS, GCP, Kubernetes, Argo CD

---

## 5. Convenciones por stack

Cuando trabajas en un stack especifico, carga la skill de convenciones. Esto le da a la IA reglas, patrones y anti-patrones para ese lenguaje.

### Go

```
/go-conventions
```

Incluye: manejo de errores con wrap, validacion en entidad, SQL parametrizado, context en todo, defer despues de error check, concurrencia (worker pools, errgroup), testing (table-driven, mocks).

### React/TypeScript

```
/react-conventions
```

Incluye: hooks custom, estado (TanStack Query, Zustand), Tailwind v4 syntax, accesibilidad, testing (Vitest + RTL), anti-patrones, componentes funcionales only.

### Flutter/Dart

```
/flutter-conventions
```

Incluye: BLoC/Riverpod, composicion de widgets, freezed, theming, testing.

### DevOps/Infra

```
/devops-conventions
```

Incluye: Docker, GitHub Actions, Terraform, Kubernetes, AWS, GCP, Argo CD/Workflows/Rollouts, seguridad de infra.

---

## 6. Vault de documentacion

Cada proyecto puede tener un vault de Obsidian para documentacion estructurada. Inicializalo asi:

```bash
cp -r ~/projects/anvil/vault-template/ ~/projects/mi-proyecto-knowledge-base/
```

### Donde va cada cosa

| Carpeta | Contenido | Quien escribe |
|---|---|---|
| `01-project/` | `context.md` — snapshot tecnico del proyecto | scanner |
| `02-backlog/` | `sprint-current.md` — board del sprint | pm |
| `03-tasks/<ID>/` | PRD, design, QA review por tarea | pm, architect, qa |
| `04-architecture/` | ADRs, bounded contexts, diagramas | architect |
| `05-bugs/` | Postmortems de bugs criticos | security, qa |
| `06-reports/` | `last-run.md` — reporte de ultima sesion | reporter |
| `07-references/` | Templates, links externos | manual |
| `08-design/` | Archivos de diseno (.pen, .fig) | designer |

### Configurar el proyecto en Anvil

Agrega una entrada en `~/.claude/project-registry.md`:

```markdown
| mi-proyecto | personal | ~/projects/mi-proyecto-knowledge-base/ |
```

Ahora todos los agentes saben donde leer y escribir documentacion para ese proyecto.

---

## 7. Gestion del CLI

### Actualizar agentes y skills

```bash
cd ~/projects/anvil
git pull
anvil deploy
```

### Ver que esta desplegado

```bash
anvil status
```

### Cambiar de proveedor de IA

```bash
anvil provider gemini    # Cambia modelos a Gemini
anvil deploy             # Redesplegar con nuevos modelos
```

### Fijar una skill a una version

```bash
anvil pin skills/go-conventions v1.2.0
anvil unpin skills/go-conventions    # Volver a HEAD
```

### Desinstalar

```bash
anvil uninstall    # Limpia anvil y restaura archivos originales
```

---

## 8. Backup y Restauracion

Anvil protege automaticamente los archivos que ya tenias antes de instalarse. No necesitas hacer nada manual.

### Como funciona

#### Al hacer `anvil deploy` por primera vez

Anvil escanea todos los targets (Claude, OpenCode, Gemini, Codex) buscando archivos existentes. Si encuentra algo, guarda una copia exacta antes de tocar nada:

```
.anvil/pre-install/
├── claude/
│   ├── agents/        # Tus agentes originales de ~/.claude/agents/
│   ├── skills/        # Tus skills originales de ~/.claude/skills/
│   ├── commands/      # Tus comandos originales
│   └── CLAUDE.md      # Tu CLAUDE.md original
├── opencode/
│   ├── agents/
│   └── commands/
├── gemini/
│   ├── skills/
│   ├── commands/
│   └── GEMINI.md
└── codex/
    ├── skills/
    └── AGENTS.md
```

Veras en la terminal:

```
[anvil] Saving pre-install snapshot of existing files...
  saved ~/.claude/agents
  saved ~/.claude/skills
  saved ~/.claude/CLAUDE.md
```

Si no habia nada previo, simplemente dice:

```
[anvil] No existing files found — clean install
```

#### En cada deploy posterior

Si hay directorios que no son symlinks (por ejemplo, alguien edito directamente en `~/.claude/agents/`), Anvil los mueve a un backup con timestamp antes de sobreescribir:

```
[anvil] Backing up ~/.claude/agents → ~/.claude/agents.backup.20260403143022
```

Esto previene perdida accidental de cambios manuales.

#### Al hacer `anvil uninstall`

Anvil remueve lo que desplego y **restaura los archivos originales** desde el snapshot:

```
[anvil] Claude Code:
  restored ~/.claude/agents
  restored ~/.claude/skills
  restored ~/.claude/CLAUDE.md
[anvil] OpenCode:
  removed agents (no snapshot)
  removed commands (no snapshot)
[anvil] Gemini CLI:
  removed skills (no snapshot)
[anvil] Codex:
  removed skills (no snapshot)

[anvil] Anvil uninstalled. Pre-existing files restored where snapshots existed.
```

Si no habia snapshot (instalacion limpia), simplemente borra los archivos de Anvil y listo.

### Escenarios comunes

| Situacion | Que pasa |
|---|---|
| Primera vez, nada existia | Deploy limpio, snapshot vacio |
| Primera vez, tenias tus propios agents | Snapshot guarda tus agents, deploy sobreescribe, uninstall los restaura |
| Ya usas Anvil, editas un agent manualmente | Proximo deploy hace backup con timestamp del editado |
| Quieres volver al estado pre-Anvil | `anvil uninstall` restaura todo como estaba antes |
| Perdiste el snapshot | Los backups con timestamp estan en `~/.claude/*.backup.*` |

### Donde estan los backups

| Tipo | Ubicacion | Cuando se crea |
|---|---|---|
| Snapshot original | `.anvil/pre-install/` | Primer deploy |
| Backup con timestamp | `~/.claude/*.backup.YYYYMMDDHHMMSS` | Cada deploy posterior (si hay cambios manuales) |

### Importante

- El snapshot se crea **una sola vez** — en el primer deploy. Si haces deploy, uninstall, y deploy de nuevo, el segundo deploy crea un nuevo snapshot.
- `anvil uninstall` elimina el directorio `.anvil/` completo, incluyendo el snapshot. Si quieres conservarlo, copialo antes.
- Los backups con timestamp en `~/.claude/` no se borran automaticamente. Limpialos manualmente cuando ya no los necesites.

---

## 9. Tips

- **No cargues skills que no necesitas** — cada skill consume tokens. Solo carga las relevantes a tu tarea.
- **Usa `/orchestrate` cuando no sepas por donde empezar** — el sistema clasifica y elige los agentes por ti.
- **Las convenciones se acumulan** — si cargas `/go-conventions` y despues `/devops-conventions`, ambas aplican.
- **Los agentes tienen limites estrictos** — developer no toca tests, tester no toca produccion, dba no toca logica de negocio. Esto es por diseno.
- **context-init ahorra tokens** — ejecuta `context-init` (modo: `init` o `regular`) al inicio de una sesion para que los demas agentes tengan contexto sin leer cada archivo.
- **Documentar es parte del done** — despues de que los tests pasen, invoca `reporter` para actualizar `.project-context/`. Es tan obligatorio como los tests.
- **AGENTS.md se genera automaticamente** — no lo edites a mano, se sobreescribe en cada deploy.
