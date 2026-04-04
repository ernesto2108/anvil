---
name: design-to-code
description: Translate approved designs to production code. Works with any design tool (Pencil, Figma). Syncs design tokens with CSS, maps components to code, and validates visual fidelity. Use when user says "implement this design", "code this", "translate to code", "design to code", or after a design is approved.
---

# Design to Code

> Translates approved designs into production code with visual fidelity. Tool-agnostic.

## Prerequisites

- Design must be approved (never code before design approval)
- The design file must be open — use `/design-project` first if needed

## Step 0: Detect design tool

Determine which tool the design is in:

| Signal | Tool | How to read |
|---|---|---|
| `.pen` file open | Pencil | Use Pencil MCP tools (`get_variables`, `batch_get`, `get_screenshot`) |
| Figma URL provided | Figma | Use Figma MCP tools (load `/figma-use` first) |
| Design spec / static mockup | Manual | Read dimensions and tokens from the spec document |

All subsequent steps use the appropriate tool's API but the **output is the same**: CSS, HTML, components.

## Step 0.5: Compare design vs code (MANDATORY if code already exists)

Before writing any code, create a section-by-section comparison:

| Section | Design | Code | Difference | Action |
|---|---|---|---|---|
| Hero | Linear flow, no cards | Bento grid with cards | Structure differs | Align to design |
| Nav | Hamburger + X | Hamburger only | Missing state | Add X animation |

Present this table to the user. If design and code diverge, **analyze which version is better for UX** — don't just list differences. Take the best of each, explain why, get approval before coding.

## Step 1: Sync design tokens

### Read tokens from design:

- **Pencil**: `get_variables()` — returns all variables with types and themed values
- **Figma**: read variables/styles from the Figma file via MCP

### Read tokens from code:

Read the project's CSS variables file (e.g., `global.css`, `variables.css`, `tailwind.config`)

### Diff and fix:

1. **Missing in CSS**: tokens that exist in design but not in code → add them
2. **Value mismatch**: same name but different value → flag to user
3. **Missing in design**: tokens in code not in design → may be fine (code-only utilities)

Present the diff. Fix mismatches before proceeding.

## Step 2: Read target screen

### From Pencil:
1. `batch_get` the screen frame at depth 2-3
2. `get_screenshot` of the screen

### From Figma:
1. Read the frame/page structure
2. Get a screenshot or inspect node properties

### Then (tool-agnostic):
3. Identify sections and which components are used
4. Map each section to a code component (e.g., Hero.astro, Navbar.tsx)

## Step 3: Read component structure

For each component that needs to be created or updated:

### From Pencil:
`batch_get` the reusable component at depth 3

### From Figma:
Read the component's properties, variants, and auto-layout settings

### Then map to CSS (universal):

| Design property | CSS |
|---|---|
| Vertical layout | `flex-direction: column` |
| Horizontal layout | `flex-direction: row` |
| Gap with token | `gap: var(--token-name)` |
| Fill with token | `background: var(--token-name)` |
| Corner radius | `border-radius: var(--token-name)` |
| Border/stroke | `border: 1px solid var(--token-name)` |
| Padding array | `padding: var(--top) var(--right) var(--bottom) var(--left)` |
| Fill container | `width: 100%` or `flex: 1` |
| Fit content | `width: fit-content` |
| Space between | `justify-content: space-between` |
| Center align | `align-items: center` |

## Step 4: Delegate to developer agent

The `developer` agent is the ONLY agent allowed to write production code. After steps 1-3, launch the developer agent with:

1. **Token diff** — new/changed CSS variables to add
2. **Component map** — which components to create/update, mapped from design sections
3. **Design properties** — layout, spacing, colors, typography extracted in step 3
4. **Screenshot** — design screenshot for visual reference

Include this context INLINE in the agent prompt (never tell the agent to "read file X").

Rules for the developer:
- **Use CSS custom properties**, never hardcoded values
- **Use the same semantic names** as the design tokens
- **If a design token doesn't have a CSS equivalent**, add it to the CSS file first
- **Mobile-first**: if both web and mobile designs exist, code mobile layout first, add desktop overrides with `min-width` media queries
- **Reuse existing components** — check what already exists in the codebase before creating new ones
- **Load the appropriate conventions skill** for the target stack (e.g., `astro-conventions`, `react-conventions`, `flutter-conventions`)

## Step 5: Visual QA (MANDATORY)

After implementing:

1. **Build check**: Run `build` to verify no errors
2. **Browser check**: View at target viewport
3. **Compare with design**: Open the design side by side with the browser. Check:
   - Spacing matches (padding, gap, margins)
   - Colors match (especially across themes/modes)
   - Typography matches (family, size, weight, line-height)
   - Layout matches (alignment, direction, wrapping)
4. **Check all states**: If component has interactive states, verify each one
5. **Check both modes**: If light/dark modes exist, verify both
6. **Check responsive**: If mobile + desktop, verify both viewports

**Only present to the user after ALL checks pass.**

## Design-to-Code Completeness Check (MANDATORY)

After implementing design tokens and components, verify:

1. [ ] Every CSS variable used in components has a value in both `:root` (light) and `.dark` (dark mode)
2. [ ] If dark mode exists in design → a JS mechanism toggles the `dark` class on `<html>` (hook or store)
3. [ ] If theme toggle exists in design → it's wired to the toggle mechanism and persists to `localStorage`
4. [ ] System preference: `prefers-color-scheme` is respected as the initial default
5. [ ] Every interactive element in the design (dropdowns, modals, menus) has a functional implementation, not just visual
6. [ ] Frontend request/response types match the current backend DTOs (check after any backend changes)
7. [ ] All icons use the project's icon library (e.g., `lucide-react`), not inline SVGs
8. [ ] Tailwind classes use v4 syntax if the project uses Tailwind v4: `(--var)` not `[var(--var)]`

## Rules

- **Never guess dimensions** — read them from the design file
- **Never hardcode colors** — always use CSS variables
- **Design is source of truth** — if code looks different from design, the code is wrong
- **Surgical changes** — if updating existing code, only change what the design changed
- **Token sync first** — always sync variables before coding components
- **Tool-agnostic output** — CSS is CSS regardless of whether the design came from Pencil or Figma

## Anti-Patterns

| Anti-Pattern | Fix |
|---|---|
| Coding from memory instead of reading the design | Always read the design file first |
| Hardcoded hex colors in CSS | Use `var(--color-name)` |
| Presenting code without building | Run build first |
| Presenting without visual comparison | Compare design screenshot with browser |
| Implementing mobile without checking desktop | Check both viewports |
| Assuming the design tool | Detect from file extension or URL |
