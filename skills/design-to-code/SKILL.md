---
name: design-to-code
description: Traducir diseños aprobados a código de producción. Funciona con cualquier herramienta de diseño (Pencil, Figma). Sincroniza tokens de diseño con CSS, mapea componentes a código y valida la fidelidad visual. Usar cuando el usuario diga "implementa este diseño", "codifica esto", "traduce a código", "design to code", o después de que un diseño sea aprobado.
---

# Design to Code

> Traduce diseños aprobados a código de producción con fidelidad visual. Independiente de la herramienta.

## Prerrequisitos

- El diseño debe estar aprobado (nunca codificar antes de la aprobación del diseño)
- El archivo de diseño debe estar abierto — usa `/design-project` primero si es necesario

## Paso 0: Detectar la herramienta de diseño

Determina en qué herramienta está el diseño:

| Señal | Herramienta | Cómo leer |
|---|---|---|
| Archivo `.pen` abierto | Pencil | Usa las herramientas MCP de Pencil (`get_variables`, `batch_get`, `get_screenshot`) |
| URL de Figma proporcionada | Figma | Usa las herramientas MCP de Figma (carga `/figma-use` primero) |
| Especificación de diseño / mockup estático | Manual | Lee dimensiones y tokens desde el documento de especificación |

Todos los pasos siguientes usan la API de la herramienta correspondiente pero el **output es el mismo**: CSS, HTML, componentes.

## Paso 0.5: Comparar diseño vs. código (OBLIGATORIO si ya existe código)

Antes de escribir cualquier código, crea una comparación sección por sección:

| Sección | Diseño | Código | Diferencia | Acción |
|---|---|---|---|---|
| Hero | Flujo lineal, sin tarjetas | Bento grid con tarjetas | Estructura difiere | Alinear al diseño |
| Nav | Hamburguesa + X | Solo hamburguesa | Estado faltante | Agregar animación X |

Presenta esta tabla al usuario. Si el diseño y el código divergen, **analiza qué versión es mejor para UX** — no solo liste diferencias. Toma lo mejor de cada uno, explica por qué, obtén aprobación antes de codificar.

## Paso 1: Sincronizar tokens de diseño

### Leer tokens del diseño:

- **Pencil**: `get_variables()` — devuelve todas las variables con tipos y valores temáticos
- **Figma**: leer variables/estilos desde el archivo de Figma via MCP

### Leer tokens del código:

Lee el archivo de variables CSS del proyecto (ej., `global.css`, `variables.css`, `tailwind.config`)

### Diff y corrección:

1. **Faltante en CSS**: tokens que existen en el diseño pero no en el código → agregarlos
2. **Discrepancia de valor**: mismo nombre pero valor diferente → señalar al usuario
3. **Faltante en diseño**: tokens en el código que no están en el diseño → puede estar bien (utilidades solo de código)

Presenta el diff. Corrige las discrepancias antes de continuar.

## Paso 2: Leer la pantalla objetivo

### Desde Pencil:
1. `batch_get` el frame de pantalla a profundidad 2-3
2. `get_screenshot` de la pantalla

### Desde Figma:
1. Lee la estructura de frame/página
2. Obtén una captura de pantalla o inspecciona las propiedades del nodo

### Luego (independiente de la herramienta):
3. Identifica las secciones y qué componentes se usan
4. Mapea cada sección a un componente de código (ej., Hero.astro, Navbar.tsx)

## Paso 3: Leer la estructura del componente

Para cada componente que necesita ser creado o actualizado:

### Desde Pencil:
`batch_get` el componente reutilizable a profundidad 3

### Desde Figma:
Lee las propiedades del componente, variantes y configuraciones de auto-layout

### Luego mapea a CSS (universal):

| Propiedad de diseño | CSS |
|---|---|
| Layout vertical | `flex-direction: column` |
| Layout horizontal | `flex-direction: row` |
| Gap con token | `gap: var(--token-name)` |
| Fill con token | `background: var(--token-name)` |
| Corner radius | `border-radius: var(--token-name)` |
| Border/stroke | `border: 1px solid var(--token-name)` |
| Array de padding | `padding: var(--top) var(--right) var(--bottom) var(--left)` |
| Fill container | `width: 100%` o `flex: 1` |
| Fit content | `width: fit-content` |
| Space between | `justify-content: space-between` |
| Alineación centrada | `align-items: center` |

## Paso 4: Delegar al developer del stack

Los developers de stack (`developer-frontend` para web React/TypeScript/Astro, `developer-mobile` para Flutter/Dart) son los ÚNICOS autorizados a escribir código de producción. Después de los pasos 1-3, lanza el developer del stack objetivo con:

1. **Token diff** — variables CSS nuevas/modificadas para agregar
2. **Component map** — qué componentes crear/actualizar, mapeados desde las secciones del diseño
3. **Propiedades de diseño** — layout, espaciado, colores, tipografía extraídos en el paso 3
4. **Screenshot** — captura de pantalla del diseño como referencia visual

Incluye este contexto INLINE en el prompt del agente (nunca le digas al agente "lee el archivo X").

Reglas para el developer del stack:
- **Usa CSS custom properties**, nunca valores codificados
- **Usa los mismos nombres semánticos** que los tokens de diseño
- **Si un token de diseño no tiene equivalente CSS**, agrégalo primero al archivo CSS
- **Mobile-first**: si existen diseños web y móvil, codifica el layout móvil primero, agrega overrides de desktop con media queries `min-width`
- **Reutiliza componentes existentes** — verifica qué ya existe en el codebase antes de crear nuevos
- **Carga la skill de convenciones apropiada** para el stack objetivo (ej., `astro-conventions`, `react-conventions`, `flutter-conventions`, `swift-conventions`)

## Paso 5: QA Visual de Fidelidad (OBLIGATORIO para tareas UI)

Después de implementar y antes de presentar al usuario:

1. **Verificación de build**: Ejecuta `build` para verificar que no hay errores.
2. **QA de fidelidad visual**: Invoca la skill `visual-fidelity-qa` con:
   - `frame_id` — Frame ID del diseño (viene de los inputs de la task)
   - `pen_file` — path al archivo `.pen` (viene de `Design reference`)
   - `impl_url_or_component` — la URL o componente implementado

   La skill produce un JSON con `score` e `issues` clasificados por severidad.

3. **Regla de entrega:**
   - Si `visual-fidelity-qa` reporta **BLOQUEADO** (issues críticos) → NO entregar al humano. Resolver primero los críticos (invocar `qa-fixer` si es necesario) y re-ejecutar esta skill.
   - Si reporta **APROBADO** o solo issues menores/cosméticos → incluir el reporte en el handoff.

4. **Estados, modos y viewports**: si el componente tiene estados interactivos, modos claro/oscuro, o variantes responsive, repite el paso 2 por cada uno (un frame de diseño contra su implementación correspondiente).

**Si la task NO trae `frame_id` ni `pen_file` y aun así toca UI visible:** preguntar al humano: *"Esta tarea toca UI visible pero no tiene Design reference ni Frame ID. ¿Puedes proveerlos para ejecutar QA de fidelidad visual, o confirmar que no aplica?"*. Solo proceder sin QA visual si el humano confirma explícitamente que no aplica.

**Solo presenta al usuario después de que el build pase y `visual-fidelity-qa` apruebe.**

## Checklist de Completitud Design-to-Code (OBLIGATORIO)

Después de implementar tokens de diseño y componentes, verifica:

1. [ ] Cada variable CSS usada en componentes tiene un valor tanto en `:root` (claro) como en `.dark` (modo oscuro)
2. [ ] Si existe modo oscuro en el diseño → un mecanismo JS activa/desactiva la clase `dark` en `<html>` (hook o store)
3. [ ] Si existe toggle de tema en el diseño → está conectado al mecanismo de toggle y persiste en `localStorage`
4. [ ] Preferencia del sistema: `prefers-color-scheme` se respeta como el valor inicial por defecto
5. [ ] Cada elemento interactivo en el diseño (dropdowns, modales, menús) tiene una implementación funcional, no solo visual
6. [ ] Los tipos de solicitud/respuesta del frontend coinciden con los DTOs actuales del backend (verificar después de cualquier cambio en el backend)
7. [ ] Todos los íconos usan la librería de íconos del proyecto (ej., `lucide-react`), no SVGs inline
8. [ ] Las clases de Tailwind usan sintaxis v4 si el proyecto usa Tailwind v4: `(--var)` no `[var(--var)]`

## Reglas

- **Nunca adivines dimensiones** — léelas desde el archivo de diseño
- **Nunca codifiques colores** — siempre usa variables CSS
- **El diseño es la fuente de verdad** — si el código luce diferente al diseño, el código está mal
- **Cambios quirúrgicos** — si actualizas código existente, cambia solo lo que cambió en el diseño
- **Sincronizar tokens primero** — siempre sincroniza las variables antes de codificar componentes
- **Output independiente de la herramienta** — CSS es CSS sin importar si el diseño vino de Pencil o Figma

## Anti-Patrones

| Anti-Patrón | Corrección |
|---|---|
| Codificar de memoria en lugar de leer el diseño | Siempre leer el archivo de diseño primero |
| Colores hex codificados en CSS | Usar `var(--color-name)` |
| Presentar código sin construir | Ejecutar build primero |
| Presentar sin comparación visual | Comparar captura del diseño con el browser |
| Implementar mobile sin verificar desktop | Verificar ambos viewports |
| Asumir la herramienta de diseño | Detectar desde extensión de archivo o URL |
