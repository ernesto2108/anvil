---
name: swift-conventions
description: Convenciones y estándares de código para iOS nativo con Swift/SwiftUI. Usar cuando se escriba código Swift, SwiftUI, se revise código iOS nativo, o el usuario mencione "Swift", "SwiftUI", "iOS nativo", "@Observable", "Observation", "SwiftData", "Core Data", "Swift Testing", "NavigationStack", "Swift concurrency", "actors", "SwiftLint", "SwiftFormat", "SPM", o cuando se trabaje con archivos .swift, Package.swift o proyectos .xcodeproj/.xcworkspace.
---

# Swift Conventions

Convenciones para iOS nativo con Swift 6.x + SwiftUI. Postura por defecto verificada a julio 2026 (Swift 6.3, Xcode 26, iOS 26 SDK); cada default admite escape hatch documentado.

## Filosofía

- **Single-threaded por defecto, concurrencia solo cuando se pide** — con Approachable Concurrency el código secuencial es el punto de partida; el paralelismo se introduce explícitamente, no "por si acaso"
- **El compilador es tu primera prueba** — value types, typed state (enums con associated values), `Sendable` y strict concurrency detectan en compile-time lo que de otro modo explota en runtime
- **La UI observa, el dominio decide** — las views renderizan estado `@Observable`; la lógica vive en ViewModels y en la capa de dominio, nunca en `body`
- **Claridad en el punto de uso > brevedad** — los nombres se juzgan leyendo el call site (API Design Guidelines de swift.org)

## Stack

- Swift 6.x en language mode 6, Xcode 26+, deployment target iOS 17 mínimo (iOS 18 si el negocio lo permite). Justificar cualquier target menor.
- UI: SwiftUI-first con Observation (`@Observable`); UIKit solo vía `UIViewRepresentable`/`UIViewControllerRepresentable` con justificación.
- Concurrencia: Approachable Concurrency, `defaultIsolation(MainActor.self)` en app target; dominio/datos `nonisolated`.
- Persistencia: SwiftData (iOS 17+) para proyectos nuevos; Core Data por legacy o migraciones complejas.
- Networking: `URLSession` + async/await + `Codable`. Sin Alamofire.
- DI: initializer injection con protocolos (base); `@Environment`; Factory o swift-dependencies si el grafo crece.
- Testing: Swift Testing para unit; XCTest solo para XCUITest y performance.
- Dependencias: SPM exclusivamente (CocoaPods prohibido en código nuevo).
- Tooling: SwiftLint + SwiftFormat, versiones fijadas por proyecto.

## Reglas de Código

- Value types primero — structs y enums por defecto; clases solo para identidad compartida, interop Objective-C, o modelos `@Observable`.
- Modelar estado con enums + associated values (`enum LoadState { case idle, loading, loaded([Item]), failed(Error) }`), no booleanos paralelos.
- Optionals: `guard let` para early-exit, `if let` shorthand; nunca force-unwrap en producción salvo invariante probada con comentario. Nunca `try!`; `try?` solo cuando descartar el error es la semántica correcta y documentada.
- Errores: enums que conforman `Error` (+ `LocalizedError` al llegar a UI); `throws` como mecanismo primario con async/await; typed throws (`throws(MyError)`) solo en APIs internas con contratos de error cerrados.
- Marcar `final` las clases no diseñadas para herencia; acceso mínimo (`private` por defecto).
- Naming (API Design Guidelines): `UpperCamelCase` para tipos/protocolos, `lowerCamelCase` para el resto (incluidos casos de enum); métodos con side effects en verbo imperativo (`sort()`), sin side effects en frase nominal (`sorted()`); booleans como aserciones (`isEmpty`, `canUndo`); omitir palabras redundantes; factory methods con prefijo `make`.
- Un tipo por archivo, archivo nombrado igual al tipo; `// MARK: -` para seccionar; orden: propiedades → init → lifecycle → API pública → privados.

Ver `concurrency-guide.md` para las reglas completas de Swift 6 strict concurrency.

## Reglas de Arquitectura

1. **MVVM ligero pragmático** — un ViewModel `@Observable` por pantalla con lógica; views triviales sin ViewModel. La separación crítica es UI / dominio / datos, no View/ViewModel.
2. **Capas** — `Features/` (views + view models + router de feature), `Domain/` (entidades puras, casos de uso, protocolos de repositorio, sin UI ni frameworks de datos), `Data/` (API clients, persistencia, mappers DTO→dominio), `Core/` (DesignSystem, Networking, utilidades).
3. **Modularización con SPM local packages** — dependencias fluyen hacia adentro (features → domain/core, nunca al revés); las features no se conocen entre sí (navegación cross-feature vía router/protocolos en Core); el app target es cascarón (composición + inyección + entry point).
4. **Patrón Repository** — protocolo en Domain, implementación en Data; nunca llamar APIs desde views ni ViewModels directamente.
5. **DTOs separados de entidades de dominio** en apps medianas+ con mappers `toDomain()`.
6. **DI** — protocolos (o structs de closures) en dominio; el concreto se compone en el composition root del app target.

Ver `architecture-guide.md` para capas, SPM, repository, networking y persistencia en detalle.

## Reglas de UI y Estado

- `@Observable` (macro, iOS 17+) obligatorio para modelos nuevos — nunca `ObservableObject`/`@Published` en código nuevo. Mapeo: `@StateObject`→`@State`, `@ObservedObject`→propiedad simple, `@EnvironmentObject`→`@Environment(Type.self)`, binding a objeto recibido→`@Bindable`.
- No crear objetos caros inline en `@State` (se re-evalúa en cada rebuild); inyectarlos.
- Estado lo más abajo posible en la jerarquía; `@State` para estado local, `@Binding` para mutación delegada de value types, `@Environment` para dependencias de ambiente.
- Navegación: `NavigationStack` con rutas tipadas — enum `Route: Hashable`, `navigationDestination(for:)`, Router `@Observable` inyectado por environment con push/pop/popToRoot. Empujar rutas, nunca views ni strings. `NavigationView` está deprecado.
- Composición: views pequeñas (los structs de View son gratis — extraer subviews); `ViewModifier` para estilos; evitar `AnyView` (rompe el diffing de identidad); sin lógica de negocio en `body`.
- Liquid Glass (iOS 26): controles estándar del sistema; adopción obligatoria de cara a iOS 27.
- Accesibilidad: `.accessibilityLabel/Value/Hint`, Dynamic Type con `Font` semántica y `@ScaledMetric`, respetar `accessibilityReduceMotion`/`ReduceTransparency`.
- `#Preview` en toda view; `@Previewable @State` para estado interactivo.

Ver `swiftui-guide.md` para property wrappers, navegación, UIKit interop, Liquid Glass y previews en detalle.

## Checklist Pre-Implementación

- [ ] El módulo/feature existe siguiendo la estructura por feature (SPM local package cuando aplica)
- [ ] Modelos de estado nuevos usan `@Observable`, no `ObservableObject`
- [ ] Sin force-unwrap ni `try!` en código de producción
- [ ] Estado modelado con enums + associated values, no booleanos paralelos
- [ ] App target con `defaultIsolation(MainActor.self)`; dominio/datos `nonisolated`
- [ ] Sin `DispatchQueue`/completion handlers en código nuevo (async/await)
- [ ] Navegación con `NavigationStack` + enum de rutas + Router, no strings
- [ ] Repositories como protocolos en Domain; DTOs separados del dominio
- [ ] Persistencia nueva con SwiftData y `save()` explícito en rutas críticas
- [ ] `#Preview` presente en views nuevas
- [ ] Accesibilidad: labels, Dynamic Type, reduce-motion/transparency
- [ ] SPM para dependencias (sin CocoaPods); `final` en clases no heredables

## Detección de Anti-Patrones

Ver `anti-patterns.md` para la referencia completa de detección con niveles de severidad.

**Detección pasiva:** Al revisar código Swift/SwiftUI, escanea automáticamente los patrones `error` y `warning`. Reporta como `[file:line] [severity] [category] anti-pattern-name`.

**Detección activa:** Cuando el usuario pide "improve", "refactor", "optimize" — reporta también el nivel `suggestion` y propone correcciones.

Señales de alerta que siempre deben detener el trabajo:
- Force unwrap (`!`) en código de producción sin invariante probada → force-unwrap-production (error)
- `ObservableObject`/`@Published` en código nuevo → legacy-observation (error)
- `DispatchQueue`/completion handlers en código nuevo → legacy-concurrency (error)
- Lógica de negocio en `body` de una View → logic-in-body (error)
- `#expect` dentro de `XCTestCase` (o viceversa) → mixed-test-frameworks (error)
- Strings en navegación en vez de rutas tipadas → stringly-navigation (error)

## Archivos de Soporte

- `architecture-guide.md` — MVVM ligero, capas Features/Domain/Data/Core, modularización SPM local packages, DI (3 niveles), patrón repository, DTOs vs entidades, networking URLSession+async/await, SwiftData vs Core Data
- `swiftui-guide.md` — Observation/`@Observable`, property wrappers, navegación tipada con NavigationStack + Router, composición, UIKit interop, Liquid Glass, accesibilidad, `#Preview`
- `concurrency-guide.md` — Swift 6 strict concurrency, Approachable Concurrency, `defaultIsolation`, `nonisolated`, `@concurrent`, `Sendable`, typed throws, migración incremental
- `testing-guide.md` — Swift Testing (`@Test`, `#expect`, `#require`, parametrizados, `.serialized`), XCTest para XCUITest/performance, snapshot testing, determinismo con clock/uuid/date inyectados
- `anti-patterns.md` — tabla de detección de anti-patrones con niveles de severidad y mapeo de correcciones
