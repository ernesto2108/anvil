---
name: design-system
description: Create and maintain design system foundations — variables, collections, modes, color palettes, typography, spacing, reusable components with variants, and theming. Use when user says "design system", "design tokens", "color palette", "typography scale", "spacing scale", "create variables", "define colors", "theme setup", "dark mode", "component library", or when the designer agent needs to establish visual foundations before designing screens.
---

# Design System

> **IMPORTANT:** Dispatcher only. Load reference files on demand. See routing table below.

## Philosophy

- **Understand → Plan → Approve → Build** — first understand WHAT you're designing (web? app? document?), then propose visuals, get approval, then build. Never skip steps
- **Collections → Components → Screens** — organize variables into collections, build components from those variables, assemble screens from component instances. Never skip layers
- **Iterate, never rebuild** — when the user requests a change, modify only what changed. Never delete existing work to start over
- **Modes are first-class** — light/dark, brand variants are not afterthoughts. Design for modes from day one
- **Semantic over literal** — name tokens by purpose (`color-primary`), not value (`blue-500`)
- **Components are configurable, not duplicated** — properties and variants, not 20 separate components

## Workflow

### 0. Understand the Deliverable (MANDATORY)

**Gate:** Before proposing visuals, understand WHAT you're designing.

Ask (in Spanish): "Que tipo de producto es? Pagina web estatica? App web? App movil? Landing page? Dashboard? Documento?"

This determines layout patterns, navigation, density, and component needs. A web portfolio is NOT a PDF resume. A dashboard is NOT a landing page.

#### Platform Detection

Read the PRD's **Scope → Platform** field. If missing, ask: "Para qué plataforma? Web, mobile, o ambos?"

- `web` → standard web tokens
- `mobile` → load `reference/platform-guide.md`. Use iOS/Android type scales, touch targets (44pt+), platform-native font sizes
- `both` → load `reference/platform-guide.md`. Generate BOTH web and mobile token sets. Document the mapping (abstract → web CSS → iOS Swift → Android Compose) as shown in the platform guide

### 1. Research & Inspiration (MANDATORY)

**Gate:** Before proposing visuals, research what works in the domain. Never design from scratch.

1. **Inspiration sites** — search for references on:
   - **Dribbble** / **Behance** — UI components and full case studies
   - **Awwwards** — award-winning web design
   - **Mobbin** / **Screenlane** — real mobile app patterns (essential if platform is `mobile` or `both`)
   - **Collectui** / **Landbook** — categorized UI inspiration
2. **Domain-specific search** — search "{project domain} UI design" (e.g., "healthcare SaaS dashboard design", "B2B workflow app mobile"). Find 3-5 real products in the same domain
3. **Font research** — search Google Fonts for fonts matching the project tone. Propose 2-3 pairings (heading + body). Consider:
   - Readability at small sizes (critical for mobile)
   - Variable font availability (better performance)
   - Language support (if i18n is needed)
   - If platform is `both`: verify the font works on web AND has a good mobile equivalent (or use the same)
4. **Color palette research** — search palette tools (Coolors, ColorHunt, Realtime Colors) for palettes matching the domain

Document all findings — they feed into the visual proposal.

### 1.5. Present Visual Proposal (MANDATORY)

**Gate:** Before ANY canvas work, present a proposal and get explicit approval.

Include:
- **References found** (3-5 links with what you liked about each)
- **Color direction** (2-3 options matching the domain context, with full scale preview 50→950)
- **Typography** (specific Google Fonts with links, not just "sans-serif". Show the pairing rationale)
- **Tone** (density, corners, mood, reference site)
- **Layout** (structure, navigation, mobile approach)
- **Modes** — if user wants dark+light, show both in the proposal
- **Platform considerations** — if `both`, explain how web and mobile layouts will differ

Format in Spanish. Wait for explicit approval. Iterate if needed.

### 2. Check Existing System

Read `<docs>/01-project/design-system.md`. If exists → skip to step 7.

### 3. Create Variable Collections

See `reference/primitives.md` for complete scales. See `reference/semantic-tokens.md` for mapping.

**Collection 1: Primitives** — raw values, no semantic meaning, never consumed directly
**Collection 2: Semantic** (with modes) — purpose-based names that alias primitives, values change per mode
**Collection 3: Component** (optional) — scoped to specific UI elements

#### Color Scale (MANDATORY — full 11-step ramp)

Every hue family MUST have the complete scale: **50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950**. This applies to:
- **Brand primary** — the main brand color ramp
- **Brand secondary** — if the project uses a secondary color
- **Neutral/Gray** — the gray family chosen for the project
- **Status colors** — red (danger), amber (warning), green (success), blue (info)

Never define just "primary color" as a single hex. The full ramp is required for hover states, subtle backgrounds, borders, dark mode mapping, and accessible contrast.

#### Font Selection (MANDATORY — specific font, not system stack)

Choose a specific font from **Google Fonts** (or the project's brand font if provided). The system font stack is a fallback, not a choice.

1. Select 2-3 font candidates based on the research from Step 1
2. Define the full weight set used: 300 (light), 400 (normal), 500 (medium), 600 (semibold), 700 (bold)
3. Define the full type scale: **display, 3xl, 2xl, xl, lg, base, sm, xs** — with pixel values from `reference/primitives.md`
4. If platform is `both`: define web sizes (px/rem) AND mobile sizes (pt for iOS, sp for Android) using `reference/platform-guide.md`

#### Platform-Specific Tokens

If the PRD `Platform` is `mobile` or `both`:
- Load `reference/platform-guide.md`
- Add mobile-specific tokens: touch target minimums (44pt iOS, 48dp Android), safe area awareness, platform font sizes
- Document the abstract → platform mapping table

Tool-specific implementation: load `reference/pencil-workflow.md` or `reference/figma-workflow.md`.

### 4. Build Component Library (MANDATORY)

**Gate:** Components MUST exist BEFORE any screen design.

Each component needs: **properties** (boolean, text, instance swap), **variants** (state, size, type), and **all values from variables**.

#### Library Structure

Split into **2 separate frames** positioned to the RIGHT of screens:

**Frame 1: "Design Tokens"** — visual reference, not reusable components. Uses columns.

```
Design Tokens
├── Typography                          ├── Colors                    ├── Icons
│   Type Scale (descending)             │   Grouped by family:        │   All project icons
│   ├── fs-hero / 48  Hero Display      │   ├── Neutrals (50→950)     │   at 24px with name
│   ├── fs-3xl / 30   Heading 3XL       │   ├── Brand                 │   below each one.
│   ├── fs-2xl / 24   Heading 2XL       │   ├── Status                │   Split into rows
│   ├── ...down to...                   │   Each swatch shows:        │   if they don't fit.
│   └── fs-xs / 11    Caption           │   ○ circle + name + hex     │   Document icon set
│   ─────────────                       │   Light row + Dark row      │   name for devs.
│   Weights                             │   side by side per family   │
│   ├── 700 bold  "Quick brown fox"     │                             │
│   ├── 600 semi  "Quick brown fox"     │                             │
│   ├── 500 med   "Quick brown fox"     │                             │
│   └── 400 norm  "Quick brown fox"     │                             │
```

**Frame 2: "Components"** — reusable components grouped by category.

```
Components
├── Primitives (column)        ├── Cards (column)
│   Text/Heading               │   Job/Card
│   Text/Body                  │   ├── default state
│   Text/Caption               │   ├── (hover if applicable)
│   Text/Label                 │   Project/Card
│   Section/Header             │   Stat/Card
│   Skill/Badge                │
│   Contact/Link               │
│   Divider                    │
├── Navigation (full width, below columns)
│   Navbar (full width)
│   Footer
```

#### Typography Presentation

- Show the **full type scale descending** — every font-size variable from largest to smallest
- Each line: `variable-name / px-value` (mono, muted) + sample text at that size
- Below the scale, show **weights** — same phrase in bold/semibold/medium/normal
- Show font family names (heading, body, mono) at the top

#### Color Presentation

- Group by **family**: Neutrals, Brand, Status (success/warning/danger), Accent
- For each family: show swatches with **name + hex value** below each circle
- For variables with **modes**: show Light row and Dark row side by side (use frame-level theme switching)
- If the project uses a full scale (50→950), show the entire ramp per hue
- If minimal palette, show only the semantic colors used but still grouped by purpose

#### Icons

- Show every icon used in the project at standard size (24px) with **name** below
- Split into rows if they don't fit in one line — never let icons overflow the frame
- Document the **icon set name** (Lucide, Heroicons, Phosphor) so devs install the correct npm package
- All modern icon sets use `currentColor` — icons inherit parent text color, adapting to themes automatically
- Prefer sets with **size-specific variants** (Heroicons 24/20/16px are redrawn, not scaled)

#### Component Presentation

- Each component shows its **variants side by side** where applicable (default, hover, disabled)
- Each component shows its **sizes** if it has size variants (sm, md, lg)
- Group by category with clear **labels and separators**
- Reusable text style components (Heading, Body, Caption, Label) belong in Primitives, not Typography

#### Minimum Components

Text styles, Button (primary/secondary), Input, Card, Section Header, Badge, Divider. Additional per project needs.

#### Library Documentation is NOT Optional (MANDATORY)

The "Design Tokens" frame MUST be fully populated before any screen design begins. An empty or partial token documentation frame is a blocking error.

Minimum required documentation:
- **Color palette**: All hue families with swatches showing variable name + hex value
- **Typography scale**: All sizes from Display to XS with live text samples
- **Font families**: Character set samples for each family (primary + monospace)
- **Icon inventory**: Every icon used in the project at 24px with name label
- **Spacing scale**: Visual bars showing each spacing value
- **Border radius**: Sample boxes showing each radius level

Without this documentation, developers cannot implement the design system correctly and will hardcode values instead of using tokens.

#### Placement & Sizing

- **Always to the RIGHT** of screens, never below or behind
- **Wrap Design Tokens + Components in a parent vertical auto-layout frame** ("Library") with gap. This prevents overlap automatically — never use fixed y-positions between library frames
- **Width must accommodate content** — calculate based on number of columns and largest component
- **Check for overflow** after building — if anything is clipped, widen the frame
- Use `snapshot_layout` to verify after building

### 5. Define Modes & Verify Contrast

At minimum plan light/dark. Dark mode is NOT inversion — elevated surfaces get lighter.
Verify WCAG AA (4.5:1 text, 3:1 large text) for ALL modes.

### 6. Assemble Screens from Components

Use component instances (`ref` in Pencil, instances in Figma). Override content via `descendants`, never via `U()` on the component mother.

**If user requests both dark and light:** design one, copy the frame, change only the theme/mode. Show both.

### 7. Validate / Extend Existing System

Check categories, verify gaps, propose additions (don't modify without approval), flag inconsistencies.

### 8. Produce Output

Create `<docs>/01-project/design-system.md`. See `reference/output-template.md`.

## Iteration Rules (CRITICAL)

- **NEVER delete work to apply a change** — identify what changed, modify only that
- **Component change** → edit the component mother → all instances update automatically
- **Variable change** → update the variable value → all nodes update automatically
- **Layout change** → modify the section structure, keep everything else
- **If a change requires restructuring** → explain the scope to the user first, get approval before touching anything

## Rules

- **Understand first** — know what you're designing before proposing colors
- **Plan before pixels** — visual proposal approved before any canvas work
- **Collections always** — primitives, semantic, component. Never flat dumps
- **Variables for everything** — fonts, weights, sizes, colors, spacing, radius. All
- **Components before screens** — with properties/variants, not raw nodes
- **Color matches context** — professional = slate/navy, playful = vibrant, tech = cool neutral
- **Iterate, don't rebuild** — change requests = surgical edits, not delete-and-redo
- **Component Library visible** — always to the right of screens, never hidden
- **Contrast mandatory** — verify all modes
- **Reuse, never recreate** — 4 cards = 1 component + 4 instances

## Anti-Pattern Detection

| Anti-Pattern | Severity | Fix |
|---|---|---|
| Designing without understanding the deliverable type | error | Ask what it is first |
| Jumping to canvas without proposal | error | Present visual proposal |
| Deleting work to apply a change | error | Edit only what changed |
| `fontFamily:"Inter"` hardcoded | error | Use `$font-body` variable |
| `fontWeight:"600"` hardcoded | error | Use `$fw-semibold` variable |
| `fill:"#22C55E"` or any hardcoded hex | error | Use `$color-link` or appropriate variable |
| 4 cards built manually | error | 1 component + 4 `ref` instances |
| `U()` on component mother from instance | error | Use `descendants` on the `ref` |
| Component Library behind/under screens | error | Position to the right, always visible |
| Red accent for professional CV | error | Match colors to domain |
| Only showing dark, user asked for both | warning | Show dark AND light side by side |
| No contrast check | error | Verify WCAG AA 4.5:1 |
| Writing code before design is approved | error | Design → approve → code. Never skip |
| Mocked/invented data (fake repos, fake stats) | error | Only show real data. Better 2 real items than 4 fake ones |
| Color scale with only 1-3 values per hue (e.g., just `primary: #2563eb`) | error | Full 11-step ramp required (50→950) for every hue family |
| Using system font stack without selecting a specific font | error | Choose a specific font from Google Fonts. System stack is fallback only |
| Type scale missing sizes (e.g., only base and heading) | error | Full scale required: display, 3xl, 2xl, xl, lg, base, sm, xs |
| Designing screens before variables + component library exist | error | Variables → Components → Screens. Never skip layers |
| No design research/inspiration before proposing visuals | error | Research references in Dribbble/Behance/Mobbin before proposing |
| No mobile tokens when PRD Platform is `both` or `mobile` | error | Load `reference/platform-guide.md` and define platform-specific tokens |
| Mobile components without 44pt+ touch targets | error | All interactive elements must meet minimum touch target size |

## Tool Limitations (Pencil)

- **Variable types are immutable** — plan types (color/string/number) before creating. If you need to change type, use a new variable name
- **`fontWeight` requires string type** — create weight variables as `{"type": "string", "value": "600"}`, not number
- **Font family warnings** — string variables for `fontFamily` show "invalid" in Pencil. This is cosmetic, not an error
- **No native aliasing** — semantic and primitive variables are independent. Update both when changing values

For text wrapping, grid harmony, i18n limitations, and other Pencil-specific guardrails, see `reference/pencil-workflow.md`.

## Reference Files

| Working on... | Load |
|---|---|
| Raw value scales (color, type, spacing, radius, shadow) | `reference/primitives.md` |
| Semantic token mapping | `reference/semantic-tokens.md` |
| Output template for design-system.md | `reference/output-template.md` |
| Platform guidance (web vs mobile vs both) | `reference/platform-guide.md` |
| **Pencil** — variables, components, instances | `reference/pencil-workflow.md` |
| **Figma** — collections, modes, variants, Dev Mode | `reference/figma-workflow.md` |
