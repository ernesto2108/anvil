# Guía DESIGN.md

Formato estándar para el archivo `DESIGN.md` que vive en la raíz del repo. Es el contrato portable del design system — cualquier agente AI lo lee automáticamente al abrir el proyecto, sin onboarding manual de tokens.

## Cuándo generarlo

- El DTD incluye tokens nuevos o modificados (colores, tipografía, spacing, componentes) → generar/actualizar `DESIGN.md`
- El DTD solo modifica pantallas sin cambiar tokens → omitir este paso
- Siempre en sync con el archivo `.pen` — si cambian tokens en una sesión futura, actualizar `DESIGN.md` en la misma invocación

## Cómo derivar los valores

Tomar los valores directamente de los tokens definidos en el DTD (sección "Design Tokens"). No re-leer el archivo `.pen` — los valores ya están en contexto.

## Template

```markdown
---
version: "1.0"
generated: "<fecha ISO>"
tool: "pencil"
colors:
  primary:
    "50": "<hex>"
    "100": "<hex>"
    "200": "<hex>"
    "300": "<hex>"
    "400": "<hex>"
    "500": "<hex>"
    "600": "<hex>"
    "700": "<hex>"
    "800": "<hex>"
    "900": "<hex>"
    "950": "<hex>"
  neutral:
    "50": "<hex>"
    "100": "<hex>"
    "200": "<hex>"
    "300": "<hex>"
    "400": "<hex>"
    "500": "<hex>"
    "600": "<hex>"
    "700": "<hex>"
    "800": "<hex>"
    "900": "<hex>"
    "950": "<hex>"
  status:
    success: "<hex>"
    warning: "<hex>"
    danger: "<hex>"
    info: "<hex>"
typography:
  fontFamily:
    heading: "<Google Font name>"
    body: "<Google Font name>"
    mono: "<mono font>"
  scale:
    display: { size: "<px>", lineHeight: "<px>", weight: "<num>" }
    "3xl":   { size: "<px>", lineHeight: "<px>", weight: "<num>" }
    "2xl":   { size: "<px>", lineHeight: "<px>", weight: "<num>" }
    xl:      { size: "<px>", lineHeight: "<px>", weight: "<num>" }
    lg:      { size: "<px>", lineHeight: "<px>", weight: "<num>" }
    base:    { size: "<px>", lineHeight: "<px>", weight: "<num>" }
    sm:      { size: "<px>", lineHeight: "<px>", weight: "<num>" }
    xs:      { size: "<px>", lineHeight: "<px>", weight: "<num>" }
spacing:
  "1": "<px>"
  "2": "<px>"
  "4": "<px>"
  "6": "<px>"
  "8": "<px>"
  "12": "<px>"
  "16": "<px>"
borderRadius:
  none: "0"
  sm: "<px>"
  md: "<px>"
  lg: "<px>"
  full: "9999px"
components:
  Button:
    backgroundColor: "{colors.primary.500}"
    textColor: "#ffffff"
    padding: "{spacing.3} {spacing.6}"
    borderRadius: "{borderRadius.md}"
    variants:
      hover:
        backgroundColor: "{colors.primary.600}"
      disabled:
        backgroundColor: "{colors.neutral.200}"
        textColor: "{colors.neutral.400}"
  # ... resto de componentes del DTD
modes:
  light:
    background: "{colors.neutral.50}"
    surface: "#ffffff"
    text: "{colors.neutral.900}"
    textMuted: "{colors.neutral.500}"
    border: "{colors.neutral.200}"
  dark:
    background: "{colors.neutral.950}"
    surface: "{colors.neutral.900}"
    text: "{colors.neutral.50}"
    textMuted: "{colors.neutral.400}"
    border: "{colors.neutral.800}"
---

# Design System — <Project Name>

## Overview

<2-3 líneas: dominio del proyecto, audiencia, personalidad de marca y tono visual.>

## Colors

<Lógica de la paleta: por qué se eligió ese color primario, qué transmite, cómo se usa en contexto.>

## Typography

**Heading:** <Font name> — <justificación>
**Body:** <Font name> — <justificación>

La escala tipográfica va de `display` (hero sections) hasta `xs` (captions y metadata).

## Layout

<Sistema de espaciado: base unit, cuándo usar cada token, ritmo visual.>

## Elevation & Depth

<Sombras y elevación. Omitir sección si no se definieron.>

## Shapes

<Border radius: cuándo usar sm vs md vs lg. Tono general (cuadrado vs redondeado) y por qué.>

## Components

<Lista los componentes principales con sus variantes y cuándo usar cada uno.>

## Do's and Don'ts

**Do:**
- Usar referencias de token (`{colors.primary.500}`), nunca hex hardcodeados
- Verificar contraste WCAG AA en todos los modos antes de publicar
- Reutilizar componentes — instancias con overrides, no duplicados

**Don't:**
- No mezclar escalas (usar spacing tokens, no px arbitrarios)
- No usar colores de estado (success/warning/danger) fuera de su propósito semántico
- No crear variantes de componente cuando un override de instancia alcanza
```

## Validación opcional

Si `design-md` CLI está disponible en el entorno:

```bash
npx design-md lint DESIGN.md
```

Reporta: referencias rotas, contraste WCAG insuficiente, secciones faltantes, orden incorrecto. Si el CLI no está disponible, omitir — el archivo es válido sin el lint.

## Routing de referencia

| Trabajando en... | Cargar |
|---|---|
| Generar DESIGN.md por primera vez | Este archivo completo |
| Actualizar tokens existentes | Solo la sección YAML frontmatter |
| Validar estructura | Sección "Validación opcional" |
