# Figma — Flujo de Trabajo del Design System

Cómo crear un design system completo en Figma usando herramientas MCP (`use_figma` via skill figma-use).

## Orden de Operaciones

```
1. Crear Variable Collections    →  Primitivos, Semántico (con modos), Componente
2. Construir Componentes         →  Con propiedades, variantes y bindings de variables
3. Ensamblar Pantallas           →  Desde instancias de componentes, cambiando modos por frame
4. Marcar Listo para Dev         →  Dev Mode recoge variables como code tokens
```

## Prerequisitos

- Carga el skill `/figma:figma-use` ANTES de cada llamada a `use_figma`
- Carga `/figma:figma-generate-library` para el flujo completo de creación de librería
- Para escribir en el canvas de Figma, siempre usa la herramienta `use_figma` (nunca ediciones directas de archivo)

## Paso 1: Crear Variable Collections

Figma organiza las variables en **Collections**. Cada collection puede tener **Modes**.

### Collection: Primitivos (sin modos)

Valores crudos. Nunca se usan directamente en componentes o pantallas.

```javascript
// Create "Primitives" collection with color scales
// Colors
createVariable("Primitives", "color/brand/50",  "COLOR", "#EFF6FF")
createVariable("Primitives", "color/brand/100", "COLOR", "#DBEAFE")
createVariable("Primitives", "color/brand/500", "COLOR", "#3B82F6")
createVariable("Primitives", "color/brand/600", "COLOR", "#2563EB")
createVariable("Primitives", "color/brand/700", "COLOR", "#1D4ED8")
createVariable("Primitives", "color/brand/900", "COLOR", "#1E3A8A")

// Neutrals
createVariable("Primitives", "color/gray/50",  "COLOR", "#F8FAFC")
createVariable("Primitives", "color/gray/200", "COLOR", "#E2E8F0")
createVariable("Primitives", "color/gray/400", "COLOR", "#94A3B8")
createVariable("Primitives", "color/gray/600", "COLOR", "#475569")
createVariable("Primitives", "color/gray/900", "COLOR", "#0F172A")

// Spacing
createVariable("Primitives", "spacing/1",  "FLOAT", 4)
createVariable("Primitives", "spacing/4",  "FLOAT", 16)
createVariable("Primitives", "spacing/6",  "FLOAT", 24)
createVariable("Primitives", "spacing/8",  "FLOAT", 32)

// Radius
createVariable("Primitives", "radius/sm",   "FLOAT", 4)
createVariable("Primitives", "radius/md",   "FLOAT", 8)
createVariable("Primitives", "radius/lg",   "FLOAT", 12)
createVariable("Primitives", "radius/full", "FLOAT", 9999)

// Typography
createVariable("Primitives", "font/size/xs",   "FLOAT", 12)
createVariable("Primitives", "font/size/sm",   "FLOAT", 14)
createVariable("Primitives", "font/size/base", "FLOAT", 16)
createVariable("Primitives", "font/size/lg",   "FLOAT", 18)
createVariable("Primitives", "font/size/xl",   "FLOAT", 20)
createVariable("Primitives", "font/size/2xl",  "FLOAT", 24)
createVariable("Primitives", "font/size/3xl",  "FLOAT", 30)
createVariable("Primitives", "font/size/4xl",  "FLOAT", 36)

createVariable("Primitives", "font/weight/normal",   "FLOAT", 400)
createVariable("Primitives", "font/weight/medium",   "FLOAT", 500)
createVariable("Primitives", "font/weight/semibold", "FLOAT", 600)
createVariable("Primitives", "font/weight/bold",     "FLOAT", 700)

createVariable("Primitives", "font/family/heading", "STRING", "Space Grotesk")
createVariable("Primitives", "font/family/body",    "STRING", "Inter")
createVariable("Primitives", "font/family/mono",    "STRING", "JetBrains Mono")

createVariable("Primitives", "line-height/tight",   "FLOAT", 1.25)
createVariable("Primitives", "line-height/normal",  "FLOAT", 1.5)
createVariable("Primitives", "line-height/relaxed", "FLOAT", 1.625)
```

### Collection: Semántico (con modos: light, dark)

Aliases basados en propósito. Los valores cambian por modo via aliasing a primitivos.

```javascript
// Create "Semantic" collection with 2 modes: "light" and "dark"
// Each variable aliases a different primitive per mode

// Brand
createVariable("Semantic", "color/primary",       "COLOR", {light: alias("Primitives/color/brand/600"), dark: alias("Primitives/color/brand/400")})
createVariable("Semantic", "color/primary-hover",  "COLOR", {light: alias("Primitives/color/brand/700"), dark: alias("Primitives/color/brand/300")})

// Surface
createVariable("Semantic", "color/background",     "COLOR", {light: "#FFFFFF", dark: alias("Primitives/color/gray/900")})
createVariable("Semantic", "color/surface",         "COLOR", {light: "#FFFFFF", dark: "#1E293B"})
createVariable("Semantic", "color/surface-elevated","COLOR", {light: "#FFFFFF", dark: "#334155"})

// Text
createVariable("Semantic", "color/text-primary",   "COLOR", {light: alias("Primitives/color/gray/900"), dark: alias("Primitives/color/gray/50")})
createVariable("Semantic", "color/text-secondary",  "COLOR", {light: alias("Primitives/color/gray/600"), dark: alias("Primitives/color/gray/400")})
createVariable("Semantic", "color/text-muted",      "COLOR", {light: alias("Primitives/color/gray/400"), dark: alias("Primitives/color/gray/600")})
createVariable("Semantic", "color/text-inverse",    "COLOR", {light: "#FFFFFF", dark: alias("Primitives/color/gray/900")})

// Border
createVariable("Semantic", "color/border-default", "COLOR", {light: alias("Primitives/color/gray/200"), dark: "#334155"})
createVariable("Semantic", "color/border-focus",   "COLOR", {light: alias("Primitives/color/brand/500"), dark: alias("Primitives/color/brand/400")})

// Status
createVariable("Semantic", "color/success", "COLOR", {light: "#16A34A", dark: "#4ADE80"})
createVariable("Semantic", "color/warning", "COLOR", {light: "#D97706", dark: "#FBBF24"})
createVariable("Semantic", "color/danger",  "COLOR", {light: "#DC2626", dark: "#F87171"})
createVariable("Semantic", "color/info",    "COLOR", {light: "#2563EB", dark: "#60A5FA"})

// Spacing (aliases — same for both modes)
createVariable("Semantic", "space/component-gap", "FLOAT", alias("Primitives/spacing/4"))
createVariable("Semantic", "space/section-gap",   "FLOAT", alias("Primitives/spacing/8"))
createVariable("Semantic", "space/card-padding",  "FLOAT", alias("Primitives/spacing/6"))
createVariable("Semantic", "space/page-padding",  "FLOAT", alias("Primitives/spacing/6"))
createVariable("Semantic", "space/input-padding",  "FLOAT", alias("Primitives/spacing/3"))
```

### Scoping de Variables

Restringe dónde se pueden aplicar las variables para prevenir uso incorrecto:

```javascript
// Color variables → scoped to fills, strokes
setVariableScoping("Semantic/color/primary", ["FILL_COLOR", "STROKE_COLOR"])

// Spacing variables → scoped to gap, padding, dimensions
setVariableScoping("Semantic/space/component-gap", ["GAP", "PADDING"])

// Radius variables → scoped to corner radius only
setVariableScoping("Primitives/radius/md", ["CORNER_RADIUS"])

// Font size variables → scoped to font size only
setVariableScoping("Primitives/font/size/base", ["FONT_SIZE"])
```

## Paso 2: Construir Componentes

### Componente con Propiedades y Variantes

**Ejemplo de Button** — un componente maneja todos los tipos de botón:

```javascript
// Create component set "Button" with:
// Variant properties: type (primary, secondary, ghost), size (sm, md, lg), state (default, hover, disabled)
// Boolean property: hasIcon (show/hide leading icon)
// Text property: label (editable button text)
// Instance swap property: icon (swappable icon instance)

// All visual values bound to variables:
// Fill: bound to Semantic/color/primary (for primary type)
// Text: bound to Semantic/color/text-inverse
// Corner radius: bound to Primitives/radius/md
// Padding: bound to Primitives/spacing/3 (vertical), Primitives/spacing/5 (horizontal)
// Font: bound to Primitives/font/family/body
// Font size: bound to Primitives/font/size/sm
// Font weight: bound to Primitives/font/weight/semibold
```

### Tabla Rápida de Propiedades de Componentes

| Tipo de propiedad | Usar para | Ejemplo |
|---|---|---|
| **Boolean** | Mostrar/ocultar elementos opcionales | `hasIcon`: toggle de visibilidad del ícono |
| **Text** | Strings de texto editables | `label`: "Submit", "Cancel", "Save" |
| **Instance swap** | Componentes anidados intercambiables | `leadingIcon`: intercambiar entre 20+ componentes de ícono |
| **Variant** | Estados mutuamente excluyentes | `state`: default / hover / disabled |

### Principio clave: Las propiedades reducen la explosión de variantes

```
SIN propiedades: Button/Primary/Large/WithIcon/Hover = 1 variante (de 24+)
CON propiedades: Button → type:primary, size:lg, hasIcon:true, state:hover = mismo componente
```

### Página de la Librería de Componentes

Crea una página dedicada "Component Library" organizada en secciones con etiquetas — no un volcado plano:

```
Component Library
├── Typography        → Heading (h1-h4), Body, Caption, Label con muestras de tamaño
├── Colors            → Swatches de todos los colores semánticos (primary, accent, success, etc.) con nombres
├── Icons             → Todos los iconos del proyecto en tamaños estándar (16, 20, 24px) con nombres
│                       Documenta el set de iconos (Lucide, Heroicons, Phosphor) para que los devs instalen el correcto
├── Primitives        → Botones (primary/secondary/ghost), Badges, Links, Inputs, Dividers
├── Cards             → Job/Card, Project/Card, Stat/Card, etc.
├── Navigation        → Navbar, tabs, breadcrumbs
└── Feedback          → Alert, toast, empty state
```

Cada sección tiene un frame de etiqueta visible. Esta es la referencia del desarrollador — deben poder ver la librería y entender cada elemento visual disponible.

#### Iconos para Web

Figma usa redes vectoriales para iconos. Para handoff web, documenta qué paquete de iconos usar:

| Fuente de icono Figma | Paquete web |
|---|---|
| Plugin Lucide icons | `lucide-react` / `lucide-vue` |
| Plugin Heroicons | `@heroicons/react` |
| Plugin Phosphor | `@phosphor-icons/react` |
| Material Symbols | `@mui/icons-material` |

Nombra cada icono en la librería para que el desarrollador pueda encontrarlo en el paquete (ej., `ArrowRight`, `Mail`, `Github`).

## Paso 3: Ensamblar Pantallas

### Usando Instancias de Componentes

```javascript
// Insert button instance
const submitBtn = createInstance("Button")
submitBtn.setProperties({
  type: "primary",
  size: "md",
  label: "Sign In",
  hasIcon: false
})

// Insert input instance
const emailInput = createInstance("Input")
emailInput.setProperties({
  label: "Email",
  placeholder: "you@company.com",
  state: "default",
  required: true
})
```

### Aplicando Modos a Frames

```javascript
// Set entire screen to dark mode
const loginScreen = figma.currentPage.findOne(n => n.name === "Login Screen")
loginScreen.setExplicitVariableModeForCollection(semanticCollectionId, darkModeId)

// Set just one section to dark mode (e.g., sidebar)
const sidebar = loginScreen.findChild(n => n.name === "Sidebar")
sidebar.setExplicitVariableModeForCollection(semanticCollectionId, darkModeId)
```

Esta es la característica killer: **un diseño, múltiples temas** — sin duplicación.

## Paso 4: Handoff a Dev Mode

Cuando el diseño esté completo:

1. **Marca los frames como "Ready for Dev"**
2. Dev Mode muestra variables como code tokens:
   - `fill: var(--color-primary)` en lugar de `fill: #2563EB`
   - `padding: var(--space-card-padding)` en lugar de `padding: 24px`
   - `font-family: var(--font-family-body)` en lugar de `font-family: Inter`
3. **Code Connect** mapea componentes a snippets reales de código React/Swift/Kotlin
4. Los desarrolladores ven la cadena completa de aliases: `color-primary → brand-600 → #2563EB`

## Patrones Comunes

### Multi-marca con Modos

Collection "Semantic" con modos: `brand-a-light`, `brand-a-dark`, `brand-b-light`, `brand-b-dark`

Cada variable resuelve a primitivos específicos de marca por modo. Un archivo de diseño sirve múltiples marcas.

### Responsivo con Modos

Collection "Sizing" con modos: `mobile`, `tablet`, `desktop`

```
space/page-padding:  mobile=16, tablet=24, desktop=48
font/size/display:   mobile=30, tablet=36, desktop=48
```

Aplica modo al frame → todo el layout se ajusta.

### Modos Anidados

Una página en modo `dark` puede tener un frame hijo en modo `light` (ej., un modal claro sobre una página oscura). Los modos se cascadean hacia abajo pero pueden sobreescribirse en cualquier nivel de frame.

## Errores Comunes

| Error | Enfoque correcto |
|---|---|
| Crear variables sin collections | Agrupar en collections Primitives / Semantic / Component |
| Todas las variables en un modo | Usa modos para light/dark como mínimo |
| Sin aliasing (semántico = hex crudo) | Las variables semánticas hacen alias a variables primitivas |
| Sin scoping | Aplica scope a vars de color para fills, spacing para gaps/padding |
| Componentes separados para cada estado | Usa variant property para el eje de estado |
| Componente separado para icono/sin-icono | Usa boolean property `hasIcon` |
| Texto hardcodeado en componente | Usa text property para strings editables |
| Duplicar pantallas para modo oscuro | Aplica dark mode al frame, misma pantalla |
| Variables no aparecen en Dev Mode | Asegura que las variables estén vinculadas a propiedades del nodo, no solo definidas |
