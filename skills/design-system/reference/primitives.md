# Referencia de Tokens Primitivos

Los primitivos son valores crudos SIN significado semántico. Son la capa de fundación — nunca son consumidos directamente por componentes.

Fuentes: Material Design 3, Adobe Spectrum, GitHub Primer, Tailwind CSS, IBM Carbon, Apple HIG, especificación W3C Design Tokens.

## Paleta de Colores

### Estructura

Cada familia de tono usa una escala de 11 pasos. El número indica luminosidad (50 = más claro, 950 = más oscuro).

```
{hue}-50   →  Más claro (fondos, fills sutiles)
{hue}-100  →
{hue}-200  →  Claro (hover states, bordes claros)
{hue}-300  →
{hue}-400  →  Medio-claro (elementos secundarios)
{hue}-500  →  Base (uso primario, iconos)
{hue}-600  →  Medio-oscuro (interactivo primario, botones)
{hue}-700  →  Oscuro (hover en botones oscuros, texto fuerte)
{hue}-800  →
{hue}-900  →  Muy oscuro (headings, alto contraste)
{hue}-950  →  Más oscuro (casi negro, superficies en modo oscuro)
```

### Familias de Tono Requeridas

| Familia | Propósito | Notas |
|---|---|---|
| **Brand primary** | Color principal de marca | El usuario debe proveer o elegir |
| **Brand secondary** | Color de marca de soporte | Opcional, deriva del primary si no se provee |
| **Neutral / Gray** | Texto, bordes, fondos | Elige una familia de grises: pure, cool, warm, slate, zinc |
| **Red** | Errores, acciones destructivas | El semántico `danger` mapea aquí |
| **Amber/Yellow** | Advertencias, precaución | El semántico `warning` mapea aquí |
| **Green** | Éxito, confirmación | El semántico `success` mapea aquí |
| **Blue** | Información, links | El semántico `info` mapea aquí (a menos que la marca sea azul) |

### Escala de Grises por Defecto (Tailwind neutral)

```
gray-50:   #fafafa
gray-100:  #f5f5f5
gray-200:  #e5e5e5
gray-300:  #d4d4d4
gray-400:  #a3a3a3
gray-500:  #737373
gray-600:  #525252
gray-700:  #404040
gray-800:  #262626
gray-900:  #171717
gray-950:  #0a0a0a
```

### Colores de Status por Defecto (Tailwind)

```
red-500:    #ef4444    (danger)
red-600:    #dc2626    (danger-hover)
amber-500:  #f59e0b    (warning)
amber-600:  #d97706    (warning-hover)
green-500:  #22c55e    (success)
green-600:  #16a34a    (success-hover)
blue-500:   #3b82f6    (info)
blue-600:   #2563eb    (info-hover)
```

---

## Escala Tipográfica

### Escala de Tamaños de Fuente

Basada en los valores por defecto de Tailwind. Cada tamaño se combina con un line-height recomendado.

| Token | Tamaño | Line-height | Uso típico |
|---|---|---|---|
| `text-xs` | 0.75rem (12px) | 1.333 (16px) | Captions, badges, texto de ayuda |
| `text-sm` | 0.875rem (14px) | 1.429 (20px) | Labels, texto secundario, celdas de tabla |
| `text-base` | 1rem (16px) | 1.5 (24px) | Texto de cuerpo (base) |
| `text-lg` | 1.125rem (18px) | 1.556 (28px) | Cuerpo grande, énfasis |
| `text-xl` | 1.25rem (20px) | 1.4 (28px) | Heading 4, labels de sección |
| `text-2xl` | 1.5rem (24px) | 1.333 (32px) | Heading 3 |
| `text-3xl` | 1.875rem (30px) | 1.2 (36px) | Heading 2 |
| `text-4xl` | 2.25rem (36px) | 1.111 (40px) | Heading 1 |
| `text-5xl` | 3rem (48px) | 1.0 (48px) | Display, hero |
| `text-6xl` | 3.75rem (60px) | 1.0 (60px) | Display grande |

### Escala de Pesos de Fuente

| Token | Valor | Uso |
|---|---|---|
| `font-light` | 300 | Decorativo, texto de display |
| `font-normal` | 400 | Texto de cuerpo (por defecto) |
| `font-medium` | 500 | Labels, énfasis |
| `font-semibold` | 600 | Headings, botones |
| `font-bold` | 700 | Headings fuertes, CTAs |

### Font Families por Defecto

| Token | Stack | Cuándo usar |
|---|---|---|
| `font-sans` | `-apple-system, BlinkMacSystemFont, 'Segoe UI', 'Noto Sans', Helvetica, Arial, sans-serif` | Por defecto para UI |
| `font-mono` | `ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Consolas, monospace` | Código, técnico |
| `font-serif` | `Georgia, Cambria, 'Times New Roman', Times, serif` | Editorial, decorativo |

**Nota:** Reemplaza con fuentes específicas del proyecto cuando la marca lo requiera (ej., Inter, IBM Plex, personalizada).

---

## Escala de Espaciado

Unidad base: **4px** (0.25rem). Cada valor de espaciado es un múltiplo de 4px.

| Token | Valor | px | Uso típico |
|---|---|---|---|
| `spacing-0` | 0 | 0 | Reset |
| `spacing-0.5` | 0.125rem | 2px | Gaps de hairline |
| `spacing-1` | 0.25rem | 4px | Padding interno ajustado |
| `spacing-1.5` | 0.375rem | 6px | Gap icono-texto |
| `spacing-2` | 0.5rem | 8px | Padding compacto, gaps pequeños |
| `spacing-3` | 0.75rem | 12px | Padding de input, gaps de lista |
| `spacing-4` | 1rem | 16px | Padding de componente (por defecto) |
| `spacing-5` | 1.25rem | 20px | Gap medio |
| `spacing-6` | 1.5rem | 24px | Gaps de sección, padding de card |
| `spacing-8` | 2rem | 32px | Gaps grandes |
| `spacing-10` | 2.5rem | 40px | Separación de sección |
| `spacing-12` | 3rem | 48px | Espaciado de sección grande |
| `spacing-16` | 4rem | 64px | Espaciado a nivel de página |
| `spacing-20` | 5rem | 80px | Espaciado hero |
| `spacing-24` | 6rem | 96px | Cortes de sección mayores |

---

## Escala de Border Radius

| Token | Valor | Uso |
|---|---|---|
| `radius-none` | 0 | Esquinas filosas |
| `radius-xs` | 0.125rem (2px) | Redondeado sutil (badges) |
| `radius-sm` | 0.25rem (4px) | Inputs, cards pequeñas |
| `radius-md` | 0.375rem (6px) | Botones, redondeado por defecto |
| `radius-lg` | 0.5rem (8px) | Cards, modales |
| `radius-xl` | 0.75rem (12px) | Cards grandes, contenedores |
| `radius-2xl` | 1rem (16px) | Feature cards, secciones hero |
| `radius-3xl` | 1.5rem (24px) | Paneles prominentes |
| `radius-full` | 9999px | Pills, avatares, circular |

---

## Escala de Sombra / Elevación

Sombras de doble capa (key + ambient) siguiendo el patrón de Material Design 3.

| Token | Valor | Uso |
|---|---|---|
| `shadow-xs` | `0 1px 2px 0 rgb(0 0 0 / 0.05)` | Elevación sutil (cards en reposo) |
| `shadow-sm` | `0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)` | Botones, cards pequeñas |
| `shadow-md` | `0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)` | Dropdowns, popovers |
| `shadow-lg` | `0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)` | Modales, diálogos |
| `shadow-xl` | `0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)` | Notificaciones toast, paneles flotantes |

---

## Breakpoints

| Token | Valor | Objetivo |
|---|---|---|
| `breakpoint-sm` | 640px | Teléfonos grandes (landscape) |
| `breakpoint-md` | 768px | Tablets (portrait) |
| `breakpoint-lg` | 1024px | Tablets (landscape), laptops pequeñas |
| `breakpoint-xl` | 1280px | Desktops |
| `breakpoint-2xl` | 1536px | Desktops grandes |

**Enfoque mobile-first:** los estilos base apuntan a mobile. Usa breakpoints `min-width` para escalar hacia arriba.

---

## Capas de Z-Index

| Token | Valor | Uso |
|---|---|---|
| `z-base` | 0 | Contenido por defecto |
| `z-dropdown` | 10 | Dropdowns, autocompletado |
| `z-sticky` | 20 | Headers sticky, toolbars |
| `z-overlay` | 30 | Backdrops de overlay |
| `z-modal` | 40 | Diálogos modales |
| `z-popover` | 50 | Popovers, tooltips |
| `z-toast` | 60 | Notificaciones toast |

---

## Escala de Duración / Timing

| Token | Valor | Uso |
|---|---|---|
| `duration-instant` | 0ms | Inmediato (sin transición) |
| `duration-fast` | 150ms | Micro-interacciones (hover, focus) |
| `duration-normal` | 300ms | Transiciones estándar (abrir/cerrar, fade) |
| `duration-slow` | 500ms | Animaciones complejas (transiciones de página, modales) |
| `ease-default` | `cubic-bezier(0.4, 0, 0.2, 1)` | Propósito general (estándar Material) |
| `ease-in` | `cubic-bezier(0.4, 0, 1, 1)` | Elementos saliendo |
| `ease-out` | `cubic-bezier(0, 0, 0.2, 1)` | Elementos entrando |
| `ease-in-out` | `cubic-bezier(0.4, 0, 0.2, 1)` | Elementos moviéndose |
