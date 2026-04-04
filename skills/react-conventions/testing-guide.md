# React Testing Guide

## Testing Strategy

| Layer | Tool | Scope | Speed |
|---|---|---|---|
| Unit/Component | **Vitest + React Testing Library** | Components, hooks, utilities | Fast |
| API mocking | **MSW** (Mock Service Worker) | Intercept fetch in unit and E2E | - |
| E2E | **Playwright** | 3-5 critical user flows | Slow |
| Accessibility | **axe-core + eslint-plugin-jsx-a11y** | Automated a11y in tests and CI | Fast |

**CI pipeline**: fast tests (Vitest/RTL/MSW) run first; slow tests (Playwright) only if fast pass.

---

## Component Testing with Vitest + RTL

### Philosophy

**Test user behavior, not implementation details.**

- Query by role, label, text — not by test ID or class name
- Fire user events, not internal state changes
- Assert what the user sees, not internal state

### Basic Component Test

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

### Query Priority (RTL)

1. `getByRole` — accessible roles (button, heading, textbox)
2. `getByLabelText` — form elements by label
3. `getByPlaceholderText` — form elements
4. `getByText` — non-interactive elements
5. `getByTestId` — **last resort** only

### Testing Async Operations

```tsx
it('loads and displays users', async () => {
  render(<UserList />)

  // wait for loading to finish
  expect(await screen.findByText('John Doe')).toBeInTheDocument()
  expect(screen.queryByText('Loading...')).not.toBeInTheDocument()
})
```

### Testing Custom Hooks

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

## API Mocking with MSW

Mock Service Worker intercepts at the network level — works in tests AND development.

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

### Override Handlers per Test

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

## E2E Testing with Playwright

3-5 critical user flows. Don't duplicate what unit tests already cover.

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

### Rules

- Test critical user journeys only (login, checkout, sign-up)
- Use `getByRole`/`getByLabel` — same query philosophy as RTL
- Run in CI after unit tests pass
- Use Playwright's UI Mode for debugging

---

## Accessibility Testing

### In Tests (axe-core)

```tsx
import { axe, toHaveNoViolations } from 'jest-axe'

expect.extend(toHaveNoViolations)

it('has no accessibility violations', async () => {
  const { container } = render(<LoginForm />)
  const results = await axe(container)
  expect(results).toHaveNoViolations()
})
```

### In CI (eslint-plugin-jsx-a11y)

```json
{
  "extends": ["plugin:jsx-a11y/recommended"]
}
```

### What to Test

- All interactive elements are keyboard-accessible
- Images have meaningful `alt` text
- Form inputs have associated labels
- Color contrast meets WCAG AA (4.5:1 for text)
- Focus management after navigation/modals

---

## Test File Organization

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

### Custom Render with Providers

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

## Anti-Patterns

| Anti-Pattern | Fix |
|---|---|
| Testing implementation details (`useState` value) | Test what the user sees |
| Querying by CSS class or test ID first | Use `getByRole`, `getByLabelText` |
| Mocking fetch/axios directly | Use MSW for network-level mocking |
| No async handling (missing `findBy`/`waitFor`) | Use `findBy` for async content |
| Testing library internals (TanStack Query cache) | Test the component that consumes it |
| Snapshot tests as primary testing strategy | Use for visual regression only, not logic |
