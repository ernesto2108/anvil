# Anti-Patrones — Swift/SwiftUI

Formato de reporte: `[file:line] [severity] [category] anti-pattern-name`

**Detección pasiva:** `error` + `warning` siempre. **Detección activa** (solo en "improve/refactor/optimize"): también `suggestion`.

| Anti-Patrón | Señal | Severidad | Categoría | Corrección |
|---|---|---|---|---|
| force-unwrap-production | `!` de force unwrap en código de producción sin invariante probada | error | safety | `guard let`/`if let`, o comentario justificando la invariante |
| force-try | `try!` en código de producción | error | safety | Manejar el error con `do/catch` o propagar con `throws` |
| legacy-observation | `ObservableObject`/`@Published` en código nuevo | error | ui-state | Migrar a `@Observable`; ver tabla de mapeo en swiftui-guide |
| legacy-concurrency | `DispatchQueue`/completion handlers en código nuevo | error | concurrency | `async/await` + `@MainActor`; `@concurrent` para background explícito |
| logic-in-body | Lógica de negocio (branching, cálculo, side effects) en `body` de una View | error | ui-state | Mover a ViewModel `@Observable` o computed property de dominio |
| stringly-navigation | Strings en navegación en vez de rutas tipadas | error | navigation | Enum `Route: Hashable` + `navigationDestination(for:)` + Router |
| mixed-test-frameworks | `#expect` dentro de `XCTestCase`, o `XCTAssert` en `@Test` | error | testing | Un solo framework por test; Swift Testing para unit nuevo |
| cocoapods-new-project | `Podfile`/`Pods/` en proyecto nuevo | error | tooling | SPM exclusivamente; migrar dependencias a `Package.swift` |
| swiftdata-no-save | Mutación de SwiftData en ruta crítica sin `try context.save()` explícito | error | data | Llamar `save()` explícito; el autosave puede fallar silenciosamente |
| any-view | `AnyView` en composición de views | warning | ui-state | `@ViewBuilder`, computed `some View`, o generics; `AnyView` rompe el diffing |
| expensive-state-inline | `@State` con objeto caro creado inline en el init del view | warning | performance | Inyectar el objeto; `@State` puede re-evaluar el init en cada rebuild |
| seeded-task | `Task {}`/actors sembrados sin necesidad real de concurrencia | warning | concurrency | Código secuencial por defecto; concurrencia solo cuando se pide |
| non-final-class | Clase no diseñada para herencia sin `final` | warning | code | Marcar `final` (mejor performance y claridad de intención) |
| boolean-state-soup | Múltiples booleanos paralelos para un estado de carga/flujo | warning | code | Enum con associated values (`enum LoadState { case idle, loading, ... }`) |
| default-isolation-domain | `defaultIsolation(MainActor.self)` en target de dominio/datos | warning | concurrency | Dominio/datos `nonisolated`; `MainActor` solo en app/features UI |
| dto-domain-mixed | DTO `Codable` usado como entidad de dominio en app mediana+ | warning | architecture | Separar DTO de entidad con mapper `toDomain()` |
| feature-cross-import | Una feature importa otra feature directamente | warning | architecture | Navegación cross-feature vía router/protocolos en Core |
| zstack-decoration | `ZStack` para decorar UN elemento (badge, fondo) | warning | ui-state | `.overlay(alignment:)`/`.background()` — el view base define el tamaño; `ZStack` solo para capas co-iguales. Ver swiftui-guide |
| redundant-stack-nesting | `VStack`/`HStack` con un solo hijo, o anidado sin cambiar eje/alignment/spacing | warning | ui-state | Aplanar o usar modifiers |
| conditional-identity | `if/else` que devuelve la misma vista con distinto valor | warning | ui-state | Ternario dentro del modifier (`.tint(x ? .red : .gray)`); el if/else crea dos identidades y destruye estado/transiciones |
| viewbuilder-overflow | Contenedor llegando al límite de 10 hijos de ViewBuilder | suggestion | ui-state | Descomponer en subvistas con nombre (no `Group`; `Group` solo agrupa sin imponer layout) |
| missing-preview | View nueva sin `#Preview` | suggestion | ui-state | Agregar `#Preview`; usar `@Previewable @State` si es interactivo |
| redundant-labels | Argument labels redundantes / palabras redundantes en nombres | suggestion | naming | API Design Guidelines: claridad en el call site, omitir redundancia |
| missing-accessibility | View interactiva sin `accessibilityLabel`/Dynamic Type | suggestion | accessibility | Agregar labels semánticos y soporte Dynamic Type |
