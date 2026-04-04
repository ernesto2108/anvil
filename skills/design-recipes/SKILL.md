---
name: design-recipes
description: Reusable design patterns for efficient screen building in Pencil or Figma. Reduces operations per screen by providing tested recipes. Load during the Design Execution GATE phase. Use when building screens from design system components in any design tool.
---

# Design Recipes

> Tested patterns for assembling screens efficiently. Tool-agnostic patterns with tool-specific implementations.

## When to Load

Load this skill during the **Design Execution GATE** in the orchestration pipeline — after the designer produces ui-spec.md and before executing the visual design.

## Workflow

### Step 0 — Load Pencil Guidelines (if .pen file)

Before building ANY screen, load the Pencil guidelines that match the project type:

1. **Detect project type** from the PRD/ui-spec:
   - SaaS dashboard, admin panel, CRM → `get_guidelines("guide", "Web App")`
   - Landing page, marketing site → `get_guidelines("guide", "Landing Page")`
   - Mobile app → `get_guidelines("guide", "Mobile App")`
   - Data-heavy screens, tables → `get_guidelines("guide", "Table")`
   - Design system work → `get_guidelines("guide", "Design System")`

2. **Explore visual styles** — Pencil offers curated style archetypes. Before starting:
   - Run `get_guidelines()` to see available styles
   - Pick a style that matches the domain (e.g., "Soft Bento" for friendly SaaS, "Aerial Gravitas" for enterprise, "Editorial Scientific" for data-heavy apps)
   - Load it: `get_guidelines("style", "<chosen style>")`

3. **Apply guidelines alongside recipes** — Pencil guidelines define principles (hierarchy, density, feedback). Recipes below define structure. Use both.

### Step 1 — Build Screens

1. Detect the design tool:
   - `.pen` file → load `reference/pencil.md`
   - Figma URL or `.fig` → load `reference/figma.md`
2. Identify which screen types you're building
3. Follow the recipe for each type
4. Verify with screenshot after each screen

## Screen Type Recipes

### Recipe 1: Auth Screen (Login, Register, Verify)

**Pattern:** Split layout — brand panel (left/top) + form panel (right/bottom)

**Structure:**
```
Desktop (1440×900):
┌──────────────┬─────────────────────────┐
│ Brand Panel  │     Form Panel          │
│ 560px        │     fill_container      │
│ Primary bg   │     Centered card       │
│ Logo+tagline │     Title+fields+CTA    │
└──────────────┴─────────────────────────┘

Mobile (375×812):
┌─────────────────────┐
│ Brand Header (compact) │
│ Primary bg, 1 line     │
├─────────────────────┤
│ Form (full width)      │
│ Title+fields+CTA       │
└─────────────────────┘
```

**Components needed:** InputGroup, InputPassword, Button/Primary, Button/Ghost
**Operations:** ~12 desktop, ~10 mobile
**Reuse tip:** Build Login first, then Copy for Register/Verify and modify content

### Recipe 2: App Shell (Nav + Content)

**Pattern:** Top nav + scrollable content area

**Structure:**
```
Desktop:
┌─────────────────────────────────────────┐
│ NavBar (ref, fill_container width, 64h) │
├─────────────────────────────────────────┤
│ Content area (padding 32-48, vertical)  │
│ ├── Page header (title + actions)       │
│ ├── Content sections                    │
│ └── Pagination (if table)               │
└─────────────────────────────────────────┘

Mobile:
┌─────────────────────┐
│ Mobile Nav (56h)     │
│ ☰  Logo  👤         │
├─────────────────────┤
│ Content (padding 16) │
└─────────────────────┘
```

**Components needed:** NavBar (desktop) or MobileNav, Avatar
**Reuse tip:** Build one app shell, Copy for each page, replace content only

### Recipe 3: Data Table Page

**Pattern:** Header + filters + table + pagination

**Structure:**
```
Content area:
├── Page header: Title (left) + Primary button (right)
├── Filters row: Dropdown selects (horizontal)
├── Table card (surface bg, rounded, border):
│   ├── Header row (neutral-50 bg): column labels
│   ├── Data rows: cells with text/badges
│   └── Last row: no bottom border
└── Pagination: info text (left) + prev/next buttons (right)
```

**Table row pattern (CRITICAL):**
```
Row (horizontal, fill_container, padding [14, 20], bottom border)
├── Cell 1 (frame, width: fill_container or fixed) → content
├── Cell 2 (frame, width: fill_container or fixed) → content
├── Cell N (frame, width: fixed for badges/dates) → badge/text
```

**Column width guide:**
- Name/title: fill_container
- Description: fill_container
- Status badge: 120-140px
- Date: 140-160px
- Count/number: 80-100px
- Actions: 80-100px

**Operations:** ~25 for header+table+3 rows. Split into 2 calls.

### Recipe 4: Detail Page

**Pattern:** Breadcrumb + header with actions + multi-column content

**Structure:**
```
Content area:
├── Breadcrumb: parent > current
├── Header: title + badge + edit button
├── Description text
├── Metadata row (horizontal, gap 24)
├── Two-column layout (horizontal, gap 24):
│   ├── Left column (fill_container): primary info
│   └── Right column (fill_container): secondary info
└── Related items section
```

### Recipe 5: Wizard/Stepper

**Pattern:** Breadcrumb + stepper + form card + navigation buttons

**Stepper states:**
- Completed: green circle + check icon + green text + green line
- Active: primary circle + number + primary text
- Upcoming: neutral circle + number + secondary text + neutral line

**Structure:**
```
Content area:
├── Breadcrumb
├── Title
├── Stepper (horizontal, 3 steps)
├── Form card (surface bg, rounded, padding 32):
│   ├── Section title
│   ├── Description
│   ├── Form fields
│   └── Actions: Back (left) + Next/Create (right)
```

**Reuse tip:** Build step 1, Copy for steps 2-3, update stepper states + form content

### Recipe 6: Mobile Card List (replaces tables)

**Pattern:** Vertical stack of cards instead of table rows

**Card structure:**
```
Card (surface bg, rounded-lg, border, padding 16):
├── Top row (horizontal, space_between):
│   ├── Name/title (semibold)
│   └── Status badge
├── Description (secondary text, xs)
└── Metadata (disabled text, xs)
```

### Recipe 7: Hamburger Menu (Mobile)

**Pattern:** Full-screen overlay with sections

**Structure:**
```
Full screen (375×812):
├── Nav bar: X (close) + Logo + Avatar
├── Content (vertical, padding):
│   ├── SECTION: "NAVEGACIÓN" (label, uppercase, xs)
│   │   ├── Menu item (active: primary bg + primary text)
│   │   ├── Menu item (icon + text)
│   │   └── Menu item
│   ├── Divider
│   ├── SECTION: "APARIENCIA"
│   │   └── Theme toggle (icon + text + switch)
│   ├── Divider
│   ├── SECTION: "CUENTA"
│   │   ├── Profile
│   │   ├── Settings
│   │   └── Logout (error color)
├── Spacer (fill_container)
└── User bar (bottom): avatar + name + email + role badge
```

### Recipe 8: Avatar Dropdown (Desktop)

**Pattern:** Floating card anchored to avatar

**Structure:**
```
Dropdown (260w, surface bg, rounded-lg, shadow-lg):
├── User info row: avatar + name/email + role badge
├── Divider
├── Menu items: icon (18px) + text
│   ├── Profile
│   ├── Settings
│   └── Theme toggle (icon + text + switch)
├── Divider
└── Logout (error color)
```

**Placement:** absolute positioned, x = nav_width - dropdown_width - 24, y = nav_height - 4

## Dark Mode Recipe

1. Build light version first
2. Copy the frame: `C("lightFrameId", document, {name: "Dark: ...", positionDirection: "bottom", positionPadding: 100, theme: {"mode": "dark"}})`
3. Override tool-specific elements:
   - Theme toggle: icon moon→sun, text "Modo oscuro"→"Modo claro", switch OFF→ON
4. Verify with screenshot — check contrast on badges and text

## Efficiency Rules

1. **Components FIRST** — build all reusable components before any screen
2. **Copy, don't rebuild** — first screen of each type is built, variants are copied
3. **Max 25 ops per batch** — split large screens into logical sections
4. **Verify after each screen** — screenshot to catch issues early
5. **Use refs, not raw frames** — every repeated pattern should be a component instance
6. **Batch related updates** — group all overrides for one instance in a single call
