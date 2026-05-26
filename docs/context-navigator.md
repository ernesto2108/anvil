# Context Navigator — Plan de Diseño

## El problema

El scanner actual produce un `context.md` útil pero estático — snapshot de estructura de archivos y stack. Cada sesión nueva lo sobreescribe y los agentes leen el proyecto desde cero. No hay acumulación de conocimiento sobre patrones usados, decisiones tomadas, contratos entre servicios, o deuda técnica detectada.

El resultado: el arquitecto redescubre lo que ya sabe, el developer re-infiere patrones que ya están establecidos, y el orquestador no puede guiar a los agentes con el contexto real del sistema.

---

## Qué es el Context Navigator

Un sistema de contexto **vivo, granular y acumulativo** que vive en `.project-context/` al lado de `.handoff/`. No reemplaza `context.md` del scanner — lo extiende. Mientras el scanner produce el snapshot técnico del repositorio, el Context Navigator captura el **conocimiento operativo** del sistema: qué patrones se usan y dónde, qué contratos existen, qué decisiones se tomaron y por qué.

**Analogía:** el scanner es el mapa del territorio. El navigator es el cuaderno de bitácora del capitán — lo que aprendió navegando ese territorio.

---

## Estructura de archivos

```
.project-context/
├── NAVIGATOR.md           # Índice + cómo leer este sistema
├── project.md             # Visión, restricciones, stack, arquitectura general
├── patterns.md            # Patrones de diseño en uso con referencias a archivos
├── contracts.md           # APIs, queues, eventos, webhooks, servicios externos
├── business-rules.md      # Invariantes de negocio que cruzan dominios
├── dependencies.md        # Grafo de dependencias entre dominios con tabla de impacto
├── domains/               # Un archivo por bounded context / paquete clave
│   └── <domain>.md
├── decisions/             # ADRs-lite: decisiones arquitectónicas con contexto
│   └── NNN-<slug>.md
└── risks.md               # Gotchas, deuda técnica, restricciones conocidas
```

### NAVIGATOR.md — índice del sistema

```markdown
# Context Navigator — <ProjectName>
last_full_scan: <fecha>
last_updated: <fecha>
coverage: bootstrap | partial | full

## Índice
- [Proyecto](project.md) — stack, arquitectura, restricciones
- [Patrones](patterns.md) — diseño, SOLID, estructurales
- [Contratos](contracts.md) — APIs, queues, eventos
- [Dominios](domains/) — <domain1>, <domain2>
- [Decisiones](decisions/) — NNN archivos
- [Riesgos](risks.md) — gotchas, deuda
```

### project.md — fundamentos

Contiene lo que el arquitecto necesita leer primero:
- Objetivo del sistema (1 párrafo)
- Restricciones no negociables
- Stack con versiones
- Estilo arquitectónico detectado (hexagonal / layered / monolítico / microservicios)
- Módulo raíz y convenciones de paths
- Reglas de SOLID aplicadas (cuáles y dónde se verifican)

### patterns.md — patrones vivos

Por cada patrón detectado o introducido:
```markdown
## Factory — <NombreFactory>
- Archivo: `internal/pipeline/factory.go:34`
- Qué construye: runners por tipo de agente
- Cuándo usar: al agregar nuevo tipo de agente
- Anti-pattern: NO instanciar runners directamente fuera de la factory
```

Patrones a detectar automáticamente:
- **Creacionales**: Factory, Builder, Singleton, Provider
- **Estructurales**: Adapter, Decorator, Repository, Facade
- **Comportamiento**: Strategy, Observer, Command, Pipeline
- **Go-específicos**: functional options, middleware chain, table-driven tests
- **SOLID**: SRP (ficheros con una sola responsabilidad), OCP (interfaces), LSP, ISP, DIP

### contracts.md — bordes del sistema

```markdown
## REST API
### GET /api/v1/runs
- Handler: `internal/api/runs.go:Handler`
- Auth: Bearer token
- Response: `RunListResponse` (paginado)

## Message Queues
### pipeline.events (NATS / RabbitMQ / Redis Streams)
- Producers: `internal/runner/emit.go`
- Consumers: `internal/memory/digest.go`
- Schema: `RunEvent` — ver `pkg/events/run_event.go`

## Servicios externos
- Claude API: `internal/ai/claude/client.go`
- Ollama: `internal/ai/ollama/client.go`
```

### domains/<domain>.md — contexto por dominio

Un archivo por bounded context significativo. Solo se crea si hay código relevante.

```markdown
# Dominio: memory
last_updated: <fecha>
archivos_clave:
  - internal/memory/digest.go
  - internal/memory/store.go
  - internal/memory/transcript/

## Responsabilidad
Captura, procesa y persiste digests de sesiones de agentes.

## Flujo principal
SessionEnd → parser → haiku/summarizer → embeddings → sqlite-vec

## Patrones usados
- Strategy: summarizer intercambiable (claude API / claude CLI / ollama)
- Repository: store.go encapsula todo acceso a DB

## Decisiones tomadas
- D1: usar claude CLI como summarizer primario (ver decisions/002-claude-cli-summarizer.md)
- D2: embeddings vía Ollama para evitar latencia de red en modo local

## Gotchas
- SessionEnd tiene timeout de 30s — ver migration 000035 + fix en fbf0987
- sqlite-vec requiere CGO — Dockerfile debe tener gcc
```

### decisions/NNN-slug.md — ADRs-lite

```markdown
# 002 — Claude CLI como summarizer primario

fecha: 2026-05-01
estado: activo
afecta: internal/memory/

## Contexto
Necesitábamos un summarizer que funcionara sin API key en entornos de desarrollo local.

## Decisión
Usar claude CLI cuando está disponible; fallback a API key si no.

## Consecuencias
+ Funciona offline / sin credits en dev
- Requiere que el usuario tenga claude CLI instalado
- La CLI no expone control de modelo — siempre usa el modelo configurado

## Alternativas descartadas
- Solo API: descartado por costo en dev
- Solo Ollama: calidad insuficiente para decisiones de arquitectura
```

---

## Cómo se genera y actualiza

### Modo Bootstrap — proyecto nuevo o sin `.project-context/`

Trigger: scanner no encuentra `.project-context/NAVIGATOR.md` o el usuario pide explícitamente.

**Agente responsable:** `context-init` en `mode: deep`

Pasos:
1. Detectar stack y árbol (ya lo hace)
2. Grep por patrones: factory, builder, repository, strategy, etc.
3. Grep por contratos: handlers HTTP, suscriptores de cola, clientes externos
4. Grep por interfaces (Go) o abstract classes — inferir ISP/DIP
5. Leer `go.mod` / `package.json` / etc. para dependencias externas
6. Identificar bounded contexts por estructura de directorios (`internal/<domain>/`)
7. Escribir todos los archivos de `.project-context/` desde templates
8. Marcar `coverage: bootstrap` y `last_full_scan`

Bootstrap produce contexto de calidad media-alta automáticamente. El arquitecto lo refina en el siguiente paso del pipeline.

### Modo Incremental — después de cada implementación

Trigger: al final de cada pipeline run, como último paso del reporter.

**Agente responsable:** `reporter` con instrucción explícita de actualizar `.project-context/`

El reporter recibe el diff de la implementación y actualiza **solo las secciones afectadas**:

```
Si la implementación toca internal/memory/ → actualizar domains/memory.md
Si introduce un nuevo patrón → agregar entrada en patterns.md
Si agrega endpoint o queue → agregar entrada en contracts.md
Si documenta una decisión en el SPEC → crear decisions/NNN-slug.md
Si detecta gotcha o deuda → agregar entrada en risks.md
Siempre → actualizar last_updated en NAVIGATOR.md
```

**Regla de actualización:** delta, no sobrescritura. El reporter agrega o modifica secciones específicas usando Edit. Nunca sobreescribe el archivo completo.

### Staleness detection

NAVIGATOR.md tiene `last_updated`. El orquestador compara esa fecha contra el último commit del repo:

```
git log -1 --format=%ci -- internal/  (fecha del último cambio en código)
NAVIGATOR.md.last_updated

Si diff > 3 días → marcar como STALE en el brief del agente
Si diff > 7 días → sugerir al usuario correr scanner en mode: deep
```

---

## Integración con el pipeline existente

### Paso 0.5 — Context load (nuevo)

Antes del architect, el orquestador carga `.project-context/`:

```
Contexto disponible en .project-context/:
- project.md → inyectar completo (< 200 líneas)
- patterns.md → inyectar completo
- contracts.md → inyectar completo
- domains/<dominios afectados por la tarea> → inyectar solo los relevantes
```

El orquestador decide qué domains inyectar basándose en los archivos que la tarea va a tocar.

### Architect recibe

Antes del Context Navigator, el architect tenía que explorar el repo para entender el contexto. Ahora recibe inline:

```
## Contexto del sistema (pre-cargado desde .project-context/)

### Patrones en uso
[contenido de patterns.md]

### Contratos existentes
[contenido de contracts.md]

### Dominio afectado: memory
[contenido de domains/memory.md]
```

Esto elimina el ciclo de exploración del architect y mejora la calidad de sus decisiones porque trabaja con conocimiento acumulado, no inferido.

### Reporter escribe

Al final del pipeline, el reporter tiene dos responsabilidades:
1. Su tarea actual: resumen de ejecución
2. Nueva: delta a `.project-context/` basado en el diff de la implementación

---

## Integración con el modo directo (sin agentes)

En modo directo (el usuario no pide pipeline), el orquestador lee `.project-context/` al inicio y lo inyecta en la conversación principal:

> "Contexto del sistema cargado: patrones en uso (Factory, Repository, Strategy), 3 endpoints REST, dominio memory activo. Puedes preguntar sobre cualquier sección."

Esto ahorra el ciclo de exploración incluso en modo directo.

---

## Templates iniciales

Al bootstrap, se generan desde templates en `skills/context-nav/templates/`:
- `project.tmpl.md`
- `patterns.tmpl.md`
- `contracts.tmpl.md`
- `domain.tmpl.md`
- `decision.tmpl.md`
- `risks.tmpl.md`

Los templates tienen secciones marcadas con `<!-- TODO: detectar -->` que el scanner completa con grep/análisis real.

---

## Archivos a crear / modificar

### Nuevos
| Archivo | Propósito |
|---------|-----------|
| `skills/context-nav/SKILL.md` | Skill principal — spec de lectura/escritura |
| `skills/context-nav/bootstrap.md` | Guía de bootstrap desde cero |
| `skills/context-nav/update.md` | Guía de actualización incremental |
| `skills/context-nav/staleness.md` | Reglas de detección de staleness |
| `skills/context-nav/templates/*.tmpl.md` | Templates para cada tipo de archivo |
| `agents/context-nav.md` | Agente dedicado si se necesita fuera del scanner |

### Modificados
| Archivo | Cambio |
|---------|--------|
| `skills/scan-project/SKILL.md` | Agregar Paso 5: generar `.project-context/` en modo deep |
| `skills/scan-project/guides/deep-scan.md` | Agregar sección de detección de patrones y contratos |
| `agents/scanner.md` | Mencionar responsabilidad de bootstrap de `.project-context/` |
| `skills/orchestrate/SKILL.md` | Agregar Paso 0.5 de context load antes del architect |
| `agents/reporter.md` | Agregar responsabilidad de delta a `.project-context/` |
| `docs/token-optimization.md` | Agregar regla: inyectar `.project-context/` vs explorar repo |

---

## Estimación de ahorro de tokens

| Escenario | Sin Navigator | Con Navigator |
|-----------|--------------|---------------|
| Architect explorando dominio memory | ~8K tokens en reads | ~2K inyectado inline |
| Developer infiriendo patrones | ~5K en grep/reads | 0 (patrones en context) |
| Orquestador decidiendo convention files | ~3K buscando | 0 (stack en project.md) |
| Scanner re-descubriendo contratos | ~10K cada sesión | 0 si < 3 días |
| **Total por sesión medium** | ~26K extra | ~2K fijo |

Ahorro estimado: **15-24K tokens por sesión** en proyectos con > 3 meses de historia.

---

## Fases de implementación

### Fase 1 — Estructura base y bootstrap (Small, 3 pts)
- Skill `context-nav/SKILL.md` con spec completa
- Templates para todos los tipos de archivo
- Actualizar `scan-project` para generar `.project-context/` en modo deep

### Fase 2 — Integración con orquestador (Small, 2 pts)
- Agregar Paso 0.5 en orchestrate: cargar y filtrar `.project-context/`
- Regla de staleness detection

### Fase 3 — Actualización incremental por reporter (Small, 2 pts)
- Actualizar `agents/reporter.md` con delta responsibilities
- Guía `update.md` con reglas de qué sección actualizar según diff

### Fase 4 — Integración modo directo (Trivial, 1 pt)
- Regla en CLAUDE.md: al inicio de sesión, si `.project-context/NAVIGATOR.md` existe → leerlo y mencionar cobertura

---

## Preguntas abiertas antes de implementar

1. **¿Dónde vive `.project-context/` en proyectos con vault de Obsidian?** — ¿Al lado del código o dentro del vault? Propuesta: siempre al lado del código (`.project-context/` en el repo), no en el vault — es contexto técnico, no documentación.

2. **¿Se commitea `.project-context/` al repo?** — Propuesta: sí, en `.gitignore` no. Es conocimiento del equipo. El git history de `.project-context/` es trazabilidad de evolución del sistema.

3. **¿Qué pasa si el reporter falla o se salta?** — El sistema debe ser resiliente a updates perdidos. La staleness detection cubre esto — si el código cambió pero el context no, se marca STALE.

4. **¿Context Navigator para anvil-dashboard también?** — Propuesta: sí, mismo sistema, `.project-context/` en cada repo. El cross-service skill ya sabe de múltiples repos.
