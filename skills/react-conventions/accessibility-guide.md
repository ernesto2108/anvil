# React Accessibility Guide

Target: **WCAG 2.2 AA** compliance (mandatory in EU since June 2025).

## Principles

1. **Semantic HTML first** — ARIA only when native semantics are insufficient
2. **Keyboard accessible** — every interactive element reachable and operable via keyboard
3. **Screen reader friendly** — content meaningful when read aloud
4. **Sufficient contrast** — 4.5:1 for text, 3:1 for large text and UI components

---

## Tooling

| Tool | When | What |
|---|---|---|
| `eslint-plugin-jsx-a11y` | Development | Lint-time a11y checks |
| `axe-core` / `jest-axe` | Tests | Automated a11y assertions |
| React DevTools Accessibility tab | Development | Inspect accessible tree |
| Playwright + axe | CI | E2E accessibility checks |
| **React Aria** (Adobe) | Components | Headless accessible primitives |
| **Radix UI** | Components | Accessible component primitives |
| **ARIAKit** | Components | Accessible component primitives |

---

## Semantic HTML

```tsx
// bad: div soup
<div onClick={handleClick}>Click me</div>
<div className="heading">Title</div>
<div className="list">
  <div>Item 1</div>
</div>

// good: semantic elements
<button onClick={handleClick}>Click me</button>
<h2>Title</h2>
<ul>
  <li>Item 1</li>
</ul>
```

### Common Semantic Elements

| Need | Element | Not |
|---|---|---|
| Clickable action | `<button>` | `<div onClick>` |
| Navigation link | `<a href>` | `<span onClick>` |
| Page section | `<section>`, `<article>`, `<nav>`, `<aside>` | `<div>` |
| Page heading | `<h1>`–`<h6>` (in order) | `<div className="title">` |
| List of items | `<ul>`/`<ol>` + `<li>` | Nested `<div>`s |
| Form input | `<input>` + `<label>` | `<input placeholder="Email">` |
| Table data | `<table>` + `<thead>` + `<tbody>` | Grid of `<div>`s |

---

## ARIA Patterns

### Live Regions (Dynamic Content)

```tsx
// announce content changes to screen readers
<div role="status" aria-live="polite">
  {notification && <p>{notification}</p>}
</div>

// for urgent updates (errors)
<div role="alert" aria-live="assertive">
  {error && <p>{error}</p>}
</div>
```

### Labels

```tsx
// visible label (preferred)
<label htmlFor="email">Email</label>
<input id="email" type="email" />

// hidden label (icon-only buttons)
<button aria-label="Close dialog">
  <CloseIcon />
</button>

// described by (additional context)
<input aria-describedby="password-help" type="password" />
<p id="password-help">Must be at least 8 characters</p>
```

### Dialog/Modal

```tsx
<dialog
  open={isOpen}
  aria-labelledby="dialog-title"
  aria-describedby="dialog-desc"
>
  <h2 id="dialog-title">Confirm Deletion</h2>
  <p id="dialog-desc">This action cannot be undone.</p>
  <button onClick={onConfirm}>Delete</button>
  <button onClick={onCancel} autoFocus>Cancel</button>
</dialog>
```

---

## Keyboard Navigation

### Focus Management

```tsx
// manage focus after client-side navigation
function Page() {
  const headingRef = useRef<HTMLHeadingElement>(null)

  useEffect(() => {
    headingRef.current?.focus()
  }, [])

  return <h1 ref={headingRef} tabIndex={-1}>Dashboard</h1>
}
```

### Focus Trap (Modals)

```tsx
// use a library for focus trapping
import { FocusTrap } from 'focus-trap-react'

function Modal({ isOpen, children }: ModalProps) {
  if (!isOpen) return null
  return (
    <FocusTrap>
      <div role="dialog" aria-modal="true">
        {children}
      </div>
    </FocusTrap>
  )
}
```

### Skip Links

```tsx
<a href="#main-content" className="skip-link">
  Skip to main content
</a>
{/* ... navigation ... */}
<main id="main-content">
  {children}
</main>
```

---

## Images

```tsx
// informative image — describe the content
<img src="/chart.png" alt="Revenue grew 23% in Q3 2025" />

// decorative image — empty alt
<img src="/divider.png" alt="" />

// complex image — link to full description
<figure>
  <img src="/architecture.png" alt="System architecture diagram" aria-describedby="arch-desc" />
  <figcaption id="arch-desc">
    Three-tier architecture with React frontend, Go API, and PostgreSQL database.
  </figcaption>
</figure>
```

### Rules

- Always include `alt` on `<img>` tags
- Avoid generic words: "image", "photo", "icon", "picture"
- Decorative images: `alt=""`
- Complex images: `aria-describedby` pointing to detailed description

---

## Forms

```tsx
// good: every input has a label
<div>
  <label htmlFor="name">Full name</label>
  <input id="name" type="text" required aria-required="true" />
</div>

// good: error messages linked to input
<div>
  <label htmlFor="email">Email</label>
  <input
    id="email"
    type="email"
    aria-invalid={!!errors.email}
    aria-describedby={errors.email ? 'email-error' : undefined}
  />
  {errors.email && (
    <p id="email-error" role="alert">{errors.email.message}</p>
  )}
</div>

// good: group related fields
<fieldset>
  <legend>Shipping Address</legend>
  {/* address fields */}
</fieldset>
```

---

## Color and Contrast

- **Text contrast**: minimum 4.5:1 ratio (AA)
- **Large text** (18pt+): minimum 3:1 ratio
- **UI components**: minimum 3:1 ratio against adjacent colors
- **Never use color alone** to convey information — add icons, patterns, or text

```tsx
// bad: only color indicates error
<input className={hasError ? 'border-red' : 'border-gray'} />

// good: color + icon + text
<input
  className={hasError ? 'border-red' : 'border-gray'}
  aria-invalid={hasError}
  aria-describedby={hasError ? 'error-msg' : undefined}
/>
{hasError && (
  <p id="error-msg" role="alert">
    <ErrorIcon /> This field is required
  </p>
)}
```

---

## Testing Accessibility

### In Component Tests

```tsx
import { axe, toHaveNoViolations } from 'jest-axe'

expect.extend(toHaveNoViolations)

it('LoginForm has no a11y violations', async () => {
  const { container } = render(<LoginForm />)
  expect(await axe(container)).toHaveNoViolations()
})
```

### In E2E Tests

```tsx
import { test, expect } from '@playwright/test'
import AxeBuilder from '@axe-core/playwright'

test('home page is accessible', async ({ page }) => {
  await page.goto('/')
  const results = await new AxeBuilder({ page }).analyze()
  expect(results.violations).toEqual([])
})
```

---

## Anti-Patterns

| Anti-Pattern | Fix |
|---|---|
| `<div onClick>` for buttons | Use `<button>` |
| Missing `alt` on images | Add descriptive `alt` or `alt=""` for decorative |
| `tabIndex > 0` | Use `tabIndex={0}` or `tabIndex={-1}` only |
| `accessKey` attribute | Remove — conflicts with assistive technology |
| Color-only error indication | Add icon + text alongside color |
| No focus management after navigation | Focus heading on page change |
| Autoplaying media | Add controls, no autoplay, or `prefers-reduced-motion` |
