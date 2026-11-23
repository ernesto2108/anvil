# Referencia de Tokens Semánticos

Los tokens semánticos mapean primitivos a propósitos. Son la capa que consumen los componentes. Cuando el tema cambia (modo oscuro, rebranding), solo cambian las referencias a primitivos — los nombres semánticos permanecen estables.

Fuentes: Material Design 3 (arquitectura de 3 capas), GitHub Primer (functional tokens), Adobe Spectrum (alias tokens), IBM Carbon (role-based tokens), Apple HIG (semantic colors).

## Color — Brand

| Token | Propósito | Light por defecto | Dark por defecto |
|---|---|---|---|
| `color-primary` | Acción principal de marca, botones primarios, estados activos | `{brand-600}` | `{brand-400}` |
| `color-primary-hover` | Estado hover del primario | `{brand-700}` | `{brand-300}` |
| `color-primary-subtle` | Acento de fondo claro | `{brand-50}` | `{brand-950}` |
| `color-secondary` | Acciones de marca secundarias | `{secondary-600}` | `{secondary-400}` |
| `color-secondary-hover` | Estado hover del secundario | `{secondary-700}` | `{secondary-300}` |
| `color-accent` | Highlights, badges, decorativo | `{accent-500}` | `{accent-400}` |

## Color — Semántico (Status)

| Token | Propósito | Light por defecto | Dark por defecto |
|---|---|---|---|
| `color-success` | Confirmación positiva | `{green-600}` | `{green-400}` |
| `color-success-subtle` | Fondos de éxito | `{green-50}` | `{green-950}` |
| `color-warning` | Precaución, atención requerida | `{amber-600}` | `{amber-400}` |
| `color-warning-subtle` | Fondos de advertencia | `{amber-50}` | `{amber-950}` |
| `color-danger` | Errores, acciones destructivas | `{red-600}` | `{red-400}` |
| `color-danger-subtle` | Fondos de error | `{red-50}` | `{red-950}` |
| `color-info` | Informacional, links, acciones neutras | `{blue-600}` | `{blue-400}` |
| `color-info-subtle` | Fondos de info | `{blue-50}` | `{blue-950}` |

## Color — Superficie

Siguiendo el modelo de capas de IBM Carbon + el patrón de superficies elevadas de Apple HIG.

| Token | Propósito | Light por defecto | Dark por defecto |
|---|---|---|---|
| `color-background` | Fondo de página (capa más baja) | `{white}` o `{gray-50}` | `{gray-950}` |
| `color-surface` | Cards, paneles (capa 1) | `{white}` | `{gray-900}` |
| `color-surface-elevated` | Modales, popovers (capa 2) | `{white}` | `{gray-800}` |
| `color-surface-overlay` | Backdrop detrás de modales | `{black / 50%}` | `{black / 60%}` |
| `color-surface-subtle` | Rayas de zebra, fondos de hover | `{gray-50}` | `{gray-800}` |

**Regla del modo oscuro (de Apple + Carbon):** las superficies elevadas se vuelven MÁS CLARAS, no más oscuras. Esto transmite profundidad.

## Color — Texto

Siguiendo el patrón de texto jerárquico de GitHub Primer + la jerarquía de labels de Apple HIG.

| Token | Propósito | Light por defecto | Dark por defecto | Req. de contraste |
|---|---|---|---|---|
| `color-text-primary` | Texto de contenido principal | `{gray-900}` | `{gray-50}` | 4.5:1 sobre surface |
| `color-text-secondary` | Texto de soporte, descripciones | `{gray-600}` | `{gray-400}` | 4.5:1 sobre surface |
| `color-text-tertiary` | Placeholder, hints deshabilitados | `{gray-400}` | `{gray-500}` | 3:1 sobre surface |
| `color-text-disabled` | Controles deshabilitados | `{gray-300}` | `{gray-600}` | Sin mínimo (intencionalmente bajo) |
| `color-text-inverse` | Texto en botones filled, superficies oscuras | `{white}` | `{gray-950}` | 4.5:1 sobre primary |
| `color-text-link` | Hipervínculos | `{brand-600}` | `{brand-400}` | 4.5:1 sobre surface |
| `color-text-on-primary` | Texto sobre fondos de color primary | `{white}` | `{white}` | 4.5:1 sobre primary |
| `color-text-on-danger` | Texto sobre fondos de danger | `{white}` | `{white}` | 4.5:1 sobre danger |

## Color — Borde

Siguiendo la guía de uso de grises de Adobe Spectrum + los tokens de border de GitHub Primer.

| Token | Propósito | Light por defecto | Dark por defecto |
|---|---|---|---|
| `color-border-default` | Bordes estándar (cards, inputs) | `{gray-200}` | `{gray-700}` |
| `color-border-strong` | Bordes enfatizados, inputs activos | `{gray-400}` | `{gray-500}` |
| `color-border-subtle` | Divisores, separadores | `{gray-100}` | `{gray-800}` |
| `color-border-focus` | Anillos de foco (accesibilidad) | `{brand-500}` | `{brand-400}` |
| `color-border-danger` | Inputs en estado de error | `{red-500}` | `{red-400}` |

---

## Tipografía — Roles

Cada rol mapea un propósito semántico a una combinación primitiva de tamaño + peso + line-height.

Basado en los roles de tipo de Material Design 3, adaptado para web.

| Token | Tamaño de fuente | Peso | Line-height | Uso |
|---|---|---|---|---|
| `type-display` | `{text-5xl}` 48px | `{font-bold}` 700 | 1.0 | Headlines hero, landing page |
| `type-heading-1` | `{text-4xl}` 36px | `{font-bold}` 700 | 1.111 | Título de página (h1) |
| `type-heading-2` | `{text-3xl}` 30px | `{font-semibold}` 600 | 1.2 | Título de sección (h2) |
| `type-heading-3` | `{text-2xl}` 24px | `{font-semibold}` 600 | 1.333 | Título de subsección (h3) |
| `type-heading-4` | `{text-xl}` 20px | `{font-medium}` 500 | 1.4 | Título de card, sub-subsección (h4) |
| `type-body` | `{text-base}` 16px | `{font-normal}` 400 | 1.5 | Texto de contenido por defecto |
| `type-body-small` | `{text-sm}` 14px | `{font-normal}` 400 | 1.429 | Contenido secundario, celdas de tabla |
| `type-label` | `{text-sm}` 14px | `{font-medium}` 500 | 1.429 | Labels de formulario, labels de tab, botones |
| `type-caption` | `{text-xs}` 12px | `{font-normal}` 400 | 1.333 | Texto de ayuda, timestamps, badges |
| `type-code` | `{text-sm}` 14px | `{font-normal}` 400 | 1.5 | Bloques de código, datos técnicos |

---

## Espaciado — Semántico

| Token | Primitivo | Uso |
|---|---|---|
| `space-input-padding-x` | `{spacing-3}` 12px | Padding horizontal dentro de inputs/botones |
| `space-input-padding-y` | `{spacing-2}` 8px | Padding vertical dentro de inputs/botones |
| `space-input-gap` | `{spacing-2}` 8px | Gap entre label e input |
| `space-component-gap` | `{spacing-4}` 16px | Gap entre componentes hermanos |
| `space-card-padding` | `{spacing-6}` 24px | Padding interno de card |
| `space-section-gap` | `{spacing-8}` 32px | Gap entre secciones de página |
| `space-page-padding` | `{spacing-4}` 16px mobile, `{spacing-6}` 24px desktop | Padding en bordes de página |
| `space-stack-sm` | `{spacing-2}` 8px | Stack vertical ajustado (campos de formulario) |
| `space-stack-md` | `{spacing-4}` 16px | Stack vertical normal |
| `space-stack-lg` | `{spacing-6}` 24px | Stack vertical suelto (secciones) |
| `space-inline-sm` | `{spacing-1}` 4px | Horizontal ajustado (icono + texto) |
| `space-inline-md` | `{spacing-2}` 8px | Horizontal normal (botones en fila) |
| `space-inline-lg` | `{spacing-4}` 16px | Horizontal suelto (items de nav) |

---

## Estados Interactivos

Estos tokens modifican colores base para diferentes estados de interacción.

| Token | Propósito | Patrón |
|---|---|---|
| `state-hover` | Mouse hover | Color base + 1 tono más oscuro (light) / más claro (dark) |
| `state-active` | Mouse down, presionado | Color base + 2 tonos más oscuro/claro |
| `state-focus` | Foco de teclado | Anillo `{color-border-focus}`, 2px offset |
| `state-disabled` | No interactivo | 40% de opacidad del base, sin pointer events |
| `state-selected` | Ítem seleccionado/activo | Fondo `{color-primary-subtle}` + indicador `{color-primary}` |

---

## Lista de Verificación de Contraste

Cada combinación texto/superficie DEBE cumplir:

| Combinación | Ratio requerido | Estándar |
|---|---|---|
| `text-primary` sobre `surface` | 4.5:1 | WCAG AA texto normal |
| `text-primary` sobre `background` | 4.5:1 | WCAG AA texto normal |
| `text-secondary` sobre `surface` | 4.5:1 | WCAG AA texto normal |
| `text-inverse` sobre `primary` | 4.5:1 | WCAG AA texto normal |
| `heading-1` sobre `surface` | 3:1 | WCAG AA texto grande (>24px) |
| `text-tertiary` sobre `surface` | 3:1 | WCAG AA texto grande / componentes UI |
| Cualquier icono sobre su fondo | 3:1 | WCAG AA componentes UI |
| Anillo de foco contra surface | 3:1 | WCAG 2.2 apariencia de foco |

**Herramienta:** Usa verificador de contraste (WebAIM o similar) para verificar. Nunca asumas — siempre calcula.
