# Guía de Arquitectura — Swift/iOS nativo

## Patrón por defecto: MVVM ligero pragmático

Un ViewModel `@Observable` por **pantalla con lógica**, no por cada view. Views triviales (una lista que solo muestra datos ya resueltos) no necesitan ViewModel. El híbrido MV/MVVM es válido y deseado.

La separación que importa no es View/ViewModel sino **UI / dominio / datos**. Un ViewModel-Dios que hace networking, mapeo y navegación es un anti-patrón tan grave como no tener ViewModel.

```swift
@Observable
final class CheckoutViewModel {
    private(set) var state: LoadState<Order> = .idle
    private let repository: OrderRepository

    init(repository: OrderRepository) { self.repository = repository }

    func load() async {
        state = .loading
        do { state = .loaded(try await repository.fetchOrder()) }
        catch { state = .failed(error) }
    }
}
```

TCA (Point-Free) es una decisión explícita de equipo para apps con estado complejo compartido y exigencia de exhaustive testing — no el default. VIPER/Clean estricto se considera sobre-ingeniería para SwiftUI.

## Capas

| Capa | Contenido | Reglas |
|---|---|---|
| **Features/** | Un módulo por feature: views, view models, router de la feature | No importan otras features |
| **Domain/** | Entidades (structs puros), casos de uso/servicios, protocolos de repositorio | Sin dependencias de UI ni de frameworks de datos |
| **Data/** | API clients, persistencia (SwiftData/Core Data), mappers DTO→dominio | Implementa los protocolos de Domain |
| **Core/ (o Shared/)** | DesignSystem, Networking, extensiones, utilidades | Base compartida; sin lógica de feature |

Las dependencias fluyen **hacia adentro**: features → domain/core, nunca al revés.

## Modularización con SPM local packages

Práctica estándar 2025-2026. Paquetes locales sin versionado — un `Package.swift` por módulo, o un mega-paquete con múltiples targets/products.

Reglas:
- Las features **no se conocen entre sí**; la navegación cross-feature va vía router/protocolos definidos en Core.
- Un target `ExternalDependencies` que agrega los third-party; el resto depende de él, no directamente de las librerías.
- Módulos pequeños y cohesivos; `internal` por defecto impone los límites de dominio en compile-time.
- El app target queda como cascarón: composición + inyección + entry point.

Beneficios: recompilación incremental por módulo, previews más rápidos, tests por módulo, límites impuestos por el compilador. **Buildable folders** (Xcode 16+): la estructura del navigator es la del filesystem — prerequisito sano para extraer módulos y elimina conflictos de `.pbxproj`.

Organización **por feature, no por tipo**: `Features/Checkout/{CheckoutView, CheckoutViewModel, ...}`, no `Views/`/`ViewModels/` globales.

## Patrón Repository

```swift
// Domain/
protocol OrderRepository {
    func fetchOrder() async throws -> Order
}

// Data/
final class RemoteOrderRepository: OrderRepository {
    private let client: APIClient
    init(client: APIClient) { self.client = client }
    func fetchOrder() async throws -> Order {
        let dto: OrderDTO = try await client.get("/order")
        return dto.toDomain()
    }
}
```

Nunca llamar APIs desde views ni ViewModels directamente. Los repositories no se llaman entre sí — combinar datos en casos de uso de dominio.

## DTOs vs entidades de dominio

En apps medianas+, separar DTOs (`Codable`, forma del JSON) de entidades de dominio (structs puros) con mappers `toDomain()`. `Codable` con `CodingKeys` explícitas cuando el JSON difiere; fijar una convención de proyecto para snake_case (ej. `JSONDecoder.keyDecodingStrategy = .convertFromSnakeCase`).

## Networking

`URLSession` + async/await es el estándar:

```swift
let (data, response) = try await session.data(for: request)
```

Cliente API como protocolo en Domain + implementación en Data; errores tipados del transporte mapeados a errores de dominio. **Sin Alamofire** en apps nuevas. Para APIs con spec OpenAPI: `swift-openapi-generator` (Apple) es práctica creciente.

## Persistencia: SwiftData vs Core Data

**SwiftData** (iOS 17+, `@Model`, `@Query`, `ModelContainer/ModelContext`) para proyectos nuevos con target iOS 17+. API declarativa, `#Predicate` type-safe, integración nativa con SwiftUI.

Limitaciones reales a considerar:
- Más lento que Core Data en cargas masivas.
- **Autosave puede fallar silenciosamente** — llamar `try context.save()` explícito en operaciones críticas.
- Solo migraciones lightweight + `SchemaMigrationPlan` por etapas.

**Core Data** sigue correcto para: soporte iOS 15/16, schemas complejos con migraciones custom pesadas, apps con millones de registros, o codebases existentes. **GRDB** (SQLite tipado) es escape hatch para offline-first serio, no default.

## Inyección de dependencias (3 niveles)

1. **Initializer injection + protocolos** (baseline universal): protocolo en Domain, real en Data, mock en tests. Compile-time safe, cero dependencias. Composición en el app target. Suficiente para apps pequeñas-medianas.
2. **`@Environment` con claves custom** (idiomático SwiftUI): `EnvironmentKey` + `extension EnvironmentValues`; override trivial en previews y tests. Limitación: solo fluye por la jerarquía de views, sin garantía compile-time de registro.
3. **Librerías** cuando el grafo crece: **Factory** (ligera, container-based, compile-time safe) o **swift-dependencies** (Point-Free; trae `date`/`uuid`/`clock` controlables para tests deterministas, integra con Swift Testing). Evitar service locators globales tipo singleton mutable.

Regla transversal: las dependencias se definen como protocolos (o structs de closures) en la capa de dominio; el detalle concreto vive afuera.
