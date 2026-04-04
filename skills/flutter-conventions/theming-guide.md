# Flutter Theming & Design System Guide

## Material 3 Setup

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

Generates a harmonious, accessible color palette from a single seed color. All Material 3 components automatically use these colors.

---

## Custom Design Tokens via ThemeExtension

For project-specific tokens (spacing, border radius, shadows) not covered by Material 3.

### Define

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

### Register

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

### Use

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

## Responsive Layouts

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

### Never Hardcode Dimensions

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

## Accessibility in Theming

- **Color contrast**: Material 3 `ColorScheme.fromSeed` generates accessible palettes by default
- **Text scaling**: use `Theme.of(context).textTheme` — respects user's font size preference
- **Never hardcode font sizes** — always reference `textTheme`
- **Dark mode**: provide `darkTheme` for users who prefer reduced light

---

## Rules

1. **Components consume token references, never literal values** — `spacing.md` not `16.0`
2. **Use `ColorScheme` colors** — `colors.primary`, `colors.surface`, never hardcoded hex
3. **Use `textTheme` styles** — `textTheme.bodyLarge`, never inline `TextStyle`
4. **`ThemeExtension` for custom tokens** — spacing, radius, shadows, durations
5. **Test both light and dark themes** in widget tests and golden tests
