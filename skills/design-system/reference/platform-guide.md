# Platform-Specific Design Token Guide

## Web Only

### Defaults
- **Base unit:** 4px (0.25rem)
- **Font size base:** 16px (1rem) — browser default, accessible
- **Font family:** System stack or custom web font (Inter, Geist, etc.)
- **Breakpoints:** 640, 768, 1024, 1280, 1536px (Tailwind standard)
- **Approach:** Mobile-first with `min-width` media queries

### Web-Specific Tokens
- Breakpoints (sm through 2xl)
- Container max-widths
- Focus ring styles (2px solid, 2px offset — WCAG 2.2)
- Scrollbar styling tokens (if custom)
- Print-specific overrides (optional)

### Framework Integration

**Tailwind CSS:**
- Tokens map directly to `@theme` CSS custom properties
- Semantic tokens become Tailwind utilities via `theme.extend`
- Example: `color-primary` → `--color-primary` → `bg-primary`, `text-primary`

**CSS Custom Properties (vanilla):**
```css
:root {
  --color-primary: #2563eb;
  --spacing-4: 1rem;
  --text-base: 1rem;
}
[data-theme="dark"] {
  --color-primary: #60a5fa;
}
```

---

## Mobile Only (iOS / Android)

### Defaults
- **Base unit:** 8pt (Apple), 4dp (Material)
- **Font size base:** 17pt iOS (Body), 16sp Android (Body1)
- **Font family:** SF Pro (iOS), Roboto (Android), or custom
- **Touch targets:** 44x44pt minimum

### iOS-Specific Tokens (Apple HIG)

**Text styles** — use semantic names, not fixed sizes:
| Role | iOS name | Default size |
|---|---|---|
| Display | Large Title | 34pt |
| Heading 1 | Title 1 | 28pt |
| Heading 2 | Title 2 | 22pt |
| Heading 3 | Title 3 | 20pt |
| Body | Body | 17pt |
| Body small | Subheadline | 15pt |
| Caption | Caption 1 | 12pt |
| Label | Footnote | 13pt |

**Dynamic Type is mandatory** — all text must scale with user accessibility settings (xSmall through AX5).

**Color approach:**
- Use semantic system colors (`label`, `secondaryLabel`, `systemBackground`)
- Dark mode adapts automatically
- Elevated surfaces get lighter in dark mode (opposite of light)
- Safe areas vary by device — never hardcode margins

### Android-Specific Tokens (Material Design 3)

**Type scale:**
| Role | M3 name | Default size |
|---|---|---|
| Display | Display Large | 57sp |
| Heading 1 | Headline Large | 32sp |
| Heading 2 | Headline Medium | 28sp |
| Heading 3 | Title Large | 22sp |
| Body | Body Large | 16sp |
| Body small | Body Medium | 14sp |
| Caption | Body Small | 12sp |
| Label | Label Large | 14sp |

**Shape scale:**
- Extra-Small: 4dp (inputs)
- Small: 8dp (buttons)
- Medium: 12dp (cards)
- Large: 16dp (modals)
- Extra-Large: 28dp (bottom sheets)
- Full: circular

---

## Both (Web + Mobile)

### Strategy
Define tokens at an abstract level, then map to platform-specific values.

```
Abstract token        →  Web (CSS)           →  iOS (Swift)        →  Android (Compose)
color-primary         →  --color-primary     →  .accentColor       →  MaterialTheme.colorScheme.primary
type-body             →  font-size: 1rem     →  .body              →  Typography.bodyLarge
spacing-4             →  1rem (16px)         →  16pt               →  16.dp
radius-md             →  0.375rem (6px)      →  6pt                →  8.dp (M3 Small)
```

### Shared Decisions
- Color palette: identical across platforms (same hex values)
- Typography roles: same semantic names, platform-specific sizes
- Spacing scale: same ratios, may differ in absolute values (pt vs px vs dp)
- Shadows: similar visual weight, platform-native implementation

### Platform Differences to Document

| Concern | Web | iOS | Android |
|---|---|---|---|
| Font size base | 16px | 17pt | 16sp |
| Touch target min | 44px | 44pt | 48dp |
| Safe areas | None (viewport) | Dynamic Island, home indicator | Status bar, nav bar |
| Dark mode | CSS `prefers-color-scheme` or `data-theme` | `UITraitCollection.userInterfaceStyle` | `isSystemInDarkTheme()` |
| Dynamic text | `rem` units + media queries | Dynamic Type (mandatory) | `sp` units (scalable) |
| Focus indicators | Visible focus ring (WCAG) | VoiceOver cursor | TalkBack focus |
| Haptics | N/A | `UIImpactFeedbackGenerator` | `HapticFeedback` |

### Flutter-Specific Notes

When the project uses Flutter for mobile:
- Tokens map to `ThemeData` and `ColorScheme`
- Typography uses `TextTheme` with named styles
- Spacing via `EdgeInsets` and `SizedBox`
- Material 3 via `useMaterial3: true`
- Adaptive widgets handle iOS/Android differences

```dart
// Token → Flutter mapping
final colorScheme = ColorScheme(
  primary: Color(0xFF2563EB),    // color-primary
  onPrimary: Color(0xFFFFFFFF),  // color-text-on-primary
  surface: Color(0xFFFFFFFF),    // color-surface
  // ...
);
```
