# Dominio: ai-system

last_updated: 2026-05-17

## Responsabilidad

Configuración completa del sistema de IA del proyecto: los **agentes** (roles, permisos, modelos), las **skills** (comportamiento reutilizable), los **pipelines** (orquestaciones multi-agente) y los **commands** (puntos de entrada del CLI). Este dominio no contiene código de aplicación — son los artefactos declarativos que rigen cómo Claude orquesta el trabajo en el repo.

Cualquier modificación de archivos en `agents/`, `skills/*/SKILL.md`, `pipelines/*.yaml` o `commands/*.md` debe pasar por el agente `agent-designer`.

## Archivos clave

```
agents/                       — 28 specs de comportamiento de agente
skills/                       — 54 skills (cada una en su propio directorio con SKILL.md)
pipelines/                    — 8 pipelines YAML (DAGs multi-agente)
commands/                     — 1 slash command del CLI
```

## Agentes

### Orquestación

| Agente | Rol | Notas |
|---|---|---|
| `leader` | Orquestador central con 4 modos (Explorador, Planeación, Integración, Pruebas) | No ejecuta trabajo concreto. Produce plan de run y vault del proyecto |
| `explorer` | Exploración e investigación del repo | Produce hallazgos en `.context/runs/explorer-<topic>.md` |

### Contexto

| Agente | Rol | Notas |
|---|---|---|
| `context-bootstrap` | Crea estructura vacía de `.context/` | Idempotente |
| `scanner` | Escanea el repo y puebla `.context/` | Modos: bootstrap y deep |
| `reporter` | Aplica delta a `.context/` al final de cada run | Siempre el ÚLTIMO agente del pipeline |

### Planeación

| Agente | Rol | Notas |
|---|---|---|
| `pm` | PRDs accionables en español | Único autorizado para crear PRDs |
| `requirements` | Requirements EARS con IDs trazables | — |
| `architect` | Decisiones técnicas, contratos API, ADRs, ARD | — |
| `spec-writer` | Transforma ARD + requirements en `spec.md` implementable | — |
| `task-writer` | Escribe archivos de task atómicos para el backlog a partir de un spec | — |
| `designer` | UX/UI: sistemas de diseño, tokens, wireframes, flujos | — |
| `diagrammer` | Diagramas técnicos `.drawio`: flujos de datos, arquitecturas, pipelines | — |

### Implementación

| Agente | Rol | Notas |
|---|---|---|
| `developer` | ÚNICO autorizado para escribir código de producción | Go, React, Flutter, Astro, Python, TS, Rust |
| `tester` | ÚNICO autorizado para crear/modificar archivos de tests | — |
| `dba` | Migraciones y schema relacional | PostgreSQL, SQLite, MySQL |
| `dba-nosql` | Persistencia no-relacional | Document DBs, vector DBs, time-series, search |
| `dba-cache` | Estrategias de caché con Redis | Keyspace, TTL, eviction |
| `dba-broker` | Message brokers | Kafka, RabbitMQ, NATS. Topics, schemas, DLQ |
| `devops` | CI/CD, Docker, Kubernetes, Terraform | ÚNICO en `.github/workflows` |
| `committer` | Commits, push y PRs en GitHub | 2 fases: pre-review y post-QA |

### Calidad / Revisión

| Agente | Rol | Notas |
|---|---|---|
| `reviewer` | Revisión post-desarrollo de diffs/PRs | SOLO LECTURA |
| `qa` | Gate de calidad para tareas ≥5 pts | SOLO LECTURA |
| `qa-fixer` | Correcciones quirúrgicas post-QA/security | — |
| `security` | SAST, SCA, secretos, auth | Puede bloquear con CVE crítico. SOLO LECTURA |
| `dba-reader` | Auditoría de persistencia: schemas, EXPLAIN plans, índices | SOLO LECTURA |

### Documentación / Diseño del sistema

| Agente | Rol | Notas |
|---|---|---|
| `tech-writer` | Documentación, READMEs, docs de API, diagramas Mermaid, changelogs | — |
| `agent-designer` | ÚNICO autorizado para modificar `agents/*.md`, `skills/*/SKILL.md`, `commands/*.md`, `pipelines/*.yaml` | — |

### Contenido

| Agente | Rol | Notas |
|---|---|---|
| `mkt-content` | Contenido para redes sociales y copywriting | — |

## Skills

### Orquestación del Líder

| Skill | Propósito | Invocada por |
|---|---|---|
| `leader/output-formats` | Templates de output al cerrar cada modo del Líder | `leader` |
| `mode-gate` | Detecta el modo correcto del Líder según el prompt | `leader` |
| `run-init` | Inicializa el plan y estructura de un run | `leader` |
| `agent-teams` | Mapeo de equipos de agentes por tipo de tarea | `leader` |
| `budget-tracker` | Tracking de presupuesto de tokens por run | `leader` |
| `handoff` | Formato y reglas de archivos `.handoff/<TASK-ID>.md` | Todos los agentes que participan en un pipeline |
| `integration-close` | Cierre de runs en modo Integración | `leader` |

### Contexto del Proyecto

| Skill | Propósito | Invocada por |
|---|---|---|
| `context-nav` | Navegación y mantenimiento de `.context/` (incluye `update.md`) | `reporter`, `scanner` |
| `scan-project` | Escaneo de repo para poblar `.context/` | `scanner`, `context-bootstrap` |
| `service-map` | Mapeo de servicios y dependencias entre repos | `architect`, `explorer` |
| `read-files` | Lectura eficiente de archivos del repo | `explorer`, agentes de lectura |
| `write-files` | Escritura segura de archivos | Agentes con permiso write |
| `git-diff` | Análisis de diffs para reviewer/reporter | `reviewer`, `reporter`, `committer` |

### Implementación

| Skill | Propósito | Invocada por |
|---|---|---|
| `go-conventions` | Convenciones de Go del proyecto | `developer` |
| `react-conventions` | Convenciones de React | `developer` |
| `flutter-conventions` | Convenciones de Flutter | `developer` |
| `astro-conventions` | Convenciones de Astro | `developer` |
| `python-conventions` | Convenciones de Python | `developer` |
| `typescript-conventions` | Convenciones de TypeScript | `developer` |
| `rust-conventions` | Convenciones de Rust | `developer` |
| `wails-recipes` | Recipes para apps Wails (Go + frontend) | `developer` |
| `cross-service-dev` | Coordinar cambios cross-repo | `developer`, `architect` |
| `task-complete` | Checklist de cierre de tarea | `developer` |
| `summarize-changes` | Resumen de cambios para handoff/commit | `developer`, `committer` |

### Testing y Calidad

| Skill | Propósito | Invocada por |
|---|---|---|
| `run-tests` | Ejecución de la suite de tests del proyecto | `tester`, `developer`, `qa` |
| `test-api` | Tests de contratos de API | `tester` |
| `e2e-test-run` | Ejecución de tests end-to-end | `tester`, `qa` |
| `code-review-rubric` | Rúbrica de revisión de código | `reviewer`, `qa` |
| `post-review` | Procesamiento de hallazgos post-review | `reviewer`, `qa-fixer` |
| `lint` | Ejecución de linters del proyecto | `developer`, `qa` |
| `dependency-check` | Auditoría de dependencias y vulnerabilidades | `security`, `devops` |
| `perf` | Análisis de performance | `qa`, `developer` |
| `a11y-check` | Verificación de accesibilidad | `qa`, `designer`, `developer` |
| `visual-diff` | Diffs visuales para cambios de UI | `qa`, `designer` |
| `bundle-analyzer` | Análisis del bundle de frontend | `developer`, `qa` |
| `ui-component-scan` | Escaneo de componentes UI existentes | `designer`, `developer` |
| `design-review` | Revisión de cambios de diseño | `designer`, `qa` |

### Base de Datos

| Skill | Propósito | Invocada por |
|---|---|---|
| `db-engines` | Características y trade-offs por motor de DB | `dba`, `dba-nosql`, `dba-cache`, `dba-broker`, `architect` |
| `db-schema-scan` | Escaneo de schema existente | `dba`, `dba-reader` |
| `db-optimize` | Optimización de queries e índices | `dba`, `dba-reader` |

### Arquitectura y Diseño

| Skill | Propósito | Invocada por |
|---|---|---|
| `architecture-views` | Vistas arquitectónicas (C4, deployment, runtime) | `architect`, `tech-writer` |
| `architecture-boundary-guardrails` | Guardrails de límites entre módulos/dominios | `architect`, `developer` |
| `domain-entity-guardrails` | Guardrails para entidades de dominio | `architect`, `developer` |
| `document-architecture` | Documentación arquitectónica formal | `architect`, `tech-writer` |
| `generate-diagram` | Generación de diagramas Mermaid | `tech-writer`, `architect` |
| `drawio` | Generación de archivos `.drawio` editables | `diagrammer` |

### Diseño UX/UI

| Skill | Propósito | Invocada por |
|---|---|---|
| `design-system` | Diseño y mantenimiento de sistemas de diseño | `designer` |
| `design-project` | Estructura y artefactos de un proyecto de diseño | `designer` |
| `design-to-code` | Traducción de diseños a código | `designer`, `developer` |
| `design-recipes` | Recipes de patrones de diseño comunes | `designer` |

### Gestión de Proyecto

| Skill | Propósito | Invocada por |
|---|---|---|
| `prd-template` | Template de PRD del proyecto | `pm` |
| `backlog-management` | Gestión del backlog y tareas | `task-writer`, `pm` |
| `devops-conventions` | Convenciones de DevOps del proyecto | `devops` |

### Sistema de Agentes

| Skill | Propósito | Invocada por |
|---|---|---|
| `skill-standards` | Estándares para crear/modificar skills | `agent-designer` |

### Contenido

| Skill | Propósito | Invocada por |
|---|---|---|
| `social-content` | Contenido para redes sociales y copy | `mkt-content` |

## Pipelines

| Pipeline | Propósito | Cadena de agentes |
|---|---|---|
| `feat.yaml` | Feature estándar | architect → developer → tester → qa |
| `design.yaml` | Feature con UI | designer → architect → developer → tester → qa |
| `epic.yaml` | Épica completa | pm → designer → architect → dba → developer → tester → qa → security |
| `bug.yaml` | Bug fix | developer → tester → qa |
| `db.yaml` | Cambio de persistencia | dba → developer → tester |
| `infra.yaml` | Infraestructura | devops → security |
| `quick.yaml` | Fast-path simplificado | — |
| `example.yaml` | Pipeline de referencia | — |

## Commands

| Command | Propósito |
|---|---|
| `cross-service.md` | Orquestar features cross-repo en microservicios |

## Cambios recientes

<!-- el reporter actualiza esta sección -->
