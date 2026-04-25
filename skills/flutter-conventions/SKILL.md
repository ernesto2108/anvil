---
name: flutter-conventions
description: Convenciones y estándares de código para Flutter/Dart móvil. Usar cuando se escriban widgets de Flutter, se revise código Dart, o el usuario mencione "Flutter patterns", "BLoC", "Riverpod", "widget composition", "freezed", o cuando se trabaje con archivos .dart.
---

# Flutter Conventions

## Filosofía

- **Los widgets son baratos, los rebuilds no** — compón widgets pequeños, pero controla cuándo se reconstruyen
- **El tipado seguro es tu primera prueba** — si el compilador puede detectarlo, no lo dejes para runtime
- **El estado pertenece fuera de la UI** — los widgets renderizan, los BLoCs/Notifiers deciden
- **Flujo de datos unidireccional** — el estado fluye hacia abajo, los eventos fluyen hacia arriba

## Stack

- Flutter + Dart (null safety obligatorio)
- Gestión de estado: BLoC o Riverpod (verificar preferencia del proyecto)
- Generación de código: freezed + json_serializable + build_runner
- DI: get_it + injectable (o providers de Riverpod)
- Navegación: GoRouter

## Reglas de Código

- Null safety — sin tipos `dynamic` a menos que sea absolutamente necesario
- Objetos de estado inmutables (usar `freezed` o `@immutable`)
- Composición de widgets — widgets pequeños y enfocados con responsabilidad única
- Separar widgets UI de la lógica (BLoC/Cubit/Notifier)
- Constructores `const` en todos lados donde sea posible — hasta 30% de mejora en renderizado
- Material 3 como línea base de diseño
- Extraer widgets a clases separadas, nunca métodos helper `Widget _buildX()`

## Reglas de Arquitectura

1. **MVVM + Clean Architecture** — UI Layer → Domain Layer → Data Layer (imports en dirección hacia adentro)
2. **Estructura de carpetas feature-first** — `lib/src/features/{name}/{presentation,application,domain,data}`
3. **Patrón Repository** — nunca llames APIs desde widgets o ViewModels directamente
4. **Patrón Result para errores** — los repositories retornan `Result<T>`, nunca lanzan excepciones. Los ViewModels hacen switch, nunca try/catch
5. **Dos capas DTO** — entidades de dominio (`freezed`) separadas de DTOs (`json_serializable`) con mappers `toDomain()`
6. **DI via constructores** — get_it + injectable, o providers de Riverpod
7. **Los repositories nunca se llaman entre sí** — combina datos en ViewModels o casos de uso de dominio

## Reglas de Gestión de Estado

| Alcance | Simple | Medio | Complejo |
|---|---|---|---|
| Widget único | `setState` | `setState` | Cubit |
| Feature | `ValueNotifier` | Cubit | BLoC |
| Cross-feature | Provider | Riverpod | BLoC |
| Global | Riverpod | Riverpod | BLoC |

Escalada: `setState` → Provider → Riverpod → BLoC. Ver `state-management-guide.md` para patrones completos.

## Checklist Pre-Implementación

- [ ] La carpeta de feature existe siguiendo la estructura feature-first
- [ ] La gestión de estado coincide con la convención del proyecto (BLoC vs Riverpod)
- [ ] Los modelos de dominio usan `freezed` o `@immutable`
- [ ] Sin tipos `dynamic` en la capa de dominio
- [ ] El widget tiene constructor `const` si es stateless
- [ ] Los streams y suscripciones tienen limpieza en `dispose()`
- [ ] El manejo de errores usa el patrón Result (sin try/catch en ViewModels)
- [ ] La navegación usa GoRouter con parámetros tipados
- [ ] Accesibilidad: Semantics, tooltips, probado con screen reader
- [ ] Theming: usa tokens `ColorScheme`, `textTheme`, `ThemeExtension` — sin valores codificados

## Detección de Anti-Patrones

Ver `anti-patterns.md` para la referencia completa de detección con niveles de severidad.

**Detección pasiva:** Al revisar código Flutter/Dart, escanea automáticamente los patrones `error` y `warning`. Reporta como `[file:line] [severity] [category] anti-pattern-name`.

**Detección activa:** Cuando el usuario pide "improve", "refactor", "optimize" — reporta también el nivel `suggestion` y propone correcciones.

Señales de alerta que siempre deben detener el trabajo:
- `setState` después de dispose sin verificación `mounted` → setState-after-dispose (error)
- `Timer`/`StreamSubscription` sin `cancel()` en `dispose()` → resource-leak (error)
- `dynamic` en modelos de dominio → untyped-domain (error)
- `BuildContext` a través de gaps async → context-across-async (error)
- `try/catch` en ViewModel/BLoC → error-swallowing (error)

## Archivos de Soporte

- `architecture-guide.md` — Arquitectura (MVVM, Clean Arch, patrón Result, generación de código, DI, GoRouter, código específico de plataforma, patrones de empresa)
- `state-management-guide.md` — Gestión de estado (BLoC, Riverpod, Provider, Cubit, setState, ValueNotifier)
- `testing-guide.md` — Pirámide de testing (pruebas unitarias, de widget, golden, de integración)
- `performance-guide.md` — Optimización de rendimiento (const, ListView.builder, RepaintBoundary, rebuilds granulares, patrones Alibaba/ByteDance)
- `theming-guide.md` — Theming (Material 3, ColorScheme.fromSeed, tokens ThemeExtension, layouts responsive)
- `anti-patterns.md` — Tabla de detección de anti-patrones con niveles de severidad y mapeo de correcciones
