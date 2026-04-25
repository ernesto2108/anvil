# Referencia de Detección de Anti-Patrones Flutter

## Modos de Detección

**Detección pasiva:** Al revisar código Flutter/Dart, escanear automáticamente en busca de patrones `error` y `warning`. Reportar como `[file:line] [severity] [category] anti-pattern-name`.

**Detección activa:** Cuando el usuario pide "mejorar", "refactorizar", "optimizar" — también reportar nivel `suggestion` y proponer correcciones referenciando la guía correspondiente.

---

## Error (Debe Corregirse)

| Patrón de código | Anti-patrón | Categoría | Corrección |
|---|---|---|---|
| `setState` en `dispose()` o después de un gap async sin verificar `mounted` | setState-after-dispose | crashes | Verificar `mounted` antes de `setState` |
| `Timer`/`StreamSubscription` sin `cancel()` en `dispose()` | resource-leak | memory | Cancelar en `dispose()`, usar `autoDispose` con Riverpod |
| Tipo `dynamic` en modelos de dominio | untyped-domain | types | Tipos concretos, clases `freezed` |
| `BuildContext` usado entre gaps async | context-across-async | crashes | Capturar navigator/theme antes del `await` |
| `try/catch` en ViewModel/BLoC en vez del patrón Result | error-swallowing | architecture | El repositorio retorna `Result<T>`, el ViewModel hace switch. Ver `architecture-guide.md` |

---

## Warning (Debería Corregirse)

| Patrón de código | Anti-patrón | Categoría | Corrección |
|---|---|---|---|
| Lógica dentro del método `build()` (llamadas API, cómputo pesado) | logic-in-build | performance | Mover a BLoC/Cubit/Notifier. Ver `performance-guide.md` |
| Árbol de widgets con >4 niveles de anidado en un único `build()` | deep-widget-tree | readability | Extraer widgets hijos a clases separadas |
| `setState` para estado compartido entre widgets | local-state-abuse | state | BLoC, Riverpod, o InheritedWidget. Ver `state-management-guide.md` |
| `http.get`/`http.post` directos en un widget | network-in-widget | architecture | Patrón Repository, inyectar data source. Ver `architecture-guide.md` |
| Métodos helper `Widget buildSomething()` | helper-methods | performance | Extraer a clase widget separada para aislamiento de reconstrucción. Ver `performance-guide.md` |
| Repositorios que se llaman entre sí | repo-coupling | architecture | Combinar en ViewModel o caso de uso del dominio. Ver `architecture-guide.md` |
| `BlocBuilder` envolviendo toda la pantalla | wide-rebuild | performance | Usar `BlocSelector` para reconstrucciones granulares. Ver `performance-guide.md` |

---

## Suggestion (Considerar Corregir)

| Patrón de código | Anti-patrón | Categoría | Corrección |
|---|---|---|---|
| Falta `const` en constructores de widgets stateless | missing-const | performance | Agregar constructor `const`. Ver `performance-guide.md` |
| `print()` para depuración | print-debugging | observability | `debugPrint()` o paquete `logger` |
| Strings hardcodeadas en la UI | hardcoded-strings | i18n | `l10n` / `intl` para localización |
| `MediaQuery.of(context)` en widgets profundamente anidados | excessive-media-query | performance | Pasar valores hacia abajo o usar `LayoutBuilder`. Ver `theming-guide.md` |
| Tests de widget sin `pumpAndSettle()` | missing-pump | testing | `await tester.pumpAndSettle()` después de interacciones. Ver `testing-guide.md` |
| `GetX` en código de producción | getx-in-prod | architecture | Migrar a BLoC o Riverpod. Ver `state-management-guide.md` |
| Falta `Semantics`/`tooltip` en botones con íconos | missing-a11y | accessibility | Agregar `tooltip` y wrapper `Semantics` |
| Colores hardcodeados (`Color(0xFF...)`) | hardcoded-colors | theming | Usar `Theme.of(context).colorScheme`. Ver `theming-guide.md` |
| Tamaños de fuente hardcodeados (`TextStyle(fontSize: 16)`) | hardcoded-fonts | theming | Usar `Theme.of(context).textTheme`. Ver `theming-guide.md` |
| Espaciado hardcodeado (`EdgeInsets.all(16)`) | hardcoded-spacing | theming | Usar tokens `ThemeExtension`. Ver `theming-guide.md` |

---

## Patrones de Detección

```
# setState-after-dispose
setState\( → dentro de dispose() o después de await sin verificación de mounted

# resource-leak
Timer\.|Stream.*listen → sin el correspondiente .cancel() en dispose()

# context-across-async
await.*\n.*context\. → BuildContext usado después de gap async

# helper-methods
Widget _build\w+\( → método helper de widget en vez de clase separada

# wide-rebuild
BlocBuilder.*\n.*Scaffold → BlocBuilder envolviendo toda la pantalla

# hardcoded-colors
Color\(0x → color directo en vez de referencia al tema

# hardcoded-spacing
EdgeInsets\.\w+\(\d → número directo en vez de token
```

---

## Mapeo de Correcciones

| Anti-patrón | Patrón recomendado | Guía |
|---|---|---|
| error-swallowing | Patrón Result | `architecture-guide.md` |
| logic-in-build | BLoC/Cubit/Notifier | `state-management-guide.md` |
| local-state-abuse | Ruta de escalado | `state-management-guide.md` |
| network-in-widget | Repository + DI | `architecture-guide.md` |
| helper-methods | Clase widget separada | `performance-guide.md` |
| repo-coupling | Composición en ViewModel | `architecture-guide.md` |
| wide-rebuild | BlocSelector/Consumer | `performance-guide.md` |
| hardcoded-colors/fonts/spacing | Tokens de tema | `theming-guide.md` |
| missing-pump | pumpAndSettle | `testing-guide.md` |
| getx-in-prod | BLoC o Riverpod | `state-management-guide.md` |
