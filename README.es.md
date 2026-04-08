[Read in English](README.md)

# Anvil

Sistema de orquestacion de agentes IA. Define agentes, skills y convenciones una sola vez — despliega a Claude Code, OpenCode, Gemini CLI, Codex y Cursor. Compatible con el estandar abierto [AGENTS.md](https://agents.md/).

<p align="center">
  <img src="assets/anvil-flow.svg" alt="Flujo de Anvil: Definir → Explorar → Desplegar → Usar" width="700"/>
</p>

## Que es Anvil?

Anvil es una coleccion de **agentes** (roles especializados de IA), **skills** (conocimiento de dominio y convenciones) y un **CLI** que los despliega a tus herramientas de desarrollo con IA. Escribes tus agentes y skills en markdown, ejecutas `anvil browse`, y cada herramienta recibe el mismo conocimiento.

Anvil genera automaticamente un archivo `AGENTS.md` — el [estandar abierto](https://agents.md/) mantenido por la Linux Foundation y adoptado por Codex, Cursor, Copilot y otros.

## Manual de Uso

Para una guia completa de como usar Anvil en el dia a dia — invocar skills, usar agentes, flujos de trabajo tipicos, convenciones por stack y tips — lee el **[Manual de Uso](docs/manual.es.md)**.

## Inicio Rapido

```bash
# 1. Clonar
git clone https://github.com/ernesto2108/anvil.git ~/projects/anvil
cd ~/projects/anvil

# 2. Compilar el CLI (requiere Go 1.25+)
make build

# 3. Hacerlo disponible globalmente (opcional)
ln -sf ~/projects/anvil/anvil /usr/local/bin/anvil

# 4. Configuracion inicial
anvil init

# 5. Explorar, seleccionar e instalar agentes/skills/comandos
anvil browse
```

> Despues del paso 3, puedes ejecutar `anvil` desde cualquier lugar. Si lo omites, usa `./anvil` desde el directorio de anvil.

## TUI Interactivo

`anvil browse` lanza una interfaz interactiva de terminal donde puedes explorar todos los agentes, skills y comandos — e instalarlos o desinstalarlos por target con una sola tecla.

<p align="center">
  <img src="assets/tui-browse.svg" alt="Interfaz TUI de anvil browse" width="700"/>
</p>

Caracteristicas:
- **Cambio de pestanas** entre vistas de Agentes, Skills y Comandos
- **Busqueda/filtro** para encontrar rapidamente lo que necesitas
- **Toggle por target** — instalar solo en Claude, o en todos los targets a la vez
- **Modos de deploy** — copy (permanente) o symlink (siempre sincronizado)
- **Controlado por teclado** — `enter` para instalar, `d` para desinstalar, `t` para alternar targets

## Como Funciona

```mermaid
flowchart TD
    User([Usuario]) -->|describe tarea| Orch[Orquestador]
    Orch -->|clasifica| Decision{Complejidad?}

    Decision -->|Trivial 1-2 archivos| Direct[Ejecucion directa]
    Decision -->|Media 3-8 archivos| Pipeline1[Developer → Tester → QA]
    Decision -->|Alta 8+ archivos| Pipeline2[Pipeline completo]

    Pipeline2 --> PM[Agente PM]
    PM -->|PRD| Arch[Agente Architect]
    Arch -->|design.md| Des[Agente Designer]
    Des -->|ui-spec.md| Dev[Agente Developer]
    Dev -->|codigo| Test[Agente Tester]
    Test -->|tests| QA[Agente QA]
    QA -->|score >= 7?| Gate{Quality Gate}
    Gate -->|Pasa| Rep[Agente Reporter]
    Gate -->|Falla| Dev

    style Orch fill:#6366f1,color:#fff
    style Gate fill:#f59e0b,color:#000
    style Direct fill:#10b981,color:#fff
```

Cada agente tiene limites estrictos:
- **developer** solo escribe codigo de produccion
- **tester** solo escribe archivos de test
- **dba** solo gestiona migraciones
- **devops** solo gestiona infra/CI
- Los agentes nunca cruzan limites

## Estructura del Proyecto

```
anvil/
├── cmd/anvil/             # CLI de despliegue (Go)
├── internal/
│   ├── cli/               # Comandos del CLI (init, status, doctor, etc.)
│   ├── deploy/            # Proveedores de deploy por target
│   └── tui/               # TUI interactivo (Bubble Tea)
├── anvil.yaml             # Manifiesto de despliegue (targets, componentes)
├── anvil.config.yaml      # Mapeo de proveedores y modelos
├── agents/                # 13 agentes especializados
├── skills/                # 44 skills de dominio y convenciones
├── commands/              # Comandos invocables por el usuario
├── docs/                  # Documentacion (en + es)
├── examples/              # Template de CLAUDE.md para proyectos
└── vault-template/        # Template de vault Obsidian para documentacion
```

## Agentes

Cada agente es un archivo markdown con frontmatter YAML que define su rol, permisos y nivel de modelo.

| Agente | Rol | Permiso | Nivel |
|--------|-----|---------|-------|
| **pm** | Requisitos, PRDs, backlog, planificacion de sprints | write | high |
| **architect** | Diseno de sistema, contratos API, ADRs | write | high |
| **designer** | Diseno UX/UI, design system, flujos de usuario | write | high |
| **developer** | Codigo de produccion (Go, React, Flutter, Astro, Python, TypeScript, Rust) | execute | medium |
| **tester** | Archivos de test en todos los stacks | execute | medium |
| **dba** | Migraciones, diseno de schema, optimizacion de queries | execute | medium |
| **devops** | CI/CD, Docker, Terraform, K8s, infra cloud | execute | medium |
| **qa** | Code review, quality gate (bloquea si score < 7) | execute | medium |
| **security** | SAST, SCA, auditoria de secretos, revision de auth | execute | medium |
| **scanner** | Escaneo de repositorio, generacion de contexto | execute | medium |
| **tech-writer** | Documentacion, README, API docs, changelogs | write | medium |
| **reporter** | Reportes de ejecucion de sesion | execute | low |
| **mkt-content** | Marketing de contenido, copywriting, assets visuales | execute | high |

### Como funcionan los agentes

- El orquestador (tu o `/orchestrate`) clasifica la complejidad de la tarea
- **Trivial**: se ejecuta directo, sin agentes
- **Medium+**: los agentes corren en secuencia con gates entre fases
- Cada agente tiene limites estrictos — developer no toca tests, tester no toca codigo de produccion

### Permisos

| Nivel | Herramientas disponibles |
|-------|-------------------------|
| **read** | Glob, Grep, LS, Read |
| **write** | + Write, Edit |
| **execute** | + Bash |

### Niveles de modelo

| Nivel | Uso | Ejemplo Claude | Ejemplo Gemini |
|-------|-----|----------------|----------------|
| **high** | Decisiones complejas (PM, Architect) | Opus | gemini-2.5-pro |
| **medium** | Implementacion (Developer, Tester) | Sonnet | gemini-2.5-flash |
| **low** | Tareas simples (Reporter) | Haiku | gemini-2.5-flash-lite |

## Skills

Las skills son modulos de conocimiento que se cargan bajo demanda segun la tarea.

### Convenciones por Stack

| Skill | Cubre |
|-------|-------|
| `/go-conventions` | Manejo de errores, validacion, SQL, concurrencia, testing, Kafka, RabbitMQ |
| `/react-conventions` | Hooks, estado, Tailwind v4, accesibilidad, testing, anti-patrones |
| `/flutter-conventions` | BLoC/Riverpod, composicion de widgets, theming, testing |
| `/astro-conventions` | Islands, content collections, componentes, estilos |
| `/python-conventions` | Type hints 3.12+, Pydantic v2, pytest, numpy, async, seguridad |
| `/typescript-conventions` | Strict mode, discriminated unions, Zod, Vitest, ESLint v8, Node.js ESM |
| `/rust-conventions` | Edition 2024, tokio, clap, Solana/Anchor, async, patrones CLI |
| `/devops-conventions` | Docker, GitHub Actions, Terraform, K8s, AWS, GCP, Argo CD/Workflows/Rollouts |

### Skills de Workflow

| Skill | Proposito |
|-------|-----------|
| `/orchestrate` | Clasificar complejidad, seleccionar agentes, gestionar gates |
| `/lint` | Auto-detecta stack, corre linters y formatters |
| `/run-tests` | Auto-detecta stack, corre tests con cobertura |
| `/perf` | Load/stress testing con Vegeta, k6, Locust |
| `/design-system` | Crear tokens, variables, componentes (Pencil/Figma) |
| `/design-project` | Punto de entrada rapido para proyectos de diseno, auto-detecta herramienta |
| `/design-recipes` | Patrones de diseno reutilizables para construir pantallas eficientemente |
| `/design-review` | Auditoria de calidad de disenos con puntaje |
| `/design-to-code` | Traducir disenos a codigo de produccion |
| `/prd-template` | Escritura de PRD con cuestionario de descubrimiento |
| `/backlog-management` | Dividir PRDs en tickets, gestionar sprints |
| `/handoff` | Continuidad de sesion — crear/retomar notas de handoff entre sesiones |
| `/scan-project` | Escanear estructura del repo y generar context.md |
| `/cross-service-dev` | Orquestar cambios a traves de multiples repos de microservicios |

### Skills de Guardia

| Skill | Proposito |
|-------|-----------|
| `/architecture-boundary-guardrails` | Enforzar bounded contexts, prevenir leaks entre dominios |
| `/domain-entity-guardrails` | Tipado estricto, sin punteros para campos opcionales |
| `/code-review-rubric` | Criterios de puntuacion para reviews de QA |
| `/skill-standards` | Estandares y checklist para crear nuevas skills |

### Skills de Utilidad

| Skill | Proposito |
|-------|-----------|
| `/dependency-check` | Auditar paquetes por vulnerabilidades y licencias |
| `/bundle-analyzer` | Analisis de impacto en tamano de bundle frontend |
| `/db-schema-scan` | Inspeccion read-only de schema via migraciones |
| `/db-optimize` | Identificar queries lentos, sugerir indices |
| `/generate-diagram` | Diagramas Mermaid.js (C4, ERD, secuencia, flujo) |
| `/git-diff` | Resumir cambios del repositorio |
| `/summarize-changes` | Escribir resumen legible de la sesion al vault |
| `/service-map` | Dependencias entre microservicios |
| `/a11y-check` | Auditoria de accesibilidad WCAG 2.1 |
| `/test-api` | Validacion de contratos de API endpoints |
| `/e2e-test-run` | Tests end-to-end (Playwright, Cypress) |
| `/ui-component-scan` | Escanear libreria de componentes para reusar |
| `/visual-diff` | Comparacion de screenshots para regresiones visuales |
| `/document-architecture` | Auto-documentar arquitectura de servicios |
| `/social-content` | Creacion de contenido para redes sociales (LinkedIn, Instagram, X) |
| `/task-complete` | Marcar tareas como completadas, actualizar tablero Kanban |

## Referencia del CLI

<p align="center">
  <img src="assets/anvil-status.svg" alt="salida de anvil status" width="600"/>
</p>

```bash
# Configuracion
anvil init                       # Configuracion inicial — muestra config y abre navegador
anvil browse                     # TUI interactivo para gestionar agentes/skills/comandos
anvil update                     # Pull del ultimo codigo + recompilar binario

# Targets (a que herramientas desplegar)
anvil targets                    # Mostrar targets activos
anvil targets claude opencode    # Definir targets exactos
anvil targets --add gemini       # Habilitar un target
anvil targets --rm cursor        # Deshabilitar un target
anvil targets all                # Habilitar todos

# Provider (mapeo de modelos)
anvil provider                   # Mostrar provider actual
anvil provider gemini            # Cambiar a modelos Gemini
anvil provider local             # Cambiar a modelos locales/Ollama

# Diagnosticos
anvil status                     # Mostrar version, rama, targets, tags
anvil doctor                     # Diagnosticar salud del despliegue
anvil diff                       # Mostrar cambios desde ultimo deploy

# Versionado
anvil pin skills/go-conventions v1.2.0    # Fijar a un git tag
anvil unpin skills/go-conventions         # Volver a seguir HEAD

# Mantenimiento
anvil uninstall                  # Remover de todos los targets
```

## Configuracion

### `anvil.yaml` — Manifiesto de despliegue

```yaml
targets:
  claude:
    enabled: true
    path: ~/.claude
  opencode:
    enabled: true
    path: ~/.config/opencode
  gemini:
    enabled: true
    path: ~/.gemini
  codex:
    enabled: true
    path: ~/.codex
  cursor:
    enabled: true
    path: per-project

components:
  agents:
    tag: "HEAD"
  skills:
    tag: "HEAD"
  commands:
    tag: "HEAD"
```

### `anvil.config.yaml` — Mapeo de proveedores y modelos

```yaml
provider: claude

providers:
  claude:
    high: opus
    medium: sonnet
    low: haiku
  cursor:
    high: claude-opus-4-20250514
    medium: claude-sonnet-4-20250514
    low: claude-haiku-4-5-20251001
  gemini:
    high: gemini-2.5-pro
    medium: gemini-2.5-flash
    low: gemini-2.5-flash-lite
  local:
    high: qwen3:32b
    medium: qwen3:14b
    low: qwen3:8b
```

## Crear Nuevos Agentes

Crear `agents/{nombre}.md`:

```markdown
---
name: mi-agente
description: Descripcion en una linea que el sistema usa para decidir cuando invocar este agente
permission: execute    # read | write | execute
model: medium          # high | medium | low
---

# Agent Spec — Titulo del Rol

## Role
Que hace este agente y que NO hace.

## Input
Que le proporciona el orquestador.

## Rules
Restricciones y permisos especificos.

## Output
Que produce y donde lo escribe.
```

## Crear Nuevas Skills

Crear `skills/{nombre}/SKILL.md`:

```markdown
---
name: mi-skill
description: Descripcion en una linea de lo que ensena esta skill
---

# Nombre de la Skill

## When to Load
Condiciones que disparan la carga de esta skill.

## Contenido
El conocimiento, convenciones y patrones.
```

Para skills complejas, usar subdirectorios con tabla de ruteo:

```
skills/mi-skill/
├── SKILL.md           # Dispatcher con tabla de ruteo
├── rules/             # Archivos de referencia rapida
├── guides/            # Patrones detallados
└── examples/          # Patrones buenos y malos
```

## Vault de Documentacion

Usar `vault-template/` para inicializar un vault Obsidian en cualquier proyecto:

```bash
cp -r vault-template/ ~/projects/mi-proyecto-knowledge-base/
```

Estructura:
```
01-project/context.md         # Output del scanner
02-backlog/sprint-current.md  # Board del sprint
03-tasks/<ID>/                # PRD, design, QA por tarea
04-architecture/              # ADRs, bounded contexts
05-bugs/                      # Postmortems
06-reports/last-run.md        # Reportes de sesion
07-references/                # Templates, links externos
08-design/                    # Archivos de diseno (.pen, .fig)
```

## Backup y Restauracion

Anvil protege tus archivos existentes automaticamente:

- **Primer deploy**: guarda snapshot de todo lo que encuentra en `~/.claude/`, `~/.codex/`, etc.
- **Cada deploy**: hace backup con timestamp si detecta cambios manuales
- **Uninstall**: restaura los archivos originales desde el snapshot

Ver [seccion completa en el manual](docs/manual.es.md#8-backup-y-restauracion).

## Compatibilidad con AGENTS.md

[AGENTS.md](https://agents.md/) es un estandar abierto mantenido por la Linux Foundation para configurar agentes de IA en proyectos de software. Anvil genera automaticamente este archivo cada vez que instalas agentes.

### Que herramientas lo leen?

| Herramienta | Lee AGENTS.md | Archivo nativo |
|---|---|---|
| **OpenAI Codex** | Si (primario) | `~/.codex/AGENTS.md` |
| **Cursor** | Si (en raiz del repo) | `.cursor/rules/*.mdc` |
| **GitHub Copilot** | Via `.github/copilot-instructions.md` | `.github/agents/*.agent.md` |
| **OpenCode** | Si (primario) | — |
| **Claude Code** | No (usa `CLAUDE.md`) | `~/.claude/agents/*.md` |
| **Gemini CLI** | Discusion activa | `GEMINI.md` |

### Como funciona en Anvil

1. Defines agentes en `agents/*.md` con frontmatter (rol, permisos, nivel)
2. Al instalar via `anvil browse`, Anvil:
   - Despliega agentes nativos a cada target (Claude, OpenCode, Gemini, etc.)
   - Genera `AGENTS.md` compacto a `~/.codex/` para Codex
3. Cualquier herramienta compatible con AGENTS.md puede leer el archivo generado

## Licencia

MIT
