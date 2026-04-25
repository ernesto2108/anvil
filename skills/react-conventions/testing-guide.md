# Guía de Testing en React

## Estrategia de Testing

| Capa | Herramienta | Alcance | Velocidad |
|---|---|---|---|
| Unit/Componente | **Vitest + React Testing Library** | Componentes, hooks, utilidades | Rápida |
| Mock de API | **MSW** (Mock Service Worker) | Interceptar fetch en unit y E2E | - |
| E2E | **Playwright** | 3-5 flujos críticos de usuario | Lenta |
| Accesibilidad | **axe-core + eslint-plugin-jsx-a11y** | a11y automatizado en tests y CI | Rápida |

**Pipeline CI**: los tests rápidos (Vitest/RTL/MSW) se ejecutan primero; los tests lentos (Playwright) solo si los rápidos pasan.

---

## Testing de Componentes con Vitest + RTL

### Filosofía

**Testear comportamiento del usuario, no detalles de implementación.**

- Consultar por rol, label, texto — no por test ID o nombre de clase
- Disparar eventos de usuario, no cambios de estado internos
- Afirmar lo que el usuario ve, no el estado interno

### Test de Componente Básico

```tsx
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect } from 'vitest'

describe('LoginForm', () => {
  it('shows error when submitted empty', async () => {
    render(<LoginForm />)

    await userEvent.click(screen.getByRole('button', { name: /submit/i }))

    expect(screen.getByRole('alert')).toHaveTextContent(/required/i)
  })

  it('calls onSubmit with form data', async () => {
    const onSubmit = vi.fn()
    render(<LoginForm onSubmit={onSubmit} />)

    await userEvent.type(screen.getByLabelText(/email/i), 'test@test.com')
    await userEvent.type(screen.getByLabelText(/password/i), 'password123')
    await userEvent.click(screen.getByRole('button', { name: /submit/i }))

    expect(onSubmit).toHaveBeenCalledWith({
      email: 'test@test.com',
      password: 'password123',
    })
  })
})
```

### Prioridad de Queries (RTL)

1. `getByRole` — roles accesibles (button, heading, textbox)
2. `getByLabelText` — elementos de formulario por label
3. `getByPlaceholderText` — elementos de formulario
4. `getByText` — elementos no interactivos
5. `getByTestId` — **último recurso** solamente

### Testing de Operaciones Async

```tsx
it('loads and displays users', async () => {
  render(<UserList />)

  // wait for loading to finish
  expect(await screen.findByText('John Doe')).toBeInTheDocument()
  expect(screen.queryByText('Loading...')).not.toBeInTheDocument()
})
```

### Testing de Custom Hooks

```tsx
import { renderHook, act } from '@testing-library/react'

describe('useCounter', () => {
  it('increments count', () => {
    const { result } = renderHook(() => useCounter())

    act(() => {
      result.current.increment()
    })

    expect(result.current.count).toBe(1)
  })
})
```

---

## Mock de API con MSW

Mock Service Worker intercepta a nivel de red — funciona en tests Y en desarrollo.

### Setup

```tsx
// src/testing/handlers.ts
import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/users', () => {
    return HttpResponse.json([
      { id: '1', name: 'John Doe', email: 'john@test.com' },
      { id: '2', name: 'Jane Doe', email: 'jane@test.com' },
    ])
  }),

  http.post('/api/login', async ({ request }) => {
    const body = await request.json()
    if (body.email === 'test@test.com') {
      return HttpResponse.json({ token: 'fake-token' })
    }
    return HttpResponse.json({ error: 'Invalid credentials' }, { status: 401 })
  }),
]

// src/testing/server.ts
import { setupServer } from 'msw/node'
import { handlers } from './handlers'

export const server = setupServer(...handlers)

// vitest.setup.ts
import { server } from './testing/server'

beforeAll(() => server.listen())
afterEach(() => server.resetHandlers())
afterAll(() => server.close())
```

### Sobreescribir Handlers por Test

```tsx
import { http, HttpResponse } from 'msw'
import { server } from '../testing/server'

it('shows error when API fails', async () => {
  server.use(
    http.get('/api/users', () => {
      return HttpResponse.json({ error: 'Server error' }, { status: 500 })
    })
  )

  render(<UserList />)
  expect(await screen.findByRole('alert')).toHaveTextContent(/error/i)
})
```

---

## Testing E2E con Playwright

3-5 flujos críticos de usuario. No duplicar lo que los tests unitarios ya cubren.

### Setup

```tsx
// e2e/login.spec.ts
import { test, expect } from '@playwright/test'

test.describe('Login flow', () => {
  test('user can log in and see dashboard', async ({ page }) => {
    await page.goto('/login')

    await page.getByLabel('Email').fill('test@test.com')
    await page.getByLabel('Password').fill('password123')
    await page.getByRole('button', { name: 'Sign in' }).click()

    await expect(page).toHaveURL('/dashboard')
    await expect(page.getByRole('heading', { name: 'Welcome' })).toBeVisible()
  })

  test('shows error on invalid credentials', async ({ page }) => {
    await page.goto('/login')

    await page.getByLabel('Email').fill('wrong@test.com')
    await page.getByLabel('Password').fill('wrong')
    await page.getByRole('button', { name: 'Sign in' }).click()

    await expect(page.getByRole('alert')).toContainText('Invalid credentials')
  })
})
```

### Reglas

- Testear solo journeys críticos de usuario (login, checkout, sign-up)
- Usar `getByRole`/`getByLabel` — misma filosofía de queries que RTL
- Ejecutar en CI después de que los tests unitarios pasen
- Usar el modo UI de Playwright para depuración

---

## Testing de Accesibilidad

### En Tests (axe-core)

```tsx
import { axe, toHaveNoViolations } from 'jest-axe'

expect.extend(toHaveNoViolations)

it('has no accessibility violations', async () => {
  const { container } = render(<LoginForm />)
  const results = await axe(container)
  expect(results).toHaveNoViolations()
})
```

### En CI (eslint-plugin-jsx-a11y)

```json
{
  "extends": ["plugin:jsx-a11y/recommended"]
}
```

### Qué Testear

- Todos los elementos interactivos son accesibles con teclado
- Las imágenes tienen texto `alt` significativo
- Los inputs de formulario tienen labels asociados
- El contraste de color cumple WCAG AA (4.5:1 para texto)
- Gestión del foco después de navegación/modales

---

## Organización de Archivos de Test

```
src/
  features/
    auth/
      components/
        LoginForm.tsx
        LoginForm.test.tsx     # co-located with component
      hooks/
        useAuth.ts
        useAuth.test.ts        # co-located with hook
  testing/
    handlers.ts                # MSW handlers
    server.ts                  # MSW server setup
    test-utils.tsx             # custom render with providers
    factories.ts               # test data factories
```

### Render Personalizado con Providers

```tsx
// testing/test-utils.tsx
import { render, RenderOptions } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

function AllProviders({ children }: { children: ReactNode }) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  )
}

function customRender(ui: ReactElement, options?: RenderOptions) {
  return render(ui, { wrapper: AllProviders, ...options })
}

export { customRender as render }
```

---

## Mocking de Servicios/Repositorios

Para mockear interfaces TypeScript de servicios, hooks, o repositorios, usar `jest-mock-extended` (ver `typescript-conventions/guides/testing/vitest.md` para setup y ejemplos completos). Nunca crear objetos mock manuales con `vi.fn()` para interfaces — pueden divergir de la interfaz real.

`vi.fn()` sigue siendo válido para callbacks de props (ej: `onSubmit`, `onClick`).

---

## Anti-Patrones

| Anti-Patrón | Corrección |
|---|---|
| Testear detalles de implementación (valor de `useState`) | Testear lo que el usuario ve |
| Consultar por clase CSS o test ID primero | Usar `getByRole`, `getByLabelText` |
| Mockear fetch/axios directamente | Usar MSW para mocking a nivel de red |
| Sin manejo async (falta `findBy`/`waitFor`) | Usar `findBy` para contenido async |
| Testear internos de librerías (caché de TanStack Query) | Testear el componente que lo consume |
| Tests de snapshot como estrategia principal | Usar solo para regresión visual, no para lógica |
| Mocks manuales (`{ save: vi.fn() }`) para interfaces | Usar `mock<Interface>()` de jest-mock-extended |
