# Islands Architecture

## Core Principle

Pages render as static HTML. Interactive components ("islands") hydrate independently with their own JS bundle. The rest of the page is zero JS.

## Client Directives

| Directive | When it hydrates | Use for |
|---|---|---|
| `client:load` | Immediately on page load | Critical: auth state, nav dropdowns |
| `client:idle` | Browser is idle | Medium: search bar, forms |
| `client:visible` | Scrolled into viewport | Below-fold: comments, carousels, maps |
| `client:media="(max-width: 768px)"` | Media query matches | Mobile-only: hamburger menu |
| `client:only="react"` | Client only, skip SSR | Components that can't render server-side |

```astro
---
import Counter from '../components/islands/Counter.tsx';
import Comments from '../components/islands/Comments.tsx';
import MobileNav from '../components/islands/MobileNav.tsx';
---
<!-- Critical — hydrate immediately -->
<Counter client:load initialCount={0} />

<!-- Below fold — hydrate when visible -->
<Comments client:visible postId={post.id} />

<!-- Mobile only -->
<MobileNav client:media="(max-width: 768px)" />
```

## Server Islands

`server:defer` isolates slow/dynamic server content so it doesn't block page render:

```astro
---
import UserGreeting from '../components/UserGreeting.astro';
---
<UserGreeting server:defer>
  <p slot="fallback">Loading...</p>  <!-- shown until server responds -->
</UserGreeting>
```

Use for: personalized content, slow API calls, auth-dependent UI.

## Multi-Framework Islands

Different frameworks coexist because each island is isolated:

```astro
---
import ReactSearch from '../components/islands/Search.tsx';
import VueChart from '../components/islands/Chart.vue';
import SvelteToggle from '../components/islands/Toggle.svelte';
---
<ReactSearch client:idle />
<VueChart client:visible data={chartData} />
<SvelteToggle client:load />
```

Each ships only its own framework runtime — React islands don't load Vue, etc.

## Decision Matrix

```
Is it interactive? (needs browser events, state, effects)
├── NO → Astro component (.astro) — zero JS
└── YES
    ├── Is it above the fold / critical for first interaction?
    │   ├── YES → client:load
    │   └── NO
    │       ├── Is it visible on initial viewport?
    │       │   ├── YES → client:idle
    │       │   └── NO → client:visible
    │       └── Is it device-specific?
    │           └── YES → client:media
    └── Can it render on the server?
        ├── YES → client:load/idle/visible (pick one)
        └── NO → client:only="framework"
```

## Anti-Patterns

| Anti-Pattern | Why it's bad | Fix |
|---|---|---|
| `client:load` on everything | Defeats zero-JS purpose | Audit each: idle? visible? |
| Layout-level framework wrapper | Ships entire framework for static layout | Break into small islands |
| `{items.map(() => <Island client:load />)}` | N islands = N JS bundles | Static list + 1 hydrated controller |
| Island for static content | Unnecessary JS for content that doesn't change | Use .astro component |
| No `client:*` on framework component | Component renders but isn't interactive | Add directive or convert to .astro |
