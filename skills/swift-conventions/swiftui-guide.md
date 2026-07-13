# Guía de SwiftUI

## Observation framework — el estándar actual

Usar `@Observable` (macro, iOS 17+), **no** `ObservableObject`/`@Published`, para código nuevo. Ventaja de performance concreta: SwiftUI rastrea qué propiedades lee cada view y solo re-renderiza las views que dependen de la propiedad que cambió. Con `ObservableObject`, cualquier `@Published` invalida todas las views suscritas.

```swift
@Observable
final class ProfileModel {
    var name: String = ""
    var isEditing: Bool = false
}
```

Mapeo de property wrappers al migrar desde `ObservableObject`:

| ObservableObject | Observation |
|---|---|
| `@StateObject var m = M()` | `@State var m = M()` |
| `@ObservedObject var m` | `let m: M` (propiedad simple) |
| `@EnvironmentObject var m` | `@Environment(M.self) var m` |
| binding a objeto recibido | `@Bindable var m` |

**Gotcha de migración:** `@State` no tiene el `@autoclosure` de `@StateObject` — el inicializador puede evaluarse en cada rebuild (aunque solo se use el primer valor). No crear objetos caros inline; inyectarlos.

`ObservableObject` queda solo para: soporte iOS 15/16, o pipelines Combine legacy. iOS 26 añadió `Observations` (streams asíncronos de cambios) para observar modelos fuera de SwiftUI.

## Property wrappers de estado

- `@State` — estado local del view y modelos `@Observable` que el view **posee**.
- `@Binding` — mutación delegada de value types.
- `@Bindable` — bindings a un modelo `@Observable` recibido.
- `@Environment(Type.self)` — dependencias de ambiente y modelos compartidos.

Mantener el estado lo más abajo posible en la jerarquía ("state ownership at the lowest common ancestor").

## Navegación tipada

`NavigationStack` (iOS 16+) con rutas tipadas. Nunca empujar views directamente ni usar strings — empujar **rutas**.

```swift
enum Route: Hashable { case detail(Item.ID), settings }

@Observable
final class Router {
    var path = NavigationPath()
    func push(_ route: Route) { path.append(route) }
    func popToRoot() { path = NavigationPath() }
}

struct AppView: View {
    @State private var router = Router()
    var body: some View {
        NavigationStack(path: $router.path) {
            HomeView()
                .navigationDestination(for: Route.self) { route in
                    switch route { /* ... */ }
                }
        }
        .environment(router)
    }
}
```

El Router `@Observable` se inyecta por environment. Este patrón habilita deep linking y state restoration gratis (añadir `Codable` a `Route` si se persiste el path). `NavigationSplitView` para iPad/Mac/vision. `NavigationView` está deprecado.

## Composición

Views pequeñas y baratas — los structs de View son gratis, extraer subviews agresivamente. Usar computed properties `some View` para fragmentos y `ViewModifier` para estilos reutilizables. **Evitar `AnyView`** (rompe el diffing de identidad). No meter lógica de negocio en `body`.

## UIKit interop

SwiftUI-first. UIKit todavía necesario, vía `UIViewRepresentable`/`UIViewControllerRepresentable`, para: webviews complejos, cámaras (`AVCaptureVideoPreviewLayer`), text editing avanzado, collection views con performance extrema, y APIs sin equivalente SwiftUI. Siempre con justificación — la dirección es SwiftUI-first.

## Liquid Glass (iOS 26)

Rediseño visual más grande desde iOS 7. APIs: `.glassEffect()`, `GlassEffectContainer`, `glassButtonStyle`, formas concéntricas, toolbars y tab bars rediseñados. Compilar contra el SDK de iOS 26 adopta gran parte automáticamente si se usan componentes estándar.

**Crítico:** el opt-out (`UIDesignRequiresCompatibility`) desaparece con iOS 27 — la adopción no es opcional. Recomendación: controles estándar del sistema, sin fondos custom en barras.

## Accesibilidad

- `.accessibilityLabel`, `.accessibilityValue`, `.accessibilityHint`.
- `.accessibilityElement(children: .combine)` para agrupar.
- Dynamic Type con `Font` semántica (`.body`, `.headline`) y `@ScaledMetric`.
- Respetar `accessibilityReduceMotion` y `ReduceTransparency` (relevante con Liquid Glass).
- Auditar con Accessibility Inspector y `performAccessibilityAudit()` de XCUITest.

## Previews

`#Preview("nombre") { ... }` (macro, reemplaza `PreviewProvider`). `@Previewable @State` para estado interactivo en previews. Usar `PreviewModifier`/traits para inyectar contenedores SwiftData o dependencias mock. Objetivo: `#Preview` en toda view.
