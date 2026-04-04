# Content Collections

## Definition

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

| Loader | Source | Use for |
|---|---|---|
| `glob()` | Markdown/MDX files from disk | Blog posts, docs, projects |
| `file()` | Single JSON/YAML file | Authors, config, navigation |
| Custom | API, CMS, database | Headless CMS content |

## Querying

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

## Dynamic Routes

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

## Typed Props

```typescript
import type { CollectionEntry } from 'astro:content';

interface Props {
  post: CollectionEntry<'blog'>;
}
```

## Rules

- Always define Zod schemas — catches errors at dev time
- `z.coerce.date()` for dates — handles string→Date
- `reference('collection')` for cross-collection links
- Sort manually — query order is non-deterministic
- Filter early with `getCollection()` callbacks
- Collections live outside `src/pages/` — create routes via `getStaticPaths()`
- `strictNullChecks: true` required for proper typing
