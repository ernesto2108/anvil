---
name: flutter-conventions
description: Flutter/Dart mobile conventions and coding standards. Use when writing Flutter widgets, reviewing Dart code, or user mentions "Flutter patterns", "BLoC", "Riverpod", "widget composition", "freezed", or working with .dart files.
---

# Flutter Conventions

## Philosophy

- **Widgets are cheap, rebuilds are not** — compose small widgets, but control when they rebuild
- **Type safety is your first test** — if the compiler can catch it, don't leave it to runtime
- **State belongs outside the UI** — widgets render, BLoCs/Notifiers decide
- **Unidirectional data flow** — state flows down, events flow up

## Stack

- Flutter + Dart (null safety enforced)
- State management: BLoC or Riverpod (check project preference)
- Code generation: freezed + json_serializable + build_runner
- DI: get_it + injectable (or Riverpod providers)
- Navigation: GoRouter

## Coding Rules

- Null safety — no `dynamic` types unless absolutely necessary
- Immutable state objects (use `freezed` or `@immutable`)
- Widget composition — small, focused widgets with single responsibility
- Separate UI widgets from logic (BLoC/Cubit/Notifier)
- `const` constructors everywhere possible — up to 30% rendering improvement
- Material 3 as design baseline
- Extract widgets to separate classes, never `Widget _buildX()` helper methods

## Architecture Rules

1. **MVVM + Clean Architecture** — UI Layer → Domain Layer → Data Layer (imports directional inward)
2. **Feature-first folder structure** — `lib/src/features/{name}/{presentation,application,domain,data}`
3. **Repository pattern** — never call APIs from widgets or ViewModels directly
4. **Result pattern for errors** — repositories return `Result<T>`, never throw. ViewModels switch, never try/catch
5. **Two DTO layers** — domain entities (`freezed`) separate from DTOs (`json_serializable`) with `toDomain()` mappers
6. **DI via constructors** — get_it + injectable, or Riverpod providers
7. **Repositories never call each other** — combine data in ViewModels or domain use cases

## State Management Rules

| Scope | Simple | Medium | Complex |
|---|---|---|---|
| Single Widget | `setState` | `setState` | Cubit |
| Feature | `ValueNotifier` | Cubit | BLoC |
| Cross-feature | Provider | Riverpod | BLoC |
| Global | Riverpod | Riverpod | BLoC |

Escalation: `setState` → Provider → Riverpod → BLoC. See `state-management-guide.md` for full patterns.

## Pre-Implementation Checklist

- [ ] Feature folder exists following feature-first structure
- [ ] State management matches project convention (BLoC vs Riverpod)
- [ ] Domain models use `freezed` or `@immutable`
- [ ] No `dynamic` types in domain layer
- [ ] Widget has `const` constructor if stateless
- [ ] Streams and subscriptions have cleanup in `dispose()`
- [ ] Error handling uses Result pattern (no try/catch in ViewModels)
- [ ] Navigation uses GoRouter with typed parameters
- [ ] Accessibility: Semantics, tooltips, tested with screen reader
- [ ] Theming: uses `ColorScheme`, `textTheme`, `ThemeExtension` tokens — no hardcoded values

## Anti-Pattern Detection

See `anti-patterns.md` for the full detection reference with severity levels.

**Passive detection:** When reviewing Flutter/Dart code, automatically scan for `error` and `warning` patterns. Report as `[file:line] [severity] [category] anti-pattern-name`.

**Active detection:** When user asks to "improve", "refactor", "optimize" — also report `suggestion` level and propose fixes.

Red flags that should always stop work:
- `setState` after dispose without `mounted` check → setState-after-dispose (error)
- `Timer`/`StreamSubscription` without `cancel()` in `dispose()` → resource-leak (error)
- `dynamic` in domain models → untyped-domain (error)
- `BuildContext` across async gaps → context-across-async (error)
- `try/catch` in ViewModel/BLoC → error-swallowing (error)

## Support Files

- `architecture-guide.md` — Architecture (MVVM, Clean Arch, Result pattern, code generation, DI, GoRouter, platform-specific code, company patterns)
- `state-management-guide.md` — State management (BLoC, Riverpod, Provider, Cubit, setState, ValueNotifier)
- `testing-guide.md` — Testing pyramid (unit, widget, golden, integration tests)
- `performance-guide.md` — Performance optimization (const, ListView.builder, RepaintBoundary, granular rebuilds, Alibaba/ByteDance patterns)
- `theming-guide.md` — Theming (Material 3, ColorScheme.fromSeed, ThemeExtension tokens, responsive layouts)
- `anti-patterns.md` — Anti-pattern detection table with severity levels and fix mapping
