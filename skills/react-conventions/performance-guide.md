# Guía de Performance en React

## React Compiler (React 19.2+)

El React Compiler realiza memoización en tiempo de compilación automáticamente — analiza tu código y aplica `useMemo`/`React.memo`/`useCallback` con granularidad fina donde se necesite. Entrega una reducción del 30-60% en re-renders innecesarios.

**Cuando el compilador está activo:** escribe código directo y deja que el compilador optimice. No memorices manualmente a menos que el profiling demuestre un cuello de botella.

**Cuándo seguir memoizando manualmente:**
- Computaciones costosas confirmadas por profiling
- Proyectos que aún no usan el React Compiler
- Librerías de terceros que no pueden compilarse

---

## Checklist de Performance (Orden de Prioridad)

### 1. Code Splitting por Ruta (Mayor ROI)

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

### 2. Perfilar Antes de Optimizar

Usar React DevTools Profiler, no suposiciones:
- Resaltar renders para encontrar re-renders innecesarios
- Flame graph para encontrar componentes lentos
- Solo optimizar lo que el profiler marque

### 3. Server Components (Next.js)

Mantener dependencias pesadas completamente fuera del bundle del cliente:

```tsx
// This component renders on the server — markdown lib never ships to client
import { marked } from 'marked'

async function BlogPost({ slug }: { slug: string }) {
  const post = await getPost(slug)
  return <article dangerouslySetInnerHTML={{ __html: marked(post.content) }} />
}
```

### 4. Optimización de Imágenes

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

Para proyectos sin Next.js: imágenes responsivas con `srcset`, lazy loading con `loading="lazy"`.

### 5. Arquitectura de Estado

- **Mantener el estado local** — elevar solo cuando sea necesario
- **Evitar "todo en el store global"** — la mayor parte del estado es server state (TanStack Query)
- **Selectores en Zustand** — suscribirse solo a lo que necesitas

```tsx
// bad: subscribes to entire store, re-renders on any change
const store = useCartStore()

// good: only re-renders when items change
const items = useCartStore(state => state.items)
```

### 6. Estados de Carga con Skeleton

Hacer coincidir la estructura del contenido para prevenir Cumulative Layout Shift (CLS):

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

## Patrones de la Industria

### Netflix — Server-Driven UI

La estructura de UI viene del servidor, habilitando A/B testing rápido sin deploys del cliente:
- Arquitectura micro-frontend — secciones independientes para home, búsqueda, perfil
- Code splitting por ruta con `import()` dinámico
- SSR vía Node.js para componentes React pre-renderizados

### Spotify — Design System of Systems (Encore)

- No un design system monolítico sino una familia de subsistemas
- **Design tokens** en la base (tipo, color, movimiento, espaciado)
- Múltiples capas: tokens → componentes primitivos → componentes compuestos → componentes de producto
- ~400+ ingenieros tocando el frontend — "sistema de sistemas" previene cuellos de botella

### Shopify — Web Components

- Polaris migró de componentes React a **Web Components** (agnóstico al framework)
- Lección: a escala, las librerías de componentes agnósticas al framework reducen el mantenimiento a largo plazo

---

## Cuellos de Botella Comunes y Correcciones

| Cuello de botella | Síntoma | Corrección |
|---|---|---|
| Bundle grande | Carga inicial lenta | Code splitting, tree shaking, eliminar barrel exports |
| Re-renders innecesarios | UI lenta | Selectores, `memo` (o React Compiler), separar contexts |
| Computación pesada en render | Scroll entrecortado | `useMemo`, Web Workers, virtualización |
| Listas grandes | Alta memoria, scroll lento | `react-window` o `@tanstack/virtual` para virtualización |
| Imágenes no optimizadas | LCP lento | `next/image`, imágenes responsivas, lazy loading |
| Layout shifts | Mala puntuación CLS | Skeleton loading, dimensiones explícitas, `priority` en imágenes hero |
| Fetching de datos en cascada | Cargas de página lentas | Queries paralelos, prefetching, Server Components |

---

## Reglas de Memoización (Sin React Compiler)

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

### Cuándo NO Memoizar

- Computaciones simples (concatenación de strings, matemáticas básicas)
- Componentes que siempre se re-renderizan de todos modos (nuevas props en cada render)
- Valores únicos que no necesitan estabilidad
- Cuando el React Compiler está activo — lo maneja automáticamente
