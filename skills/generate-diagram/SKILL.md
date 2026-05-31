---
name: generate-diagram
description: Crear diagramas Mermaid.js embebidos dentro de archivos Markdown (Architecture Views / ADRs, READMEs, documentación). Usar cuando el usuario diga "diagrama embebido", "Mermaid", "diagrama en el markdown", "flowchart en el doc", "incluir diagrama en Architecture Views / ADRs", "diagrama de secuencia", "ERD", o cuando un documento markdown se beneficie de una visualización inline. Para diagramas standalone editables, usar la skill `drawio` en su lugar.
---

# Skill — generate-diagram (Mermaid embebido)

Produce bloques Mermaid.js sintácticamente válidos para embeber en archivos Markdown. La skill cubre Mermaid v10+ y se enfoca en evitar los errores de sintaxis que más rompen el renderizado.

## Cuándo usar esta skill vs `drawio`

| Necesidad | Skill correcta |
|---|---|
| Diagrama dentro de un `.md` (Architecture Views / ADRs, README, spec) | `generate-diagram` (Mermaid) |
| Diagrama standalone editable (`.drawio`) para arquitectura técnica detallada con shapes ricos, message brokers, gateways | `drawio` |
| Diagrama que viaja con la documentación y debe renderizarse en GitHub/Obsidian/Outline | `generate-diagram` |
| Diagrama que se exporta a PNG/SVG para slides o whiteboards técnicos | `drawio` |

Regla: si el output final es un archivo `.md` que renderiza el diagrama inline, usar Mermaid. Si el artefacto es un archivo independiente que se edita visualmente, usar drawio.

## Tipos de diagrama soportados

Usar el keyword exacto de apertura. Mermaid v10+ deprecó `graph` — usar siempre `flowchart`.

| Tipo | Keyword de apertura | Caso de uso |
|---|---|---|
| Flowchart horizontal | `flowchart LR` | Flujos de proceso, pipelines, flujo de datos izquierda→derecha |
| Flowchart vertical | `flowchart TD` | Árboles de decisión, jerarquías top-down |
| Diagrama de secuencia | `sequenceDiagram` | Interacción async/sync entre componentes, llamadas API |
| Diagrama de clases | `classDiagram` | Estructura de tipos, herencia, interfaces |
| Diagrama Entidad-Relación | `erDiagram` | Schema de base de datos, relaciones entre tablas |
| Máquina de estados | `stateDiagram-v2` | FSMs, ciclos de vida, transiciones |
| C4 Contexto | `C4Context` | Vista de sistema de alto nivel (requiere plugin en algunos renderers) |
| C4 Containers | `C4Container` | Vista de contenedores (requiere plugin en algunos renderers) |
| Git graph | `gitGraph` | Estrategia de branching, releases |

**Nota sobre C4:** algunos renderers (GitHub básico, Obsidian sin plugin) no soportan `C4Context`/`C4Container`. Si el documento se va a renderizar fuera de un entorno con plugin Mermaid actualizado, preferir `flowchart LR` con `subgraph` para representar contextos y contenedores.

---

## Reglas de caracteres especiales en labels (CRÍTICO)

La causa más común de Mermaid roto son caracteres especiales sin escapar dentro de labels. Aplicar estas reglas siempre.

### Dentro de `[]`, `()`, `{}` (formas de nodo)

Caracteres prohibidos sin envolver en comillas: `:`, `(`, `)`, `"`, `/`, `#`, `&`.

Reglas:

- Si el label tiene **paréntesis** → envolver el texto completo en comillas: `A["Crear Orden (v2)"]`
- Si el label tiene **dos puntos** → reemplazar por ` -` o reformular: `A[Estado - activo]` en vez de `A[Estado: activo]`
- Si el label tiene **slash** → envolver en comillas: `A["Lectura/Escritura"]`
- Si el label tiene **comillas internas** → escapar con `&quot;` o reformular
- **Regla general:** ante cualquier duda, envolver el label en comillas dobles. Un label con comillas siempre es seguro.

### Dentro de labels de flecha (`-- texto -->`)

- Caracteres problemáticos: `:`, `-` solo (puede confundirse con flecha), `|` (delimita labels en `-->|texto|`)
- Dos puntos en label de flecha → reformular sin `:`
- Guion solo → reemplazar por `—` (em dash) o palabra equivalente
- Para labels complejos, preferir la forma con pipes: `A -- "texto: con caracteres" --> B` también requiere comillas si tiene `:`

### Comillas escapadas

Si el label necesita comillas internas: `A["Etiqueta con \"interna\""]` no es válido en Mermaid. Usar `&quot;`: `A["Etiqueta con &quot;interna&quot;"]`.

---

## Reglas de subgraph

```mermaid
flowchart LR
    subgraph backend [Backend Service]
        api[API Handler]
        store[Store]
    end
    subgraph database [PostgreSQL]
        users[(users)]
    end
    api --> store --> users
```

Reglas:

- El **ID del subgraph** (primer token después de `subgraph`) NO puede tener espacios, paréntesis ni caracteres especiales. Usar `camelCase`, `snake_case` o `kebab-case`.
- El **título legible** va entre `[...]` después del ID. Aplican las mismas reglas de caracteres que para nodos.
- Siempre cerrar con `end` en su propia línea.
- Los IDs deben ser únicos en todo el diagrama — un nodo y un subgraph no pueden compartir ID.

---

## Snippets mínimos funcionales

### Flowchart (flujo de datos)

```mermaid
flowchart LR
    client[Cliente Web] --> gateway[API Gateway]
    gateway --> auth[Auth Service]
    gateway --> orders[Orders Service]
    orders --> db[(PostgreSQL)]
    orders -- "evento OrderCreated" --> bus[Event Bus]
    bus --> notify[Notification Worker]
```

### Diagrama de secuencia

```mermaid
sequenceDiagram
    participant C as Cliente
    participant API as API Service
    participant DB as PostgreSQL
    C->>API: POST /orders
    API->>DB: INSERT order
    DB-->>API: order_id
    API-->>C: 201 Created
    Note over API,DB: Transacción dura menos de 50ms
```

### Diagrama Entidad-Relación

```mermaid
erDiagram
    USER ||--o{ ORDER : "crea"
    ORDER ||--|{ ORDER_ITEM : "contiene"
    PRODUCT ||--o{ ORDER_ITEM : "referencia"
    USER {
        uuid id PK
        string email UK
        timestamp created_at
    }
    ORDER {
        uuid id PK
        uuid user_id FK
        string status
        decimal total
    }
```

### Diagrama de clases

```mermaid
classDiagram
    class Order {
        +UUID id
        +UUID userId
        +OrderStatus status
        +calculateTotal() Money
        +cancel() void
    }
    class OrderItem {
        +UUID productId
        +int quantity
        +Money unitPrice
    }
    Order "1" *-- "many" OrderItem
```

---

## Checklist de validación (OBLIGATORIO antes de entregar)

Antes de cerrar cualquier documento que incluya un bloque Mermaid, verificar cada punto:

- [ ] El bloque abre con el keyword correcto (`flowchart LR`/`TD`, `sequenceDiagram`, `erDiagram`, `classDiagram`, `stateDiagram-v2`, `C4Context`/`C4Container`, `gitGraph`). **Nunca `graph`** — está deprecated en Mermaid v10+.
- [ ] Todos los nodos referenciados en edges/flechas están definidos en el diagrama (no hay edges huérfanos a IDs inexistentes).
- [ ] Ningún label sin comillas contiene `:`, `(`, `)`, `/`, `#`, `&` o `"` interno.
- [ ] Todos los subgraphs tienen ID sin espacios ni caracteres especiales y cierran con `end`.
- [ ] El bloque está encerrado en triple backtick con `mermaid` como language tag: ` ```mermaid ... ``` `.
- [ ] El diagrama cabe en una pantalla estándar (objetivo ≤15 nodos, ≤20 edges). Si excede, partir en varios diagramas con un foco distinto cada uno.
- [ ] Si se usa `C4Context`/`C4Container`, está documentado en el texto que requiere plugin Mermaid actualizado.

Si algún ítem falla → corregir antes de entregar.

---

## Anti-patrones comunes

| Anti-patrón | Forma correcta | Por qué |
|---|---|---|
| `graph LR` | `flowchart LR` | `graph` está deprecated en Mermaid v10+ |
| `A[Create Order (v2)]` | `A["Create Order (v2)"]` | Paréntesis sin comillas rompen el parser |
| `A -- texto: con dos puntos --> B` | `A -- "texto con guión" --> B` | Dos puntos en labels confunden al parser |
| `A[Estado: activo]` | `A[Estado - activo]` o `A["Estado: activo"]` | Dos puntos rompen labels de nodo |
| `subgraph My Backend` | `subgraph backend [My Backend]` | El ID no admite espacios |
| `A[Read/Write]` | `A["Read/Write"]` | Slash en label sin comillas |
| `A --> B --> ` (edge incompleto) | `A --> B` | Edges colgantes producen errores opacos |
| Definir `B` solo en un edge `A --> B` y nunca darle forma | `A --> B[Etiqueta de B]` | Sin forma definida, el nodo aparece como ID literal |
| Mezclar `participant` y nodos de flowchart | Elegir UNO: o `sequenceDiagram` o `flowchart` | Cada keyword tiene su propio léxico |

---

## Reglas operativas

- Colocar siempre el diagrama dentro de un bloque ` ```mermaid ... ``` `.
- Un diagrama por bloque — no concatenar dos tipos distintos en el mismo bloque.
- Preferir **diagramas de secuencia** para flujos asíncronos complejos (orden de mensajes importa).
- Preferir **flowchart** para vistas de arquitectura cuando el renderer no soporta C4.
- Preferir **ER** para schema de base de datos — nunca usar flowchart para representar tablas.
- Mantener el diagrama **simple y enfocado**: una idea por diagrama. Si necesitas mostrar dos vistas (datos y secuencia), produce dos diagramas separados.
