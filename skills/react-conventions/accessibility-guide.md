# Guía de Accesibilidad en React

Objetivo: cumplimiento **WCAG 2.2 AA** (obligatorio en la UE desde junio de 2025).

## Principios

1. **HTML semántico primero** — ARIA solo cuando la semántica nativa es insuficiente
2. **Accesible con teclado** — cada elemento interactivo alcanzable y operable mediante teclado
3. **Amigable para lectores de pantalla** — contenido significativo al leerlo en voz alta
4. **Contraste suficiente** — 4.5:1 para texto, 3:1 para texto grande y componentes de UI

---

## Herramientas

| Herramienta | Cuándo | Qué hace |
|---|---|---|
| `eslint-plugin-jsx-a11y` | Desarrollo | Verificaciones a11y en tiempo de lint |
| `axe-core` / `jest-axe` | Tests | Aserciones a11y automatizadas |
| Pestaña Accessibility de React DevTools | Desarrollo | Inspeccionar árbol accesible |
| Playwright + axe | CI | Verificaciones de accesibilidad E2E |
| **React Aria** (Adobe) | Componentes | Primitivos accesibles sin estilos |
| **Radix UI** | Componentes | Primitivos de componentes accesibles |
| **ARIAKit** | Componentes | Primitivos de componentes accesibles |

---

## HTML Semántico

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

### Elementos Semánticos Comunes

| Necesidad | Elemento | En lugar de |
|---|---|---|
| Acción clickeable | `<button>` | `<div onClick>` |
| Enlace de navegación | `<a href>` | `<span onClick>` |
| Sección de página | `<section>`, `<article>`, `<nav>`, `<aside>` | `<div>` |
| Encabezado de página | `<h1>`–`<h6>` (en orden) | `<div className="title">` |
| Lista de ítems | `<ul>`/`<ol>` + `<li>` | `<div>`s anidados |
| Input de formulario | `<input>` + `<label>` | `<input placeholder="Email">` |
| Datos tabulares | `<table>` + `<thead>` + `<tbody>` | Grid de `<div>`s |

---

## Patrones ARIA

### Regiones Live (Contenido Dinámico)

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

## Navegación con Teclado

### Gestión del Foco

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

### Focus Trap (Modales)

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

## Imágenes

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

### Reglas

- Siempre incluir `alt` en etiquetas `<img>`
- Evitar palabras genéricas: "image", "photo", "icon", "picture"
- Imágenes decorativas: `alt=""`
- Imágenes complejas: `aria-describedby` apuntando a descripción detallada

---

## Formularios

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

## Color y Contraste

- **Contraste de texto**: ratio mínimo 4.5:1 (AA)
- **Texto grande** (18pt+): ratio mínimo 3:1
- **Componentes de UI**: ratio mínimo 3:1 contra colores adyacentes
- **Nunca usar solo color** para transmitir información — agregar íconos, patrones o texto

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

## Testing de Accesibilidad

### En Tests de Componentes

```tsx
import { axe, toHaveNoViolations } from 'jest-axe'

expect.extend(toHaveNoViolations)

it('LoginForm has no a11y violations', async () => {
  const { container } = render(<LoginForm />)
  expect(await axe(container)).toHaveNoViolations()
})
```

### En Tests E2E

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

## Anti-Patrones

| Anti-Patrón | Corrección |
|---|---|
| `<div onClick>` para botones | Usar `<button>` |
| `alt` faltante en imágenes | Agregar `alt` descriptivo o `alt=""` para decorativas |
| `tabIndex > 0` | Usar solo `tabIndex={0}` o `tabIndex={-1}` |
| Atributo `accessKey` | Eliminar — entra en conflicto con tecnología asistiva |
| Indicación de error solo por color | Agregar ícono + texto junto al color |
| Sin gestión de foco después de navegación | Enfocar encabezado al cambiar de página |
| Medios con autoplay | Agregar controles, sin autoplay, o `prefers-reduced-motion` |
