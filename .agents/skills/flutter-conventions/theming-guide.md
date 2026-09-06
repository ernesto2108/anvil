# Guía de Temas y Sistema de Diseño Flutter

## Configuración de Material 3

```dart
final lightTheme = ThemeData(
  useMaterial3: true,
  colorScheme: ColorScheme.fromSeed(
    seedColor: const Color(0xFF6750A4),
    brightness: Brightness.light,
  ),
  textTheme: const TextTheme(
    headlineLarge: TextStyle(fontSize: 32, fontWeight: FontWeight.bold),
    headlineMedium: TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
    bodyLarge: TextStyle(fontSize: 16),
    bodyMedium: TextStyle(fontSize: 14),
    labelLarge: TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
  ),
);

final darkTheme = ThemeData(
  useMaterial3: true,
  colorScheme: ColorScheme.fromSeed(
    seedColor: const Color(0xFF6750A4),
    brightness: Brightness.dark,
  ),
);

// app
MaterialApp(
  theme: lightTheme,
  darkTheme: darkTheme,
  themeMode: ThemeMode.system,
)
```

### `ColorScheme.fromSeed`

Genera una paleta de colores armoniosa y accesible a partir de un único color semilla. Todos los componentes de Material 3 usan estos colores automáticamente.

---

## Tokens de Diseño Personalizados via ThemeExtension

Para tokens específicos del proyecto (espaciado, radio de borde, sombras) no cubiertos por Material 3.

### Definición

```dart
class AppSpacing extends ThemeExtension<AppSpacing> {
  final double xs;
  final double sm;
  final double md;
  final double lg;
  final double xl;

  const AppSpacing({
    this.xs = 4,
    this.sm = 8,
    this.md = 16,
    this.lg = 24,
    this.xl = 32,
  });

  @override
  AppSpacing copyWith({double? xs, double? sm, double? md, double? lg, double? xl}) {
    return AppSpacing(
      xs: xs ?? this.xs,
      sm: sm ?? this.sm,
      md: md ?? this.md,
      lg: lg ?? this.lg,
      xl: xl ?? this.xl,
    );
  }

  @override
  AppSpacing lerp(covariant AppSpacing? other, double t) {
    if (other == null) return this;
    return AppSpacing(
      xs: lerpDouble(xs, other.xs, t)!,
      sm: lerpDouble(sm, other.sm, t)!,
      md: lerpDouble(md, other.md, t)!,
      lg: lerpDouble(lg, other.lg, t)!,
      xl: lerpDouble(xl, other.xl, t)!,
    );
  }
}

class AppRadius extends ThemeExtension<AppRadius> {
  final double sm;
  final double md;
  final double lg;
  final double full;

  const AppRadius({
    this.sm = 4,
    this.md = 8,
    this.lg = 16,
    this.full = 999,
  });

  @override
  AppRadius copyWith({double? sm, double? md, double? lg, double? full}) {
    return AppRadius(
      sm: sm ?? this.sm,
      md: md ?? this.md,
      lg: lg ?? this.lg,
      full: full ?? this.full,
    );
  }

  @override
  AppRadius lerp(covariant AppRadius? other, double t) {
    if (other == null) return this;
    return AppRadius(
      sm: lerpDouble(sm, other.sm, t)!,
      md: lerpDouble(md, other.md, t)!,
      lg: lerpDouble(lg, other.lg, t)!,
      full: lerpDouble(full, other.full, t)!,
    );
  }
}
```

### Registro

```dart
final theme = ThemeData(
  useMaterial3: true,
  colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xFF6750A4)),
  extensions: const [
    AppSpacing(),
    AppRadius(),
  ],
);
```

### Uso

```dart
@override
Widget build(BuildContext context) {
  final spacing = Theme.of(context).extension<AppSpacing>()!;
  final radius = Theme.of(context).extension<AppRadius>()!;
  final colors = Theme.of(context).colorScheme;

  return Container(
    padding: EdgeInsets.all(spacing.md),
    decoration: BoxDecoration(
      color: colors.surface,
      borderRadius: BorderRadius.circular(radius.md),
    ),
    child: Text(
      'Hello',
      style: Theme.of(context).textTheme.bodyLarge,
    ),
  );
}
```

---

## Layouts Responsivos

### LayoutBuilder

```dart
LayoutBuilder(
  builder: (context, constraints) {
    if (constraints.maxWidth > 900) {
      return const DesktopLayout();
    } else if (constraints.maxWidth > 600) {
      return const TabletLayout();
    } else {
      return const MobileLayout();
    }
  },
)
```

### Nunca Hardcodear Dimensiones

```dart
// bad
Container(width: 375) // assumes phone width

// good
Container(
  width: double.infinity,
  constraints: const BoxConstraints(maxWidth: 600),
)
```

---

## Accesibilidad en el Tema

- **Contraste de colores**: `ColorScheme.fromSeed` de Material 3 genera paletas accesibles por defecto
- **Escalado de texto**: usar `Theme.of(context).textTheme` — respeta la preferencia de tamaño de fuente del usuario
- **Nunca hardcodear tamaños de fuente** — siempre referenciar `textTheme`
- **Modo oscuro**: proveer `darkTheme` para usuarios que prefieren menos luz

---

## Reglas

1. **Los componentes consumen referencias a tokens, nunca valores literales** — `spacing.md` no `16.0`
2. **Usar colores de `ColorScheme`** — `colors.primary`, `colors.surface`, nunca hex hardcodeado
3. **Usar estilos de `textTheme`** — `textTheme.bodyLarge`, nunca `TextStyle` inline
4. **`ThemeExtension` para tokens personalizados** — espaciado, radio, sombras, duraciones
5. **Testear tanto tema claro como oscuro** en widget tests y golden tests
