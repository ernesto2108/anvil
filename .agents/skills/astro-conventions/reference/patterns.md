# Patrones Avanzados de Astro

## TypeScript

```json
// tsconfig.json
{
  "extends": "astro/tsconfigs/strict"
}
```

- Usa `strict` o `strictest` — nunca `base` para producción
- Interfaz `Props` en cada componente `.astro`
- Ejecuta `astro check` en CI: `"build": "astro check && astro build"`
- `import type` para imports solo de tipos
- Utilidades de tipos: `HTMLAttributes<"div">`, `ComponentProps`, `InferGetStaticParamsType`
- `src/env.d.ts` para extensiones de tipos globales (`Astro.locals`, propiedades de window)

## Optimización de Imágenes

```astro
---
import { Image, Picture } from 'astro:assets';
import heroImg from '../images/hero.jpg';
---
<!-- Optimized: auto width/height, lazy loading, format conversion -->
<Image src={heroImg} alt="Hero" />

<!-- Multi-format: avif + webp + fallback -->
<Picture src={heroImg} formats={['avif', 'webp']} alt="Hero" />

<!-- Remote (requires image.domains config) -->
<Image src="https://example.com/photo.jpg" width={800} height={400} alt="Remote" />
```

**Reglas:**
- Imágenes en `src/` → optimizadas en build (recomendado)
- Imágenes en `public/` → servidas tal cual (solo para assets sin procesar)
- Remotas → necesita `image.domains` en `astro.config.mjs`
- `alt` es obligatorio — aplicado por Astro
- Usa el helper `image()` en esquemas de Content Collection para referencias de imágenes tipadas

## View Transitions

```astro
---
// src/layouts/BaseLayout.astro
import { ViewTransitions } from 'astro:transitions';
---
<html>
  <head>
    <ViewTransitions />
  </head>
  <body>
    <slot />
  </body>
</html>
```

- Agrégalo una vez en el layout base → navegación tipo SPA en todo el sitio
- Incorporados: `fade`, `slide`, `none`
- `transition:name="hero"` → persiste elemento entre páginas
- `transition:animate="slide"` → animación por elemento
- Preserva MPA (cada página tiene su propia URL, indexable)

## SSR / Renderizado Híbrido

```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';
import vercel from '@astrojs/vercel';

export default defineConfig({
  output: 'hybrid',        // mostly static + some SSR
  adapter: vercel(),
});
```

Opt-in/out por página:

```astro
---
// This page renders on every request (SSR)
export const prerender = false;
---
```

| Modo | Por defecto | Override |
|---|---|---|
| `static` (por defecto) | Todas las páginas pre-renderizadas | N/A |
| `hybrid` | Todas pre-renderizadas | `prerender = false` para páginas SSR |
| `server` | Todo SSR | `prerender = true` para páginas estáticas |

## Middleware

```typescript
// src/middleware.ts
import { defineMiddleware, sequence } from "astro:middleware";

const auth = defineMiddleware((context, next) => {
  const token = context.cookies.get("token");
  if (!token && context.url.pathname.startsWith("/dashboard")) {
    return context.redirect("/login");
  }
  context.locals.user = decodeToken(token);
  return next();
});

const logging = defineMiddleware(async (context, next) => {
  const start = Date.now();
  const response = await next();
  console.log(`${context.request.method} ${context.url.pathname} ${Date.now() - start}ms`);
  return response;
});

export const onRequest = sequence(auth, logging);
```

- `context.locals` → comparte datos entre middleware y páginas
- `sequence()` → encadena múltiples middleware
- `context.rewrite()` → muestra contenido diferente sin redireccionar
- Solo corre para páginas SSR (no estáticas)

## Testing

**Unitario (Vitest + Container API):**

```typescript
// src/components/__tests__/Card.test.ts
import { experimental_AstroContainer as AstroContainer } from 'astro/container';
import { expect, test } from 'vitest';
import Card from '../Card.astro';

test('renders card with title', async () => {
  const container = await AstroContainer.create();
  const result = await container.renderToString(Card, {
    props: { title: 'Test' },
    slots: { default: '<p>Content</p>' },
  });
  expect(result).toContain('Test');
  expect(result).toContain('Content');
});
```

**E2E (Playwright):**

```typescript
// playwright.config.ts
export default {
  webServer: {
    command: 'npm run preview',
    url: 'http://localhost:4321',
  },
};
```

```typescript
// tests/blog.spec.ts
import { test, expect } from '@playwright/test';

test('blog page loads', async ({ page }) => {
  await page.goto('/blog');
  await expect(page.locator('h1')).toContainText('Blog');
});
```

## API Endpoints (Solo SSR)

```typescript
// src/pages/api/posts.ts
import type { APIRoute } from 'astro';

export const GET: APIRoute = async ({ request }) => {
  const posts = await fetchPosts();
  return new Response(JSON.stringify(posts), {
    headers: { 'Content-Type': 'application/json' },
  });
};

export const POST: APIRoute = async ({ request }) => {
  const body = await request.json();
  // validate, process...
  return new Response(JSON.stringify({ ok: true }), { status: 201 });
};
```
