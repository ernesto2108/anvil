# Guía de Testing — Swift

> Los tests los escribe el agente `tester`; esta guía documenta las convenciones que aplican al código Swift bajo test y las decisiones de framework, para coherencia con el resto de la skill.

## Swift Testing es el estándar para unit tests nuevos

Xcode 16+, open source, multiplataforma.

- `@Test` en funciones libres o structs (sin herencia de clase); `@Suite` para agrupar.
- `#expect(a == b)` — mensajes de fallo con la expresión completa y valores reales (muy superior a `XCTAssertEqual`).
- `#require` para condiciones que deben abortar el test (unwrap seguro de optionals: `let user = try #require(result)`).
- `withKnownIssue` para expected failures que no deben romper el suite.

```swift
@Test func totalIncludesTax() {
    let cart = Cart(items: [.init(price: 100)])
    #expect(cart.total == 110)
}
```

## Tests parametrizados nativos

Cada caso corre y reporta por separado — elimina loops manuales de XCTest y muestra exactamente cuál combinación falló:

```swift
@Test(arguments: [
    (subtotal: 100, rate: 0.10, expected: 110),
    (subtotal: 0,   rate: 0.10, expected: 0),
    (subtotal: 50,  rate: 0.00, expected: 50),
])
func appliesTax(subtotal: Int, rate: Double, expected: Int) {
    #expect(Invoice(subtotal: subtotal).total(taxRate: rate) == expected)
}
```

Se pueden pasar dos colecciones con `arguments: a, b` para producto cartesiano, o `zip(a, b)` para pares alineados.

## Organización de suites y tags

```swift
@Suite("Checkout")
struct CheckoutTests {
    @Test(.tags(.critical)) func placesOrder() async { ... }
}

extension Tag { @Tag static var critical: Self }
```

- `@Suite` agrupa y permite estado compartido vía `init`/`deinit` (reemplazan `setUp`/`tearDown`).
- `.tags(...)` para filtrar en CI (correr solo `critical` en pre-merge, todo en nightly).
- Otros traits: `.enabled(if:)`, `.disabled(...)`, `.timeLimit(...)`, `.bug(...)`.

## Paralelización y estado compartido

Swift Testing paraleliza por defecto (sync y async, in-process). Implicaciones:

- Los tests **no deben** depender de orden ni compartir estado mutable global.
- Si un suite toca un recurso compartido no aislable (un singleton, un archivo temporal fijo), aplicar el trait `.serialized` a nivel de `@Suite` para forzar ejecución secuencial dentro de él.
- Preferir aislar el estado (inyectar dependencias frescas por test) antes que recurrir a `.serialized`.

## Testing de código @MainActor y actors

- Un `@Test` que ejerce tipos `@MainActor` (ViewModels, modelos `@Observable` de UI) debe correr en el main actor: marcar la función o el `@Suite` con `@MainActor`.
- Los tests `async` esperan directamente con `await`; no hay expectations manuales estilo XCTest para async.
- Para código detrás de un `actor`, `await` sobre sus métodos serializa el acceso naturalmente — no hace falta sincronización extra en el test.
- Para forzar estados de carrera o cancelación, usar `Task` explícitas y `await` sobre sus resultados; verificar con `#expect`.

```swift
@MainActor
@Test func loadPopulatesState() async {
    let vm = ProfileViewModel(repository: StubRepository())
    await vm.load()
    #expect(vm.state == .loaded(.stub))
}
```

## Determinismo (clock/uuid/date inyectados)

Nunca leer `Date()`, `UUID()` ni relojes del sistema directamente en la lógica bajo test — hacen los tests no deterministas y flaky:

- Inyectar un proveedor de fecha (`() -> Date`), un generador de `UUID` y un `Clock` como dependencias.
- **swift-dependencies** trae `date`, `uuid` y `clock` controlables out-of-the-box y exige override en tests (falla si usas la versión live), lo que hace el determinismo el default.
- Para tiempo, usar un `TestClock` (adelantar el reloj manualmente) en vez de `Task.sleep` real, que ralentiza y desestabiliza la suite.

## Snapshot testing

**pointfreeco/swift-snapshot-testing** (v1.17+ con soporte Swift Testing) es la librería de facto. Especialmente valioso en SwiftUI, donde no hay acceso programático al view tree.

- Estrategias: imagen (`.image`), texto/`.description` de un value, y `.recursiveDescription` para jerarquías UIKit.
- **Inline snapshots**: el snapshot vive en el código fuente (`assertInlineSnapshot`), revisable en el diff del PR sin abrir archivos aparte — preferirlos para valores textuales pequeños.
- **Fijar device y OS de referencia en CI** (p. ej. iPhone 16, un iOS SDK concreto): renders en simuladores/OS distintos producen diffs de píxeles falsos.
- Primera corrida genera el snapshot y "falla" — commitear el snapshot generado; correr en `record` mode solo cuando el cambio visual es intencional.

## Swift Testing vs XCTest — qué va dónde

| Necesidad | Framework |
|---|---|
| Unit tests de dominio, ViewModels, mappers, lógica | **Swift Testing** |
| UI end-to-end (`XCUIApplication`, tap/scroll, screenshots de flujo) | **XCTest** (XCUITest) |
| Performance / benchmarks (`XCTMetric`, `measure`) | **XCTest** |
| Auditoría de accesibilidad (`performAccessibilityAudit()`) | **XCTest** (XCUITest) |

**No mezclar frameworks en un mismo test:** `#expect` dentro de `XCTestCase` no registra fallos, y `XCTAssert` dentro de `@Test` tampoco. Migración incremental: tests nuevos en Swift Testing, los viejos se migran al tocarlos (recomendación oficial de Apple); ambos runners coexisten en el mismo target.

Pirámide: grueso en unit tests sin UI (rápidos, paralelos); XCUITest solo para flujos críticos (login, checkout). Cobertura vía plan de tests de Xcode (`-enableCodeCoverage YES` / `swift test --enable-code-coverage`).
