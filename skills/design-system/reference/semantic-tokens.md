# Semantic Tokens Reference

Semantic tokens map primitives to purpose. They are the layer that components consume. When theming changes (dark mode, rebranding), only the primitive references change — semantic names stay stable.

Sources: Material Design 3 (3-layer architecture), GitHub Primer (functional tokens), Adobe Spectrum (alias tokens), IBM Carbon (role-based tokens), Apple HIG (semantic colors).

## Color — Brand

| Token | Purpose | Light default | Dark default |
|---|---|---|---|
| `color-primary` | Main brand action, primary buttons, active states | `{brand-600}` | `{brand-400}` |
| `color-primary-hover` | Hover state for primary | `{brand-700}` | `{brand-300}` |
| `color-primary-subtle` | Light background accent | `{brand-50}` | `{brand-950}` |
| `color-secondary` | Supporting brand actions | `{secondary-600}` | `{secondary-400}` |
| `color-secondary-hover` | Hover state for secondary | `{secondary-700}` | `{secondary-300}` |
| `color-accent` | Highlights, badges, decorative | `{accent-500}` | `{accent-400}` |

## Color — Semantic (Status)

| Token | Purpose | Light default | Dark default |
|---|---|---|---|
| `color-success` | Positive confirmation | `{green-600}` | `{green-400}` |
| `color-success-subtle` | Success backgrounds | `{green-50}` | `{green-950}` |
| `color-warning` | Caution, attention needed | `{amber-600}` | `{amber-400}` |
| `color-warning-subtle` | Warning backgrounds | `{amber-50}` | `{amber-950}` |
| `color-danger` | Errors, destructive actions | `{red-600}` | `{red-400}` |
| `color-danger-subtle` | Error backgrounds | `{red-50}` | `{red-950}` |
| `color-info` | Informational, links, neutral actions | `{blue-600}` | `{blue-400}` |
| `color-info-subtle` | Info backgrounds | `{blue-50}` | `{blue-950}` |

## Color — Surface

Following IBM Carbon's layering model + Apple HIG's elevated surfaces pattern.

| Token | Purpose | Light default | Dark default |
|---|---|---|---|
| `color-background` | Page background (lowest layer) | `{white}` or `{gray-50}` | `{gray-950}` |
| `color-surface` | Cards, panels (layer 1) | `{white}` | `{gray-900}` |
| `color-surface-elevated` | Modals, popovers (layer 2) | `{white}` | `{gray-800}` |
| `color-surface-overlay` | Backdrop behind modals | `{black / 50%}` | `{black / 60%}` |
| `color-surface-subtle` | Zebra stripes, hover backgrounds | `{gray-50}` | `{gray-800}` |

**Dark mode rule (from Apple + Carbon):** elevated surfaces get LIGHTER, not darker. This conveys depth.

## Color — Text

Following GitHub Primer's hierarchical text pattern + Apple HIG's label hierarchy.

| Token | Purpose | Light default | Dark default | Contrast req |
|---|---|---|---|---|
| `color-text-primary` | Main content text | `{gray-900}` | `{gray-50}` | 4.5:1 on surface |
| `color-text-secondary` | Supporting text, descriptions | `{gray-600}` | `{gray-400}` | 4.5:1 on surface |
| `color-text-tertiary` | Placeholder, disabled hints | `{gray-400}` | `{gray-500}` | 3:1 on surface |
| `color-text-disabled` | Disabled controls | `{gray-300}` | `{gray-600}` | No minimum (intentionally low) |
| `color-text-inverse` | Text on filled buttons, dark surfaces | `{white}` | `{gray-950}` | 4.5:1 on primary |
| `color-text-link` | Hyperlinks | `{brand-600}` | `{brand-400}` | 4.5:1 on surface |
| `color-text-on-primary` | Text on primary color backgrounds | `{white}` | `{white}` | 4.5:1 on primary |
| `color-text-on-danger` | Text on danger backgrounds | `{white}` | `{white}` | 4.5:1 on danger |

## Color — Border

Following Adobe Spectrum's gray usage guide + GitHub Primer's border tokens.

| Token | Purpose | Light default | Dark default |
|---|---|---|---|
| `color-border-default` | Standard borders (cards, inputs) | `{gray-200}` | `{gray-700}` |
| `color-border-strong` | Emphasized borders, active inputs | `{gray-400}` | `{gray-500}` |
| `color-border-subtle` | Dividers, separators | `{gray-100}` | `{gray-800}` |
| `color-border-focus` | Focus rings (accessibility) | `{brand-500}` | `{brand-400}` |
| `color-border-danger` | Error state inputs | `{red-500}` | `{red-400}` |

---

## Typography — Roles

Each role maps a semantic purpose to a primitive combination of size + weight + line-height.

Based on Material Design 3's type roles, adapted for web.

| Token | Font size | Weight | Line-height | Use |
|---|---|---|---|---|
| `type-display` | `{text-5xl}` 48px | `{font-bold}` 700 | 1.0 | Hero headlines, landing page |
| `type-heading-1` | `{text-4xl}` 36px | `{font-bold}` 700 | 1.111 | Page title (h1) |
| `type-heading-2` | `{text-3xl}` 30px | `{font-semibold}` 600 | 1.2 | Section title (h2) |
| `type-heading-3` | `{text-2xl}` 24px | `{font-semibold}` 600 | 1.333 | Subsection title (h3) |
| `type-heading-4` | `{text-xl}` 20px | `{font-medium}` 500 | 1.4 | Card title, sub-subsection (h4) |
| `type-body` | `{text-base}` 16px | `{font-normal}` 400 | 1.5 | Default content text |
| `type-body-small` | `{text-sm}` 14px | `{font-normal}` 400 | 1.429 | Secondary content, table cells |
| `type-label` | `{text-sm}` 14px | `{font-medium}` 500 | 1.429 | Form labels, tab labels, buttons |
| `type-caption` | `{text-xs}` 12px | `{font-normal}` 400 | 1.333 | Helper text, timestamps, badges |
| `type-code` | `{text-sm}` 14px | `{font-normal}` 400 | 1.5 | Code blocks, technical data |

---

## Spacing — Semantic

| Token | Primitive | Use |
|---|---|---|
| `space-input-padding-x` | `{spacing-3}` 12px | Horizontal padding inside inputs/buttons |
| `space-input-padding-y` | `{spacing-2}` 8px | Vertical padding inside inputs/buttons |
| `space-input-gap` | `{spacing-2}` 8px | Gap between label and input |
| `space-component-gap` | `{spacing-4}` 16px | Gap between sibling components |
| `space-card-padding` | `{spacing-6}` 24px | Internal card padding |
| `space-section-gap` | `{spacing-8}` 32px | Gap between page sections |
| `space-page-padding` | `{spacing-4}` 16px mobile, `{spacing-6}` 24px desktop | Page edge padding |
| `space-stack-sm` | `{spacing-2}` 8px | Tight vertical stack (form fields) |
| `space-stack-md` | `{spacing-4}` 16px | Normal vertical stack |
| `space-stack-lg` | `{spacing-6}` 24px | Loose vertical stack (sections) |
| `space-inline-sm` | `{spacing-1}` 4px | Tight horizontal (icon + text) |
| `space-inline-md` | `{spacing-2}` 8px | Normal horizontal (buttons in a row) |
| `space-inline-lg` | `{spacing-4}` 16px | Loose horizontal (nav items) |

---

## Interactive States

These tokens modify base colors for different interaction states.

| Token | Purpose | Pattern |
|---|---|---|
| `state-hover` | Mouse hover | Base color + 1 shade darker (light) / lighter (dark) |
| `state-active` | Mouse down, pressed | Base color + 2 shades darker/lighter |
| `state-focus` | Keyboard focus | `{color-border-focus}` ring, 2px offset |
| `state-disabled` | Non-interactive | 40% opacity of base, no pointer events |
| `state-selected` | Selected/active item | `{color-primary-subtle}` background + `{color-primary}` indicator |

---

## Contrast Verification Checklist

Every text/surface combination MUST pass:

| Combination | Required ratio | Standard |
|---|---|---|
| `text-primary` on `surface` | 4.5:1 | WCAG AA normal text |
| `text-primary` on `background` | 4.5:1 | WCAG AA normal text |
| `text-secondary` on `surface` | 4.5:1 | WCAG AA normal text |
| `text-inverse` on `primary` | 4.5:1 | WCAG AA normal text |
| `heading-1` on `surface` | 3:1 | WCAG AA large text (>24px) |
| `text-tertiary` on `surface` | 3:1 | WCAG AA large text / UI components |
| Any icon on its background | 3:1 | WCAG AA UI components |
| Focus ring against surface | 3:1 | WCAG 2.2 focus appearance |

**Tool:** Use contrast checker (WebAIM or similar) to verify. Never assume — always calculate.
