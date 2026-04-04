# Output Template — design-system.md

Use this exact structure when producing `<docs>/01-project/design-system.md`.

```markdown
# Design System — <Project Name>

> Last updated: <date>
> Platform: <web | mobile | both>
> Framework: <Tailwind | Material | custom | etc.>

## Foundations

### Color Palette (Primitives)

#### Brand

| Token | Value | Swatch |
|---|---|---|
| `brand-50` | #... | |
| `brand-100` | #... | |
| ... | | |
| `brand-950` | #... | |

#### Neutral

| Token | Value | Swatch |
|---|---|---|
| `gray-50` | #... | |
| ... | | |
| `gray-950` | #... | |

#### Status

| Token | Value | Use |
|---|---|---|
| `red-500` | #... | Danger base |
| `red-600` | #... | Danger hover |
| `amber-500` | #... | Warning base |
| `amber-600` | #... | Warning hover |
| `green-500` | #... | Success base |
| `green-600` | #... | Success hover |
| `blue-500` | #... | Info base |
| `blue-600` | #... | Info hover |

### Typography

**Font family:** <primary font>, <fallback stack>
**Mono:** <mono font>, <fallback stack>

| Role | Size | Weight | Line-height | Use |
|---|---|---|---|---|
| Display | 48px | 700 | 1.0 | Hero, landing |
| Heading 1 | 36px | 700 | 1.111 | Page title |
| Heading 2 | 30px | 600 | 1.2 | Section title |
| Heading 3 | 24px | 600 | 1.333 | Subsection |
| Heading 4 | 20px | 500 | 1.4 | Card title |
| Body | 16px | 400 | 1.5 | Default text |
| Body small | 14px | 400 | 1.429 | Secondary text |
| Label | 14px | 500 | 1.429 | Form labels, buttons |
| Caption | 12px | 400 | 1.333 | Helper text |
| Code | 14px | 400 | 1.5 | Code blocks |

### Spacing

Base unit: 4px (0.25rem)

| Token | Value | Use |
|---|---|---|
| `spacing-1` | 4px | Tight gaps |
| `spacing-2` | 8px | Compact padding |
| `spacing-3` | 12px | Input padding |
| `spacing-4` | 16px | Component padding |
| `spacing-6` | 24px | Card padding |
| `spacing-8` | 32px | Section gaps |
| `spacing-12` | 48px | Large sections |
| `spacing-16` | 64px | Page spacing |

### Border Radius

| Token | Value | Use |
|---|---|---|
| `radius-sm` | 4px | Inputs, small elements |
| `radius-md` | 6px | Buttons (default) |
| `radius-lg` | 8px | Cards |
| `radius-xl` | 12px | Large cards, containers |
| `radius-full` | 9999px | Pills, avatars |

### Shadows

| Token | Value | Use |
|---|---|---|
| `shadow-xs` | `0 1px 2px 0 rgb(0 0 0 / 0.05)` | Subtle lift |
| `shadow-sm` | `0 1px 3px ...` | Buttons, cards |
| `shadow-md` | `0 4px 6px ...` | Dropdowns |
| `shadow-lg` | `0 10px 15px ...` | Modals |
| `shadow-xl` | `0 20px 25px ...` | Toasts, floating |

### Breakpoints

| Token | Value | Target |
|---|---|---|
| `sm` | 640px | Phone landscape |
| `md` | 768px | Tablet portrait |
| `lg` | 1024px | Tablet landscape |
| `xl` | 1280px | Desktop |
| `2xl` | 1536px | Large desktop |

### Z-Index

| Token | Value | Use |
|---|---|---|
| `z-dropdown` | 10 | Dropdowns |
| `z-sticky` | 20 | Sticky headers |
| `z-overlay` | 30 | Backdrops |
| `z-modal` | 40 | Modals |
| `z-popover` | 50 | Tooltips |
| `z-toast` | 60 | Notifications |

---

## Semantic Tokens

### Color Roles

#### Brand

| Token | Light | Dark | Use |
|---|---|---|---|
| `color-primary` | `{brand-600}` | `{brand-400}` | Primary actions |
| `color-primary-hover` | `{brand-700}` | `{brand-300}` | Primary hover |
| `color-primary-subtle` | `{brand-50}` | `{brand-950}` | Light accents |
| `color-secondary` | ... | ... | Secondary actions |
| `color-accent` | ... | ... | Decorative |

#### Status

| Token | Light | Dark | Use |
|---|---|---|---|
| `color-success` | `{green-600}` | `{green-400}` | Positive |
| `color-warning` | `{amber-600}` | `{amber-400}` | Caution |
| `color-danger` | `{red-600}` | `{red-400}` | Error/destructive |
| `color-info` | `{blue-600}` | `{blue-400}` | Informational |

#### Surfaces

| Token | Light | Dark | Use |
|---|---|---|---|
| `color-background` | `{gray-50}` | `{gray-950}` | Page bg |
| `color-surface` | `{white}` | `{gray-900}` | Cards |
| `color-surface-elevated` | `{white}` | `{gray-800}` | Modals |

#### Text

| Token | Light | Dark | Contrast |
|---|---|---|---|
| `color-text-primary` | `{gray-900}` | `{gray-50}` | 4.5:1 |
| `color-text-secondary` | `{gray-600}` | `{gray-400}` | 4.5:1 |
| `color-text-disabled` | `{gray-300}` | `{gray-600}` | — |
| `color-text-inverse` | `{white}` | `{gray-950}` | 4.5:1 |

#### Borders

| Token | Light | Dark | Use |
|---|---|---|---|
| `color-border-default` | `{gray-200}` | `{gray-700}` | Cards, inputs |
| `color-border-strong` | `{gray-400}` | `{gray-500}` | Active inputs |
| `color-border-subtle` | `{gray-100}` | `{gray-800}` | Dividers |
| `color-border-focus` | `{brand-500}` | `{brand-400}` | Focus rings |

### Spacing Roles

| Token | Value | Use |
|---|---|---|
| `space-input-padding-x` | `{spacing-3}` | Input horizontal |
| `space-input-padding-y` | `{spacing-2}` | Input vertical |
| `space-component-gap` | `{spacing-4}` | Between components |
| `space-card-padding` | `{spacing-6}` | Inside cards |
| `space-section-gap` | `{spacing-8}` | Between sections |
| `space-page-padding` | `{spacing-4}` mobile / `{spacing-6}` desktop | Page edges |

---

## Accessibility

### Contrast Verification

| Combination | Ratio | Pass? |
|---|---|---|
| text-primary on surface | X:1 | ✓/✗ |
| text-secondary on surface | X:1 | ✓/✗ |
| text-inverse on primary | X:1 | ✓/✗ |
| heading on surface | X:1 | ✓/✗ |
| text-primary on background | X:1 | ✓/✗ |

### Focus States
- Focus ring: 2px solid `{color-border-focus}`, 2px offset
- All interactive elements must have visible focus indicator

### Touch Targets
- Minimum: 44x44px (mobile), 24x24px (desktop with pointer)

---

## Assumptions & Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Base font size | 16px | Browser default, accessible |
| Spacing base | 4px | Industry standard (Tailwind, Material) |
| Gray family | neutral | <reason> |
| Primary color | <value> | <brand / user preference> |
| Dark mode | yes/no/later | <reason> |

## Open Questions

- [ ] <Pending decisions>
```
