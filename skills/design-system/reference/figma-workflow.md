# Figma — Design System Workflow

How to create a complete design system in Figma using MCP tools (`use_figma` via figma-use skill).

## Order of Operations

```
1. Create Variable Collections    →  Primitives, Semantic (with modes), Component
2. Build Components               →  With properties, variants, and variable bindings
3. Assemble Screens               →  From component instances, switching modes per frame
4. Mark Ready for Dev             →  Dev Mode picks up variables as code tokens
```

## Prerequisites

- Load `/figma:figma-use` skill BEFORE every `use_figma` call
- Load `/figma:figma-generate-library` for full library creation workflow
- For writing to Figma canvas, always use `use_figma` tool (never direct file edits)

## Step 1: Create Variable Collections

Figma organizes variables into **Collections**. Each collection can have **Modes**.

### Collection: Primitives (no modes)

Raw values. Never used directly in components or screens.

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

### Collection: Semantic (with modes: light, dark)

Purpose-based aliases. Values change per mode via aliasing to primitives.

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

### Variable Scoping

Restrict where variables can be applied to prevent misuse:

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

## Step 2: Build Components

### Component with Properties and Variants

**Button example** — one component handles all button types:

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

### Component Properties Cheat Sheet

| Property type | Use for | Example |
|---|---|---|
| **Boolean** | Show/hide optional elements | `hasIcon`: toggles icon visibility |
| **Text** | Editable text strings | `label`: "Submit", "Cancel", "Save" |
| **Instance swap** | Swappable nested components | `leadingIcon`: swap between 20+ icon components |
| **Variant** | Mutually exclusive states | `state`: default / hover / disabled |

### Key principle: Properties reduce variant explosion

```
WITHOUT properties: Button/Primary/Large/WithIcon/Hover = 1 variant (of 24+)
WITH properties:    Button → type:primary, size:lg, hasIcon:true, state:hover = same component
```

### Component Library Page

Create a dedicated page "Component Library" organized in labeled sections — not a flat dump:

```
Component Library
├── Typography        → Heading (h1-h4), Body, Caption, Label with size samples
├── Colors            → Swatches of all semantic colors (primary, accent, success, etc.) with names
├── Icons             → All project icons at standard sizes (16, 20, 24px) with names
│                       Document the icon set (Lucide, Heroicons, Phosphor) so devs install the right package
├── Primitives        → Buttons (primary/secondary/ghost), Badges, Links, Inputs, Dividers
├── Cards             → Job/Card, Project/Card, Stat/Card, etc.
├── Navigation        → Navbar, tabs, breadcrumbs
└── Feedback          → Alert, toast, empty state
```

Each section has a visible label frame. This is the developer's reference — they should be able to look at the library and understand every visual element available.

#### Icons for Web

Figma uses vector networks for icons. For web handoff, document which icon package to use:

| Figma icon source | Web package |
|---|---|
| Lucide icons plugin | `lucide-react` / `lucide-vue` |
| Heroicons plugin | `@heroicons/react` |
| Phosphor plugin | `@phosphor-icons/react` |
| Material Symbols | `@mui/icons-material` |

Name each icon in the library so the developer can find it in the package (e.g., `ArrowRight`, `Mail`, `Github`).

## Step 3: Assemble Screens

### Using Component Instances

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

### Applying Modes to Frames

```javascript
// Set entire screen to dark mode
const loginScreen = figma.currentPage.findOne(n => n.name === "Login Screen")
loginScreen.setExplicitVariableModeForCollection(semanticCollectionId, darkModeId)

// Set just one section to dark mode (e.g., sidebar)
const sidebar = loginScreen.findChild(n => n.name === "Sidebar")
sidebar.setExplicitVariableModeForCollection(semanticCollectionId, darkModeId)
```

This is the killer feature: **one design, multiple themes** — no duplication.

## Step 4: Dev Mode Handoff

When the design is complete:

1. **Mark frames as "Ready for Dev"**
2. Dev Mode shows variables as code tokens:
   - `fill: var(--color-primary)` instead of `fill: #2563EB`
   - `padding: var(--space-card-padding)` instead of `padding: 24px`
   - `font-family: var(--font-family-body)` instead of `font-family: Inter`
3. **Code Connect** maps components to actual React/Swift/Kotlin code snippets
4. Developers see the full alias chain: `color-primary → brand-600 → #2563EB`

## Common Patterns

### Multi-brand with Modes

Collection "Semantic" with modes: `brand-a-light`, `brand-a-dark`, `brand-b-light`, `brand-b-dark`

Each variable resolves to brand-specific primitives per mode. One design file serves multiple brands.

### Responsive with Modes

Collection "Sizing" with modes: `mobile`, `tablet`, `desktop`

```
space/page-padding:  mobile=16, tablet=24, desktop=48
font/size/display:   mobile=30, tablet=36, desktop=48
```

Apply mode to frame → entire layout adjusts.

### Nested Modes

A page in `dark` mode can have a child frame in `light` mode (e.g., a light modal on a dark page). Modes cascade down but can be overridden at any frame level.

## Common Mistakes

| Mistake | Correct approach |
|---|---|
| Creating variables without collections | Group into Primitives / Semantic / Component collections |
| All variables in one mode | Use modes for light/dark at minimum |
| No aliasing (semantic = raw hex) | Semantic variables alias primitive variables |
| No scoping | Scope color vars to fills, spacing to gaps/padding |
| Separate components for each state | Use variant property for state axis |
| Separate component for icon/no-icon | Use boolean property `hasIcon` |
| Hardcoded text in component | Use text property for editable strings |
| Duplicating screens for dark mode | Apply dark mode to frame, same screen |
| Variables not showing in Dev Mode | Ensure variables are bound to node properties, not just defined |
