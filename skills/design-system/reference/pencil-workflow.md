# Pencil Design Tool — Design System Workflow

How to create a complete design system inside a `.pen` file using Pencil MCP tools.

## Pencil Limitations vs Figma

Pencil does NOT have native Collections or Modes. Simulate them:
- **Collections** → use naming convention prefixes: `prim-*`, `sem-*`, `comp-*`
- **Modes** → use Pencil's theme axis system for light/dark
- **Component Properties** → Pencil supports `reusable: true` but not boolean/text/instance-swap properties natively. Use descendant overrides on instances instead
- **Variants** → create separate reusable components per variant (Button-Primary, Button-Secondary) or use theme-based switching
- **Scoping** → not available. Rely on naming discipline

## Order of Operations

```
1. set_variables()     →  Create ALL variables (simulating collections via naming)
2. batch_design()      →  Build reusable components using $variables
3. batch_design()      →  Assemble screens from component instances (ref)
4. get_screenshot()    →  Verify visually after each section
```

Never skip to step 3 without completing 1 and 2.

## Step 1: Create Variables

Use `set_variables` to define everything at once. Use prefixed names to simulate collections.

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

**Critical:**
- `font-family-*` MUST be string variables. `fontFamily:"$font-body"`, never `fontFamily:"Inter"`
- `font-weight-*` MUST be **string** type (not number). Pencil's `fontWeight` expects a string. Use `{"type":"string","value":"600"}`, not `{"type":"number","value":600}`
- Variable types are **immutable** once created. If you created a variable as number and need string, create a new variable with a different name. `replace:true` in `set_variables` does NOT change types of existing variables

### Simulating Collections via Naming

Since Pencil has no native collections, use prefixes to group variables:

```
Primitives:  color-brand-500, color-gray-200, spacing-4, radius-md, font-size-base
Semantic:    color-primary, color-text-primary, color-surface, space-component-gap
Component:   (optional) button-bg, card-padding, input-border
```

The semantic variables reference the same values as primitives but with purpose-based names. In Pencil there is no aliasing — both are independent variables with the same hex/number value. Update primitives AND semantics when changing values.

### Themed Variables (Modes via Theme Axis)

Pencil supports themes via its theme axis system. Use this to simulate Figma's modes:

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

This registers a `mode` theme axis with `light` and `dark` values automatically. Apply theme to frames via the `theme` property: `{theme: {"mode": "dark"}}`.

## Step 2: Build Reusable Components

Create a component library frame, then build each component inside it.

### Component Library Structure

The library is organized in **labeled vertical sections**, not a flat row. Each section has a title label and its components below.

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

### Color Swatches

Create a swatch for each semantic color so the designer and developer can see the palette:

```javascript
// One swatch = colored circle + name label
swatch=I(colorRow,{type:"frame",layout:"vertical",gap:"$sp-2",alignItems:"center"})
swatchCircle=I(swatch,{type:"ellipse",width:40,height:40,fill:"$color-primary"})
swatchName=I(swatch,{type:"text",content:"primary",fontFamily:"$font-mono",fontSize:"$fs-xs",fill:"$color-text-muted"})
```

### Icon Samples

Show every icon used in the project. This is the developer's reference for which icons to import.

Pencil uses icon fonts (Lucide, Material Symbols, Phosphor, Feather). For web implementation, these become SVG packages:

| Pencil icon font | Web package | Install |
|---|---|---|
| `lucide` | `lucide-react` or `lucide-vue` | `npm i lucide-react` |
| `feather` | `react-feather` | `npm i react-feather` |
| `Material Symbols Outlined` | `@mui/icons-material` | `npm i @mui/icons-material` |
| `phosphor` | `@phosphor-icons/react` | `npm i @phosphor-icons/react` |

```javascript
// Icon sample = icon + name label below
iconSample=I(iconRow,{type:"frame",layout:"vertical",gap:"$sp-2",alignItems:"center"})
I(iconSample,{type:"icon_font",iconFontName:"mail",iconFontFamily:"lucide",width:24,height:24,fill:"$color-text-primary"})
I(iconSample,{type:"text",content:"mail",fontFamily:"$font-mono",fontSize:"$fs-xs",fill:"$color-text-muted"})
```

Create samples at the standard sizes used in the project (typically 16px for inline, 20px for buttons, 24px for standalone).

### Text Style Components

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

### Button Components

```javascript
// Button/Primary
btnPrimary=I(lib,{type:"frame",name:"Button/Primary",reusable:true,layout:"horizontal",width:"fit_content",height:"fit_content",padding:["$spacing-3","$spacing-5"],fill:"$color-primary",cornerRadius:"$radius-md",justifyContent:"center",alignItems:"center",gap:"$spacing-2"})
btnPrimaryText=I(btnPrimary,{type:"text",content:"Button",fontFamily:"$font-family-body",fontSize:"$font-size-sm",fontWeight:"$font-weight-semibold",fill:"$color-text-inverse"})

// Button/Secondary
btnSecondary=I(lib,{type:"frame",name:"Button/Secondary",reusable:true,layout:"horizontal",width:"fit_content",height:"fit_content",padding:["$spacing-3","$spacing-5"],fill:"$color-surface",stroke:{thickness:1,fill:"$color-border-default"},cornerRadius:"$radius-md",justifyContent:"center",alignItems:"center",gap:"$spacing-2"})
btnSecText=I(btnSecondary,{type:"text",content:"Button",fontFamily:"$font-family-body",fontSize:"$font-size-sm",fontWeight:"$font-weight-medium",fill:"$color-text-primary"})
```

### Input Component

```javascript
// Input/Field (label + input frame + placeholder)
inputField=I(lib,{type:"frame",name:"Input/Field",reusable:true,layout:"vertical",width:320,gap:"$spacing-2"})
inputLabel=I(inputField,{type:"text",content:"Label",fontFamily:"$font-family-body",fontSize:"$font-size-sm",fontWeight:"$font-weight-medium",fill:"$color-text-primary"})
inputBox=I(inputField,{type:"frame",layout:"horizontal",width:"fill_container",height:44,padding:["$spacing-3","$spacing-4"],fill:"$color-surface",stroke:{thickness:1,fill:"$color-border-default"},cornerRadius:"$radius-md",alignItems:"center"})
inputPlaceholder=I(inputBox,{type:"text",content:"Placeholder",fontFamily:"$font-family-body",fontSize:"$font-size-base",fontWeight:"$font-weight-normal",fill:"$color-text-muted"})
```

### Section Header Component

```javascript
// Section/Header (title + accent line)
sectionHeader=I(lib,{type:"frame",name:"Section/Header",reusable:true,layout:"vertical",width:"fit_content",gap:"$spacing-3"})
sectionTitle=I(sectionHeader,{type:"text",content:"SECTION TITLE",fontFamily:"$font-family-heading",fontSize:"$font-size-xs",fontWeight:"$font-weight-semibold",fill:"$color-text-primary",letterSpacing:2})
sectionLine=I(sectionHeader,{type:"rectangle",width:48,height:3,fill:"$color-primary"})
```

### Card Component

```javascript
// Card
card=I(lib,{type:"frame",name:"Card",reusable:true,layout:"vertical",width:400,padding:"$spacing-6",fill:"$color-surface",stroke:{thickness:1,fill:"$color-border-default"},cornerRadius:"$radius-lg",gap:"$spacing-4"})
```

### Divider Component

```javascript
// Divider
divider=I(lib,{type:"rectangle",name:"Divider",reusable:true,width:400,height:1,fill:"$color-border-default"})
```

## Step 3: Assemble Screens from Components

Use `ref` to instantiate components. Override properties via the root or `descendants`.

### Example: Using a component instance

**CRITICAL: Use `descendants` on the `ref`, NOT `U()` after inserting.**
Using `U(instance+"/childId")` modifies the COMPONENT MOTHER, corrupting all instances.

```javascript
// CORRECT — customize via descendants at insert time
header1=I(mainContent,{type:"ref",ref:"sectionHeaderId",descendants:{"titleTextId":{content:"EXPERIENCE"}}})

// CORRECT — input with label + placeholder overrides
emailInput=I(formFrame,{type:"ref",ref:"inputFieldId",descendants:{"labelId":{content:"Email"},"placeholderId":{content:"you@company.com"}}})

// CORRECT — button with text override and size change
submitBtn=I(formFrame,{type:"ref",ref:"btnPrimaryId",width:"fill_container",descendants:{"btnTextId":{content:"Submit"}}})
```

```javascript
// WRONG — this modifies the component mother, not the instance!
header1=I(mainContent,{type:"ref",ref:"sectionHeaderId"})
U(header1+"/titleTextId",{content:"EXPERIENCE"})  // ← CORRUPTS THE COMPONENT
```

### Key Rules for Instances

- Customize content: use `descendants` in the `ref` insert call, OR `U(instance+"/childId")` after insertion
- Resize: override `width` or `height` directly on the `ref` node
- Hide a child: `descendants:{"childId":{enabled:false}}`
- Replace a child: `R(instance+"/childId", {type:"text",...})` (only for structural replacement)
- `U(instance+"/childId")` is SAFE — it only modifies the instance's override, not the mother component. Example: `U("YkHfO/MNS4B",{content:"New text"})` changes text in instance `YkHfO` only
- `U("childId")` WITHOUT instance prefix modifies the mother component — NEVER do this to customize an instance
- NEVER recreate a component manually — always use `ref`

### Instance Replacement Gotchas

- When you `R(instance+"/childId")`, the replacement creates a NEW node with a new ID. The old ID is gone
- If you need to modify a previously-replaced node, use the NEW node ID, not the original: `R("YkHfO/newNodeId")` not `R("YkHfO/originalId")`
- If `R()` fails with "No such node", the node was already replaced or deleted. Use `batch_get` to find the current ID
- Alternative pattern when R() fails: `D(nodeId)` + `I(parentId, {...})` + `M(newNode, parentId, position)`

### Component Library Placement

- Position the library frame to the **RIGHT** of all screens (e.g., `x:3200`)
- Never place it below screens where it gets hidden
- After assembling screens, **verify components are intact**: `get_screenshot(libraryFrameId)`

## Step 4: Verify

After each major section:

```
get_screenshot(nodeId) → visually inspect
```

Check for:
- Text visibility (all text has `fill` set via variable)
- Alignment (flexbox layout, no hardcoded x/y in flex children)
- Spacing consistency (gaps and padding use `$spacing-*` variables)
- Color appropriateness (matches the approved proposal)
- Component reuse (no duplicated node structures)

## Text Inside Containers (CRITICAL — #1 recurring bug)

Every text node inside a frame with limited width (cards, callouts, bento tiles):

```javascript
// CORRECT — text wraps inside card
I(card,{type:"text",content:"Long text...",textGrowth:"fixed-width",width:"fill_container",fontFamily:"$font-sans",fontSize:"$fs-base",fill:"$color-text-secondary",lineHeight:"$lh-relaxed"})

// WRONG — text overflows/truncates
I(card,{type:"text",content:"Long text...",fontFamily:"$font-sans",fontSize:"$fs-base",fill:"$color-text-secondary"})
```

- **MUST** set `textGrowth: "fixed-width"` + `width: "fill_container"` on any text inside a width-limited container
- **Exception**: short labels, badges, buttons (1 line max) — keep default `textGrowth: "auto"`
- **Verify immediately** with `get_screenshot` after inserting text

## Grid Harmony

When creating cards in a horizontal layout (bento grids, card rows):

- Cards with `height: "fit_content"` will have **uneven heights** if content varies
- **PREFER fixed equal height** for sibling cards (e.g., all 220px)
- Screenshot immediately after creating any card grid to verify alignment

## Additional Pencil Limitations

- **String variables in `content` resolve ONLY within theme context** — `content: "$txt-title"` works on instance descendants (inherits parent theme) but NOT on nodes created via `R()` (Replace). Replaced nodes lose theme context. For reliable i18n, use `U(instance+"/childId",{content:"$txt-var"})` on existing descendants, or hardcode text on replaced nodes
- **Copied node descendants get new IDs** — after `C()`, child IDs are regenerated. Never `U()` with original child IDs on a copy. Either use `descendants` in the Copy operation itself, or `batch_get` the copy to read new IDs first
- **Component changes cascade to instances — unless overridden** — if an instance replaced a node (e.g. text → frame with bullets), deleting and re-creating that node in the component WILL cascade to instances that hadn't overridden it. Instances with overrides keep their overrides (now orphaned). Plan component restructuring carefully

## Interactive Component States (MANDATORY for mobile/app design)

Pencil has NO prototyping. To compensate, every component with interactive states MUST have a dedicated **States frame** showing all states side by side.

### When to create state frames

- **Navigation**: closed + open (hamburger menu)
- **Buttons**: default + hover + disabled + loading
- **Inputs**: empty + focused + filled + error
- **Cards**: collapsed + expanded (if expandable)
- **Modals/Sheets**: the overlay + the content
- **Toggles**: on + off
- **Dropdowns**: closed + open with options

### How to structure

Create a top-level frame named `{Component} — States` with all states labeled:

```javascript
// Example: Navbar mobile states
states=I(document,{type:"frame",name:"Navbar/Mobile — States",layout:"vertical",width:390,gap:"$sp-8",padding:"$sp-6",fill:"$color-bg",theme:{mode:"dark"}})
closedLabel=I(states,{type:"text",content:"State: Closed",fill:"$color-text-muted",fontFamily:"$font-mono",fontSize:"$fs-xs",letterSpacing:2})
closedNav=I(states,{type:"ref",ref:"navMobileId",width:"fill_container"})
openLabel=I(states,{type:"text",content:"State: Open",fill:"$color-text-muted",fontFamily:"$font-mono",fontSize:"$fs-xs",letterSpacing:2})
// ... build the open state
```

### Rules

- **Create states automatically** when designing the component — don't wait for the user to ask
- **Label each state** clearly with a `State: {name}` text above it
- **Place near the component** in the canvas, not in the Library frame
- **Use the same theme** as the target screen (dark/light)
- **Both modes if applicable** — if the component appears in dark + light screens, show states for both

## Canvas Organization (MANDATORY)

Keep the canvas organized chronologically and by type. Always leave ~200px gaps between rows and between frames horizontally so the user can iterate without frames colliding.

### Layout order (top to bottom)

```
ROW 1 — Library + Component States + Loose Layers
  Library frame (design tokens + components)
  Component state frames (Navbar states, expanded cards, modals, etc.)
  Always at the TOP — this is the reference layer

ROW 2 — v1 (first iteration of screens)
  Oldest design iteration, kept for history

ROW 3 — v2 (current iteration of screens)
  Latest web screens: dark, light, language variants, blog, etc.

ROW 4 — Mobile screens
  Mobile dark, mobile light, mobile menu open, etc.

(Add more rows below as new iterations or platforms are added)
```

### Rules

- **~200px gap** between every row and between horizontal frames — never pack frames tight
- **Chronological top-to-bottom** — oldest at top, newest at bottom
- **Library always ROW 1** — it's the reference, not a screen
- **States/layers next to Library** in the same row, not scattered across the canvas
- **After reorganizing**, run `snapshot_layout(problemsOnly: true)` to verify no overlaps
- **Never delete frames to reorganize** — only move with `U(id, {x, y})`

## Cross-Screen Content Sync (MANDATORY)

When updating content in one screen (text, badges, labels, data), check ALL screens that share the same data:

1. `batch_get` with a pattern matching the old value across the entire document
2. Update every instance — dark, light, EN, ES, mobile, desktop
3. If the content comes from a variable (`$txt-*`), update the variable instead — all screens update automatically
4. If it's hardcoded text, search and replace in every screen manually

**Why:** K8s was changed to Kafka in mobile dark but not in mobile light or web screens, creating inconsistency. Content must be consistent across all screens.

## Icon Registry (MANDATORY)

Every icon used in any screen or component MUST appear in the Library's Icons section. When adding a new icon to a design:

1. Use the icon in the component/screen
2. **Immediately** add it to the Icons section in the Library frame
3. Each icon sample: icon at 24px + name label below, inside a vertical frame with `alignItems: center`

Never leave an icon undocumented. The Library is the developer's reference for which icons to install.

## Common Mistakes to Avoid

| Mistake | Correct approach |
|---|---|
| `fontFamily:"Inter"` | `fontFamily:"$font-family-body"` |
| `fontWeight:"600"` | `fontWeight:"$font-weight-semibold"` |
| `fill:"#1E40AF"` | `fill:"$color-primary"` |
| `fontSize:14` | `fontSize:"$font-size-sm"` |
| `cornerRadius:8` | `cornerRadius:"$radius-md"` |
| `gap:16` | `gap:"$spacing-4"` |
| `padding:24` | `padding:"$spacing-6"` |
| Building same card 4 times | Create card component once, use 4 `ref` instances |
| Designing without a plan | Present color/type/tone proposal first |
| Text in card without `textGrowth` | Add `textGrowth:"fixed-width"` + `width:"fill_container"` |
| Sibling cards with uneven heights | Set fixed equal height for all cards in a row |
| `content:"$txt-var"` on replaced nodes | Use `U(instance+"/descendantId")` for variables, or hardcode on replaced nodes |
| `U(copy+"/originalChildId")` | Read copy's new IDs first, or use descendants in `C()` |
| `R(instance+"/oldReplacedId")` | Use current node ID from `batch_get`, not the original component ID |
| `U("childId")` to customize instance | Use `U(instance+"/childId")` — without prefix you modify the mother |
| Adding all info at once (tags, links, metadata) | Start minimal, verify, then add layers. Secondary info at low opacity (0.4-0.6) |
| Inventing content for designs | Use real data from CV, LinkedIn, or user-provided docs |
