---
name: design-to-code
description: Traducir diseños aprobados a código de producción con fidelidad visual. Funciona con cualquier herramienta de diseño (Pencil, Figma) y stack (web CSS/React/Astro, Flutter, SwiftUI). Produce un inventario de diseño verificable antes de codificar, sincroniza tokens y mapea componentes a código. La carga el propio developer del stack. Usar cuando el usuario diga "implementa este diseño", "codifica esto", "traduce a código", "design to code", o después de que un diseño sea aprobado.
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->


# Design to Code

> Traduce diseños aprobados a código de producción con fidelidad visual. Independiente de la herramienta y del stack. La carga el developer del stack (web/Flutter/SwiftUI) que la ejecuta directamente — no delega a nadie.

## Filosofía

1. **Entender antes de codificar** — los gaps visuales nacen de leer el frame "a ojo". El inventario de diseño (Paso 2) es el mecanismo central anti-gaps: convierte el diseño en una checklist verificable antes de escribir una línea.
2. **El diseño es la fuente de verdad** — si el código luce distinto al diseño, el código está mal hasta que el humano diga lo contrario.
3. **Tokens antes que componentes** — sincroniza siempre las variables/tokens antes de mapear componentes; codificar valores sueltos es deuda.

## Prerrequisitos

- El diseño debe estar aprobado (nunca codificar antes de la aprobación del diseño).
- El archivo/referencia de diseño debe estar accesible (`.pen` abierto, URL de Figma, o spec/mockup).

## Paso 0: Detectar la herramienta de diseño

| Señal | Herramienta | Cómo leer |
|---|---|---|
| Archivo `.pen` abierto | Pencil | Herramientas MCP de Pencil (`get_variables`, `batch_get`, `get_screenshot`) — **solo lectura** |
| URL de Figma proporcionada | Figma | Herramientas MCP de Figma (carga `/figma-use` primero) |
| Especificación / mockup estático | Manual | Leer dimensiones y tokens desde el documento |

## Paso 1: Comparar diseño vs. código existente (OBLIGATORIO si ya existe código)

Antes de escribir código, crea una comparación sección por sección:

| Sección | Diseño | Código | Diferencia | Acción |
|---|---|---|---|---|
| Hero | Flujo lineal | Bento grid | Estructura difiere | Alinear al diseño |
| Nav | Hamburguesa + X | Solo hamburguesa | Estado faltante | Agregar animación X |

Presenta la tabla. Si divergen, analiza qué versión es mejor para UX (no solo listes diferencias), explica por qué y obtén aprobación antes de codificar.

## Paso 2: Inventario de diseño (OBLIGATORIO — antes de escribir código)

Extrae del diseño una lista verificable. Este inventario se **declara** (en el handoff si existe, o en el run) y se usa como **checklist de cierre**: cada ítem termina "implementado" o "no aplica + razón". No adivines — léelo del frame.

Categorías a inventariar:

- **Secciones / elementos** del frame (cada bloque visual).
- **Estados interactivos:** hover, focus, empty, loading, error, disabled, selected.
- **Variantes:** dark/light, breakpoints (web) o tamaños de pantalla/orientación (mobile).
- **Tokens usados:** colores, tipografía, espaciado, radios, sombras.
- **Interacciones / animaciones:** transiciones, gestos, feedback.
- **Assets:** iconos, imágenes, ilustraciones.

> Los gaps de UI casi siempre son ítems de este inventario que nunca se leyeron. Si un ítem no se implementa, debe quedar "no aplica + razón" explícita — nunca simplemente omitido.

## Paso 3: Sincronizar tokens de diseño

**Leer tokens del diseño:**
- Pencil: `get_variables()` — variables con tipos y valores temáticos.
- Figma: variables/estilos desde el archivo via MCP.

**Leer tokens del código** según stack (ver Paso 4 para el destino por stack).

**Diff y corrección:**
1. Faltante en código: token en diseño ausente en código → agregarlo.
2. Discrepancia de valor: mismo nombre, valor distinto → señalar al usuario.
3. Faltante en diseño: token solo en código → puede estar bien (utilidades).

Presenta el diff. Corrige las discrepancias antes de continuar.

## Paso 4: Leer pantalla + mapear componentes al stack

**Leer la pantalla objetivo:**
- Pencil: `batch_get` del frame a profundidad 2-3 + `get_screenshot`.
- Figma: estructura de frame/página + captura o inspección de nodos.

Identifica las secciones y mapea cada una a un componente/widget/view del stack. Usa la tabla del stack objetivo:

### Web (CSS)

| Propiedad de diseño | CSS |
|---|---|
| Layout vertical / horizontal | `flex-direction: column` / `row` |
| Gap con token | `gap: var(--token)` |
| Fill con token | `background: var(--token)` |
| Corner radius | `border-radius: var(--token)` |
| Border/stroke | `border: 1px solid var(--token)` |
| Padding (array) | `padding: var(--t) var(--r) var(--b) var(--l)` |
| Fill container / fit | `width: 100%` o `flex: 1` / `width: fit-content` |
| Space between / centrado | `justify-content: space-between` / `align-items: center` |

### Flutter

| Propiedad de diseño | Flutter |
|---|---|
| Tokens (color/tipografía) | `ThemeExtension` / `ColorScheme` / `TextTheme` |
| Layout vertical / horizontal | `Column` / `Row` |
| Gap con token | `spacing:` (Flex) o `SizedBox` entre hijos |
| Fill con token | `Container(color:)` / `BoxDecoration` |
| Corner radius | `BorderRadius.circular(token)` |
| Border/stroke | `Border.all(color:, width:)` |
| Padding | `EdgeInsets.fromLTRB(...)` / `.all(token)` |
| Fill container / fit | `Expanded` / `double.infinity` / `MainAxisSize.min` |
| Space between / centrado | `MainAxisAlignment.spaceBetween` / `.center` |

### SwiftUI

| Propiedad de diseño | SwiftUI |
|---|---|
| Tokens (color) | Asset Catalog / extensiones de `Color` |
| Layout vertical / horizontal | `VStack` / `HStack` |
| Gap con token | `VStack(spacing: token)` / `HStack(spacing:)` |
| Fill con token | `.background(Color.token)` |
| Corner radius | `.clipShape(RoundedRectangle(cornerRadius:))` / `.cornerRadius()` |
| Border/stroke | `.overlay(RoundedRectangle().stroke())` |
| Padding | `.padding(.init(top:leading:bottom:trailing:))` |
| Fill container / fit | `.frame(maxWidth: .infinity)` / `.fixedSize()` |
| Dark mode | `@Environment(\.colorScheme)` |
| Space between / centrado | `Spacer()` / `.frame(alignment: .center)` |

## Paso 5: Implementar (el propio developer del stack escribe el código)

Esta skill la carga el developer del stack (`developer-frontend` para web, `developer-mobile` para Flutter/SwiftUI), que es quien escribe el código directamente con el inventario del Paso 2 y el mapeo del Paso 4. Reglas de implementación:

- **Usa tokens/variables semánticas**, nunca valores codificados. Si un token no tiene equivalente en el destino, agrégalo primero.
- **Reutiliza componentes existentes** — grep/verifica qué ya existe antes de crear.
- **Mobile-first (web):** codifica el layout móvil primero, agrega overrides desktop con `min-width`.
- **Carga la skill de convenciones del stack** (`astro-conventions`, `react-conventions`, `flutter-conventions`, `swift-conventions`).
- Recorre el inventario del Paso 2: cada ítem debe quedar "implementado" o "no aplica + razón".

### Insumos que esta skill deja preparados para el Auto-QA

El QA visual **NO ocurre dentro de esta skill** — ocurre en el `## Auto-QA` del agente host. Al terminar, deja listos:

- La referencia de diseño (`frame_id` + `pen_file`, URL Figma o screenshots).
- `impl_url_or_component` — dónde vive la implementación.
- El inventario del Paso 2 como checklist de cierre.

## Checklist de Completitud (recorrer antes de cerrar la implementación)

### Web

1. [ ] Cada variable CSS usada tiene valor en `:root` (claro) y en `.dark` (oscuro) si existe modo oscuro.
2. [ ] Si existe modo oscuro → un mecanismo JS activa/desactiva la clase `dark` en `<html>`.
3. [ ] Si existe toggle de tema → conectado al mecanismo y persistido en `localStorage`.
4. [ ] `prefers-color-scheme` se respeta como valor inicial por defecto.
5. [ ] Cada elemento interactivo (dropdowns, modales, menús) tiene implementación funcional, no solo visual.
6. [ ] Tipos request/response del frontend coinciden con los DTOs actuales del backend.
7. [ ] Íconos usan la librería del proyecto (ej. `lucide-react`), no SVGs inline.
8. [ ] Clases Tailwind usan sintaxis v4 si el proyecto usa v4: `(--var)` no `[var(--var)]`.

### Mobile (Flutter / SwiftUI)

1. [ ] Tokens definidos en el mecanismo de tema del stack (`ThemeExtension`/`ColorScheme` en Flutter; Asset Catalog / extensiones de `Color` en SwiftUI).
2. [ ] Si existe modo oscuro → theme switching por plataforma (Flutter `ThemeMode`/`MediaQuery.platformBrightness`; SwiftUI `@Environment(\.colorScheme)`).
3. [ ] Safe areas respetadas (`SafeArea` en Flutter; `.safeAreaInset`/layout guides en SwiftUI).
4. [ ] Tamaños de pantalla y orientación soportados según el inventario (breakpoints/`LayoutBuilder`, size classes).
5. [ ] Cada estado interactivo del inventario (loading, empty, error, disabled) tiene implementación real.
6. [ ] Íconos/assets del catálogo del proyecto, no hardcodeados.

## Reglas

- **Nunca adivines dimensiones** — léelas del diseño.
- **Nunca codifiques colores** — usa tokens/variables del stack.
- **El diseño es la fuente de verdad** — si el código luce distinto, el código está mal.
- **Cambios quirúrgicos** — al actualizar código existente, cambia solo lo que cambió en el diseño.
- **Tokens primero** — sincroniza variables antes de codificar componentes.
- **Inventario primero** — no escribas código sin el inventario del Paso 2.

## Anti-Patrones

| Anti-Patrón | Corrección |
|---|---|
| Codificar de memoria en vez de leer el diseño | Leer el diseño y producir el inventario del Paso 2 primero |
| Saltarse el inventario de diseño | Es obligatorio: los gaps nacen de leer el frame a ojo |
| Colores/dimensiones hardcodeados | Usar tokens del stack |
| Asumir el stack o la herramienta de diseño | Detectar desde marcadores del repo y extensión/URL |
| Implementar solo el estado default | Recorrer todos los estados/variantes del inventario |
