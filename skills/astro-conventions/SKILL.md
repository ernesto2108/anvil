---
name: astro-conventions
description: Astro framework conventions and coding standards for static and content-driven sites. Use when writing Astro components, reviewing Astro code, or user mentions "Astro patterns", "islands architecture", "content collections", "static site", ".astro files", "astro components", "client directives", or working with .astro files.
---

# Astro Conventions

> **IMPORTANT:** Dispatcher only. Load reference files on demand. See routing table below.

## Philosophy

- **Content-first, JS-last** — pages render as static HTML with zero JS by default. Add JS only where genuine interactivity is needed
- **Islands over SPAs** — interactive components are isolated islands that hydrate independently. The page is a document with embedded widgets, not an app
- **Static until proven otherwise** — start with SSG. Only add SSR for pages that truly need request-time data
- **Ship what you use** — no framework runtime, no unused CSS, no speculative JS. Every byte earns its place

## Stack

- Astro 5+ with TypeScript (strict mode)
- Tailwind CSS v4 via `@tailwindcss/vite` plugin
- Content Collections with Zod schemas
- Vitest + Playwright for testing
- Deploy: Vercel, Netlify, or Cloudflare Pages

## When to Use Astro (vs Next.js)

| Use Astro | Use Next.js |
|---|---|
| Blogs, docs, portfolios, marketing | Full SPAs, heavy client routing |
| Content-driven, minimal interactivity | Real-time dashboards, complex auth |
| SEO and performance non-negotiable | React Server Components needed |
| Static hosting, low cost at scale | Server-heavy with streaming SSR |

## Project Structure

```
src/
├── components/
│   ├── common/           # Shared: Button, Card, Badge
│   ├── features/         # Domain: BlogCard, ProjectGrid
│   └── islands/          # Interactive with client:* directives
├── content/
│   ├── blog/             # Markdown/MDX collection
│   ├── projects/         # Content collection
│   └── config.ts         # Zod schemas
├── layouts/              # BaseLayout.astro, BlogLayout.astro
├── pages/                # File-based routing (only reserved dir)
├── styles/               # global.css (@import "tailwindcss")
├── lib/                  # Utilities, helpers
└── env.d.ts              # Global type extensions
public/                   # Static assets (favicon, robots.txt)
```

**Rules:**
- Interactive components with `client:*` → `src/components/islands/` (makes JS boundaries explicit)
- Content with Zod schemas → `src/content/` (never raw unvalidated markdown)
- Path aliases: `@components/*`, `@layouts/*`, `@lib/*`
- `public/` is for static assets ONLY — never CSS/JS (those go in `src/`)

## Red Flags (always stop work)

- `client:load` on below-fold content → use `client:visible`
- Entire layout wrapped in React/Vue → extract granular islands
- Large array mapped to spawn islands → render statically, hydrate one controller
- Raw `<img>` for local assets → use `<Image />` from `astro:assets`
- Content without Zod schema → define schema in `config.ts`
- Fetching own API via HTTP at build → import function directly
- Using Astro for full SPA → wrong tool, use Next.js

## Pre-Implementation Checklist

- [ ] Astro is the right choice (content-driven, not SPA)
- [ ] TypeScript strict mode (`extends: "astro/tsconfigs/strict"`)
- [ ] Content Collections with Zod schemas defined
- [ ] Tailwind v4 via `@tailwindcss/vite`
- [ ] Base layout with `<ViewTransitions />`
- [ ] Interactive components isolated in `islands/`
- [ ] All images use `<Image />` from `astro:assets`
- [ ] Path aliases in tsconfig

## New Pattern Validation

When introducing a pattern that doesn't exist yet in the project (icon system, script organization, CSS architecture, state management, i18n approach):

1. **Check if the project already has a convention** — read existing code first
2. **If no convention exists**, search the framework's official docs and community best practices before implementing. Don't assume — verify
3. **If a convention exists**, follow it. Don't introduce a competing pattern

This does NOT mean search the web for every change. Only for **new architectural patterns** that will be used project-wide.

## Anti-Pattern Detection

| Anti-Pattern | Severity | Fix |
|---|---|---|
| `client:load` on non-critical widget | warning | Use `client:idle` or `client:visible` |
| Layout wrapped in framework component | error | Extract granular islands |
| Array mapped to spawn islands | error | Static list + one hydrated controller |
| Raw `<img>` for local asset | warning | `<Image />` from `astro:assets` |
| Content without Zod schema | error | Define in `content/config.ts` |
| HTTP fetch to own API at build | error | Import function directly |
| `public/` used for CSS/JS | error | Move to `src/` |
| No `Props` interface in .astro file | warning | Add `interface Props` |
| Using Astro for full SPA | error | Wrong tool — use Next.js |
| Interactive component not in `islands/` | warning | Move to `src/components/islands/` |

## Reference Files

Load ONLY when needed:

| Working on... | Load |
|---|---|
| .astro component model, Props, slots, template expressions | `reference/components.md` |
| client:* directives, hydration strategy, server islands | `reference/islands.md` |
| Content Collections, Zod schemas, loaders, MDX | `reference/content-collections.md` |
| Scoped CSS, Tailwind v4, define:vars, preprocessors | `reference/styling.md` |
| TypeScript, images, View Transitions, SSR, middleware, testing | `reference/patterns.md` |
| Pods architecture, mappers, defensive fetching, API consumption | `reference/api-patterns.md` |

## Post-Implementation Gate

After ANY change to `.astro` files:
1. Run `astro check` for type errors
2. Run `astro build` to verify generation succeeds
3. Invoke `/lint` skill
