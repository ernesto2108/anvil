# Styling in Astro

## Scoped CSS (Default)

`<style>` in `.astro` files is automatically scoped via data attributes:

```astro
<style>
  /* Only applies to THIS component's <h1> */
  h1 { color: navy; }
</style>
```

## Global Styles

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

## Dynamic CSS Variables

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

## Passing Classes to Components

Components must explicitly accept and forward `class`:

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

## Tailwind CSS v4 Integration

```bash
npx astro add tailwind
```

This installs `@tailwindcss/vite` plugin. Then create:

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

Import once in base layout:

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

## CSS Preprocessors

Install the preprocessor, use `lang` attribute:

```html
<style lang="scss">
  $primary: navy;
  h1 { color: $primary; }
</style>
```

Supported: Sass/SCSS, Stylus, Less.

## Cascade Order (lowest → highest)

1. `<link>` tags (external stylesheets)
2. Imported stylesheets (`import './styles.css'`)
3. Scoped styles (`<style>` in component)

## Production Optimization

- Stylesheets < 4kB → auto-inlined in `<head>`
- Larger stylesheets → external `<link>` tags
- Unused CSS purged by Tailwind automatically
- No CSS-in-JS runtime — all resolved at build time
