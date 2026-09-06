# Content Collections

## Definición

```typescript
// src/content/config.ts
import { defineCollection } from 'astro:content';
import { glob, file } from 'astro/loaders';
import { z } from 'astro/zod';

const blog = defineCollection({
  loader: glob({ pattern: "**/*.md", base: "./src/content/blog" }),
  schema: z.object({
    title: z.string(),
    pubDate: z.coerce.date(),
    updatedDate: z.coerce.date().optional(),
    description: z.string(),
    draft: z.boolean().default(false),
    tags: z.array(z.string()),
    image: z.string().optional(),
    author: reference('authors'),        // cross-collection ref
  })
});

const projects = defineCollection({
  loader: glob({ pattern: "**/*.md", base: "./src/content/projects" }),
  schema: z.object({
    title: z.string(),
    repo: z.string().url(),
    tech: z.array(z.string()),
    featured: z.boolean().default(false),
    order: z.number().default(0),
  })
});

const authors = defineCollection({
  loader: file("./src/content/authors.json"),
  schema: z.object({
    name: z.string(),
    avatar: z.string(),
    bio: z.string(),
  })
});

export const collections = { blog, projects, authors };
```

## Loaders

| Loader | Fuente | Usar para |
|---|---|---|
| `glob()` | Archivos Markdown/MDX del disco | Posts de blog, docs, proyectos |
| `file()` | Archivo JSON/YAML único | Autores, config, navegación |
| Custom | API, CMS, base de datos | Contenido de CMS headless |

## Consultas

```typescript
import { getCollection, getEntry, render } from 'astro:content';

// Get all entries
const allPosts = await getCollection('blog');

// Filter with callback
const published = await getCollection('blog', ({ data }) => !data.draft);

// Get single entry
const post = await getEntry('blog', 'my-first-post');

// Render Markdown to HTML
const { Content, headings } = await render(post);
```

## Rutas Dinámicas

```astro
---
// src/pages/blog/[slug].astro
import { getCollection, render } from 'astro:content';
import BlogLayout from '../../layouts/BlogLayout.astro';

export async function getStaticPaths() {
  const posts = await getCollection('blog', ({ data }) => !data.draft);
  return posts.map(post => ({
    params: { slug: post.id },
    props: { post },
  }));
}

const { post } = Astro.props;
const { Content } = await render(post);
---
<BlogLayout title={post.data.title}>
  <h1>{post.data.title}</h1>
  <time>{post.data.pubDate.toLocaleDateString()}</time>
  <Content />
</BlogLayout>
```

## Props Tipadas

```typescript
import type { CollectionEntry } from 'astro:content';

interface Props {
  post: CollectionEntry<'blog'>;
}
```

## Reglas

- Siempre define esquemas Zod — detecta errores en tiempo de desarrollo
- `z.coerce.date()` para fechas — maneja string→Date
- `reference('collection')` para links entre colecciones
- Ordena manualmente — el orden de consulta es no determinista
- Filtra temprano con callbacks de `getCollection()`
- Las colecciones viven fuera de `src/pages/` — crea rutas vía `getStaticPaths()`
- `strictNullChecks: true` requerido para tipado correcto
