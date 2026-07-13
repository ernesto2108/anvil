# Guía de Concurrencia — Swift 6

El cambio más importante del período. Filosofía: **single-threaded por defecto, concurrencia solo cuando se pide explícitamente**. Progressive disclosure — código secuencial primero → `async/await` para APIs que suspenden → actors/`Sendable` solo cuando se introduce paralelismo real. Evitar sembrar `Task {}` y actors "por si acaso".

## Language mode y strict concurrency

Código nuevo en Swift 6 language mode (strict concurrency completo). Codebases existentes: migración incremental (ver checklist abajo), nunca "flip a Swift 6" de golpe.

## Approachable Concurrency (Swift 6.2+)

Redefine los defaults recomendados. En Xcode 26 agrupa 5 upcoming features:

| Feature | SE | Efecto |
|---|---|---|
| `DisableOutwardActorInference` | SE-401 | El aislamiento de un tipo no se propaga hacia afuera por conformances/miembros |
| `GlobalActorIsolatedTypesUsability` | SE-434 | Menos fricción usando tipos aislados a un global actor (closures, stored properties) |
| `InferIsolatedConformances` | SE-470 | Conformances heredan el aislamiento del tipo en vez de exigir `nonisolated` explícito |
| `InferSendableFromCaptures` | SE-418 | Closures infieren `Sendable` desde lo que capturan, reduciendo anotaciones |
| `NonisolatedNonsendingByDefault` | SE-461 | Funciones `async` no aisladas corren en el actor del caller, no en el executor global |

Para proyectos existentes, habilitarlas **una por una**, no en bloque — cada una puede exponer diagnósticos nuevos que conviene resolver aislados.

## defaultIsolation MainActor

`SWIFT_DEFAULT_ACTOR_ISOLATION = MainActor` (en SPM: `.defaultIsolation(MainActor.self)` en `swiftSettings`): el target se aísla a `@MainActor` por defecto. Elimina la mayor parte del boilerplate de anotaciones en código UI-heavy. Es el default de Apple para proyectos nuevos en Xcode 26.

**Recomendado para app targets y features UI; NO para librerías/paquetes de dominio**, que deben ser `nonisolated` por defecto (así el dominio es reutilizable y testeable fuera del main actor).

```swift
// Package.swift — app/feature target
.target(
    name: "CheckoutFeature",
    swiftSettings: [.defaultIsolation(MainActor.self)]
)
// Package.swift — domain target: sin defaultIsolation (nonisolated por defecto)
```

## nonisolated(nonsending) y @concurrent

Con SE-461 (default), las funciones `async` no aisladas corren en el actor del **caller**, no en el executor global. Esto evita saltos de contexto inesperados. Para trabajo que sí debe ir a background, marcarlo explícitamente con **`@concurrent`**.

```swift
// Corre en el actor del caller (p. ej. MainActor si lo llama la UI): NO salta a
// otro executor, pero tampoco descarga la CPU.
func parse(_ data: Data) async throws -> Report { ... }

// Descarga explícita a un thread del pool concurrente para trabajo CPU-bound.
@concurrent
func renderThumbnail(from data: Data) async -> UIImage { ... }
```

Regla práctica: si no hay trabajo pesado que descargar, no anotar nada — deja que corra en el caller.

## Cuándo actor vs @MainActor vs nada

| Situación | Elección |
|---|---|
| Estado mutable compartido tocado por UI | `@MainActor` (el default del app target ya lo cubre) |
| Estado mutable compartido accedido concurrentemente fuera de la UI (caché, coordinador de red) | `actor` dedicado |
| Value type que solo cruza fronteras (DTO, config) | `struct` + `Sendable`, sin aislamiento |
| Función pura / cálculo sin estado compartido | nada (`nonisolated`); `@concurrent` solo si es CPU-bound y conviene descargar |

Preferir estado local y value types antes que introducir un `actor`; un `actor` es correcto solo cuando hay estado mutable compartido con acceso concurrente real.

## Sendable — gotchas

- Los **value types** cuyos campos son todos `Sendable` lo conforman automáticamente.
- Las **clases** requieren `final` + estado inmutable (`let`) o sincronización interna para ser `Sendable`; una clase con `var` mutable no puede serlo sin un mecanismo de protección.
- `@unchecked Sendable` es una promesa manual al compilador — usarlo solo con sincronización propia demostrable (lock/queue), nunca para silenciar un diagnóstico.
- `InferSendableFromCaptures` (SE-418) reduce boilerplate en closures, pero capturar una referencia mutable no-`Sendable` seguirá fallando: extraer un value type o aislar el acceso.
- Tipos aislados a un global actor (`@MainActor class`) son implícitamente seguros de cruzar hacia ese actor; el diagnóstico aparece al cruzarlos hacia otro dominio.

## Typed throws (SE-0413, Swift 6.0)

```swift
func fetch() throws(NetworkError) -> Data
```

Usar en APIs internas con contratos de error estables y cerrados (parsers, validadores, capas de dominio) y en contextos Embedded/performance. **El default sigue siendo `throws` sin tipo** para APIs públicas o donde los errores pueden evolucionar — Apple lo recomienda así.

Nota: typed throws admite **un solo** tipo de error (no existen uniones `throws(A | B)`). Si se necesitan varios, envolver en un enum.

## Checklist de migración incremental

1. Mantener Swift 5 language mode; activar `-strict-concurrency=complete` **módulo por módulo**, empezando por los módulos hoja (los que nadie importa).
2. Resolver los diagnósticos de un módulo antes de pasar al siguiente; no acumular warnings entre módulos.
3. Habilitar las 5 upcoming features de Approachable Concurrency **una por una**, no en bloque.
4. Introducir `defaultIsolation(MainActor.self)` en app/feature targets una vez que sus dependencias de dominio ya compilan en strict.
5. Cuando todo el grafo compila en strict sin warnings, promover a Swift 6 language mode.
6. Migrar `DispatchQueue`/completion handlers a async/await al tocar cada módulo (no como refactor masivo separado).

## Prohibido en código nuevo

- `DispatchQueue` / `DispatchQueue.main.async` — usar `@MainActor` / async-await.
- Completion handlers — usar `async/await`.
- `Task {}` sembrados sin necesidad real de concurrencia.
- `@unchecked Sendable` sin sincronización propia demostrable.
