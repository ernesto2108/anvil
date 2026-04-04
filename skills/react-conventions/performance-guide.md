# React Performance Guide

## React Compiler (React 19.2+)

The React Compiler performs compile-time memoization automatically — it analyzes your code and applies fine-grained `useMemo`/`React.memo`/`useCallback` where needed. Delivers 30-60% reduction in unnecessary re-renders.

**When the compiler is active:** write straightforward code and let the compiler optimize. Don't manually memoize unless profiling proves a bottleneck.

**When to still manually memoize:**
- Expensive computations confirmed by profiling
- Projects not yet using the React Compiler
- Third-party libraries that can't be compiled

---

## Performance Checklist (Priority Order)

### 1. Route-Based Code Splitting (Highest ROI)

```tsx
import { lazy, Suspense } from 'react'

const Dashboard = lazy(() => import('./features/dashboard/DashboardPage'))
const Settings = lazy(() => import('./features/settings/SettingsPage'))

function App() {
  return (
    <Suspense fallback={<PageSkeleton />}>
      <Routes>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/settings" element={<Settings />} />
      </Routes>
    </Suspense>
  )
}
```

### 2. Profile Before Optimizing

Use React DevTools Profiler, not guesswork:
- Highlight renders to find unnecessary re-renders
- Flame graph to find slow components
- Only optimize what the profiler flags

### 3. Server Components (Next.js)

Keep heavy dependencies off the client bundle entirely:

```tsx
// This component renders on the server — markdown lib never ships to client
import { marked } from 'marked'

async function BlogPost({ slug }: { slug: string }) {
  const post = await getPost(slug)
  return <article dangerouslySetInnerHTML={{ __html: marked(post.content) }} />
}
```

### 4. Image Optimization

```tsx
// Next.js: automatic optimization
import Image from 'next/image'

<Image
  src="/hero.jpg"
  width={1200}
  height={630}
  alt="Hero image"
  priority        // for above-the-fold
  placeholder="blur"
/>
```

For non-Next.js: responsive images with `srcset`, lazy loading with `loading="lazy"`.

### 5. State Architecture

- **Keep state local** — lift only when needed
- **Avoid "everything in global store"** — most state is server state (TanStack Query)
- **Selectors in Zustand** — only subscribe to what you need

```tsx
// bad: subscribes to entire store, re-renders on any change
const store = useCartStore()

// good: only re-renders when items change
const items = useCartStore(state => state.items)
```

### 6. Skeleton Loading States

Match content structure to prevent Cumulative Layout Shift (CLS):

```tsx
function UserCardSkeleton() {
  return (
    <div className="user-card">
      <div className="skeleton skeleton-avatar" />
      <div className="skeleton skeleton-text" style={{ width: '60%' }} />
      <div className="skeleton skeleton-text" style={{ width: '40%' }} />
    </div>
  )
}
```

---

## Industry Patterns

### Netflix — Server-Driven UI

UI structure comes from the server, enabling rapid A/B testing without client deploys:
- Micro-frontend architecture — independent sections for home, search, profile
- Route-based code splitting with dynamic `import()`
- SSR via Node.js for pre-rendered React components

### Spotify — Design System of Systems (Encore)

- Not one monolithic design system but a family of subsystems
- **Design tokens** at the foundation (type, color, motion, spacing)
- Multiple layers: tokens → primitive components → composed components → product components
- ~400+ engineers touching frontend — "system of systems" prevents bottlenecks

### Shopify — Web Components

- Polaris moved from React components to **Web Components** (framework-agnostic)
- Lesson: at scale, framework-agnostic component libraries reduce long-term maintenance

---

## Common Bottlenecks and Fixes

| Bottleneck | Symptom | Fix |
|---|---|---|
| Large bundle | Slow initial load | Code splitting, tree shaking, remove barrel exports |
| Unnecessary re-renders | Sluggish UI | Selectors, `memo` (or React Compiler), split contexts |
| Heavy computation in render | Janky scrolling | `useMemo`, Web Workers, virtualization |
| Large lists | High memory, slow scroll | `react-window` or `@tanstack/virtual` for virtualization |
| Unoptimized images | Slow LCP | `next/image`, responsive images, lazy loading |
| Layout shifts | Poor CLS score | Skeleton loading, explicit dimensions, `priority` on hero images |
| Waterfall data fetching | Slow page loads | Parallel queries, prefetching, Server Components |

---

## Memoization Rules (Without React Compiler)

```tsx
// useMemo: expensive computation
const sortedItems = useMemo(
  () => items.sort((a, b) => a.name.localeCompare(b.name)),
  [items]
)

// useCallback: stable function reference for child props
const handleClick = useCallback((id: string) => {
  dispatch(selectItem(id))
}, [dispatch])

// React.memo: prevent re-render when props haven't changed
const ExpensiveList = memo(function ExpensiveList({ items }: { items: Item[] }) {
  return items.map(item => <ExpensiveItem key={item.id} item={item} />)
})
```

### When NOT to Memoize

- Simple computations (string concatenation, basic math)
- Components that always re-render anyway (new props every render)
- One-off values that don't need stability
- When React Compiler is active — it handles this automatically
