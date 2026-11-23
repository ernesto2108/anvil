# Guía de Estructura de Markup (Web)

Aplica a JSX de React y a cualquier HTML generado. El objetivo es evitar **div soup / over-nesting**: wrappers sin función, `z-index` arbitrarios, modals hechos con divs, contenedores intermedios que no aportan nada. Un markup puede verse pixel-perfect en un screenshot y aun así ser estructuralmente inválido — el QA visual NO lo detecta. Estas reglas sí, y son lintables.

## Filosofía

- **Todo wrapper justifica su existencia** — semántica, layout/estilo propio o boundary funcional (scroll, stacking). Si lo borras y nada cambia visual ni semánticamente, sobraba.
- **El elemento nativo antes que el div genérico** — el HTML ya trae el rol, el foco, el teclado y el anuncio de lector de pantalla. Reconstruirlo con divs es trabajo extra y peor accesibilidad.
- **El layout vive en el padre** — Grid/Flex + `gap` en el contenedor, no divs espaciadores ni wrappers intermedios que cambian quién es el flex/grid item.

---

## A. HTML semántico y landmarks

Fuentes: MDN (accesibilidad HTML), W3C ARIA APG (Landmark Regions).

### Usar el elemento nativo del propósito

| Propósito | Elemento | Por qué |
|---|---|---|
| Acción (dispara JS) | `<button>` | Foco, teclado (Enter/Space), rol implícito |
| Navegación (cambia URL) | `<a href>` | Rol link, teclado, menú contextual |
| Colección de ítems | `<ul>` / `<ol>` | El lector anuncia "lista de N ítems" |
| Pares nombre-valor | `<dl>` (`<dt>`/`<dd>`) | Semántica de definición |
| Datos tabulares | `<table>` | Encabezados, navegación por celdas |
| Imagen + leyenda | `<figure>` + `<figcaption>` | Asociación semántica |
| Caja de búsqueda | `<search>` | Landmark de búsqueda nativo |

```jsx
// MAL — div con onClick reconstruye un botón peor
<div className="btn" onClick={save}>Guardar</div>

// BIEN
<button type="button" onClick={save}>Guardar</button>
```

### Landmarks

- Exactamente **UN `<main>`** visible por página.
- `<header>`/`<footer>` a nivel top-level = `banner`/`contentinfo`.
- `<nav>` solo para navegación mayor; si hay varios, `aria-label` único en cada uno.
- `<section>` es landmark **solo si tiene nombre accesible** (un heading o `aria-label`); sin él, es un div con nombre bonito — usa `<div>` o dale heading.
- `<article>` para contenido autocontenido.
- Todo contenido perceptible debe vivir dentro de algún landmark.
- **No agregar roles ARIA redundantes** — `<nav role="navigation">` es ruido; el elemento ya trae el rol.

### Headings

- Un solo `<h1>` por página.
- Sin saltos hacia abajo (`h2` → `h4` está prohibido; usa `h3`).
- Nunca elegir el nivel por su tamaño visual — el nivel es jerarquía, el tamaño es CSS.

---

## B. Anti-div-soup — presupuestos y reglas

Fuente: Lighthouse "Avoid an excessive DOM size" (Chrome for Developers).

### Presupuestos DOM (Lighthouse)

| Métrica | Warning | Error | Objetivo interno |
|---|---|---|---|
| Nodos totales | > ~800 | > ~1.400 | < 800 |
| Profundidad de anidamiento | — | > 32 | ≤ 20 |
| Hijos directos por nodo | — | > ~60 | virtualizar/paginar si se supera |

### Regla central

Cada wrapper debe aportar una de tres cosas: **semántica**, **layout/estilo propio**, o **boundary funcional** (scroll container, stacking context). Test rápido: borra el div — si nada cambia visual ni semánticamente, sobraba.

### Presupuesto por componente

- Máx **~4 niveles** de anidamiento propio y **un solo wrapper raíz** (o ninguno con Fragment). Más que eso → extraer subcomponentes.
- Usa **Fragment** (`<>...</>`) en vez de un div wrapper: no crea nodo DOM y no rompe el flex/grid del padre.
- El layout se define en el padre con Grid/Flex + `gap` — **nunca** divs espaciadores ni contenedores intermedios que cambian quién es el flex/grid item.
- Lo decorativo sin contenido real (overlays de oscurecimiento, separadores) va en pseudo-elementos `::before`/`::after`, no en divs.

```jsx
// MAL — wrapper inútil + div espaciador
<div>
  <div className="wrapper">
    <Avatar />
    <div className="spacer" />
    <Name />
  </div>
</div>

// BIEN — el padre define el layout, sin wrappers de sobra
<div className="flex items-center gap-3">
  <Avatar />
  <Name />
</div>
```

---

## C. Superposición correcta

Fuentes: MDN (`<dialog>`, Popover API), CSS-Tricks (grid stacking).

### Tabla de decisión

| Necesidad | Solución | Por qué |
|---|---|---|
| Modal que bloquea la página | `<dialog>` + `showModal()` | Top layer + `inert`, focus trap, Esc, `::backdrop` gratis |
| Tooltip / menú / popover | Popover API (`popover` + `popovertarget`) | Top layer + light-dismiss nativo |
| Badge sobre avatar, texto sobre imagen | Padre `position: relative` + hijo `absolute` (alcance mínimo), o Grid stacking | Alcance acotado, sin z-index global |
| Capas del mismo tamaño (hero, crossfade) | CSS Grid, todos los hijos en `grid-area: 1/1` | Se apilan sin position absolute |

**Por qué nativo gana:** `dialog.showModal()` y los popovers van al **top layer** del navegador — por encima de todo stacking context, imposible de tapar con `z-index`. Un modal de divs con `z-index: 9999` puede quedar atrapado dentro de un stacking context ancestro (un padre con `transform` u `opacity`) y renderizar por debajo de otra cosa.

### z-index

- Solo con **escala de tokens**, nunca valores arbitrarios:

```css
--z-dropdown: 100;
--z-sticky: 200;
--z-overlay: 300;
--z-toast: 400;
```

- Modals y popovers nativos **no necesitan token** — ya están en el top layer.
- `isolation: isolate` en componentes con capas internas, para que sus z-index no compitan con la página.
- Recuerda: `transform` y `opacity` crean stacking contexts por accidente — es la causa #1 de "mi z-index no funciona".

```jsx
// MAL — modal de divs, z-index mágico, puede quedar tapado
<div className="modal" style={{ zIndex: 9999 }}>...</div>

// BIEN — top layer, focus trap y Esc gratis
<dialog ref={ref}>...</dialog>
// ref.current.showModal()
```

---

## D. Verificación lintable

### ESLint (`eslint-plugin-jsx-a11y`)
- `no-static-element-interactions`
- `no-noninteractive-element-interactions`
- `click-events-have-key-events`

### axe-core (jest-axe / @axe-core/playwright)
- `landmark-one-main`, `landmark-unique`
- `heading-order`, `page-has-heading-one`
- `region`, `list` / `listitem`

### Lighthouse CI
- Presupuesto DOM: < 800 nodos, profundidad ≤ 20.

### Checklist de review
- [ ] Un solo wrapper raíz por componente (o Fragment).
- [ ] ≤ 4 niveles de anidamiento propio.
- [ ] Cero `<div onClick>` / `<span onClick>` — elemento nativo.
- [ ] Cero z-index fuera de la escala de tokens.
- [ ] Todo `<section>` tiene heading o `aria-label`.
- [ ] `<nav>` repetido con `aria-label` único.
- [ ] Un `<main>`, un `<h1>`, headings sin saltos.
- [ ] Overlays con `<dialog>` / Popover / grid stacking, no divs + z-index.
