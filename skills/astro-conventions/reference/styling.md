# Estilos en Astro

## CSS Scoped (Por Defecto)

`<style>` en archivos `.astro` se aplica automáticamente con scope via atributos de datos:

```astro
<style>
  /* Only applies to THIS component's <h1> */
  h1 { color: navy; }
</style>
```

## Estilos Globales

```astro
<!-- Entire block is global -->
<style is:global>
  body { margin: 0; }
</style>

<!-- Individual global rule inside scoped block -->
<style>
  :global(.nav-active) { font-weight: bold; }
</style>
```

## Variables CSS Dinámicas

```astro
---
const accentColor = "#3B82F6";
const spacing = "1.5rem";
---
<div class="card">Content</div>

<style define:vars={{ accentColor, spacing }}>
  .card {
    border-color: var(--accentColor);
    padding: var(--spacing);
  }
</style>
```

## Pasar Clases a Componentes

Los componentes deben aceptar y reenviar explícitamente `class`:

```astro
---
interface Props {
  class?: string;
}
const { class: className } = Astro.props;
---
<div class:list={["card", className]}>
  <slot />
</div>
```

## Integración de Tailwind CSS v4

```bash
npx astro add tailwind
```

Esto instala el plugin `@tailwindcss/vite`. Luego crea:

```css
/* src/styles/global.css */
@import "tailwindcss";

/* Design tokens via @theme */
@theme {
  --color-primary: #18181B;
  --color-accent: #3B82F6;
  --font-sans: "Inter", sans-serif;
  --font-mono: "JetBrains Mono", monospace;
}
```

Importa una vez en el layout base:

```astro
---
// src/layouts/BaseLayout.astro
import '../styles/global.css';
---
<html>
  <body class="bg-background text-primary">
    <slot />
  </body>
</html>
```

## Preprocesadores CSS

Instala el preprocesador, usa el atributo `lang`:

```html
<style lang="scss">
  $primary: navy;
  h1 { color: $primary; }
</style>
```

Soportados: Sass/SCSS, Stylus, Less.

## Orden de Cascada (menor → mayor)

1. Tags `<link>` (hojas de estilo externas)
2. Hojas de estilo importadas (`import './styles.css'`)
3. Estilos con scope (`<style>` en componente)

## Optimización en Producción

- Hojas de estilo < 4kB → auto-inlined en `<head>`
- Hojas de estilo más grandes → tags `<link>` externos
- CSS no usado purgado automáticamente por Tailwind
- Sin runtime de CSS-in-JS — todo resuelto en tiempo de build
