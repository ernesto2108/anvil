# Plantilla de Output — design-system.md

Usa exactamente esta estructura al producir `<docs>/01-project/design-system.md`.

```markdown
# Design System — <Nombre del Proyecto>

> Última actualización: <fecha>
> Plataforma: <web | mobile | ambas>
> Framework: <Tailwind | Material | custom | etc.>

## Foundations

### Paleta de Colores (Primitivos)

#### Brand

| Token | Valor | Swatch |
|---|---|---|
| `brand-50` | #... | |
| `brand-100` | #... | |
| ... | | |
| `brand-950` | #... | |

#### Neutral

| Token | Valor | Swatch |
|---|---|---|
| `gray-50` | #... | |
| ... | | |
| `gray-950` | #... | |

#### Status

| Token | Valor | Uso |
|---|---|---|
| `red-500` | #... | Base de danger |
| `red-600` | #... | Hover de danger |
| `amber-500` | #... | Base de warning |
| `amber-600` | #... | Hover de warning |
| `green-500` | #... | Base de success |
| `green-600` | #... | Hover de success |
| `blue-500` | #... | Base de info |
| `blue-600` | #... | Hover de info |

### Tipografía

**Font family:** <fuente principal>, <stack de fallback>
**Mono:** <fuente mono>, <stack de fallback>

| Rol | Tamaño | Peso | Line-height | Uso |
|---|---|---|---|---|
| Display | 48px | 700 | 1.0 | Hero, landing |
| Heading 1 | 36px | 700 | 1.111 | Título de página |
| Heading 2 | 30px | 600 | 1.2 | Título de sección |
| Heading 3 | 24px | 600 | 1.333 | Subsección |
| Heading 4 | 20px | 500 | 1.4 | Título de card |
| Body | 16px | 400 | 1.5 | Texto por defecto |
| Body small | 14px | 400 | 1.429 | Texto secundario |
| Label | 14px | 500 | 1.429 | Labels de formulario, botones |
| Caption | 12px | 400 | 1.333 | Texto de ayuda |
| Code | 14px | 400 | 1.5 | Bloques de código |

### Espaciado

Unidad base: 4px (0.25rem)

| Token | Valor | Uso |
|---|---|---|
| `spacing-1` | 4px | Gaps ajustados |
| `spacing-2` | 8px | Padding compacto |
| `spacing-3` | 12px | Padding de inputs |
| `spacing-4` | 16px | Padding de componentes |
| `spacing-6` | 24px | Padding de cards |
| `spacing-8` | 32px | Gaps de secciones |
| `spacing-12` | 48px | Secciones grandes |
| `spacing-16` | 64px | Espaciado de página |

### Border Radius

| Token | Valor | Uso |
|---|---|---|
| `radius-sm` | 4px | Inputs, elementos pequeños |
| `radius-md` | 6px | Botones (por defecto) |
| `radius-lg` | 8px | Cards |
| `radius-xl` | 12px | Cards grandes, contenedores |
| `radius-full` | 9999px | Pills, avatares |

### Sombras

| Token | Valor | Uso |
|---|---|---|
| `shadow-xs` | `0 1px 2px 0 rgb(0 0 0 / 0.05)` | Elevación sutil |
| `shadow-sm` | `0 1px 3px ...` | Botones, cards |
| `shadow-md` | `0 4px 6px ...` | Dropdowns |
| `shadow-lg` | `0 10px 15px ...` | Modales |
| `shadow-xl` | `0 20px 25px ...` | Toasts, flotantes |

### Breakpoints

| Token | Valor | Objetivo |
|---|---|---|
| `sm` | 640px | Teléfono landscape |
| `md` | 768px | Tablet portrait |
| `lg` | 1024px | Tablet landscape |
| `xl` | 1280px | Desktop |
| `2xl` | 1536px | Desktop grande |

### Z-Index

| Token | Valor | Uso |
|---|---|---|
| `z-dropdown` | 10 | Dropdowns |
| `z-sticky` | 20 | Headers sticky |
| `z-overlay` | 30 | Backdrops |
| `z-modal` | 40 | Modales |
| `z-popover` | 50 | Tooltips |
| `z-toast` | 60 | Notificaciones |

---

## Tokens Semánticos

### Roles de Color

#### Brand

| Token | Light | Dark | Uso |
|---|---|---|---|
| `color-primary` | `{brand-600}` | `{brand-400}` | Acciones primarias |
| `color-primary-hover` | `{brand-700}` | `{brand-300}` | Hover primario |
| `color-primary-subtle` | `{brand-50}` | `{brand-950}` | Acentos suaves |
| `color-secondary` | ... | ... | Acciones secundarias |
| `color-accent` | ... | ... | Decorativo |

#### Status

| Token | Light | Dark | Uso |
|---|---|---|---|
| `color-success` | `{green-600}` | `{green-400}` | Positivo |
| `color-warning` | `{amber-600}` | `{amber-400}` | Precaución |
| `color-danger` | `{red-600}` | `{red-400}` | Error/destructivo |
| `color-info` | `{blue-600}` | `{blue-400}` | Informacional |

#### Superficies

| Token | Light | Dark | Uso |
|---|---|---|---|
| `color-background` | `{gray-50}` | `{gray-950}` | Fondo de página |
| `color-surface` | `{white}` | `{gray-900}` | Cards |
| `color-surface-elevated` | `{white}` | `{gray-800}` | Modales |

#### Texto

| Token | Light | Dark | Contraste |
|---|---|---|---|
| `color-text-primary` | `{gray-900}` | `{gray-50}` | 4.5:1 |
| `color-text-secondary` | `{gray-600}` | `{gray-400}` | 4.5:1 |
| `color-text-disabled` | `{gray-300}` | `{gray-600}` | — |
| `color-text-inverse` | `{white}` | `{gray-950}` | 4.5:1 |

#### Bordes

| Token | Light | Dark | Uso |
|---|---|---|---|
| `color-border-default` | `{gray-200}` | `{gray-700}` | Cards, inputs |
| `color-border-strong` | `{gray-400}` | `{gray-500}` | Inputs activos |
| `color-border-subtle` | `{gray-100}` | `{gray-800}` | Divisores |
| `color-border-focus` | `{brand-500}` | `{brand-400}` | Anillos de foco |

### Roles de Espaciado

| Token | Valor | Uso |
|---|---|---|
| `space-input-padding-x` | `{spacing-3}` | Horizontal en inputs |
| `space-input-padding-y` | `{spacing-2}` | Vertical en inputs |
| `space-component-gap` | `{spacing-4}` | Entre componentes |
| `space-card-padding` | `{spacing-6}` | Dentro de cards |
| `space-section-gap` | `{spacing-8}` | Entre secciones |
| `space-page-padding` | `{spacing-4}` mobile / `{spacing-6}` desktop | Bordes de página |

---

## Accesibilidad

### Verificación de Contraste

| Combinación | Ratio | ¿Pasa? |
|---|---|---|
| text-primary sobre surface | X:1 | ✓/✗ |
| text-secondary sobre surface | X:1 | ✓/✗ |
| text-inverse sobre primary | X:1 | ✓/✗ |
| heading sobre surface | X:1 | ✓/✗ |
| text-primary sobre background | X:1 | ✓/✗ |

### Estados de Foco
- Anillo de foco: 2px solid `{color-border-focus}`, 2px offset
- Todos los elementos interactivos deben tener indicador de foco visible

### Targets Táctiles
- Mínimo: 44x44px (móvil), 24x24px (desktop con puntero)

---

## Supuestos y Decisiones

| Decisión | Elección | Justificación |
|---|---|---|
| Tamaño base de fuente | 16px | Por defecto del navegador, accesible |
| Base de espaciado | 4px | Estándar de la industria (Tailwind, Material) |
| Familia de grises | neutral | <razón> |
| Color primario | <valor> | <marca / preferencia del usuario> |
| Modo oscuro | sí/no/después | <razón> |

## Preguntas Abiertas

- [ ] <Decisiones pendientes>
```
