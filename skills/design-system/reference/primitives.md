# Primitive Tokens Reference

Primitives are raw values with NO semantic meaning. They are the foundation layer — never consumed directly by components.

Sources: Material Design 3, Adobe Spectrum, GitHub Primer, Tailwind CSS, IBM Carbon, Apple HIG, W3C Design Tokens spec.

## Color Palette

### Structure

Each hue family uses an 11-step scale. The number indicates lightness (50 = lightest, 950 = darkest).

```
{hue}-50   →  Lightest (backgrounds, subtle fills)
{hue}-100  →
{hue}-200  →  Light (hover states, light borders)
{hue}-300  →
{hue}-400  →  Mid-light (secondary elements)
{hue}-500  →  Base (primary usage, icons)
{hue}-600  →  Mid-dark (primary interactive, buttons)
{hue}-700  →  Dark (hover on dark buttons, strong text)
{hue}-800  →
{hue}-900  →  Very dark (headings, high contrast)
{hue}-950  →  Darkest (near-black, dark mode surfaces)
```

### Required Hue Families

| Family | Purpose | Notes |
|---|---|---|
| **Brand primary** | Main brand color | User must provide or choose |
| **Brand secondary** | Supporting brand color | Optional, derive from primary if not provided |
| **Neutral / Gray** | Text, borders, backgrounds | Choose one gray family: pure, cool, warm, slate, zinc |
| **Red** | Errors, destructive actions | `danger` semantic maps here |
| **Amber/Yellow** | Warnings, caution | `warning` semantic maps here |
| **Green** | Success, confirmation | `success` semantic maps here |
| **Blue** | Information, links | `info` semantic maps here (unless brand is blue) |

### Default Gray Scale (Tailwind neutral)

```
gray-50:   #fafafa
gray-100:  #f5f5f5
gray-200:  #e5e5e5
gray-300:  #d4d4d4
gray-400:  #a3a3a3
gray-500:  #737373
gray-600:  #525252
gray-700:  #404040
gray-800:  #262626
gray-900:  #171717
gray-950:  #0a0a0a
```

### Default Status Colors (Tailwind)

```
red-500:    #ef4444    (danger)
red-600:    #dc2626    (danger-hover)
amber-500:  #f59e0b    (warning)
amber-600:  #d97706    (warning-hover)
green-500:  #22c55e    (success)
green-600:  #16a34a    (success-hover)
blue-500:   #3b82f6    (info)
blue-600:   #2563eb    (info-hover)
```

---

## Typography Scale

### Font Size Scale

Based on Tailwind defaults. Each size pairs with a recommended line-height.

| Token | Size | Line-height | Typical use |
|---|---|---|---|
| `text-xs` | 0.75rem (12px) | 1.333 (16px) | Captions, badges, helper text |
| `text-sm` | 0.875rem (14px) | 1.429 (20px) | Labels, secondary text, table cells |
| `text-base` | 1rem (16px) | 1.5 (24px) | Body text (base) |
| `text-lg` | 1.125rem (18px) | 1.556 (28px) | Large body, emphasis |
| `text-xl` | 1.25rem (20px) | 1.4 (28px) | Heading 4, section labels |
| `text-2xl` | 1.5rem (24px) | 1.333 (32px) | Heading 3 |
| `text-3xl` | 1.875rem (30px) | 1.2 (36px) | Heading 2 |
| `text-4xl` | 2.25rem (36px) | 1.111 (40px) | Heading 1 |
| `text-5xl` | 3rem (48px) | 1.0 (48px) | Display, hero |
| `text-6xl` | 3.75rem (60px) | 1.0 (60px) | Display large |

### Font Weight Scale

| Token | Value | Use |
|---|---|---|
| `font-light` | 300 | Decorative, display text |
| `font-normal` | 400 | Body text (default) |
| `font-medium` | 500 | Labels, emphasis |
| `font-semibold` | 600 | Headings, buttons |
| `font-bold` | 700 | Strong headings, CTAs |

### Font Family Defaults

| Token | Stack | When |
|---|---|---|
| `font-sans` | `-apple-system, BlinkMacSystemFont, 'Segoe UI', 'Noto Sans', Helvetica, Arial, sans-serif` | Default for UI |
| `font-mono` | `ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Consolas, monospace` | Code, technical |
| `font-serif` | `Georgia, Cambria, 'Times New Roman', Times, serif` | Editorial, decorative |

**Note:** Replace with project-specific fonts when brand requires it (e.g., Inter, IBM Plex, custom).

---

## Spacing Scale

Base unit: **4px** (0.25rem). Every spacing value is a multiple of 4px.

| Token | Value | px | Typical use |
|---|---|---|---|
| `spacing-0` | 0 | 0 | Reset |
| `spacing-0.5` | 0.125rem | 2px | Hairline gaps |
| `spacing-1` | 0.25rem | 4px | Tight internal padding |
| `spacing-1.5` | 0.375rem | 6px | Icon-to-text gap |
| `spacing-2` | 0.5rem | 8px | Compact padding, small gaps |
| `spacing-3` | 0.75rem | 12px | Input padding, list gaps |
| `spacing-4` | 1rem | 16px | Component padding (default) |
| `spacing-5` | 1.25rem | 20px | Medium gap |
| `spacing-6` | 1.5rem | 24px | Section gaps, card padding |
| `spacing-8` | 2rem | 32px | Large gaps |
| `spacing-10` | 2.5rem | 40px | Section separation |
| `spacing-12` | 3rem | 48px | Large section spacing |
| `spacing-16` | 4rem | 64px | Page-level spacing |
| `spacing-20` | 5rem | 80px | Hero spacing |
| `spacing-24` | 6rem | 96px | Major section breaks |

---

## Border Radius Scale

| Token | Value | Use |
|---|---|---|
| `radius-none` | 0 | Sharp corners |
| `radius-xs` | 0.125rem (2px) | Subtle rounding (badges) |
| `radius-sm` | 0.25rem (4px) | Inputs, small cards |
| `radius-md` | 0.375rem (6px) | Buttons, default rounding |
| `radius-lg` | 0.5rem (8px) | Cards, modals |
| `radius-xl` | 0.75rem (12px) | Large cards, containers |
| `radius-2xl` | 1rem (16px) | Feature cards, hero sections |
| `radius-3xl` | 1.5rem (24px) | Prominent panels |
| `radius-full` | 9999px | Pills, avatars, circular |

---

## Shadow / Elevation Scale

Dual-layer shadows (key + ambient) following Material Design 3 pattern.

| Token | Value | Use |
|---|---|---|
| `shadow-xs` | `0 1px 2px 0 rgb(0 0 0 / 0.05)` | Subtle lift (cards at rest) |
| `shadow-sm` | `0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)` | Buttons, small cards |
| `shadow-md` | `0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)` | Dropdowns, popovers |
| `shadow-lg` | `0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)` | Modals, dialogs |
| `shadow-xl` | `0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)` | Toast notifications, floating panels |

---

## Breakpoints

| Token | Value | Target |
|---|---|---|
| `breakpoint-sm` | 640px | Large phones (landscape) |
| `breakpoint-md` | 768px | Tablets (portrait) |
| `breakpoint-lg` | 1024px | Tablets (landscape), small laptops |
| `breakpoint-xl` | 1280px | Desktops |
| `breakpoint-2xl` | 1536px | Large desktops |

**Mobile-first approach:** base styles target mobile. Use `min-width` breakpoints to scale up.

---

## Z-Index Layers

| Token | Value | Use |
|---|---|---|
| `z-base` | 0 | Default content |
| `z-dropdown` | 10 | Dropdowns, autocomplete |
| `z-sticky` | 20 | Sticky headers, toolbars |
| `z-overlay` | 30 | Backdrop overlays |
| `z-modal` | 40 | Modal dialogs |
| `z-popover` | 50 | Popovers, tooltips |
| `z-toast` | 60 | Toast notifications |

---

## Duration / Timing Scale

| Token | Value | Use |
|---|---|---|
| `duration-instant` | 0ms | Immediate (no transition) |
| `duration-fast` | 150ms | Micro-interactions (hover, focus) |
| `duration-normal` | 300ms | Standard transitions (open/close, fade) |
| `duration-slow` | 500ms | Complex animations (page transitions, modals) |
| `ease-default` | `cubic-bezier(0.4, 0, 0.2, 1)` | General purpose (Material standard) |
| `ease-in` | `cubic-bezier(0.4, 0, 1, 1)` | Elements exiting |
| `ease-out` | `cubic-bezier(0, 0, 0.2, 1)` | Elements entering |
| `ease-in-out` | `cubic-bezier(0.4, 0, 0.2, 1)` | Elements moving |
