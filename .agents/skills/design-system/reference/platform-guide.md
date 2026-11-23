# Guía de Design Tokens por Plataforma

## Solo Web

### Valores por Defecto
- **Unidad base:** 4px (0.25rem)
- **Tamaño base de fuente:** 16px (1rem) — por defecto del navegador, accesible
- **Font family:** System stack o fuente web personalizada (Inter, Geist, etc.)
- **Breakpoints:** 640, 768, 1024, 1280, 1536px (estándar Tailwind)
- **Enfoque:** Mobile-first con media queries `min-width`

### Tokens Específicos de Web
- Breakpoints (sm hasta 2xl)
- Max-widths de contenedor
- Estilos de anillo de foco (2px solid, 2px offset — WCAG 2.2)
- Tokens de estilo de scrollbar (si es personalizado)
- Overrides específicos para impresión (opcional)

### Integración con Frameworks

**Tailwind CSS:**
- Los tokens se mapean directamente a custom properties CSS de `@theme`
- Los tokens semánticos se convierten en utilidades de Tailwind via `theme.extend`
- Ejemplo: `color-primary` → `--color-primary` → `bg-primary`, `text-primary`

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

## Solo Mobile (iOS / Android)

### Valores por Defecto
- **Unidad base:** 8pt (Apple), 4dp (Material)
- **Tamaño base de fuente:** 17pt iOS (Body), 16sp Android (Body1)
- **Font family:** SF Pro (iOS), Roboto (Android), o personalizada
- **Touch targets:** mínimo 44x44pt

### Tokens Específicos de iOS (Apple HIG)

**Estilos de texto** — usa nombres semánticos, no tamaños fijos:
| Rol | Nombre iOS | Tamaño por defecto |
|---|---|---|
| Display | Large Title | 34pt |
| Heading 1 | Title 1 | 28pt |
| Heading 2 | Title 2 | 22pt |
| Heading 3 | Title 3 | 20pt |
| Body | Body | 17pt |
| Body small | Subheadline | 15pt |
| Caption | Caption 1 | 12pt |
| Label | Footnote | 13pt |

**Dynamic Type es obligatorio** — todo el texto debe escalar con la configuración de accesibilidad del usuario (xSmall hasta AX5).

**Enfoque de color:**
- Usa colores semánticos del sistema (`label`, `secondaryLabel`, `systemBackground`)
- El modo oscuro se adapta automáticamente
- Las superficies elevadas se vuelven más claras en modo oscuro (opuesto a la luz)
- Las safe areas varían por dispositivo — nunca hardcodees márgenes

### Tokens Específicos de Android (Material Design 3)

**Escala de tipos:**
| Rol | Nombre M3 | Tamaño por defecto |
|---|---|---|
| Display | Display Large | 57sp |
| Heading 1 | Headline Large | 32sp |
| Heading 2 | Headline Medium | 28sp |
| Heading 3 | Title Large | 22sp |
| Body | Body Large | 16sp |
| Body small | Body Medium | 14sp |
| Caption | Body Small | 12sp |
| Label | Label Large | 14sp |

**Escala de formas:**
- Extra-Small: 4dp (inputs)
- Small: 8dp (botones)
- Medium: 12dp (cards)
- Large: 16dp (modales)
- Extra-Large: 28dp (bottom sheets)
- Full: circular

---

## Ambos (Web + Mobile)

### Estrategia
Define tokens a un nivel abstracto, luego mapea a valores específicos de plataforma.

```
Token abstracto      →  Web (CSS)           →  iOS (Swift)        →  Android (Compose)
color-primary        →  --color-primary     →  .accentColor       →  MaterialTheme.colorScheme.primary
type-body            →  font-size: 1rem     →  .body              →  Typography.bodyLarge
spacing-4            →  1rem (16px)         →  16pt               →  16.dp
radius-md            →  0.375rem (6px)      →  6pt                →  8.dp (M3 Small)
```

### Decisiones Compartidas
- Paleta de colores: idéntica entre plataformas (mismos valores hex)
- Roles tipográficos: mismos nombres semánticos, tamaños específicos de plataforma
- Escala de espaciado: mismas proporciones, pueden diferir en valores absolutos (pt vs px vs dp)
- Sombras: peso visual similar, implementación nativa de plataforma

### Diferencias de Plataforma a Documentar

| Aspecto | Web | iOS | Android |
|---|---|---|---|
| Tamaño base de fuente | 16px | 17pt | 16sp |
| Touch target mínimo | 44px | 44pt | 48dp |
| Safe areas | Ninguna (viewport) | Dynamic Island, home indicator | Status bar, nav bar |
| Modo oscuro | CSS `prefers-color-scheme` o `data-theme` | `UITraitCollection.userInterfaceStyle` | `isSystemInDarkTheme()` |
| Texto dinámico | Unidades `rem` + media queries | Dynamic Type (obligatorio) | Unidades `sp` (escalables) |
| Indicadores de foco | Anillo de foco visible (WCAG) | VoiceOver cursor | TalkBack focus |
| Haptics | N/A | `UIImpactFeedbackGenerator` | `HapticFeedback` |

### Notas Específicas de Flutter

Cuando el proyecto usa Flutter para mobile:
- Los tokens se mapean a `ThemeData` y `ColorScheme`
- La tipografía usa `TextTheme` con estilos nombrados
- Espaciado via `EdgeInsets` y `SizedBox`
- Material 3 via `useMaterial3: true`
- Los widgets adaptativos manejan las diferencias iOS/Android

```dart
// Token → Flutter mapping
final colorScheme = ColorScheme(
  primary: Color(0xFF2563EB),    // color-primary
  onPrimary: Color(0xFFFFFFFF),  // color-text-on-primary
  surface: Color(0xFFFFFFFF),    // color-surface
  // ...
);
```
