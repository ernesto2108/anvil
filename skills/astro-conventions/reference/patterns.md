# Advanced Astro Patterns

## TypeScript

```json
// tsconfig.json
{
  "extends": "astro/tsconfigs/strict"
}
```

- Use `strict` or `strictest` — never `base` for production
- `Props` interface in every `.astro` component
- Run `astro check` in CI: `"build": "astro check && astro build"`
- `import type` for type-only imports
- Type utilities: `HTMLAttributes<"div">`, `ComponentProps`, `InferGetStaticParamsType`
- `src/env.d.ts` for global type extensions (`Astro.locals`, window properties)

## Image Optimization

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

**Rules:**
- `src/` images → optimized at build (recommended)
- `public/` images → served as-is (only for unprocessed assets)
- Remote → needs `image.domains` in `astro.config.mjs`
- `alt` is mandatory — enforced by Astro
- Use `image()` helper in Content Collection schemas for typed image refs

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

- Add once in base layout → site-wide SPA-like navigation
- Built-in: `fade`, `slide`, `none`
- `transition:name="hero"` → persist element across pages
- `transition:animate="slide"` → per-element animation
- Preserves MPA (each page has its own URL, crawlable)

## SSR / Hybrid Rendering

```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';
import vercel from '@astrojs/vercel';

export default defineConfig({
  output: 'hybrid',        // mostly static + some SSR
  adapter: vercel(),
});
```

Per-page opt-in/out:

```astro
---
// This page renders on every request (SSR)
export const prerender = false;
---
```

| Mode | Default | Override |
|---|---|---|
| `static` (default) | All pages prerendered | N/A |
| `hybrid` | All prerendered | `prerender = false` for SSR pages |
| `server` | All SSR | `prerender = true` for static pages |

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

- `context.locals` → share data between middleware and pages
- `sequence()` → chain multiple middleware
- `context.rewrite()` → show different content without redirect
- Only runs for SSR pages (not static)

## Testing

**Unit (Vitest + Container API):**

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

## API Endpoints (SSR only)

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
