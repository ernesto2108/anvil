---
name: e2e-test-run
description: Convenciones para escribir y ejecutar pruebas E2E. Cubre Playwright (web + desktop), Maestro (mobile cross-platform: Flutter, React Native e iOS nativo) y XCUITest (iOS nativo). Usar cuando el tester necesite escribir tests de flujo completo, visual regression, o accesibilidad automatizada.
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->


# E2E Test Conventions

Guía de convenciones para tests end-to-end. El tester carga este skill en PASO 0 cuando el handoff incluye tests E2E.

---

## Alcance

Esta guía cubre **suites E2E permanentes y versionadas** — tests que viven en el repo, se commitean junto al código y corren en CI de forma recurrente (Playwright, Maestro, XCUITest).

### Lo que esta skill NO cubre

- **Smoke desechable de flujos HTTP** — colecciones Postman portables generadas pre-handoff para validar un flujo encadenado a mano, sin versionarse como suite ni correr en CI → skill `test-api`.
- Cualquier artefacto de prueba cuyo ciclo de vida termine al validar el cambio, en vez de mantenerse como regresión permanente.

Regla de frontera: si el artefacto se commitea y corre en CI de forma recurrente → aplica esta skill. Si es efímero, portable y sirve para verificar un flujo una vez → no aplica.

---

## Selección de herramienta

| Plataforma | Herramienta | Cuándo |
|---|---|---|
| Web (React, Astro, cualquier SPA/SSR) | **Playwright** | Siempre para web |
| Desktop (Electron, Tauri, Wails) | **Playwright** | Conecta al webview |
| Mobile (Flutter) | **Maestro** | YAML flows, usa Semantics |
| Mobile (React Native) | **Maestro** | YAML flows, usa testID |
| Mobile (iOS nativo) | **Maestro** o **XCUITest** | Maestro para flujos declarativos cross-platform; XCUITest para integración profunda con Xcode/CI de Apple y APIs del sistema (ver abajo) |

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

Maestro es **cross-platform**: soporta Flutter, React Native e **iOS nativo** (Swift/SwiftUI) con los mismos flows YAML. Para iOS nativo, los selectores por `id` mapean al `accessibilityIdentifier` de las views. Elegir Maestro cuando se quieren flujos declarativos legibles y reutilizables entre plataformas; elegir XCUITest (abajo) cuando se necesita integración profunda con el ecosistema de Apple.

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

### iOS nativo (SwiftUI)

Exponer un identificador estable con `.accessibilityIdentifier(...)`:

```swift
Button("Checkout") { ... }
    .accessibilityIdentifier("btn_checkout")
```
```yaml
- tapOn: { id: "btn_checkout" }
```

Preferir `accessibilityIdentifier` a apoyarse en el texto visible (que cambia con i18n). No confundir con `.accessibilityLabel`, que es para VoiceOver — usar `identifier` para automatización.

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

---

## XCUITest (iOS nativo)

Framework de UI testing de Apple, parte de XCTest. Elegirlo sobre Maestro cuando se necesita integración profunda con Xcode/CI de Apple, acceso a APIs del sistema (permisos, notificaciones, `springboard`) o auditoría de accesibilidad nativa. Vive en el target de UI tests, en subclases de `XCTestCase`.

### Estructura y launch

```swift
final class CheckoutUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    @MainActor
    func testCheckoutFlow() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-uiTesting", "-resetState"]   // estado de test
        app.launchEnvironment = ["API_BASE": "http://localhost:8080"]
        app.launch()

        app.buttons["btn_checkout"].tap()
        let confirmation = app.staticTexts["order_confirmation"]
        XCTAssertTrue(confirmation.waitForExistence(timeout: 5))
    }
}
```

Usar los `launchArguments`/`launchEnvironment` para poner la app en un estado determinista (resetear datos, apuntar a un backend local, desactivar animaciones). La app lee esos flags en su entry point.

### Queries y expectations

- Localizar por tipo + `accessibilityIdentifier`: `app.buttons["btn_checkout"]`, `app.textFields["email_input"]`, `app.staticTexts["title"]`.
- **Preferir `accessibilityIdentifier`** sobre texto visible (estable ante i18n).
- Esperar con `waitForExistence(timeout:)` — nunca `sleep`. Para condiciones compuestas, usar `XCTNSPredicateExpectation` + `wait(for:timeout:)`.
- Acciones: `.tap()`, `.typeText(...)`, `.swipeUp()`, `.press(forDuration:)`.

### Auditoría de accesibilidad automatizada

```swift
@MainActor
func testHomeAccessibility() throws {
    let app = XCUIApplication()
    app.launch()
    try app.performAccessibilityAudit()   // falla ante issues de a11y detectables
}
```

`performAccessibilityAudit()` (Xcode 15+) es el equivalente nativo de axe-core para web: detecta contraste insuficiente, elementos sin label, hit targets pequeños, texto que no soporta Dynamic Type. Filtrar tipos con el parámetro `for:` cuando haga falta acotar.

### Ejecución

```bash
# Listar schemes disponibles primero
xcodebuild -list

# Correr el scheme de UI tests en un simulador
xcodebuild test \
  -scheme <UITestsScheme> \
  -destination 'platform=iOS Simulator,name=iPhone 16'
```

En CI, fijar el `-destination` (device + OS) para determinismo; subir el `.xcresult` como artifact.

### Cuándo XCUITest vs Maestro

| Elegir **XCUITest** | Elegir **Maestro** |
|---|---|
| Integración profunda con Xcode y CI de Apple | Flujos declarativos legibles en YAML |
| Acceso a APIs del sistema (permisos, notificaciones, springboard) | Cross-platform (mismo flow para iOS, Android, Flutter, RN) |
| Auditoría de accesibilidad nativa (`performAccessibilityAudit`) | Iteración rápida sin recompilar el target de tests |
| El equipo ya vive en el ecosistema XCTest | Menor curva y menos código |

### Anti-patrones

| Prohibido | Correcto |
|---|---|
| `sleep(2)` / `Thread.sleep` | `waitForExistence(timeout:)` o predicate expectations |
| Localizar por texto visible en apps i18n | `accessibilityIdentifier` estable |
| Estado compartido entre tests | Reset vía `launchArguments` en cada test |
| `#expect` de Swift Testing dentro de `XCUITestCase` | `XCTAssert*` — XCUITest es XCTest, no Swift Testing |
| Tests que dependen del backend real de prod | Backend local vía `launchEnvironment` |
