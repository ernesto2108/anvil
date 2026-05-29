---
name: drawio
description: Generar archivos `.drawio` (XML editable) para diagramas técnicos de arquitectura, flujo de datos, conexiones entre servicios, message brokers, bases de datos, gateways y clientes. Úsalo cuando el usuario pida "diagrama", "diagrama de arquitectura", "visualiza el flujo", "drawio", "grafica las conexiones", "muéstrame cómo se comunica X con Y", o cuando un agente downstream necesite producir un `.drawio` validado. Cubre shapes, estilos, colores convencionales por rol de nodo y reglas de layout horizontal/vertical.
---

# Skill — draw.io (`.drawio`) para diagramas técnicos

## Filosofía

1. **Editable antes que bonito** — el output es XML legible. El usuario lo abre en draw.io desktop o en app.diagrams.net y lo retoca; no es un PNG terminal.
2. **Una vista lógica por archivo** — un `.drawio` muestra una capa (flujo de datos, despliegue, secuencia). Si una historia necesita dos vistas → dos archivos, no un solo archivo con dos páginas mezcladas.
3. **Color como semántica, no decoración** — el color de un shape comunica el rol del nodo (productor / consumidor / broker / DB / gateway). Quien lea el diagrama debe poder inferir el rol sin leer el label.

## Cuándo activarse

Se activa cuando el agente que carga esta skill necesita producir archivos `.drawio`. Los disparadores típicos en el prompt del usuario:

- "diagrama", "diagrámame", "diagram"
- "visualiza", "grafica", "dibuja"
- "muéstrame cómo está conectado", "cómo se comunican", "cómo fluye"
- "drawio", "draw.io"

## Estructura base del XML

Un archivo `.drawio` mínimo tiene este esqueleto. Todo shape o conector vive como `mxCell` dentro de `root`:

```xml
<mxfile host="app.diagrams.net" modified="2026-01-01T00:00:00.000Z" agent="diagrammer" version="24.0.0">
  <diagram id="diagram-1" name="Page-1">
    <mxGraphModel dx="1422" dy="762" grid="1" gridSize="10" guides="1"
      tooltips="1" connect="1" arrows="1" fold="1" page="0"
      pageScale="1" pageWidth="1169" pageHeight="827"
      math="0" shadow="0" fit="0" border="50">
      <root>
        <mxCell id="0" />
        <mxCell id="1" parent="0" />
        <!-- shapes y conectores aquí -->
      </root>
    </mxGraphModel>
  </diagram>
</mxfile>
```

**Atributos clave de `mxGraphModel`:**

- `page="0"` — desactiva el borde de página visible; el canvas se expande libremente con el contenido.
- `fit="0"` — no escala el contenido para caber en la página.
- `border="50"` — margen mínimo de 50px alrededor del contenido.
- `dx`/`dy` altos (1422×762) para viewport inicial amplio.

> **Nunca usar `page="1"` con contenido denso — fuerza compresión visual. Siempre `page="0"` para diagramas técnicos.**

**Reglas estructurales:**

- `id="0"` y `id="1"` son obligatorios y reservados (root y capa raíz). Nunca reasignarlos.
- Todo `mxCell` propio usa `parent="1"` salvo que viva dentro de un container.
- Los IDs son strings únicos en el archivo — usar slugs descriptivos (`svc-orders`, `db-events`) en vez de UUIDs cuando se pueda.

## Anatomía de un nodo (shape)

```xml
<mxCell id="svc-orders" value="Orders Service" style="rounded=1;whiteSpace=wrap;html=1;fillColor=#dae8fc;strokeColor=#6c8ebf;" vertex="1" parent="1">
  <mxGeometry x="120" y="80" width="160" height="60" as="geometry" />
</mxCell>
```

- `value` → label visible.
- `style` → string de pares `clave=valor;` separados por `;`. Es el contrato visual del shape.
- `vertex="1"` → es un nodo, no un edge.
- `mxGeometry` → posición (`x`, `y`) y tamaño (`width`, `height`) en píxeles.

## Anatomía de un conector (edge)

```xml
<mxCell id="e1" value="publishes OrderCreated" style="endArrow=classic;html=1;rounded=0;edgeStyle=orthogonalEdgeStyle;" edge="1" parent="1" source="svc-orders" target="topic-orders">
  <mxGeometry relative="1" as="geometry" />
</mxCell>
```

- `edge="1"` → es una flecha.
- `source` y `target` → IDs de los nodos conectados.
- `value` → label sobre la flecha (queda flotante en el medio).
- `edgeStyle=orthogonalEdgeStyle` → recomendado para diagramas técnicos (líneas a 90°).

### Labels intermedios en conectores

Cuando el label no debe ir centrado sino más cerca de un extremo, usar un `mxCell` hijo del edge con tipo `edgeLabel`:

```xml
<mxCell id="e1-lbl" value="async" style="edgeLabel;html=1;align=center;verticalAlign=middle;resizable=0;points=[];" vertex="1" connectable="0" parent="e1">
  <mxGeometry x="-0.3" relative="1" as="geometry">
    <mxPoint as="offset" />
  </mxGeometry>
</mxCell>
```

`x` en `mxGeometry` va entre `-1` y `1` (`-1` cerca del source, `1` cerca del target, `0` centro).

### Estilos de conector por caso de uso

**Flujo normal (mismo swimlane / grupo):**
```xml
style="edgeStyle=orthogonalEdgeStyle;rounded=0;html=1;"
```

**Cross-swimlane (entre grupos distintos):**
```xml
style="edgeStyle=elbowEdgeStyle;elbow=orthogonal;exitX=0.5;exitY=1;exitDx=0;exitDy=0;entryX=0.5;entryY=0;entryDx=0;entryDy=0;rounded=0;html=1;"
```

**Self-loop (estado a sí mismo):**
```xml
style="edgeStyle=orthogonalEdgeStyle;exitX=1;exitY=0.5;exitDx=0;exitDy=0;entryX=1;entryY=1;entryDx=0;entryDy=0;rounded=1;html=1;curved=1;"
```

**Transición crítica / riesgo:**
```xml
style="edgeStyle=orthogonalEdgeStyle;rounded=0;html=1;strokeColor=#FF0000;dashed=1;strokeWidth=2;"
```

**Regla de labels en edges:** agregar siempre `align=center;verticalAlign=bottom;` para evitar solapamiento con nodos destino. Labels > 25 chars usar `fontSize=9`.

### Reglas de separación de labels (anti-superposición)

Estas reglas existen porque el XML válido no garantiza un diagrama legible. La causa #1 de diagramas ilegibles es **labels de edges que se pisan entre sí** cuando hay múltiples conexiones cercanas. Aplicar siempre:

1. **Máximo 2 edges directos entre el mismo par de nodos.** Si hay más, **combinar en un solo edge** con label multi-línea usando `&#xa;` (entidad XML del salto de línea) como separador. Nunca 3+ edges paralelos entre `A → B`.

   ```xml
   <!-- En vez de 3 edges entre svc-orders y svc-payments -->
   <mxCell id="e1" value="createOrder&#xa;cancelOrder&#xa;refundOrder" style="endArrow=classic;html=1;edgeStyle=orthogonalEdgeStyle;" edge="1" parent="1" source="svc-orders" target="svc-payments">
     <mxGeometry relative="1" as="geometry" />
   </mxCell>
   ```

2. **Offsets obligatorios para edges paralelos** — cuando hay exactamente 2 edges entre el mismo par de nodos:
   - **Edge 1:** `exitX="0.5" exitY="0"` (sale por arriba) + `edgeLabel` hijo con `x="-0.5"` (label hacia el source).
   - **Edge 2:** `exitX="0.5" exitY="1"` (sale por abajo) + `edgeLabel` hijo con `x="0.5"` (label hacia el target).
   - Alternativa: variar `exitX` en el mismo lado (`0.3` y `0.7`) para separar los puntos de salida.

   ```xml
   <mxCell id="e1" value="" style="endArrow=classic;html=1;exitX=0.5;exitY=0;entryX=0.5;entryY=1;edgeStyle=orthogonalEdgeStyle;" edge="1" parent="1" source="A" target="B">
     <mxGeometry relative="1" as="geometry" />
   </mxCell>
   <mxCell id="e1-lbl" value="request" style="edgeLabel;html=1;align=center;" vertex="1" connectable="0" parent="e1">
     <mxGeometry x="-0.5" relative="1" as="geometry"><mxPoint as="offset" /></mxGeometry>
   </mxCell>
   <mxCell id="e2" value="" style="endArrow=classic;html=1;exitX=0.5;exitY=1;entryX=0.5;entryY=0;edgeStyle=orthogonalEdgeStyle;" edge="1" parent="1" source="A" target="B">
     <mxGeometry relative="1" as="geometry" />
   </mxCell>
   <mxCell id="e2-lbl" value="response" style="edgeLabel;html=1;align=center;" vertex="1" connectable="0" parent="e2">
     <mxGeometry x="0.5" relative="1" as="geometry"><mxPoint as="offset" /></mxGeometry>
   </mxCell>
   ```

3. **Puntos de entrada/salida escalonados** — cuando un nodo tiene 3+ edges salientes hacia distintos targets en la misma dirección, escalonar los `exitX` (o `exitY` si es vertical): `0.25`, `0.5`, `0.75`. Nunca todos por `exitX="0.5"` — eso colapsa los puntos de salida y solapa los segmentos iniciales.

4. **Labels largos en edges (> 20 caracteres)** — usar `edgeLabel` hijo con `align="center"` y considerar:
   - Abreviar el label (ej. `publishes OrderCreatedEvent` → `OrderCreated`).
   - O mover el texto completo al atributo `tooltip` del edge y dejar un label corto visible (`evt` con tooltip `publishes OrderCreated event with payload {id, customer, items}`).

   ```xml
   <mxCell id="e3" value="evt" tooltip="publishes OrderCreated event with payload {id, customer, items}" style="endArrow=classic;html=1;edgeStyle=orthogonalEdgeStyle;" edge="1" parent="1" source="svc-orders" target="topic-orders">
     <mxGeometry relative="1" as="geometry" />
   </mxCell>
   ```

5. **Self-loops** — siempre con `exitX="1" exitY="0" entryX="1" entryY="0.5"` y un `edgeLabel` hijo con offset `x="1"` (label a la derecha del loop, nunca superpuesto al nodo).

   ```xml
   <mxCell id="e-loop" value="" style="endArrow=classic;html=1;exitX=1;exitY=0;entryX=1;entryY=0.5;edgeStyle=orthogonalEdgeStyle;curved=1;" edge="1" parent="1" source="state-retry" target="state-retry">
     <mxGeometry relative="1" as="geometry" />
   </mxCell>
   <mxCell id="e-loop-lbl" value="retry" style="edgeLabel;html=1;align=left;" vertex="1" connectable="0" parent="e-loop">
     <mxGeometry x="1" relative="1" as="geometry"><mxPoint as="offset" /></mxGeometry>
   </mxCell>
   ```

6. **Cross-swimlane edges** — cuando un edge cruza de un swimlane a otro, usar **waypoints explícitos** (un `Array` dentro de `mxGeometry`) para forzar una ruta ortogonal limpia que no atraviese otros nodos:

   ```xml
   <mxCell id="e-cross" value="syncs" style="endArrow=classic;html=1;edgeStyle=orthogonalEdgeStyle;exitX=1;exitY=0.5;entryX=0;entryY=0.5;" edge="1" parent="1" source="svc-a-in-lane-1" target="svc-b-in-lane-2">
     <mxGeometry relative="1" as="geometry">
       <Array as="points">
         <mxPoint x="480" y="200" />
         <mxPoint x="480" y="360" />
       </Array>
     </mxGeometry>
   </mxCell>
   ```

## Catálogo de shapes para diagramas técnicos

| Rol del nodo | `style` recomendado | Notas |
|---|---|---|
| Servicio / microservicio | `rounded=1;whiteSpace=wrap;html=1;` | Rectángulo redondeado. Variante con bordes más suaves: agregar `arcSize=20;`. |
| API Gateway | `shape=mxgraph.networking.firewall;html=1;` o rectángulo con etiqueta `[Gateway]` | Si `shape=mxgraph.networking.firewall` no está disponible en el viewer, usar rectángulo redondeado con color morado y label `API Gateway`. |
| Base de datos (relacional o cualquier DB) | `shape=cylinder3;whiteSpace=wrap;html=1;boundedLbl=1;backgroundOutline=1;size=15;` | Forma cilíndrica nativa de draw.io. Alternativa más detallada: `shape=mxgraph.flowchart.database`. |
| Kafka / topic | `shape=mxgraph.kafka.topic;html=1;labelPosition=center;verticalLabelPosition=bottom;align=center;verticalAlign=top;` | Si la stencil Kafka no está disponible en el viewer: usar rectángulo con label `[Kafka topic: <name>]` y color naranja. |
| Queue / cola (RabbitMQ, SQS, NATS) | `shape=mxgraph.flowchart.delay;whiteSpace=wrap;html=1;` o rectángulo etiquetado `[Queue]` | El shape `delay` es un rectángulo con un lado curvo, visualmente reconocible como cola. |
| Cliente / browser | `shape=mxgraph.bootstrap.monitor;html=1;` o rectángulo con label `Browser` | Para móvil: `shape=mxgraph.android.phone2;html=1;`. |
| Nube genérica (AWS, GCP, Azure, externa) | `ellipse;shape=cloud;whiteSpace=wrap;html=1;` | Para nubes específicas de provider usar las stencils oficiales (`shape=mxgraph.aws4.group;grIcon=mxgraph.aws4.group_account;`). Si el viewer no las soporta → cloud genérico + label. |
| Container/pod (k8s) | `rounded=1;whiteSpace=wrap;html=1;dashed=1;` | Rectángulo redondeado punteado, sirve para envolver shapes. |
| Persona / actor (usuario humano) | `shape=umlActor;verticalLabelPosition=bottom;labelPosition=center;verticalAlign=top;html=1;` | Igual que en UML. |

**Regla de fallback:** si una stencil específica (Kafka, AWS, GCP) no está disponible en el viewer del usuario, degradar a un rectángulo con el color convencional y un label claro entre corchetes (`[Kafka topic: events]`). Mejor un diagrama portable que uno bonito que no abre.

## Convención de colores por rol de nodo

Esta es la regla semántica del diagrama. Usar SIEMPRE estos colores cuando el rol aplique:

| Rol | `fillColor` | `strokeColor` | Cuándo usar |
|---|---|---|---|
| Productor (publica eventos / envía mensajes) | `#dae8fc` (azul claro) | `#6c8ebf` (azul medio) | Servicios que escriben a un broker, productores Kafka, emisores de eventos. |
| Consumidor (lee eventos / recibe mensajes) | `#d5e8d4` (verde claro) | `#82b366` (verde medio) | Servicios que se suscriben a topics/queues. |
| Broker / message bus | `#ffe6cc` (naranja claro) | `#d6b656` (naranja medio) | Kafka, RabbitMQ, NATS, SQS, EventBridge. |
| Base de datos (cualquier motor) | `#f5f5f5` (gris claro) | `#666666` (gris oscuro) | Postgres, MySQL, MongoDB, DynamoDB, Redis, Elasticsearch. |
| API Gateway / edge / load balancer | `#e1d5e7` (morado claro) | `#9673a6` (morado medio) | Kong, Nginx, ALB, API Gateway de AWS, BFF. |
| Cliente externo (browser, mobile, partner) | `#fff2cc` (amarillo claro) | `#d6b656` (amarillo medio) | El usuario final, una integración externa que llama al sistema. |
| Servicio neutro (sin rol específico) | sin fill (`fillColor=none`) o blanco | `#000000` | Cuando el nodo no tiene un rol semántico claro. Evitar abusar. |

**Composición del `style` con color** (ejemplo productor Kafka):

```
rounded=1;whiteSpace=wrap;html=1;fillColor=#dae8fc;strokeColor=#6c8ebf;
```

**Cuando un nodo cumple dos roles** (ej. un servicio que es productor de un topic y consumidor de otro): elegir el rol dominante para el color y declarar el secundario en el label (`Orders (consumer of payments)`). Nunca usar gradientes — son ruido visual.

## Reglas de layout

### Horizontal (izquierda → derecha)

Usar para **flujos de datos** o **request paths**:

- Cliente → Gateway → Servicio → DB
- Producer → Broker → Consumer
- Eje X representa el avance temporal del flujo.

Posicionamiento típico (anchos de 160px, alto 60px, gap horizontal 80px):

```
Cliente (x=40)  →  Gateway (x=280)  →  Service (x=520)  →  DB (x=760)
```

### Vertical (arriba → abajo)

Usar para **stacks** o **niveles de despliegue**:

- Edge → App → Persistencia
- Frontend → BFF → Backend → DB
- Eje Y representa capas o tiers.

### Cuándo cada uno

| Tipo de diagrama | Layout |
|---|---|
| Flujo de datos entre servicios | Horizontal |
| Pipeline de mensajería (producer → broker → consumer) | Horizontal |
| Request path (cliente → backend → DB) | Horizontal |
| Stack de capas / tiers | Vertical |
| Diagrama de despliegue por entorno (dev/staging/prod) | Vertical (un entorno por fila) |
| Componentes dentro de un mismo servicio | Lo que rinda más legible — sin regla fija |

### Reglas duras de layout

1. **Alineación en grid** — `x` e `y` múltiplos de 20 (mejor: múltiplos de 40). Los conectores ortogonales se ven limpios solo si los nodos están alineados.
2. **Sin cruces de líneas evitables** — si dos conectores se cruzan, mover un nodo. Si no se puede evitar, OK; pero nunca por flojera.
3. **Labels legibles** — los labels en conectores no deben solapar shapes. Si la línea es muy corta para el label, separar más los nodos.
4. **Containers para agrupar** — cuando 3+ shapes pertenecen al mismo dominio/microservicio/cluster, envolverlos en un container (`shape=swimlane;` o rectángulo punteado) con un label de grupo.

## Patrones de layout por tipo de diagrama

### Pipeline / flujo de datos (horizontal)
- Nodos izq → der en una sola fila
- Gap horizontal entre nodos: 100px
- Usar cuando hay ≤ 8 nodos en línea recta

### State machine / máquina de estados
- Layout: **grid top-down dentro de cada swimlane o grupo**
- Columnas por swimlane: `ceil(sqrt(nodos_en_swimlane))`, mínimo 2
- Ancho de nodo: `max(180, label_chars * 9)px` — nunca truncar texto
- Alto de nodo: 60px labels cortos / 80px para labels > 20 chars
- Gap horizontal entre nodos: 120px
- Gap vertical entre nodos: 80px
- Swimlane height: `(filas * alto_nodo) + (filas * gap_vertical) + 80px de padding`
- Swimlane width: `(cols * ancho_nodo) + (cols * gap_horizontal) + 80px de padding`
- **Regla dura: nunca más de 4 nodos en una sola fila dentro de un swimlane**

### Arquitectura de microservicios (vertical por capas)
- Capas: cliente → gateway → servicios → datos
- Gap vertical entre capas: 120px
- Alinear horizontalmente servicios del mismo tier

### Despliegue / infraestructura
- Layout vertical: internet → load balancer → pods → DB
- Usar containers anidados para namespaces / VPCs

## Flujo de trabajo

1. **Identificar el tipo de diagrama** del input:
   - Flujo de mensajería → horizontal, color por rol (productor / broker / consumidor).
   - Arquitectura general → horizontal o vertical según si predomina flujo o stack.
   - Despliegue → vertical, agrupar por entorno.
2. **Enumerar nodos con su rol semántico** — cada nodo del input se clasifica en uno de los roles de la tabla de colores. Si no encaja en ninguno, marcarlo neutro.
3. **Enumerar conectores** — para cada conector identificar: source, target, label (qué transporta), si es síncrono o asíncrono. Asíncrono → label sugerido con prefijo `async:` o estilo punteado (`dashed=1`).
4. **Asignar coordenadas** en grid (paso 40) según el layout elegido.
5. **Escribir el XML** — un `mxCell` por nodo, un `mxCell` por edge, validar que cada `source`/`target` exista como ID.
6. **Auto-QA antes de entregar:**
   - ¿XML cierra todos los tags? (`mxfile`, `diagram`, `mxGraphModel`, `root`)
   - ¿Cada `mxCell` no-raíz tiene `parent`?
   - ¿Cada edge tiene `source` y `target` con IDs que existen?
   - ¿El color de cada nodo respeta la convención por rol?
   - ¿No hay nodos solapados (mismas coordenadas)?

## Reglas de output

- **Ubicación por defecto:** archivos en `{task_path}/diagrams/<nombre>.drawio`. Si el humano pasa un path explícito, usar ese.
- **Un archivo `.drawio` por diagrama lógico.** No meter dos vistas distintas en un mismo archivo aunque la API de draw.io lo permita con páginas — separar facilita el versionado y el reuso.
- **Nombre de archivo descriptivo y en kebab-case:** `order-flow.drawio`, `events-pipeline.drawio`, `deploy-prod.drawio`. NUNCA `diagram.drawio` o `untitled.drawio`.
- **Encoding UTF-8 sin BOM.** El `<?xml version="1.0" encoding="UTF-8"?>` al inicio es opcional pero recomendado.

## Reglas

- **No inventar nodos ni conexiones.** Diagramar exclusivamente lo que está en el input. Si falta información, marcarlo como pregunta abierta al humano y NO completar con asunciones.
- **No mezclar capas.** Un diagrama de flujo de datos no incluye nodos de infra (k8s pods, ALBs) salvo que sean relevantes al flujo descrito.
- **No omitir labels en conectores.** Toda flecha lleva label salvo que el contexto haga obvio el contenido (ej. `Cliente → API Gateway` sin label cuando el diagrama trata exclusivamente de routing).
- **Validar XML antes de entregar.** Si el agente downstream no puede correr un parser, releer el archivo mentalmente buscando tags sin cerrar.
- **Usar IDs descriptivos** (`svc-orders`, `db-events`, `topic-payments`) — facilita el debugging y el diff posterior.

## Anti-patrones

| Anti-patrón | Por qué | Corrección |
|---|---|---|
| Un solo archivo con flujo + despliegue + secuencia mezclados | Imposible de mantener, ilegible | Un archivo por vista |
| Colores arbitrarios (rosa, turquesa, marrón) | Pierde la semántica de la convención | Usar la tabla de colores por rol |
| Nodos sin label | No se puede leer el diagrama sin código adjunto | Todo nodo lleva `value` |
| Edges sin source/target o con IDs inexistentes | Edge "huérfano" — draw.io lo renderiza pero queda flotante | Validar que cada `source`/`target` exista |
| Coordenadas arbitrarias (x=137, y=223) | Conectores ortogonales quedan torcidos | Snap a grid (múltiplos de 20 o 40) |
| Mezclar `edgeStyle=orthogonalEdgeStyle` con cruces evitables | Confunde más de lo que ayuda | Reorganizar nodos para evitar el cruce |
| Inventar conexiones que el input no mencionó | Diagrama miente sobre la arquitectura real | Solo diagramar lo confirmado; listar gaps al humano |
| Usar `page="1"` o `fit="1"` con diagramas densos | Comprime el contenido a la hoja, nodos se solapan y labels se truncan | Usar siempre `page="0"` y `fit="0"` para diagramas técnicos |
| Más de 4 nodos en una fila horizontal dentro de un swimlane | Forma una línea apretada con labels solapados | Romper en grid top-down: `ceil(sqrt(N))` columnas |
| Ancho fijo de 160px para labels > 15 caracteres | El label se trunca o se desborda visualmente | Calcular ancho como `max(180, label_chars * 9)` |
| Edges cross-swimlane sin `exitX/exitY/entryX/entryY` explícitos | Ruteo caótico — la línea cruza el contenido del swimlane origen o destino | Usar `elbowEdgeStyle` con puntos de entrada/salida explícitos |
| Self-loops sin `exitX=1;entryX=1` | El loop queda invisible detrás del nodo | Salir y entrar por el lado derecho: `exitX=1;exitY=0.5;entryX=1;entryY=1` |
| Tres o más edges paralelos entre el mismo par de nodos | Los labels se pisan en la zona central, ilegible | Combinar en un solo edge con label multi-línea usando `&#xa;` como separador |
| Todos los edges de un nodo saliendo por `exitX=0.5` | Los segmentos iniciales colapsan en el mismo punto y se superponen | Escalonar a `exitX=0.25 / 0.5 / 0.75` (o `exitY` si es vertical) |
| Label de edge de 30+ caracteres sin abreviar | Se solapa con nodos vecinos o con otros labels de edges | Cortar a < 20 chars o mover el texto largo a `tooltip` y dejar un label corto visible |

## Checklist final antes de entregar

- [ ] XML cierra todos los tags (`mxfile`, `diagram`, `mxGraphModel`, `root`)
- [ ] `mxCell id="0"` y `mxCell id="1"` presentes
- [ ] Cada nodo tiene `vertex="1"` y `parent="1"`
- [ ] Cada edge tiene `edge="1"`, `source` y `target` válidos
- [ ] Color de cada nodo respeta la tabla de roles
- [ ] Layout elegido (horizontal / vertical) y aplicado consistentemente
- [ ] Coordenadas en grid de 20/40
- [ ] Labels presentes en nodos y en conectores con contenido no obvio
- [ ] Archivo guardado en `{task_path}/diagrams/<nombre-descriptivo>.drawio`
- [ ] Sin nodos ni conexiones inventadas — todo trazable al input
- [ ] `page="0"` y `fit="0"` presentes en `mxGraphModel`
- [ ] Ningún swimlane tiene más de 4 nodos en una fila
- [ ] Ancho de cada nodo ≥ `max(180, label_chars * 9)`px
- [ ] Self-loops usan `exitX=1;exitY=0.5;entryX=1;entryY=1`
- [ ] Edges cross-swimlane usan `elbowEdgeStyle` con `exitX/exitY/entryX/entryY` explícitos
- [ ] Labels de edges no solapan shapes (usar `verticalAlign=bottom`)
- [ ] Nunca más de 2 edges directos entre el mismo par de nodos (si hay más, combinar con label multi-línea `&#xa;`)
- [ ] Cuando un nodo tiene 3+ edges salientes en la misma dirección, los `exitX` (o `exitY`) están escalonados (`0.25/0.5/0.75`), no todos por `0.5`
- [ ] Edges paralelos entre el mismo par usan `exitX/exitY` opuestos + `edgeLabel` con offsets `x="-0.5"` y `x="0.5"` para separar labels
- [ ] Labels de edges > 20 chars están abreviados o movidos a `tooltip`
