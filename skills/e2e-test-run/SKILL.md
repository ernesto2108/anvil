---
name: e2e-test-run
description: Convenciones para escribir y ejecutar pruebas E2E. Cubre Playwright (web + desktop) y Maestro (mobile). Usar cuando el tester necesite escribir tests de flujo completo, visual regression, o accesibilidad automatizada.
---

# E2E Test Conventions

Guía de convenciones para tests end-to-end. El tester carga este skill en PASO 0 cuando el handoff incluye tests E2E.

---

## Selección de herramienta

| Plataforma | Herramienta | Cuándo |
|---|---|---|
| Web (React, Astro, cualquier SPA/SSR) | **Playwright** | Siempre para web |
| Desktop (Electron, Tauri, Wails) | **Playwright** | Conecta al webview |
| Mobile (Flutter) | **Maestro** | YAML flows, usa Semantics |
| Mobile (React Native) | **Maestro** | YAML flows, usa testID |

---

## Playwright (Web + Desktop)

### Estructura de proyecto

```
tests/
  e2e/
    auth.spec.ts
    checkout.spec.ts
  fixtures/
    index.ts           # custom fixtures
  pages/
    login.page.ts      # Page Objects
playwright.config.ts
```

Nombrar tests `<feature>.spec.ts`. Page objects `<page>.page.ts`.

### Fixtures sobre POM crudo

Definir page objects como fixtures — los tests reciben instancias listas:

```ts
// tests/fixtures/index.ts
import { test as base } from '@playwright/test'
import { LoginPage } from '../pages/login.page'

export const test = base.extend<{ loginPage: LoginPage }>({
  loginPage: async ({ page }, use) => { await use(new LoginPage(page)) },
})
```

### Selectores — orden de prioridad

1. `page.getByRole('button', { name: /submit/i })` — accesibilidad
2. `page.getByLabel('Email')` — formularios
3. `page.getByTestId('cart-total')` — contrato explícito, sin rol semántico
4. CSS — **último recurso**

**Prohibido:** XPath, `.nth()` posicional, selectores por clase CSS, selectores acoplados al DOM.

### Assertions web-first (auto-retry)

```ts
await expect(page.getByRole('alert')).toBeVisible()
await expect(page.getByRole('heading')).toHaveText('Dashboard')
await expect(page).toHaveURL(/\/dashboard/)
```

**Nunca:** `expect(await el.isVisible()).toBe(true)` — evalúa una vez sin retry.

### Visual regression

```ts
await expect(page).toHaveScreenshot('dashboard.png', {
  maxDiffPixelRatio: 0.01,
  mask: [page.getByTestId('timestamp')],  // enmascarar contenido dinámico
})
```

- `npx playwright test --update-snapshots` para crear baselines
- Commitear snapshots al repo
- Usar solo en páginas con layout crítico (landing, design system)
- `threshold: 0.2` para tolerancia en texto anti-aliased

### Accesibilidad (axe-core)

```ts
import AxeBuilder from '@axe-core/playwright'

test('sin violaciones a11y en homepage', async ({ page }) => {
  await page.goto('/')
  const results = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa'])
    .analyze()
  expect(results.violations).toEqual([])
})
```

- Ejecutar axe en cada página principal
- `.exclude('.third-party-widget')` para elementos externos
- Incluir junto a tests E2E, no en suite separada

### API testing con Playwright

Para setup/teardown o tests de API puros:

```ts
test('crear usuario via API', async ({ request }) => {
  const res = await request.post('/api/users', { data: { name: 'Ada' } })
  expect(res.ok()).toBeTruthy()
  expect(await res.json()).toMatchObject({ name: 'Ada' })
})
```

API y browser comparten cookies — usar API para seed/login, luego browser para flujo.

### Desktop — Electron

```ts
const app = await _electron.launch({ args: ['main.js'] })
const window = await app.firstWindow()
await expect(window.getByRole('heading')).toHaveText('Welcome')
```

### Desktop — Tauri / Wails

Conectar Playwright al webview via CDP. Para Tauri: `tauri-driver` o `tauri-playwright`. Para Wails: apuntar Playwright al webview directamente. Mockear IPC/invoke para aislar frontend del backend.

### CI

```ts
// playwright.config.ts
export default defineConfig({
  retries: process.env.CI ? 2 : 0,
  use: { trace: 'on-first-retry' },
  reporter: process.env.CI ? 'blob' : 'html',
})
```

- Headless por defecto en CI
- Sharding cuando la suite > 10 min: `--shard=1/4`
- Subir `test-results/` y `playwright-report/` como artifacts
- Cachear `~/.cache/ms-playwright`
- Imagen Docker: `mcr.microsoft.com/playwright`

### Anti-patrones

| Prohibido | Correcto |
|---|---|
| `page.waitForTimeout(2000)` | Web-first assertions o `waitForSelector` |
| `page.locator('.btn-primary')` | `page.getByRole('button', { name: ... })` |
| Tests que dependen del orden | Cada test monta su propio estado |
| Estado mutable compartido | Fixtures para aislamiento |
| `page.evaluate` para assertions | Locator assertions |
| Retries > 2 en CI | Arreglar tests flaky |

---

## Maestro (Mobile)

### Estructura de proyecto

```
.maestro/
  login.yaml
  checkout.yaml
  subflows/
    login_steps.yaml
    onboarding_steps.yaml
```

Nombrar flows como `<feature>.yaml`. Subflows reutilizables en `subflows/`.

### Comandos principales

```yaml
appId: com.example.app
---
- launchApp
- tapOn: "Sign In"
- inputText: "user@test.com"
- tapOn: "Password"
- inputText: "secret123"
- tapOn: "Submit"
- assertVisible: "Welcome"
- scroll
- back
```

Comandos clave: `launchApp`, `tapOn`, `inputText`, `assertVisible`, `assertNotVisible`, `scroll`, `scrollUntilVisible`, `back`, `waitForAnimationToEnd`, `swipe`, `takeScreenshot`, `clearState`.

### Selectores — orden de prioridad

1. **Por id** (más estable): `tapOn: { id: "btn_submit" }`
2. **Por accessibility label**: `tapOn: { label: "Submit button" }`
3. **Por texto** (último recurso): `tapOn: "Submit"`

### Flows reutilizables

```yaml
# subflows/login_steps.yaml
appId: com.example.app
---
- tapOn: { id: "email_input" }
- inputText: ${USERNAME}
- tapOn: { id: "btn_submit" }

# flow principal
- runFlow:
    file: subflows/login_steps.yaml
    env:
      USERNAME: "admin@test.com"
- assertVisible: "Dashboard"
```

### Flutter

Usar `Semantics(identifier: ...)` (Flutter 3.19+):

```dart
Semantics(identifier: 'submit_btn', child: ElevatedButton(...))
```
```yaml
- tapOn: { id: "submit_btn" }
```

`identifier` es preferible a `Key` — los Keys no se exponen al árbol de accesibilidad que Maestro lee.

### React Native

Usar `testID` prop:

```tsx
<TouchableOpacity testID="btn_checkout">
```
```yaml
- tapOn: { id: "btn_checkout" }
```

En Android, agregar `accessibilityLabel` como fallback.

### Variables y condiciones

```yaml
env:
  USERNAME: "test@example.com"
---
- inputText: ${USERNAME}
- runFlow:
    when:
      visible: "Onboarding"
    file: subflows/skip_onboarding.yaml
- repeat:
    times: 3
    commands:
      - swipe:
          direction: LEFT
```

CLI: `maestro test -e USERNAME=admin flow.yaml`

### CI

```bash
maestro test .maestro/                          # todos los flows
maestro test .maestro/login.yaml                # flow específico
maestro test --format junit --output report.xml .maestro/
```

Maestro Cloud para CI sin emuladores: `maestro cloud --app-file app.apk .maestro/`

### Anti-patrones

| Prohibido | Correcto |
|---|---|
| `sleep: 5000` | `assertVisible` o `waitForAnimationToEnd` |
| Selectores por texto para i18n | Usar `id` o `label` |
| Flows monolíticos (100+ pasos) | Dividir en < 30 pasos, componer con `runFlow` |
| Sin `appId` en header | Siempre declarar `appId` |
| Datos hardcodeados | Variables `env` para emails, passwords, URLs |
| Sin `clearState` | Limpiar estado antes de flows que necesitan slate limpio |
