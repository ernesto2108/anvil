---
name: react-conventions
description: React/TypeScript frontend conventions and coding standards. Use when writing React components, reviewing frontend code, or user mentions "React patterns", "component structure", "hooks best practices", "accessibility", "state management", or working with .tsx/.jsx files.
---

# React Conventions

## Philosophy

- **UI is a function of state** — components render predictably from props and state, nothing else
- **Composition over configuration** — build complex UIs from simple, focused pieces
- **Accessibility is not optional** — if it's not keyboard-navigable and screen-readable, it's not done
- **Server-first** — default to Server Components; use Client Components only when needed

## Stack

- React 19+ with TypeScript (strict mode)
- Next.js App Router preferred for new projects
- Vitest + React Testing Library for testing

## Coding Rules

- Functional components only — no class components
- Composition over inheritance — pass children, don't extend
- Small, reusable components — one responsibility per component
- Accessibility first (ARIA attributes, semantic HTML, keyboard navigation)
- Predictable state (avoid derived state, single source of truth)
- No business logic in UI components — extract to custom hooks
- No form logic (types, validation, state) in components — extract to `useXxxForm` hook
- Before creating a helper/util, grep for it — if it exists in another feature, extract to `shared/utils/`
- One error channel per action — toast for API errors, inline for field validation. Never both
- Named exports preferred over default exports (except pages/layouts in Next.js)

## TypeScript Rules

- `strict: true` with `strictNullChecks`, `noImplicitAny`, `noUncheckedIndexedAccess`
- **`unknown` over `any`** — always. Use type guards for narrowing
- **Discriminated unions** for mutually exclusive states (loading/error/success)
- **Exhaustive checking** with `never` in switch statements
- **Zod** for runtime validation with inferred types (`z.infer<typeof schema>`)
- Props defined as `interface`, not `type`
- `as const` for tuple returns from hooks

## Architecture Rules

1. **Feature-based folder structure** — `src/features/{name}/{api,components,hooks,stores,types}`
2. **Unidirectional imports** — `shared → features → app`. Never import across features
3. **The hook IS the container** — custom hooks encapsulate all logic, components are pure UI
4. **Centralized API client** — never fetch directly in components
5. **Error boundaries at route level** — use `react-error-boundary`
6. **Lazy loading / code splitting** for routes (`React.lazy` + `<Suspense>`)
7. **Server Components by default** (Next.js) — `'use client'` only for state/events/browser APIs

## State Management Rules

| State Type | Tool |
|---|---|
| Server/remote data | **TanStack Query** (or SWR) |
| URL state | **nuqs** |
| Client shared state | **Zustand** |
| Form state | **React Hook Form + Zod** |
| Low-velocity global (theme, auth) | **React Context** |

Escalation: Context → Zustand → Redux. See `state-management-guide.md` for full patterns.

## Naming Conventions (Airbnb Standard)

- **Files**: PascalCase for components (`ReservationCard.tsx`), camelCase for hooks/utils
- **One component per file** (multiple stateless components allowed)
- **Props**: camelCase names. Omit boolean `true` values (`<Foo hidden />`)
- **Key prop**: always use stable IDs, never array indexes

## Icon Rules

- **Use `lucide-react`** — never inline SVG icons in components. Lucide provides tree-shakable, typed, configurable icon components
- **Import individually** — `import { User, Moon } from 'lucide-react'` (tree-shaking only bundles imported icons)
- **Size via prop** — `<User size={18} />` not `className="w-[18px] h-[18px]"`
- **Color via `currentColor`** — icons inherit parent text color. Use Tailwind `text-*` on the parent, not `color` prop
- **Never import all** — `import * as icons from 'lucide-react'` kills tree-shaking
- **Type for props** — use `LucideIcon` type when accepting icon components as props: `icon: LucideIcon`
- **Custom icons** — if lucide doesn't have what you need, create a component in `shared/components/icons/` following the same `size` + `currentColor` pattern

```tsx
// GOOD — tree-shakable, typed, configurable
import { User, Settings, LogOut } from 'lucide-react'
<User size={18} />

// BAD — inline SVG, not configurable, duplicated
function UserIcon() {
  return <svg width="18" height="18" ...>...</svg>
}
```

## Tailwind v4 Rules

Tailwind CSS v4 changed how CSS custom properties are referenced in utility classes.

- **CSS variables use parentheses, not brackets:**
  - CORRECT (v4): `bg-(--bg-surface)`, `text-(--text-primary)`, `border-(--border-default)`
  - WRONG (v3): `bg-[var(--bg-surface)]`, `text-[var(--text-primary)]`
- **Prefer standard Tailwind classes** over arbitrary values when an equivalent exists:
  - `size-4` not `w-[16px] h-[16px]`
  - `translate-x-4` not `translate-x-[16px]`
  - `gap-3` not `gap-[12px]`
  - `h-5` not `h-[20px]`
- **Arbitrary values are acceptable** for design-specific measurements that don't have Tailwind equivalents: `max-w-[400px]`, `w-[560px]`, `text-[42px]`

## Input Validation Patterns

Browsers do NOT enforce input restrictions for `type="tel"` or `type="number"` consistently. Always filter in `onChange`.

| Input type | `inputMode` | Filter in onChange | Validation |
|-----------|-------------|-------------------|------------|
| Phone | `tel` | Strip chars not in `[0-9+\-\s()]` | 10-15 digits |
| Email | `email` | None (validate on blur) | RFC 5322 regex |
| Code/OTP | `numeric` | Only digits, auto-advance focus | Exact length match |
| Currency | `decimal` | Strip chars not in `[0-9.]` | Positive number, max 2 decimals |
| Password | — | None | Min length + complexity rules |

```tsx
// Phone input with filtering
function handlePhoneChange(value: string) {
  const filtered = value.replace(/[^0-9+\-\s()]/g, '')
  setPhone(filtered)
}
<Input type="tel" inputMode="tel" value={phone} onChange={e => handlePhoneChange(e.target.value)} />
```

## Pre-Implementation Checklist

- [ ] Similar component/hook doesn't already exist
- [ ] Component has a single responsibility
- [ ] State categorized correctly (server vs client vs URL vs form)
- [ ] No business logic in the component body
- [ ] Accessible: semantic HTML, ARIA labels, keyboard handlers
- [ ] TypeScript types are explicit (no `any`)
- [ ] Server vs Client Component decision is intentional

## Anti-Pattern Detection

See `anti-patterns.md` for the full detection reference with severity levels.

**Passive detection:** When reviewing React code, automatically scan for `error` and `warning` patterns. Report as `[file:line] [severity] [category] anti-pattern-name`.

**Active detection:** When user asks to "improve", "refactor", "optimize" — also report `suggestion` level and propose fixes.

Red flags that should always stop work:
- `useEffect` with missing/wrong deps → stale-closure (error)
- `setInterval`/`setTimeout` without cleanup → memory-leak (error)
- `addEventListener` without removal → event-leak (error)
- Direct DOM mutation → dom-bypass (error)
- Circular imports → circular-deps (error)

## Support Files

- `patterns-guide.md` — React patterns (custom hooks, compound components, facade hooks, state machine, control props, adapter, strategy, observer, decorator, factory)
- `state-management-guide.md` — State management (TanStack Query, Zustand, Context, React Hook Form + Zod, Redux)
- `testing-guide.md` — Testing strategy (Vitest + RTL, MSW, Playwright, axe-core)
- `performance-guide.md` — Performance (React Compiler, code splitting, memoization rules, Netflix/Spotify patterns)
- `accessibility-guide.md` — Accessibility (WCAG 2.2 AA, semantic HTML, ARIA, keyboard, focus management, testing)
- `anti-patterns.md` — Anti-pattern detection table with severity levels and fix mapping
