# Pencil Design Tool — Flujo de Trabajo del Design System

Cómo trabajar con archivos `.pen` usando las herramientas MCP de Pencil de forma económica en tokens. Los patrones de composición (variables, componentes reutilizables, slots, iconos, cards, tablas) viven en los guidelines nativos de Pencil — NO los dupliques aquí.

## get_guidelines — fuente de verdad nativa

Pencil expone guías ricas, actualizadas automáticamente, que NO cuestan tokens hasta que se llaman. Llámalas just-in-time según el tipo de pantalla:

| Tipo de pantalla / tarea | Guideline a cargar |
|---|---|
| Pantalla de web app (dashboard, settings, listados) | `get_guidelines("guide","Web App")` |
| Pantalla mobile | `get_guidelines("guide","Mobile App")` |
| Landing page / marketing | `get_guidelines("guide","Landing Page")` |
| Tablas, dashboards de datos | `get_guidelines("guide","Table")` |
| Cuando usas componentes del design system (slots, iconos, sidebar, cards, paginación) | `get_guidelines("guide","Design System")` |
| Generar código desde un `.pen` | `get_guidelines("guide","Code")` |

Regla: carga el guideline relevante UNA VEZ antes de construir, no por pantalla.

## Protocolo de lectura económica del `.pen`

Nunca leas el documento completo "por si acaso". Sube de profundidad solo cuando lo necesites:

1. `get_editor_state(include_schema:false)` — usa `include_schema:true` SOLO en la primera llamada por sesión (cuando el schema aún no está en contexto)
2. `snapshot_layout({maxDepth:0})` — lista de frames top-level sin bajar a hijos. Tu mapa inicial del canvas
3. `batch_get({patterns:[{reusable:true}], searchDepth:2, readDepth:1})` — solo los componentes reutilizables
4. `batch_get({nodeIds:[...]})` — solo los nodos específicos que vas a tocar

Aplica el mínimo nivel suficiente. No llames el paso 4 con patrones amplios (`["*"]`) salvo que realmente necesites todo el árbol.

## Convención de tokens (`$--*`)

El design system nativo de Pencil usa el prefijo `$--*`: `$--background`, `$--foreground`, `$--primary`, `$--muted-foreground`, `$--card`, `$--border`, `$--font-primary`, `$--font-secondary`, `$--radius-m`, `$--radius-pill`. Los tokens propios de un proyecto pueden llamarse distinto, pero al trabajar con el design system nativo usa `$--*`. Para crear o ajustar variables, sigue lo que indica `get_guidelines("guide","Design System")`.

## Flujo de Iteración (Solicitudes de Cambio)

Al modificar un diseño existente (NO creando desde cero). **Nunca elimines y recrees lo que ya existe.**

### Paso 0 — Entiende qué existe

1. `get_editor_state(include_schema:false)` → identifica el `.pen` abierto
2. `snapshot_layout({maxDepth:0})` → mapa top-level; baja con `batch_get` selectivo solo a los nodos afectados
3. `get_variables()` → entiende los tokens actuales
4. Identifica los nodos específicos afectados por la solicitud

### Paso 1 — Clasifica el cambio

| Tipo de cambio | Acción | Ejemplo |
|---|---|---|
| **Token** (color, fuente, espaciado) | `set_variables()` con solo los valores actualizados | "Primario más oscuro" → actualiza la variable |
| **Contenido** (texto, iconos) | `Update(nodeId,{content:"..."})` en cada instancia | "Cambiar heading" → actualiza nodos de texto |
| **Estructura de componente** | Modifica la **madre** — todas las instancias se actualizan | "Agregar icono al card" → edita el reutilizable |
| **Personalización de instancia** | `Update(instanceId+"/childId",{...})` o `descendants` | "Esta card necesita texto distinto" |
| **Layout** (reordenar, redimensionar, agregar/quitar) | `Update()` para reposicionar, `Insert()` solo para nuevos, `Delete()` solo para eliminados explícitos | "Mover sidebar a la derecha" |
| **Nueva pantalla/sección** | Único caso que usa flujo de creación, reutilizando componentes | "Agregar página de settings" |

### Paso 2 — Ejecuta quirúrgicamente

1. **Cambia solo lo que cambió** — no reconstruyas la pantalla
2. **Prefiere cambios de variables** — un cambio de `set_variables()` actualiza todos los nodos que la usan
3. **Prefiere ediciones de la madre** — si aplica a todas las instancias, edita la madre, no cada instancia
4. **Usa `Update` no `Replace`** — `Update` preserva el nodo; `Replace` crea uno nuevo. Solo `Replace` cuando el tipo del nodo deba cambiar
5. **Nunca `Delete`+`Insert` lo que puedes `Update`** — eso es reconstruir, no iterar

### Paso 3 — Verifica

1. `get_screenshot()` de la sección/pantalla afectada
2. Si tocaste una madre de componente, captura también las pantallas que usan sus instancias
3. Si tocaste una variable temática, verifica ambos modos (light/dark)

### Principio clave

**El cambio más rápido y seguro toca el menor número de nodos.** Variable = cero nodos (auto). Madre de componente = un nodo. Instancia = N nodos. Reconstruir = todo. Elige siempre el mayor apalancamiento.

## Limitaciones de Pencil (gotchas)

- **Sin Collections/Modes nativos** — simula Collections con prefijos de nombre y Modes con el eje de tema (`theme:{mode:"dark"}`)
- **Tipos de variables inmutables** — no puedes cambiar el tipo de una variable existente; crea una nueva con otro nombre. `font-weight` debe ser tipo **string** (`"600"`, no `600`)
- **`Update("childId")` sin prefijo de instancia modifica la MADRE** — para personalizar una instancia usa SIEMPRE `Update("instanceId/childId")`. El bare childId corrompe todas las instancias
- **Descendientes de nodos copiados obtienen nuevos IDs** — tras `Copy`, los IDs de hijos se regeneran. Usa `descendants` en la operación `Copy`, o `batch_get` la copia para leer los nuevos IDs
- **Variables string en `content` se resuelven SOLO dentro del contexto de tema** — funcionan en descendientes de instancias (heredan el tema) pero NO en nodos creados via `Replace`. Para i18n confiable, usa `Update(instance+"/childId",{content:"$--var"})`, o hardcodea texto en nodos reemplazados
- **Texto en contenedor con ancho limitado** — DEBE llevar `textGrowth:"fixed-width"` + `width:"fill_container"`, salvo labels/badges/botones de 1 línea
- **`alignItems:"baseline"`** no soportado — usa `"end"`
- **Sin prototipado** — para estados interactivos, crea un frame `{Component} — States` mostrando cada estado etiquetado, junto al componente (no en la Librería)

## Organización del Canvas (OBLIGATORIO)

Deja ~200px de separación entre filas y entre frames horizontales. Cronológico de arriba a abajo. Nunca elimines frames para reorganizar — mueve con `Update(id,{x,y})`.

```
FILA 1 — Librería + Estados de Componentes + Capas Sueltas (capa de referencia, siempre arriba)
FILA 2 — v1 (primera iteración, conservada para historia)
FILA 3 — v2 (iteración actual de pantallas)
FILA 4 — Pantallas Mobile
(más filas abajo a medida que se añaden iteraciones o plataformas)
```

- **Librería siempre FILA 1** — es la referencia, no una pantalla
- **States/capas junto a la Librería** en la misma fila
- **Después de reorganizar**, ejecuta `snapshot_layout({problemsOnly:true})` para verificar que no haya superposiciones

## Nodos de Script (Código en el Canvas)

Para patrones muy repetitivos o data-driven, un nodo de script renderiza output de JS en el canvas. Restricciones: máx 1000 nodos, timeout 2s, sin async, sin DOM/red. Para repeticiones simples (<10 ops) prefiere loops dentro de `batch_design`. Para patrones únicos personalizables, usa instancias de componentes.
