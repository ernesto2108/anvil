# Optimizacion de Tokens

## Por que importa

Cada token cuesta dinero y tiempo. Un anvil mal disenado puede quemar 10x mas tokens que uno optimizado, sin mejorar la calidad del output.

## Estrategias

> **Nota sobre nombres de agentes:** en los pipelines de este documento, `developer` refiere a la familia de developers resuelta por stack (`developer-backend` / `developer-frontend` / `developer-mobile` / `developer-ai`) y `designer` a la dupla `designer-spec` (especificacion) + `designer-visual` (construccion en Pencil). `reporter` es una skill, no un agente.

### 1. No correr agentes innecesarios

La optimizacion mas grande es no hacer trabajo que no se necesita.

| Situacion | Enfoque malo | Enfoque bueno |
|-----------|-------------|---------------|
| Fix de typo | Pipeline completo | Directo |
| Bug con repro claro | pm + architect + developer | developer + tester |
| Extender patron existente | architect + developer | developer + tester |
| Feature nueva compleja | developer solo | pm + architect + developer + tester + qa |

### 2. Pasar solo lo necesario a cada agente

Cada agente recibe SOLO los archivos que necesita. No pasar el contexto completo.

```
# MAL — developer recibe todo
"Lee prd.md, design.md, design-spec.md, qa-review.md, security-audit.md, context.md..."

# BIEN — developer recibe lo minimo
"Lee prd.md y design.md. Convention skill: go-conventions."
```

### 3. Inyectar contenido vs dejar que re-lean

Si un archivo ya esta en el contexto de la conversacion, inyectar el contenido en el prompt del agente en vez de dejarlo que lo lea de nuevo.

```
# MAL — agente gasta tokens leyendo archivos
"Lee vault/03-tasks/TASK-001/prd.md"

# BIEN — inyectar el contenido
"PRD:
## Resumen
Implementar endpoint de registro...
## Criterios de Aceptacion
1. ..."
```

### 4. Convention routing — orchestrator selecciona, agente solo lee

Las convention skills son grandes (go-conventions = 5,900 lineas). Los agentes NO navegan dispatchers ni deciden que cargar. El orchestrator lee el dispatcher, elige 2-4 archivos relevantes, y pasa rutas absolutas al agente. Para Small tasks, inyectar reglas inline (3-5 bullets).

Ver `docs/convention-routing.md` para la guia completa con ejemplos de prompts.

### 5. Skip context-init si .project-context/ es reciente

Si `.project-context/NAVIGATOR.md` se actualizo en la misma sesion, no correr `context-init` de nuevo. Ahorra ~5000 tokens.

### 6. Inyectar `.project-context/` en vez de explorar el repo

Si `.project-context/NAVIGATOR.md` existe y tiene menos de 3 dias, inyectar los archivos relevantes inline en vez de dejar que el agente explore el repo:

| Escenario | Sin Context Navigator | Con Context Navigator |
|---|---|---|
| Architect explorando un dominio | ~8K tokens en reads | ~2K inyectado inline |
| Developer infiriendo patrones existentes | ~5K en greps | 0 — patrones en `patterns.md` |
| Orquestador eligiendo convention files | ~3K buscando | 0 — stack en `project.md` |
| context-init re-descubriendo contratos | ~10K cada sesion | 0 si < 3 dias |

Regla: antes de decirle a un agente "lee `internal/X/`", verificar si ese dominio esta documentado en `.project-context/Technical domain/domain.md` (o en `business-rules.md` / `contracts.md` del mismo directorio, segun lo que se necesite). Si esta documentado y fresco, inyectar ese archivo en lugar del codigo fuente.

### 7. Reporter solo en Maximum

La skill `reporter` genera un resumen de sesion. Solo vale la pena en tareas Maximum donde hubo muchos cambios. Para trivial/medium es desperdicio.

### 8. Un QA, no N QAs

En cross-service, correr UN qa con el diff combinado de todos los servicios. No un qa por servicio.

### 9. Agentes concisos

Los agentes mismos deben ser cortos. Un agente de 200 lineas se carga en el contexto cada vez que se invoca. Mantener por debajo de 100 lineas.

### 10. Un documento por invocacion

No pedirle a un agente que produzca PRD + roadmap + sprint update en una sola corrida. Dividir en invocaciones separadas. Cada invocacion produce 1 archivo.

### 11. Presupuestos de tokens por agente

Cada agente tiene un target y un maximo. Si se excede consistentemente, revisar el prompt o dividir el trabajo.

## Presupuestos por agente

| Agente | Target | Max | Tool calls max |
|--------|--------|-----|----------------|
| pm | 15K | 25K | 5 |
| designer-spec | 20K | 40K | 10 |
| architect | 20K | 40K | 15 |
| developer-* (backend / frontend / mobile / ai) | 30K | 60K | 15 |
| tester | 20K | 40K | 10 |
| qa | 10K | 20K | 15 |
| security | 10K | 20K | 15 |
| reporter (skill) | 5K | 10K | 3 |
| context-init | 10K | 20K | 8 |

**Nota:** Estos son guidelines, no limites duros. Si un agente necesita mas, el orchestrador debe justificarlo.

**Nota sobre designer-visual:** no tiene fila propia porque su consumo lo dominan las MCP tools de Pencil, no las lecturas de contexto — su presupuesto se gestiona por pantalla segun la seccion "Optimizacion de MCP tools (Pencil/Figma)".

**Nota sobre verificacion (architect, qa, security, tester):** sus presupuestos incluyen las lecturas de verificacion (`.project-context/`, artefacto del explorer, greps de checklist, schema, paths, tipos, contratos). Verificar antes de decidir, aprobar o bloquear cuesta unas pocas lecturas; hacerlo a ciegas cuesta una re-invocacion completa — la verificacion esta dentro del presupuesto y nunca se recorta para ahorrar tool calls. Si el presupuesto no alcanza para verificar, escalar al humano en vez de aprobar o bloquear sin evidencia.

## Metricas a observar

| Metrica | Que indica |
|---------|-----------|
| Tokens por tarea trivial | Deberia ser < 5K |
| Tokens por tarea medium | Deberia ser < 30K |
| Tokens por tarea complex | Deberia ser < 100K |
| Agentes invocados vs necesarios | Si siempre corres 8 agentes, algo esta mal |
| Re-lecturas de archivos | Si el mismo archivo se lee 3+ veces, inyectar contenido |
| Tokens PM vs presupuesto | Si PM > 25K, el prompt fue muy pesado o leyo codigo |

### 12. Division spec/construccion en diseño (MCP tools)

El diseño esta dividido en dos agentes con presupuestos de naturaleza distinta:

- **`designer-spec`** produce la especificacion (`design-spec.md` + `DESIGN.md`). No usa MCP tools — su presupuesto es el de un agente de documentos (ver tabla de presupuestos).
- **`designer-visual`** construye el diseño en Pencil: declara los tools `mcp__pencil__*` en su frontmatter y ejecuta el Design Spec directamente sobre el archivo `.pen`. Su consumo lo dominan las MCP tools — aplicar las estrategias de la seccion siguiente.

```
designer-spec → design-spec.md → designer-visual construye en Pencil → architect
```

**Paso humano residual:** el Pencil MCP no abre ni crea archivos — opera sobre el documento ya abierto en el editor. El usuario es responsable de abrir/crear el `.pen` antes de invocar a `designer-visual`.

Regla general: un agente que necesita MCP tools debe declararlos explicitamente en su frontmatter (como hace `designer-visual`) — no asumir que los hereda del proceso principal.

## 13. Optimizacion de MCP tools (Pencil/Figma)

Los MCP servers de diseño (Pencil, Figma) consumen tokens masivamente. Estrategias:

### Schema caching

`get_editor_state(include_schema: true)` devuelve ~8K tokens. Solo cargarlo UNA vez por sesion. Calls posteriores: `include_schema: false`.

### Guidelines caching

`get_guidelines("guide", "Web App")` y similares son estaticos. Cargar UNA vez, no por pantalla.

### Usar design-recipes

El skill `/design-recipes` provee patrones probados que reducen operaciones por pantalla:
- Auth screen: ~18 ops (vs ~30 sin receta)
- Table page: ~25 ops en 2 batches (vs ~40 improvisando)
- Dark mode: 1 Copy + 2-3 overrides (vs reconstruir)

### Maximizar batch_design

Apuntar a 20-25 operaciones por `batch_design` call. Menos calls = menos overhead del tool description (~4K tokens cada vez que aparece en el schema).

### Componentes primero

Crear TODOS los componentes reutilizables antes de cualquier pantalla. Despues las pantallas son solo `ref` + `descendants` overrides. Esto reduce dramaticamente las operaciones por pantalla.

### Diseñar por fases

Para proyectos grandes (8+ pantallas), dividir en sesiones:
- Sesion 1: Variables + Componentes + Design System docs
- Sesion 2: Pantallas auth + dashboard
- Sesion 3: Pantallas de contenido
- Sesion 4: Mobile + Dark mode

Esto evita agotar el contexto en una sola sesion.

### Copy, nunca rebuild

Para variantes (dark mode, mobile), usar Copy del frame light y aplicar theme override. Nunca reconstruir la pantalla desde cero.

### Presupuesto estimado por pantalla

| Tipo de pantalla | Ops estimadas | batch_design calls |
|-----------------|---------------|-------------------|
| Auth (login, register) | 18-22 | 2 |
| Dashboard con tabla | 30-40 | 3 |
| Lista con tabla | 25-30 | 2-3 |
| Detalle | 20-25 | 2 |
| Wizard (por paso) | 20-25 | 2 |
| Mobile (copy + adapt) | 15-20 | 2 |
| Dark mode (copy) | 3-5 | 1 |

## Regla de oro

> El anvil mas eficiente es el que corre exactamente los agentes necesarios, les pasa exactamente la informacion que necesitan, y para exactamente cuando debe parar.
