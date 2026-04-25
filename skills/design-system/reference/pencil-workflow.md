# Pencil Design Tool — Flujo de Trabajo del Design System

Cómo crear un design system completo dentro de un archivo `.pen` usando herramientas MCP de Pencil.

## Limitaciones de Pencil vs Figma

Pencil NO tiene Collections o Modes nativos. Simúlalos:
- **Collections** → usa prefijos de convención de nombres: `prim-*`, `sem-*`, `comp-*`
- **Modes** → usa el sistema de eje de tema de Pencil para light/dark
- **Component Properties** → Pencil soporta `reusable: true` pero no propiedades boolean/text/instance-swap de forma nativa. Usa overrides de descendientes en instancias en su lugar
- **Variants** → crea componentes reutilizables separados por variante (Button-Primary, Button-Secondary) o usa switching basado en tema
- **Scoping** → no disponible. Confía en la disciplina de nomenclatura

## Orden de Operaciones

```
1. set_variables()     →  Crea TODAS las variables (simulando collections via nomenclatura)
2. batch_design()      →  Construye componentes reutilizables usando $variables
3. batch_design()      →  Ensambla pantallas desde instancias de componentes (ref)
4. get_screenshot()    →  Verifica visualmente después de cada sección
```

Nunca saltes al paso 3 sin completar el 1 y el 2.

## Flujo de Iteración (Solicitudes de Cambio)

Al modificar un diseño existente (NO creando desde cero), sigue este flujo en lugar del Orden de Operaciones anterior. **Nunca elimines y recrees lo que ya existe.**

### Paso 0 — Entiende qué existe

1. `get_editor_state()` → identifica el archivo `.pen` abierto y el estado actual
2. `batch_get({ patterns: ["*"] })` → obtén el árbol completo de nodos (o patrones específicos para archivos grandes)
3. `get_variables()` → entiende los tokens de diseño actuales
4. Identifica los nodos específicos afectados por la solicitud de cambio

### Paso 1 — Clasifica el cambio

| Tipo de cambio | Acción | Ejemplo |
|---|---|---|
| **Cambio de token** (color, fuente, espaciado) | `set_variables()` con solo los valores actualizados | "Hacer el color primario más oscuro" → actualiza la variable `color-primary` |
| **Cambio de contenido** (texto, imágenes, iconos) | `U(nodeId, { content: "new text" })` en cada instancia | "Cambiar heading a X" → actualiza nodos de texto |
| **Cambio de estructura de componente** | Modifica la **madre del componente** — todas las instancias se actualizan | "Agregar un icono al card component" → edita el componente reutilizable |
| **Personalización de instancia** | `U(instanceId+"/childId", {...})` o `descendants` | "Esta card específica necesita texto diferente" |
| **Cambio de layout** (reordenar, redimensionar, agregar/quitar secciones) | `U()` para reposicionar, `I()` solo para elementos genuinamente nuevos, `D()` solo para elementos explícitamente eliminados | "Mover sidebar a la derecha" → actualiza props x/y/layout |
| **Nueva pantalla/sección** | Solo ESTO usa el flujo de creación (Pasos 1-4 arriba) | "Agregar una página de settings" → nueva pantalla, reutilizando componentes existentes |

### Paso 2 — Ejecuta quirúrgicamente

1. **Cambia solo lo que cambió** — si el usuario dice "hacer el header más grande", actualiza la variable de font-size del header o el nodo. NO reconstruyas la pantalla
2. **Prefiere cambios de variables** — si una propiedad visual viene de una `$variable`, actualiza la variable via `set_variables()`. Todos los nodos que la usan se actualizan automáticamente
3. **Prefiere ediciones de la madre del componente** — si el cambio aplica a todas las instancias de un componente, edita la madre. NO actualices cada instancia por separado
4. **Usa `U()` no `R()`** — `Update` preserva el nodo y cambia propiedades. `Replace` crea un nuevo nodo. Solo usa `R()` cuando el tipo del nodo mismo deba cambiar (ej., intercambiando texto por icono)
5. **Nunca `D()` + `I()` lo que puedes `U()`** — eliminar y reinsertar es reconstruir, no iterar

### Paso 3 — Verifica

1. `get_screenshot()` de la sección/pantalla afectada
2. Si el cambio tocó una madre de componente, también captura pantalla de las pantallas que usan instancias de ese componente
3. Si el cambio tocó una variable, verifica puntualmente pantallas en ambos modos (light/dark)

### Principio clave

**El cambio más rápido y seguro toca el menor número de nodos.** Un cambio de variable toca cero nodos (se actualizan automáticamente). Un cambio de madre de componente toca un nodo. Las actualizaciones a nivel de instancia tocan N nodos. Reconstruir toca todo. Siempre elige el enfoque de mayor apalancamiento.

## Paso 1: Crear Variables

Usa `set_variables` para definir todo de una vez. Usa nombres con prefijos para simular collections.

```json
{
  "filePath": "path/to/file.pen",
  "variables": {
    "color-primary":        { "type": "color",  "value": "#1E40AF" },
    "color-primary-hover":  { "type": "color",  "value": "#1E3A8A" },
    "color-secondary":      { "type": "color",  "value": "#6366F1" },
    "color-success":        { "type": "color",  "value": "#16A34A" },
    "color-warning":        { "type": "color",  "value": "#D97706" },
    "color-danger":         { "type": "color",  "value": "#DC2626" },
    "color-info":           { "type": "color",  "value": "#2563EB" },
    "color-background":     { "type": "color",  "value": "#FFFFFF" },
    "color-surface":        { "type": "color",  "value": "#FFFFFF" },
    "color-surface-subtle": { "type": "color",  "value": "#F8FAFC" },
    "color-text-primary":   { "type": "color",  "value": "#0F172A" },
    "color-text-secondary": { "type": "color",  "value": "#64748B" },
    "color-text-muted":     { "type": "color",  "value": "#94A3B8" },
    "color-text-inverse":   { "type": "color",  "value": "#FFFFFF" },
    "color-border-default": { "type": "color",  "value": "#E2E8F0" },
    "color-border-strong":  { "type": "color",  "value": "#94A3B8" },
    "color-border-focus":   { "type": "color",  "value": "#1E40AF" },

    "font-family-heading":  { "type": "string", "value": "Space Grotesk" },
    "font-family-body":     { "type": "string", "value": "Inter" },
    "font-family-mono":     { "type": "string", "value": "JetBrains Mono" },
    "fw-normal":   { "type": "string", "value": "400" },
    "fw-medium":   { "type": "string", "value": "500" },
    "fw-semibold": { "type": "string", "value": "600" },
    "fw-bold":     { "type": "string", "value": "700" },

    "font-size-xs":   { "type": "number", "value": 12 },
    "font-size-sm":   { "type": "number", "value": 14 },
    "font-size-base": { "type": "number", "value": 16 },
    "font-size-lg":   { "type": "number", "value": 18 },
    "font-size-xl":   { "type": "number", "value": 20 },
    "font-size-2xl":  { "type": "number", "value": 24 },
    "font-size-3xl":  { "type": "number", "value": 30 },
    "font-size-4xl":  { "type": "number", "value": 36 },

    "line-height-tight":   { "type": "number", "value": 1.25 },
    "line-height-normal":  { "type": "number", "value": 1.5 },
    "line-height-relaxed": { "type": "number", "value": 1.625 },

    "spacing-1":  { "type": "number", "value": 4 },
    "spacing-2":  { "type": "number", "value": 8 },
    "spacing-3":  { "type": "number", "value": 12 },
    "spacing-4":  { "type": "number", "value": 16 },
    "spacing-5":  { "type": "number", "value": 20 },
    "spacing-6":  { "type": "number", "value": 24 },
    "spacing-8":  { "type": "number", "value": 32 },
    "spacing-10": { "type": "number", "value": 40 },
    "spacing-12": { "type": "number", "value": 48 },
    "spacing-16": { "type": "number", "value": 64 },

    "radius-none": { "type": "number", "value": 0 },
    "radius-sm":   { "type": "number", "value": 4 },
    "radius-md":   { "type": "number", "value": 8 },
    "radius-lg":   { "type": "number", "value": 12 },
    "radius-xl":   { "type": "number", "value": 16 },
    "radius-full": { "type": "number", "value": 9999 }
  }
}
```

**Crítico:**
- Las variables `font-family-*` DEBEN ser de tipo string. `fontFamily:"$font-body"`, nunca `fontFamily:"Inter"`
- Las variables `font-weight-*` DEBEN ser de tipo **string** (no number). El `fontWeight` de Pencil espera un string. Usa `{"type":"string","value":"600"}`, no `{"type":"number","value":600}`
- Los tipos de variables son **inmutables** una vez creados. Si creaste una variable como number y necesitas string, crea una nueva variable con un nombre diferente. `replace:true` en `set_variables` NO cambia los tipos de variables existentes

### Simulando Collections via Nomenclatura

Dado que Pencil no tiene collections nativas, usa prefijos para agrupar variables:

```
Primitivos:  color-brand-500, color-gray-200, spacing-4, radius-md, font-size-base
Semántico:   color-primary, color-text-primary, color-surface, space-component-gap
Componente:  (opcional) button-bg, card-padding, input-border
```

Las variables semánticas referencian los mismos valores que las primitivas pero con nombres basados en propósito. En Pencil no hay aliasing — ambas son variables independientes con el mismo valor hex/number. Actualiza primitivos Y semánticos al cambiar valores.

### Variables Temáticas (Modos via Eje de Tema)

Pencil soporta temas via su sistema de eje de tema. Usa esto para simular los modos de Figma:

```json
{
  "variables": {
    "color-background": {
      "type": "color",
      "value": [
        { "value": "#FFFFFF", "theme": { "mode": "light" } },
        { "value": "#0F172A", "theme": { "mode": "dark" } }
      ]
    },
    "color-surface": {
      "type": "color",
      "value": [
        { "value": "#FFFFFF", "theme": { "mode": "light" } },
        { "value": "#1E293B", "theme": { "mode": "dark" } }
      ]
    },
    "color-text-primary": {
      "type": "color",
      "value": [
        { "value": "#0F172A", "theme": { "mode": "light" } },
        { "value": "#F8FAFC", "theme": { "mode": "dark" } }
      ]
    }
  }
}
```

Esto registra un eje de tema `mode` con valores `light` y `dark` automáticamente. Aplica tema a frames via la propiedad `theme`: `{theme: {"mode": "dark"}}`.

## Paso 2: Construir Componentes Reutilizables

Crea un frame de librería de componentes, luego construye cada componente dentro de él.

### Estructura de la Librería de Componentes

La librería se organiza en **secciones verticales etiquetadas**, no en una fila plana. Cada sección tiene una etiqueta de título y sus componentes debajo.

```javascript
// Main library container — vertical, to the RIGHT of screens
lib=I(document,{type:"frame",name:"Component Library",layout:"vertical",width:1400,height:"fit_content(800)",x:3200,y:0,gap:"$sp-10",padding:"$sp-8",fill:"$color-bg"})

// --- Section: Typography ---
typoLabel=I(lib,{type:"text",content:"Typography",fontFamily:"$font-sans",fontSize:"$fs-lg",fontWeight:"$fw-semibold",fill:"$color-text-primary"})
typoRow=I(lib,{type:"frame",name:"— Typography",layout:"horizontal",width:"fill_container",gap:"$sp-8",alignItems:"end"})
// ... text components go inside typoRow

// --- Section: Colors ---
colorLabel=I(lib,{type:"text",content:"Colors",fontFamily:"$font-sans",fontSize:"$fs-lg",fontWeight:"$fw-semibold",fill:"$color-text-primary"})
colorRow=I(lib,{type:"frame",name:"— Colors",layout:"horizontal",width:"fill_container",gap:"$sp-4"})
// ... color swatches go inside colorRow

// --- Section: Icons ---
iconLabel=I(lib,{type:"text",content:"Icons",fontFamily:"$font-sans",fontSize:"$fs-lg",fontWeight:"$fw-semibold",fill:"$color-text-primary"})
iconRow=I(lib,{type:"frame",name:"— Icons",layout:"horizontal",width:"fill_container",gap:"$sp-6"})
// ... icon samples go inside iconRow

// --- Section: Primitives (buttons, badges, links) ---
primLabel=I(lib,{type:"text",content:"Primitives",fontFamily:"$font-sans",fontSize:"$fs-lg",fontWeight:"$fw-semibold",fill:"$color-text-primary"})
primRow=I(lib,{type:"frame",name:"— Primitives",layout:"horizontal",width:"fill_container",gap:"$sp-8",alignItems:"start"})

// --- Section: Cards ---
cardLabel=I(lib,{type:"text",content:"Cards",fontFamily:"$font-sans",fontSize:"$fs-lg",fontWeight:"$fw-semibold",fill:"$color-text-primary"})
cardRow=I(lib,{type:"frame",name:"— Cards",layout:"horizontal",width:"fill_container",gap:"$sp-8",alignItems:"start"})

// --- Section: Navigation ---
navLabel=I(lib,{type:"text",content:"Navigation",fontFamily:"$font-sans",fontSize:"$fs-lg",fontWeight:"$fw-semibold",fill:"$color-text-primary"})
navRow=I(lib,{type:"frame",name:"— Navigation",layout:"vertical",width:"fill_container",gap:"$sp-4"})
```

### Swatches de Color

Crea un swatch para cada color semántico para que el diseñador y el desarrollador puedan ver la paleta:

```javascript
// One swatch = colored circle + name label
swatch=I(colorRow,{type:"frame",layout:"vertical",gap:"$sp-2",alignItems:"center"})
swatchCircle=I(swatch,{type:"ellipse",width:40,height:40,fill:"$color-primary"})
swatchName=I(swatch,{type:"text",content:"primary",fontFamily:"$font-mono",fontSize:"$fs-xs",fill:"$color-text-muted"})
```

### Muestras de Iconos

Muestra cada icono usado en el proyecto. Esta es la referencia del desarrollador para saber qué iconos importar.

Pencil usa icon fonts (Lucide, Material Symbols, Phosphor, Feather). Para implementación web, estos se convierten en paquetes SVG:

| Icon font de Pencil | Paquete web | Instalación |
|---|---|---|
| `lucide` | `lucide-react` o `lucide-vue` | `npm i lucide-react` |
| `feather` | `react-feather` | `npm i react-feather` |
| `Material Symbols Outlined` | `@mui/icons-material` | `npm i @mui/icons-material` |
| `phosphor` | `@phosphor-icons/react` | `npm i @phosphor-icons/react` |

```javascript
// Icon sample = icon + name label below
iconSample=I(iconRow,{type:"frame",layout:"vertical",gap:"$sp-2",alignItems:"center"})
I(iconSample,{type:"icon_font",iconFontName:"mail",iconFontFamily:"lucide",width:24,height:24,fill:"$color-text-primary"})
I(iconSample,{type:"text",content:"mail",fontFamily:"$font-mono",fontSize:"$fs-xs",fill:"$color-text-muted"})
```

Crea muestras en los tamaños estándar usados en el proyecto (típicamente 16px para inline, 20px para botones, 24px para standalone).

### Componentes de Estilo de Texto

```javascript
// Text/Heading component
heading=I(lib,{type:"text",name:"Text/Heading",reusable:true,content:"Heading Text",fontFamily:"$font-family-heading",fontSize:"$font-size-2xl",fontWeight:"$font-weight-semibold",fill:"$color-text-primary",lineHeight:"$line-height-tight"})

// Text/Body component
body=I(lib,{type:"text",name:"Text/Body",reusable:true,content:"Body text content",fontFamily:"$font-family-body",fontSize:"$font-size-base",fontWeight:"$font-weight-normal",fill:"$color-text-primary",lineHeight:"$line-height-normal"})

// Text/Caption component
caption=I(lib,{type:"text",name:"Text/Caption",reusable:true,content:"Caption text",fontFamily:"$font-family-body",fontSize:"$font-size-xs",fontWeight:"$font-weight-normal",fill:"$color-text-muted",lineHeight:"$line-height-normal"})

// Text/Label component
label=I(lib,{type:"text",name:"Text/Label",reusable:true,content:"Label",fontFamily:"$font-family-body",fontSize:"$font-size-sm",fontWeight:"$font-weight-medium",fill:"$color-text-primary"})
```

### Componentes de Botón

```javascript
// Button/Primary
btnPrimary=I(lib,{type:"frame",name:"Button/Primary",reusable:true,layout:"horizontal",width:"fit_content",height:"fit_content",padding:["$spacing-3","$spacing-5"],fill:"$color-primary",cornerRadius:"$radius-md",justifyContent:"center",alignItems:"center",gap:"$spacing-2"})
btnPrimaryText=I(btnPrimary,{type:"text",content:"Button",fontFamily:"$font-family-body",fontSize:"$font-size-sm",fontWeight:"$font-weight-semibold",fill:"$color-text-inverse"})

// Button/Secondary
btnSecondary=I(lib,{type:"frame",name:"Button/Secondary",reusable:true,layout:"horizontal",width:"fit_content",height:"fit_content",padding:["$spacing-3","$spacing-5"],fill:"$color-surface",stroke:{thickness:1,fill:"$color-border-default"},cornerRadius:"$radius-md",justifyContent:"center",alignItems:"center",gap:"$spacing-2"})
btnSecText=I(btnSecondary,{type:"text",content:"Button",fontFamily:"$font-family-body",fontSize:"$font-size-sm",fontWeight:"$font-weight-medium",fill:"$color-text-primary"})
```

### Componente de Input

```javascript
// Input/Field (label + input frame + placeholder)
inputField=I(lib,{type:"frame",name:"Input/Field",reusable:true,layout:"vertical",width:320,gap:"$spacing-2"})
inputLabel=I(inputField,{type:"text",content:"Label",fontFamily:"$font-family-body",fontSize:"$font-size-sm",fontWeight:"$font-weight-medium",fill:"$color-text-primary"})
inputBox=I(inputField,{type:"frame",layout:"horizontal",width:"fill_container",height:44,padding:["$spacing-3","$spacing-4"],fill:"$color-surface",stroke:{thickness:1,fill:"$color-border-default"},cornerRadius:"$radius-md",alignItems:"center"})
inputPlaceholder=I(inputBox,{type:"text",content:"Placeholder",fontFamily:"$font-family-body",fontSize:"$font-size-base",fontWeight:"$font-weight-normal",fill:"$color-text-muted"})
```

### Componente de Section Header

```javascript
// Section/Header (title + accent line)
sectionHeader=I(lib,{type:"frame",name:"Section/Header",reusable:true,layout:"vertical",width:"fit_content",gap:"$spacing-3"})
sectionTitle=I(sectionHeader,{type:"text",content:"SECTION TITLE",fontFamily:"$font-family-heading",fontSize:"$font-size-xs",fontWeight:"$font-weight-semibold",fill:"$color-text-primary",letterSpacing:2})
sectionLine=I(sectionHeader,{type:"rectangle",width:48,height:3,fill:"$color-primary"})
```

### Componente de Card

```javascript
// Card
card=I(lib,{type:"frame",name:"Card",reusable:true,layout:"vertical",width:400,padding:"$spacing-6",fill:"$color-surface",stroke:{thickness:1,fill:"$color-border-default"},cornerRadius:"$radius-lg",gap:"$spacing-4"})
```

### Componente de Divider

```javascript
// Divider
divider=I(lib,{type:"rectangle",name:"Divider",reusable:true,width:400,height:1,fill:"$color-border-default"})
```

## Paso 3: Ensamblar Pantallas desde Componentes

Usa `ref` para instanciar componentes. Sobreescribe propiedades via la raíz o `descendants`.

### Ejemplo: Usando una instancia de componente

Hay dos formas correctas de personalizar instancias. Ambas son seguras:

**Opción A — `descendants` al insertar** (preferido para nuevas instancias):
```javascript
header1=I(mainContent,{type:"ref",ref:"sectionHeaderId",descendants:{"titleTextId":{content:"EXPERIENCE"}}})
emailInput=I(formFrame,{type:"ref",ref:"inputFieldId",descendants:{"labelId":{content:"Email"},"placeholderId":{content:"you@company.com"}}})
submitBtn=I(formFrame,{type:"ref",ref:"btnPrimaryId",width:"fill_container",descendants:{"btnTextId":{content:"Submit"}}})
```

**Opción B — `U(instanceId+"/childId")` después de la inserción** (preferido para actualizaciones de instancias existentes):
```javascript
// SAFE — modifies only this instance, not the mother
U("YkHfO/MNS4B",{content:"New text"})    // updates text in instance YkHfO only
U("YkHfO/MNS4B",{fill:"$color-danger"})  // changes color in instance YkHfO only
```

```javascript
// WRONG — bare childId WITHOUT instance prefix modifies the MOTHER component!
U("MNS4B",{content:"EXPERIENCE"})  // ← CORRUPTS ALL INSTANCES
U(header1+"/titleTextId",{content:"EXPERIENCE"})  // ← Also wrong if header1 resolves to a binding, not a stable ID. Use the actual instance ID from batch_get
```

**La regla:** siempre incluye el ID de instancia como prefijo. `U("instanceId/childId")` = override seguro de instancia. `U("childId")` solo = modifica la madre.

### Reglas Clave para Instancias

- Personaliza contenido al crear: usa `descendants` en la llamada de inserción `ref`
- Personaliza contenido después: usa `U(instanceId+"/childId", {props})` — esto es SEGURO y es el mecanismo principal para iterar diseños existentes
- Redimensionar: sobreescribe `width` o `height` directamente en el nodo `ref`
- Ocultar un hijo: `descendants:{"childId":{enabled:false}}`
- Reemplazar un hijo: `R(instanceId+"/childId", {type:"text",...})` (solo para reemplazo estructural)
- `U("childId")` SIN prefijo de instancia modifica la madre del componente — NUNCA hagas esto para personalizar una instancia
- NUNCA recrees un componente manualmente — siempre usa `ref`

### Gotchas al Reemplazar Instancias

- Cuando haces `R(instance+"/childId")`, el reemplazo crea un NUEVO nodo con un nuevo ID. El ID anterior ya no existe
- Si necesitas modificar un nodo previamente reemplazado, usa el ID del NUEVO nodo, no el original: `R("YkHfO/newNodeId")` no `R("YkHfO/originalId")`
- Si `R()` falla con "No such node", el nodo ya fue reemplazado o eliminado. Usa `batch_get` para encontrar el ID actual
- Patrón alternativo cuando falla `R()`: `D(nodeId)` + `I(parentId, {...})` + `M(newNode, parentId, position)`

### Ubicación de la Librería de Componentes

- Posiciona el frame de la librería a la **DERECHA** de todas las pantallas (ej., `x:3200`)
- Nunca lo pongas debajo de las pantallas donde queda oculto
- Después de ensamblar pantallas, **verifica que los componentes estén intactos**: `get_screenshot(libraryFrameId)`

## Paso 4: Verificar

Después de cada sección principal:

```
get_screenshot(nodeId) → inspeccionar visualmente
```

Verifica:
- Visibilidad del texto (todo el texto tiene `fill` establecido via variable)
- Alineación (layout flexbox, sin x/y hardcodeados en hijos flex)
- Consistencia del espaciado (gaps y padding usan variables `$spacing-*`)
- Apropiabilidad del color (coincide con la propuesta aprobada)
- Reutilización de componentes (sin estructuras de nodos duplicadas)

## Texto Dentro de Contenedores (CRÍTICO — Bug recurrente #1)

Cada nodo de texto dentro de un frame con ancho limitado (cards, callouts, bento tiles):

```javascript
// CORRECT — text wraps inside card
I(card,{type:"text",content:"Long text...",textGrowth:"fixed-width",width:"fill_container",fontFamily:"$font-sans",fontSize:"$fs-base",fill:"$color-text-secondary",lineHeight:"$lh-relaxed"})

// WRONG — text overflows/truncates
I(card,{type:"text",content:"Long text...",fontFamily:"$font-sans",fontSize:"$fs-base",fill:"$color-text-secondary"})
```

- **DEBE** establecer `textGrowth: "fixed-width"` + `width: "fill_container"` en cualquier texto dentro de un contenedor con ancho limitado
- **Excepción**: labels cortas, badges, botones (máximo 1 línea) — mantén el `textGrowth: "auto"` por defecto
- **Verifica inmediatamente** con `get_screenshot` después de insertar texto

## Armonía de Grid

Al crear cards en un layout horizontal (bento grids, filas de cards):

- Las cards con `height: "fit_content"` tendrán **alturas desiguales** si el contenido varía
- **PREFIERE altura fija igual** para cards hermanas (ej., todas 220px)
- Captura pantalla inmediatamente después de crear cualquier grid de cards para verificar la alineación

## Limitaciones Adicionales de Pencil

- **Las variables string en `content` se resuelven SOLO dentro del contexto de tema** — `content: "$txt-title"` funciona en descendientes de instancias (hereda el tema del padre) pero NO en nodos creados via `R()` (Replace). Los nodos reemplazados pierden el contexto de tema. Para i18n confiable, usa `U(instance+"/childId",{content:"$txt-var"})` en descendientes existentes, o hardcodea texto en nodos reemplazados
- **Los descendientes de nodos copiados obtienen nuevos IDs** — después de `C()`, los IDs de hijos se regeneran. Nunca `U()` con IDs originales de hijos en una copia. O usa `descendants` en la operación Copy misma, o `batch_get` la copia para leer los nuevos IDs primero
- **Los cambios en componentes se cascadean a instancias — a menos que se sobreescriban** — si una instancia reemplazó un nodo (ej., texto → frame con bullets), eliminar y recrear ese nodo en el componente SÍ se cascadeará a instancias que no lo habían sobreescrito. Las instancias con overrides mantienen sus overrides (ahora huérfanos). Planifica la reestructuración de componentes cuidadosamente

## Estados de Componentes Interactivos (OBLIGATORIO para diseño mobile/app)

Pencil NO tiene prototipado. Para compensar, cada componente con estados interactivos DEBE tener un frame de **States** dedicado mostrando todos los estados uno al lado del otro.

### Cuándo crear frames de estados

- **Navigation**: cerrado + abierto (menú hamburguesa)
- **Buttons**: default + hover + disabled + loading
- **Inputs**: vacío + focused + filled + error
- **Cards**: colapsado + expandido (si es expandible)
- **Modals/Sheets**: el overlay + el contenido
- **Toggles**: on + off
- **Dropdowns**: cerrado + abierto con opciones

### Cómo estructurar

Crea un frame de nivel superior llamado `{Component} — States` con todos los estados etiquetados:

```javascript
// Example: Navbar mobile states
states=I(document,{type:"frame",name:"Navbar/Mobile — States",layout:"vertical",width:390,gap:"$sp-8",padding:"$sp-6",fill:"$color-bg",theme:{mode:"dark"}})
closedLabel=I(states,{type:"text",content:"State: Closed",fill:"$color-text-muted",fontFamily:"$font-mono",fontSize:"$fs-xs",letterSpacing:2})
closedNav=I(states,{type:"ref",ref:"navMobileId",width:"fill_container"})
openLabel=I(states,{type:"text",content:"State: Open",fill:"$color-text-muted",fontFamily:"$font-mono",fontSize:"$fs-xs",letterSpacing:2})
// ... build the open state
```

### Reglas

- **Crea estados automáticamente** al diseñar el componente — no esperes a que el usuario lo pida
- **Etiqueta cada estado** claramente con un texto `State: {nombre}` encima
- **Coloca cerca del componente** en el canvas, no en el frame de la Librería
- **Usa el mismo tema** que la pantalla objetivo (dark/light)
- **Ambos modos si aplica** — si el componente aparece en pantallas dark + light, muestra estados para ambos

## Organización del Canvas (OBLIGATORIO)

Mantén el canvas organizado cronológicamente y por tipo. Siempre deja ~200px de separación entre filas y entre frames horizontalmente para que el usuario pueda iterar sin que los frames colisionen.

### Orden de layout (de arriba a abajo)

```
FILA 1 — Librería + Estados de Componentes + Capas Sueltas
  Frame de librería (design tokens + componentes)
  Frames de estados de componentes (estados de Navbar, cards expandidas, modales, etc.)
  Siempre en la PARTE SUPERIOR — esta es la capa de referencia

FILA 2 — v1 (primera iteración de pantallas)
  Iteración de diseño más antigua, conservada para historia

FILA 3 — v2 (iteración actual de pantallas)
  Últimas pantallas web: dark, light, variantes de idioma, blog, etc.

FILA 4 — Pantallas Mobile
  Mobile dark, mobile light, menú mobile abierto, etc.

(Agrega más filas abajo a medida que se añadan nuevas iteraciones o plataformas)
```

### Reglas

- **~200px de separación** entre cada fila y entre frames horizontales — nunca comprimas frames
- **Cronológico de arriba a abajo** — más antiguo arriba, más nuevo abajo
- **Librería siempre FILA 1** — es la referencia, no una pantalla
- **States/capas junto a la Librería** en la misma fila, no dispersas por el canvas
- **Después de reorganizar**, ejecuta `snapshot_layout(problemsOnly: true)` para verificar que no haya superposiciones
- **Nunca elimines frames para reorganizar** — solo mueve con `U(id, {x, y})`

## Sincronización de Contenido Entre Pantallas (OBLIGATORIO)

Al actualizar contenido en una pantalla (texto, badges, labels, datos), verifica TODAS las pantallas que comparten los mismos datos:

1. `batch_get` con un patrón que coincida con el valor antiguo en todo el documento
2. Actualiza cada instancia — dark, light, EN, ES, mobile, desktop
3. Si el contenido viene de una variable (`$txt-*`), actualiza la variable en su lugar — todas las pantallas se actualizan automáticamente
4. Si es texto hardcodeado, busca y reemplaza en cada pantalla manualmente

**Por qué:** K8s fue cambiado a Kafka en mobile dark pero no en mobile light o pantallas web, creando inconsistencia. El contenido debe ser consistente en todas las pantallas.

## Registro de Iconos (OBLIGATORIO)

Cada icono usado en cualquier pantalla o componente DEBE aparecer en la sección Icons de la Librería. Al agregar un nuevo icono a un diseño:

1. Usa el icono en el componente/pantalla
2. **Inmediatamente** agrégalo a la sección Icons en el frame de la Librería
3. Cada muestra de icono: icono a 24px + etiqueta de nombre abajo, dentro de un frame vertical con `alignItems: center`

Nunca dejes un icono sin documentar. La Librería es la referencia del desarrollador para saber qué iconos instalar.

## Errores Comunes a Evitar

| Error | Enfoque correcto |
|---|---|
| `fontFamily:"Inter"` | `fontFamily:"$font-family-body"` |
| `fontWeight:"600"` | `fontWeight:"$font-weight-semibold"` |
| `fill:"#1E40AF"` | `fill:"$color-primary"` |
| `fontSize:14` | `fontSize:"$font-size-sm"` |
| `cornerRadius:8` | `cornerRadius:"$radius-md"` |
| `gap:16` | `gap:"$spacing-4"` |
| `padding:24` | `padding:"$spacing-6"` |
| Construir la misma card 4 veces | Crea componente card una vez, usa 4 instancias `ref` |
| Diseñar sin un plan | Presenta propuesta de color/tipografía/tono primero |
| Texto en card sin `textGrowth` | Agrega `textGrowth:"fixed-width"` + `width:"fill_container"` |
| Cards hermanas con alturas desiguales | Establece altura fija igual para todas las cards en una fila |
| `content:"$txt-var"` en nodos reemplazados | Usa `U(instance+"/descendantId")` para variables, o hardcodea en nodos reemplazados |
| `U(copy+"/originalChildId")` | Lee primero los nuevos IDs de la copia, o usa descendants en `C()` |
| `R(instance+"/oldReplacedId")` | Usa el ID actual del nodo desde `batch_get`, no el ID original del componente |
| `U("childId")` para personalizar instancia | Usa `U(instance+"/childId")` — sin prefijo modificas la madre |
| Agregar toda la info a la vez (tags, links, metadata) | Empieza mínimo, verifica, luego agrega capas. Info secundaria con baja opacidad (0.4-0.6) |
| Inventar contenido para diseños | Usa datos reales del CV, LinkedIn o docs proporcionados por el usuario |
| Eliminar + reinsertar para actualizar | Usa `U()` para cambios de propiedades — solo `D()+I()` cuando el tipo de nodo deba cambiar |

## Slots (Flexibilidad de Componentes)

Los slots son áreas designadas dentro de un componente donde se pueden colocar elementos. Crean regiones flexibles y personalizables.

### Crear un slot

1. Crea un frame vacío dentro de un componente reutilizable
2. Márcalo como slot: `"slot": [suggestedComponentIds]`

```javascript
// Table component with a slot for rows
table=I(lib,{type:"frame",name:"Table",reusable:true,layout:"vertical",width:"fill_container"})
tableHeader=I(table,{type:"frame",layout:"horizontal",width:"fill_container",padding:[12,16],fill:"$color-surface-subtle"})
// ... header cells
tableBody=I(table,{type:"frame",layout:"vertical",width:"fill_container",slot:["tableRowComponentId"]})
```

### Usando slots

Al instanciar un componente con slots, inserta hijos directamente en el frame del slot:

```javascript
myTable=I(screen,{type:"ref",ref:"tableId",width:"fill_container"})
// Insert rows into the slot
row1=I(myTable+"/tableBodyId",{type:"ref",ref:"tableRowComponentId"})
row2=I(myTable+"/tableBodyId",{type:"ref",ref:"tableRowComponentId"})
```

**Componentes sugeridos** — marca qué componentes son recomendados para un slot. Esto ayuda tanto a diseñadores humanos como al agente de IA a saber qué insertar.

## Librerías de Diseño (.lib.pen)

Para proyectos con múltiples archivos `.pen`, extrae los componentes compartidos a un archivo de librería:

1. Crea un archivo `.pen` con componentes compartidos
2. Conviértelo en librería (se convierte en `.lib.pen` — **irreversible**)
3. Importa la librería en otros archivos `.pen` via `imports`

Los cambios en los componentes de la librería se propagan a todos los archivos que la importan. Úsalo para design systems entre archivos.

**Cuándo usar:** múltiples archivos `.pen` compartiendo los mismos componentes. Para proyectos de un solo archivo, un frame de Component Library dentro del documento es suficiente.

## Nodos de Script (Código en el Canvas)

Los nodos de script renderizan output de JavaScript directamente en el canvas. Útil para layouts basados en datos o repetitivos.

### Cuándo usar

- Repetir un patrón N veces (grid de cards, filas de datos)
- Gráficos o visualización de datos
- Layouts parametrizados que necesitan ajuste interactivo

### Cómo funcionan

```javascript
// chart.js — referenced by a script node
/**
 * @schema 2.11
 * @input rows: number(min=1, max=20) = 5
 * @input fill: color = #10B981
 */
return Array.from({length: pencil.input.rows}, (_, i) => ({
  type: "rectangle",
  width: pencil.width,
  height: 20,
  fill: pencil.input.fill,
  y: i * 24
}))
```

**Restricciones:** máximo 1000 nodos, timeout de 2s, sin async, sin acceso a DOM/red, `Math.random()` determinista.

**Convertir a capas:** una vez que el output se vea bien, convierte a capas estáticas editables para mayor personalización.

### Cuándo NO usar

- Layouts simples que batch_design maneja en <10 ops
- Pantallas únicas sin repetición
- Cuando el diseñador necesita personalizar cada elemento individualmente (usa instancias de componentes en su lugar)
