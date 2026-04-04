# Flutter Anti-Pattern Detection Reference

## Detection Modes

**Passive detection:** When reviewing Flutter/Dart code, automatically scan for `error` and `warning` patterns. Report as `[file:line] [severity] [category] anti-pattern-name`.

**Active detection:** When user asks to "improve", "refactor", "optimize" — also report `suggestion` level and propose fixes referencing the relevant guide.

---

## Error (Must Fix)

| Code Pattern | Anti-Pattern | Category | Fix |
|---|---|---|---|
| `setState` in `dispose()` or after async gap without `mounted` check | setState-after-dispose | crashes | Check `mounted` before `setState` |
| `Timer`/`StreamSubscription` without `cancel()` in `dispose()` | resource-leak | memory | Cancel in `dispose()`, use `autoDispose` with Riverpod |
| `dynamic` type in domain models | untyped-domain | types | Concrete types, `freezed` classes |
| `BuildContext` used across async gaps | context-across-async | crashes | Capture navigator/theme before `await` |
| `try/catch` in ViewModel/BLoC instead of Result pattern | error-swallowing | architecture | Repository returns `Result<T>`, ViewModel switches. See `architecture-guide.md` |

---

## Warning (Should Fix)

| Code Pattern | Anti-Pattern | Category | Fix |
|---|---|---|---|
| Logic inside `build()` method (API calls, heavy computation) | logic-in-build | performance | Move to BLoC/Cubit/Notifier. See `performance-guide.md` |
| Widget tree >4 levels of nesting in single `build()` | deep-widget-tree | readability | Extract child widgets to separate widget classes |
| `setState` for state shared across widgets | local-state-abuse | state | BLoC, Riverpod, or InheritedWidget. See `state-management-guide.md` |
| Direct `http.get`/`http.post` in widget | network-in-widget | architecture | Repository pattern, inject data source. See `architecture-guide.md` |
| `Widget buildSomething()` helper methods | helper-methods | performance | Extract to separate widget class for rebuild isolation. See `performance-guide.md` |
| Repositories calling each other | repo-coupling | architecture | Combine in ViewModel or domain use case. See `architecture-guide.md` |
| `BlocBuilder` wrapping entire screen | wide-rebuild | performance | Use `BlocSelector` for granular rebuilds. See `performance-guide.md` |

---

## Suggestion (Consider Fixing)

| Code Pattern | Anti-Pattern | Category | Fix |
|---|---|---|---|
| Missing `const` on stateless widget constructors | missing-const | performance | Add `const` constructor. See `performance-guide.md` |
| `print()` for debugging | print-debugging | observability | `debugPrint()` or `logger` package |
| Hard-coded strings in UI | hardcoded-strings | i18n | `l10n` / `intl` for localization |
| `MediaQuery.of(context)` in deeply nested widgets | excessive-media-query | performance | Pass values down or use `LayoutBuilder`. See `theming-guide.md` |
| Widget tests without `pumpAndSettle()` | missing-pump | testing | `await tester.pumpAndSettle()` after interactions. See `testing-guide.md` |
| `GetX` in production code | getx-in-prod | architecture | Migrate to BLoC or Riverpod. See `state-management-guide.md` |
| Missing `Semantics`/`tooltip` on icon buttons | missing-a11y | accessibility | Add `tooltip` and `Semantics` wrapper |
| Hardcoded colors (`Color(0xFF...)`) | hardcoded-colors | theming | Use `Theme.of(context).colorScheme`. See `theming-guide.md` |
| Hardcoded font sizes (`TextStyle(fontSize: 16)`) | hardcoded-fonts | theming | Use `Theme.of(context).textTheme`. See `theming-guide.md` |
| Hardcoded spacing (`EdgeInsets.all(16)`) | hardcoded-spacing | theming | Use `ThemeExtension` tokens. See `theming-guide.md` |

---

## Detection Patterns

```
# setState-after-dispose
setState\( → inside dispose() or after await without mounted check

# resource-leak
Timer\.|Stream.*listen → no corresponding .cancel() in dispose()

# context-across-async
await.*\n.*context\. → BuildContext used after async gap

# helper-methods
Widget _build\w+\( → widget helper method instead of separate class

# wide-rebuild
BlocBuilder.*\n.*Scaffold → BlocBuilder wrapping entire screen

# hardcoded-colors
Color\(0x → direct color instead of theme reference

# hardcoded-spacing
EdgeInsets\.\w+\(\d → direct number instead of token
```

---

## Fix Mapping

| Anti-Pattern | Recommended Pattern | Guide |
|---|---|---|
| error-swallowing | Result pattern | `architecture-guide.md` |
| logic-in-build | BLoC/Cubit/Notifier | `state-management-guide.md` |
| local-state-abuse | Escalation path | `state-management-guide.md` |
| network-in-widget | Repository + DI | `architecture-guide.md` |
| helper-methods | Separate widget class | `performance-guide.md` |
| repo-coupling | ViewModel composition | `architecture-guide.md` |
| wide-rebuild | BlocSelector/Consumer | `performance-guide.md` |
| hardcoded-colors/fonts/spacing | Theme tokens | `theming-guide.md` |
| missing-pump | pumpAndSettle | `testing-guide.md` |
| getx-in-prod | BLoC or Riverpod | `state-management-guide.md` |
