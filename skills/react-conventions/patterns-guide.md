# React Patterns Guide

## Custom Hooks — Primary Pattern

The hook IS the container — it encapsulates all logic; the component is pure UI.

```tsx
function useUserList() {
  const [users, setUsers] = useState<User[]>([])
  const [status, setStatus] = useState<'idle' | 'loading' | 'error'>('idle')

  const fetchUsers = useCallback(async () => {
    setStatus('loading')
    try {
      const data = await api.getUsers()
      setUsers(data)
      setStatus('idle')
    } catch {
      setStatus('error')
    }
  }, [])

  return { users, status, fetchUsers } as const
}

// component is pure UI — receives everything from the hook
function UserList() {
  const { users, status, fetchUsers } = useUserList()
  // ... render
}
```

**Rules:**
- Prefix with `use`
- Use `useCallback` for returned functions
- Return `as const` for tuple returns
- Custom hooks always validate context with null check + `throw new Error()`

---

## Compound Components

Related components sharing state implicitly via Context (Tabs, Accordion, Select).

```tsx
// API: <Tabs><Tabs.List><Tabs.Tab /></Tabs.List><Tabs.Panel /></Tabs>
const TabsContext = createContext<TabsContextValue | null>(null)

function useTabsContext() {
  const ctx = useContext(TabsContext)
  if (!ctx) throw new Error('Tabs components must be used within <Tabs>')
  return ctx
}

function Tabs({ children, defaultIndex = 0 }: TabsProps) {
  const [activeIndex, setActiveIndex] = useState(defaultIndex)
  return (
    <TabsContext.Provider value={{ activeIndex, setActiveIndex }}>
      <div role="tablist">{children}</div>
    </TabsContext.Provider>
  )
}

function Tab({ index, children }: TabProps) {
  const { activeIndex, setActiveIndex } = useTabsContext()
  return (
    <button
      role="tab"
      aria-selected={activeIndex === index}
      onClick={() => setActiveIndex(index)}
    >
      {children}
    </button>
  )
}

function Panel({ index, children }: PanelProps) {
  const { activeIndex } = useTabsContext()
  if (activeIndex !== index) return null
  return <div role="tabpanel">{children}</div>
}

Tabs.Tab = Tab
Tabs.Panel = Panel
```

**When to use:** UI components with multiple sub-parts sharing implicit state (accordions, tabs, selects, menus).

---

## Facade Hooks

Hide data source complexity behind a simple hook interface. UI never imports `useSelector`/`useDispatch` directly.

```tsx
// bad: component knows about Redux internals
function Profile() {
  const user = useSelector(state => state.auth.user)
  const dispatch = useDispatch()
  const handleLogout = () => dispatch(logout())
  // ...
}

// good: facade hook abstracts the data source
function useAuth() {
  const user = useSelector(state => state.auth.user)
  const dispatch = useDispatch()

  const login = useCallback((creds: Credentials) => {
    dispatch(loginThunk(creds))
  }, [dispatch])

  const logout = useCallback(() => {
    dispatch(logoutAction())
  }, [dispatch])

  return { user, login, logout, isAuthenticated: !!user } as const
}

function Profile() {
  const { user, logout, isAuthenticated } = useAuth()
  // ...
}
```

**When to use:** When components consume state from Redux, Zustand, Context, or API layers. Swap data source without touching UI.

---

## State Machine

Replace boolean hell with explicit status states using discriminated unions or `useReducer`.

```tsx
// bad: boolean hell
const [isLoading, setIsLoading] = useState(false)
const [isError, setIsError] = useState(false)
const [isSuccess, setIsSuccess] = useState(false)
const [data, setData] = useState<Data | null>(null)

// good: discriminated union
type State<T> =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success'; data: T }
  | { status: 'error'; error: Error }

type Action<T> =
  | { type: 'FETCH' }
  | { type: 'SUCCESS'; data: T }
  | { type: 'ERROR'; error: Error }
  | { type: 'RESET' }

function reducer<T>(state: State<T>, action: Action<T>): State<T> {
  switch (action.type) {
    case 'FETCH': return { status: 'loading' }
    case 'SUCCESS': return { status: 'success', data: action.data }
    case 'ERROR': return { status: 'error', error: action.error }
    case 'RESET': return { status: 'idle' }
  }
}

function useAsync<T>() {
  const [state, dispatch] = useReducer(reducer<T>, { status: 'idle' })
  // ...
  return state
}
```

**When to use:** Any flow with mutually exclusive states. Eliminates impossible states at compile time.

---

## Control Props

Controlled/uncontrolled dual-mode components — component works both ways.

```tsx
interface ToggleProps {
  isOn?: boolean          // controlled mode
  defaultIsOn?: boolean   // uncontrolled mode
  onChange?: (isOn: boolean) => void
}

function Toggle({ isOn: controlledIsOn, defaultIsOn = false, onChange }: ToggleProps) {
  const [internalIsOn, setInternalIsOn] = useState(defaultIsOn)
  const isControlled = controlledIsOn !== undefined
  const isOn = isControlled ? controlledIsOn : internalIsOn

  const handleToggle = () => {
    const next = !isOn
    if (!isControlled) setInternalIsOn(next)
    onChange?.(next)
  }

  return <button onClick={handleToggle}>{isOn ? 'ON' : 'OFF'}</button>
}
```

**When to use:** Form elements, toggles, selects — anything that should work standalone or parent-controlled.

---

## Adapter Component

Wrapping 3rd-party libraries to isolate vendor lock-in.

```tsx
// bad: 3rd party API leaks everywhere
import { Chart as ChartJS } from 'chart.js'
// used in 20 components...

// good: adapter wraps the vendor
interface ChartProps {
  data: DataPoint[]
  type: 'line' | 'bar' | 'pie'
  height?: number
}

function Chart({ data, type, height = 300 }: ChartProps) {
  // only this file imports chart.js
  return <ChartJS type={type} data={transformData(data)} height={height} />
}
```

**When to use:** Any 3rd-party UI library (charts, maps, editors, date pickers). Swap vendors by changing one file.

---

## Strategy Pattern

Replace `if/else`/`switch` blocks with interchangeable strategy objects.

```tsx
// bad: switch in component
function PricingDisplay({ plan }: { plan: Plan }) {
  switch (plan.type) {
    case 'free': return <FreePricing />
    case 'pro': return <ProPricing />
    case 'enterprise': return <EnterprisePricing />
  }
}

// good: strategy map
const pricingStrategies: Record<PlanType, ComponentType<PricingProps>> = {
  free: FreePricing,
  pro: ProPricing,
  enterprise: EnterprisePricing,
}

function PricingDisplay({ plan }: { plan: Plan }) {
  const PricingComponent = pricingStrategies[plan.type]
  return <PricingComponent plan={plan} />
}
```

**When to use:** Rendering different UI variants based on a discriminator. Adding a variant = adding one entry, no conditionals.

---

## Observer Pattern

Event bus for communication outside the React tree (toasts, WebSocket events, micro-frontends).

```tsx
type EventMap = {
  'toast:show': { message: string; type: 'success' | 'error' }
  'ws:message': { channel: string; payload: unknown }
}

class EventBus {
  private listeners = new Map<string, Set<Function>>()

  on<K extends keyof EventMap>(event: K, handler: (data: EventMap[K]) => void) {
    if (!this.listeners.has(event)) this.listeners.set(event, new Set())
    this.listeners.get(event)!.add(handler)
    return () => this.listeners.get(event)?.delete(handler)
  }

  emit<K extends keyof EventMap>(event: K, data: EventMap[K]) {
    this.listeners.get(event)?.forEach(fn => fn(data))
  }
}

export const eventBus = new EventBus()

// hook wrapper
function useEventBus<K extends keyof EventMap>(event: K, handler: (data: EventMap[K]) => void) {
  useEffect(() => eventBus.on(event, handler), [event, handler])
}
```

**When to use:** Cross-cutting events that don't fit React's component tree (toasts, analytics, WebSocket routing).

---

## Decorator Hooks

Wrapping existing hooks to add cross-cutting concerns (analytics, permissions, logging).

```tsx
// base hook
function useUsers() {
  return useQuery({ queryKey: ['users'], queryFn: api.getUsers })
}

// decorated with analytics
function useUsersWithAnalytics() {
  const result = useUsers()

  useEffect(() => {
    if (result.isSuccess) analytics.track('users_loaded', { count: result.data.length })
    if (result.isError) analytics.track('users_error', { error: result.error.message })
  }, [result.isSuccess, result.isError])

  return result
}

// decorated with permissions
function useUsersWithPermissions() {
  const { hasPermission } = useAuth()
  const result = useUsers()

  if (!hasPermission('users:read')) {
    return { ...result, data: [], isUnauthorized: true }
  }

  return { ...result, isUnauthorized: false }
}
```

**When to use:** Adding analytics, permissions, caching, or logging to existing hooks without modifying them.

---

## Factory Pattern

Generating hooks/components dynamically from configuration.

```tsx
function createResourceHook<T>(endpoint: string) {
  return function useResource(id?: string) {
    return useQuery<T>({
      queryKey: [endpoint, id],
      queryFn: () => api.get<T>(id ? `${endpoint}/${id}` : endpoint),
    })
  }
}

// usage — one line per resource
const useUsers = createResourceHook<User[]>('/users')
const useProducts = createResourceHook<Product[]>('/products')
const useOrders = createResourceHook<Order[]>('/orders')
```

**When to use:** Multiple resources with identical fetch/cache patterns. Avoid copy-pasting hooks.

---

## Pattern Relationships

```
compound-components --> provider-pattern (when state goes global)
                    --> control-props (when parent controls state)
custom-hooks --> decorator-hooks (when wrapping hooks)
             --> facade-hooks (when abstracting data source)
             --> factory-pattern (when generating hooks)
state-machine --> command-pattern (when adding undo/redo)
              --> zustand/redux (when scaling up)
adapter-component --> strategy-pattern (when multiple variants)
observer-pattern --> mediator-pattern (when coordination is needed)
```
