---
name: task-writer
description: Reglas para escribir archivos de task atómicos a partir de un archivo spec. Define los templates de task individual y archivo padre (épica/historia), categorías, inferencia del agente ejecutor y protocolo de escalación. Úsalo cuando el usuario diga "escribir tasks", "descomponer spec", "generar archivos de task", "crear tasks", "task-writer", "descomponer feature" o "descomponer épica".
disable-model-invocation: true
---

Esta skill define cómo traducir un archivo spec en un conjunto de archivos `.md` independientes — uno por task — y, cuando aplica, un archivo padre de épica que los agrupa. No cubre actualización del backlog ni transiciones de estado.

## Filosofía

1. Una task = un archivo = una preocupación.
2. El task-writer no decide implementación — solo descompone.
3. Sin confirmación del humano, no se escribe nada.

## Tipos de descomposición

Preguntar al inicio qué tipo de trabajo es:

- **feature / historia** → genera archivos individuales de task (uno por task).
- **épica** → genera 1 archivo padre + archivos individuales por cada subtask.

La elección determina si se emite un archivo padre y cómo se nombran los archivos.

## Template — archivo de task individual

Cada task es un archivo `.md` independiente con este formato exacto:

```markdown
---
name: "<FEATURE_ID>-<NN>-<slug-corto>"
type: "setup" | "implementation" | "integration" | "validation"
priority: "HIGH" | "MEDIUM" | "LOW"
agent: "developer-backend" | "developer-frontend" | "developer-mobile" | "tester" | "dba" | "dba-cache" | "dba-broker" | "dba-nosql" | "devops" | "tech-writer" | "security" | "observability" | "diagrammer"
points: 1 | 2 | 3 | 5 | 8
milestone: "<milestone>" (opcional)
feature_id: "<FEATURE_ID>"
dependencies: []
inputs:
  - "<nombre>: <tipo> (<restricción>)"
outputs:
  - "<nombre>: <tipo>"
validation_rules:
  - "<regla>: <valor>"
---

# {name}

## Objetivo
[Comportamiento observable esperado — el QUÉ, no el CÓMO]

## Contexto Técnico
[Tabla, endpoint, contrato u otra referencia técnica extraída del spec]

## Interfaces
- Llamado por: `path/caller`
- Llama a: `path/callee`

## Criterios de Aceptación
- [ ] [criterio verificable]
```

**Reglas del template:**

- `inputs`, `outputs` y `validation_rules` son opcionales — omitir si no aplican a la task.
- `## Interfaces` es opcional — omitir si la task no tiene vecinos claros.
- `dependencies: []` cuando no hay dependencias; si las hay, listar los `name` de las tasks que deben completarse primero.
- Sin sección "Pasos de Ejecución" — el task-writer no sabe implementación.
- Si el spec no declara milestones → omitir el campo o usar `tbd`.

## Template — archivo padre (épica)

```markdown
---
name: "<FEATURE_ID>-epic-<slug>"
type: "epic"
priority: "HIGH" | "MEDIUM" | "LOW"
milestone: "<milestone>" (opcional)
feature_id: "<FEATURE_ID>"
subtasks:
  - "<FEATURE_ID>-01-<slug>"
  - "<FEATURE_ID>-02-<slug>"
---

# {name}

## Objetivo
[Descripción del valor de negocio de la épica]

## Subtareas

| ID | Tipo | Agente | Pts | Depende de |
|---|---|---|---|---|
| <FEATURE_ID>-01-<slug> | implementation | developer-backend | 3 | — |
| <FEATURE_ID>-02-<slug> | integration | developer-frontend | 2 | 01 |
| <FEATURE_ID>-03-<slug> | validation | tester | 2 | 02 |
| <FEATURE_ID>-04-<slug> | setup | dba | 1 | — |

**Total pts:** N
```

Si el spec no declara milestones → omitir el campo o usar `tbd`.

## Categorías de task

| Categoría | Significado | Ejemplos |
|---|---|---|
| `setup` | Tipos, interfaces, schemas vacíos, sin lógica. Habilita el resto. | Crear interface `EventStore`, crear DTO `CreateEventRequest` |
| `implementation` | Lógica concreta encapsulada en una unidad. | Implementar método `Create` del repositorio, implementar service `BookEvent` |
| `integration` | Conectar dos componentes ya existentes. | Wirear handler con service, registrar handler en router |
| `validation` | Tests, verificación de comportamiento end-to-end. | Tests unit del service, test de contrato del endpoint |

## Reglas de descomposición

1. Una task = un archivo principal (puede tocar 1-2 adicionales de imports/config si son inevitables).
2. Si el developer necesita leer >5 archivos para entender el contexto → la task es demasiado amplia, dividir.
3. Orden topológico obligatorio: setup → implementation → integration → validation.
4. Máx 15 tasks por feature / épica. Si supera, registrar como decisión abierta y entregar las 15 primeras por prioridad.
5. Una preocupación por task — si toca backend Y frontend/mobile, separar en dos tasks.
6. Tests son SIEMPRE una task separada de la implementación.
7. Las dependencias deben ser explícitas — listar los `name` de tasks previas en `dependencies`.
8. **Puntos Fibonacci:** 1, 2, 3, 5, 8. Si una task se estima en 13+ → dividir.

| Puntos | Referencia |
|--------|-----------|
| 1 | Cambio de configuración, variable de entorno, schema trivial |
| 2 | Un handler o función simple con tests unitarios |
| 3 | Servicio o repositorio con lógica moderada |
| 5 | Integración end-to-end entre componentes existentes |
| 8 | Componente nuevo complejo; si supera 8 → considerar dividir |

Tasks de tipo `setup` raramente superan 2 pts. Tasks de tipo `validation` raramente superan 3 pts.

### Inferencia de prioridad

| Prioridad | Criterio |
|-----------|----------|
| HIGH | Task de tipo `setup` que bloquea ≥2 tasks dependientes; task en el camino crítico del sprint |
| MEDIUM | Task independiente o paralelizable; default cuando no hay criterio explícito en el spec |
| LOW | Documentación, diagramas, tareas de observabilidad o cleanup |

Si el spec declara prioridad explícita para un feature → heredar. Si no → aplicar tabla.

## Inferencia del agente ejecutor

`agent` es obligatorio en todas las tasks. Valores exactos permitidos: `developer-backend`, `developer-frontend`, `developer-mobile`, `tester`, `dba`, `dba-cache`, `dba-broker`, `dba-nosql`, `devops`, `tech-writer`, `security`, `observability`, `diagrammer`. Inferir en este orden:

1. **Por extensión/path del archivo principal o keywords de la task:**
   - `.dart` o path bajo `lib/` → `developer-mobile`
   - `.tsx`, `.jsx`, `.astro`, `.ts` en contexto frontend, o path bajo `src/components/`, `src/pages/`, `src/hooks/` → `developer-frontend`
   - `.go`, `.py`, `.rs`, o path bajo `internal/`, `cmd/`, `pkg/`, `api/` → `developer-backend`
   - `.sql` o path bajo `migrations/`, `schema/` → `dba`
   - `*redis*`, `*cache*` en nombre de archivo o descripción de la task → `dba-cache`
   - `*kafka*`, `*rabbitmq*`, `*nats*`, `*broker*`, `*topic*`, `*queue*` → `dba-broker`
   - `*mongo*`, `*firestore*`, `*dynamo*`, `*vector*`, `*embedding*`, `*elasticsearch*` → `dba-nosql`
   - Path bajo `.github/workflows/`, `Dockerfile`, `docker-compose*`, `*.tf`, `kubernetes/`, `k8s/` → `devops`
   - Task de tipo `validation` que refiere a archivos de test (`*_test.go`, `*.spec.ts`, `*.test.ts`, `*_test.dart`) → `tester`
   - Task que refiere a `*.md`, `README`, `CHANGELOG`, documentación → `tech-writer`
   - Keywords de seguridad: `auth`, `JWT`, `CVE`, `RBAC`, `CORS`, `pentest`, `vulnerability` → `security`
   - Keywords de observabilidad: `OpenTelemetry`, `metrics`, `traces`, `dashboard`, `alerting`, `Grafana`, `Prometheus` → `observability`
   - Task que produce `.drawio` o tiene keywords `diagrama`, `diagram` → `diagrammer`
2. **Si el path es ambiguo** → desempatar con el campo `Dominio` del spec (`mobile`, `frontend`, `backend`, `fullstack`).
3. **Si aún no se infiere** → marcar `developer-[?]` o `[agente-?]` según corresponda, y listar en el output de cierre.

## Design reference

Si el spec tiene sección `## Design References` con `Type != none`, agregar campo `design_reference` en el frontmatter de cada task que toque UI (componentes, pantallas, vistas, frontend visible):

```yaml
design_reference: "<valor Location copiado verbatim del spec>"
```

Omitir el campo en tasks sin UI o cuando `Type: none`.

## Protocolo de escalación

Detener y reportar al humano cuando:

| Condición | Mensaje |
|---|---|
| Tasks superan 15 | `Generé >15 tasks — entregué las 15 primeras por prioridad. ¿Partir el feature en sub-features o ampliar el límite?` |
| Dependencia circular detectada | `Ciclo detectado: [A → B → C → A]. Re-invocar spec-writer para resolver el orden.` |
| Una task requiere decisión técnica no presente en el spec | `Task [X] requiere decisión [Y] no resuelta en el spec. Re-invocar spec-writer.` |

## Flujo de Trabajo

### Inputs requeridos

Antes de comenzar, verificar que el humano haya proporcionado:

- **Path del spec** — archivo fuente (puede llamarse de cualquier forma)
- **Tipo** — feature/historia o épica
- **Destino de escritura** — path local, URL de herramienta externa, o "solo muéstralas en chat"

Si falta alguno, abrir una sección `## Necesito información` con solo las preguntas pendientes y DETENER hasta recibir respuesta.

### Pasos

1. **Leer el spec** — única fuente. No leer ADRs, Architecture Views ni requirements directamente; el spec ya los consolida. No leer código de producción amplio.
2. **Descomponer** — aplicar las reglas de descomposición definidas arriba (orden topológico setup → implementation → integration → validation, máx 15 tasks, Fibonacci 1-2-3-5-8).
3. **Enriquecer cada task** con el template: completar `name`, `type`, `priority`, `agent`, `points`, `milestone`, `feature_id`, `dependencies`, y secciones opcionales (`inputs`, `outputs`, `validation_rules`, `## Interfaces`, `design_reference`).
4. **Preview gate** — mostrar al humano la tabla resumen antes de escribir nada:

   ```
   | ID | Tipo | Agente | Pts | Depende de |
   |---|---|---|---|---|
   ```

   Preguntar literalmente: **"¿Genero los archivos?"** y DETENER hasta recibir confirmación explícita.
   - Si aprueba → continuar al paso 5.
   - Si pide ajustes → incorporarlos, regenerar tasks afectadas y volver al preview.
   - **Excepción**: si el destino fue "solo muéstralas en chat", el preview ES el output final — no preguntar "¿Genero los archivos?".

5. **Escribir los archivos** en el destino confirmado:
   - **feature/historia**: un archivo `.md` por task, nombrado `<FEATURE_ID>-<NN>-<slug>.md`.
   - **épica**: un archivo padre `<FEATURE_ID>-epic-<slug>.md` + un archivo por cada subtask.
   - **Destino externo**: generar en memoria y reportar el contenido para que el humano lo suba — no operar herramientas externas.
   - **Regla de granularidad (sin excepción):** nunca consolidar múltiples tasks en un solo documento, independientemente del destino.

Si se cumple cualquier condición de escalación, detener y reportar con el formato definido en `## Protocolo de escalación`.

## Formato de Salida

Máx 100 palabras. Los archivos ya están escritos — no repetir su contenido.

~~~
Tasks generadas — <feature_id>

**Tipo:** feature / épica
**Archivos generados:** N tasks (+ 1 archivo padre si épica)
**Total pts:** P
**Orden de ejecución:** <ID-01> → <ID-02> → ...
**Tasks críticas (bloqueadoras):** [lista]
**Decisiones abiertas:** [lista o "ninguna"]
**Acción para el humano:** [tasks developer-[?] a confirmar, o "ninguna"]

| ID | Tipo | Agente | Pts | Depende de |
|---|---|---|---|---|
| <FEATURE_ID>-01-<slug> | setup | dba | 2 | — |
| <FEATURE_ID>-02-<slug> | implementation | developer-backend | 3 | 01 |
| <FEATURE_ID>-03-<slug> | validation | tester | 2 | 02 |
~~~
