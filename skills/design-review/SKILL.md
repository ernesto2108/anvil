---
name: design-review
description: Review existing designs for quality, visual hierarchy, and anti-AI patterns. Works with Pencil (.pen files) and Figma. Use when user says "review this design", "does this look good", "improve the design", "design feedback", "design QA", or after completing visual design execution.
---

# Design Review

> Analyze an existing design for quality, propose improvements, and execute approved changes. Tool-agnostic — works with Pencil and Figma.

## When to Use

- After completing design execution (post Design Execution GATE)
- When the user asks for design feedback or improvements
- Before design-to-code translation (quality gate)
- When designs "look AI-generated" and need humanizing

## Workflow

### Step 1 — Detect Tool & Capture Current State

Detect the design tool and capture what exists:

**Pencil (.pen file):**
1. `get_editor_state` — get active file and selection
2. `batch_get` with `patterns: [{ type: "frame" }]` — list all top-level frames (screens)
3. `get_screenshot` for each screen — capture visual state
4. `get_variables` — read current design tokens

**Figma:**
1. Ask user for Figma file URL or screenshots of screens to review
2. If `use_figma` is available, inspect node structure programmatically
3. If not, work from screenshots the user provides

### Step 2 — Analyze Against Quality Checklist

Review each screen against these criteria. Score each 1-5:

#### Visual Hierarchy (weight: 25%)
- [ ] One dominant region per screen — no equal-weight competing sections
- [ ] Clear focal point — eye knows where to go first
- [ ] Action hierarchy — primary CTA is visually dominant, secondary actions reduced
- [ ] Typography hierarchy — clear size jumps between heading levels

#### Spacing & Rhythm (weight: 20%)
- [ ] Consistent spacing scale (not arbitrary pixel values)
- [ ] Tighter gaps within related content, generous whitespace between sections
- [ ] Vertical rhythm — sections have intentional, varied breathing room
- [ ] No cramped areas next to empty areas (unless intentional)

#### Color & Contrast (weight: 15%)
- [ ] Uses design token variables, not hardcoded hex values
- [ ] Sufficient contrast for text readability (WCAG AA: 4.5:1 text, 3:1 large)
- [ ] Accent color reserved for actions — not diluted across decorative elements
- [ ] Semantic colors used correctly (error for errors, success for success)

#### Anti-AI Patterns (weight: 20%)
- [ ] Intentional asymmetry — not everything is perfectly centered/mirrored
- [ ] Varied density between sections — not uniform spacing everywhere
- [ ] Real content — no "Lorem ipsum", "Item 1", "User Name" placeholders
- [ ] Progressive disclosure — complex features revealed gradually, not all at once
- [ ] Layout variation — not every section is the same card grid pattern

#### Completeness (weight: 20%)
- [ ] All interactive states designed (dropdowns open, modals visible, menus expanded)
- [ ] Loading, empty, error states exist (not just happy path)
- [ ] Mobile version exists (if responsive/both platform)
- [ ] Dark mode exists (if required)
- [ ] Every CTA has a destination screen

### Step 3 — Generate Report

Produce a structured review:

```markdown
## Design Review — <file/project name>

### Overall Score: X/10

### Screen-by-Screen Analysis

#### <Screen Name>
- **Score:** X/10
- **Strengths:** [what works well]
- **Issues:**
  1. [issue] — severity: high/medium/low — fix: [specific action]
  2. [issue] — severity: high/medium/low — fix: [specific action]

### Summary of Improvements
| # | Screen | Issue | Severity | Proposed Fix |
|---|--------|-------|----------|-------------|
| 1 | Dashboard | Equal-weight card grid | medium | Make first metric card 2x width |
| 2 | Login | Generic placeholder text | high | Use domain-specific labels |
| 3 | All screens | Uniform 24px gap everywhere | medium | Vary: 16px within sections, 32px between |

### Recommended Priority
1. [High severity fixes first]
2. [Medium fixes]
3. [Polish items]
```

### Step 4 — Propose & Confirm

Present the review to the user. Ask which improvements to apply:
- "All" — apply everything
- "High only" — only high-severity fixes
- User picks specific items

**NEVER apply changes without user confirmation.**

### Step 5 — Execute Approved Changes

**Pencil:**
- Use `batch_design` to apply changes (U for updates, R for replacements)
- Verify each change with `get_screenshot`
- Max 25 operations per batch call

**Figma:**
- Load `/figma-use` skill before any `use_figma` call
- Apply changes programmatically
- Ask user to verify visually (no screenshot tool available)

After all changes:
- Take final screenshots (Pencil) or ask user to verify (Figma)
- Compare before/after
- Report what changed

## Pencil-Specific Checks

When reviewing a Pencil file, also check:
- Load `get_guidelines("guide", "<project type>")` and verify the design follows those principles
- Check if a Pencil style was applied — if not, suggest one that fits the domain
- Verify all visual properties use `$variables`, not hardcoded values
- Check component usage — raw frames where components exist is a smell

## Figma-Specific Checks

When reviewing a Figma file:
- Check for auto-layout usage (manual positioning = fragile)
- Verify component instances vs detached copies
- Check for design token usage via styles/variables
- Verify responsive constraints

## Rules

- **Propose, don't impose** — always show the user what you'll change and get approval
- **Surgical edits** — fix specific issues, don't rebuild screens
- **Preserve intent** — improve quality without changing the design direction
- **Screenshot after each change** (Pencil) — verify no side effects
- **Score honestly** — a 10/10 means nothing needs improvement. Most designs are 6-8
