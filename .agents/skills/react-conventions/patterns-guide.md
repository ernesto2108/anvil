# Guía de Patrones React

## Custom Hooks — Patrón Principal

El hook ES el contenedor — encapsula toda la lógica; el componente es UI pura.

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

**Reglas:**
- Prefijo `use`
- Usar `useCallback` para funciones retornadas
- Retornar `as const` para retornos de tupla
- Los custom hooks siempre validan el contexto con verificación de null + `throw new Error()`

---

## Compound Components

Componentes relacionados que comparten estado implícitamente via Context (Tabs, Accordion, Select).

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

**Cuándo usar:** Componentes de UI con múltiples sub-partes que comparten estado implícito (accordions, tabs, selects, menús).

---

## Facade Hooks

Ocultan la complejidad de la fuente de datos detrás de una interfaz de hook simple. La UI nunca importa `useSelector`/`useDispatch` directamente.

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

**Cuándo usar:** Cuando los componentes consumen estado de Redux, Zustand, Context o capas de API. Cambiar la fuente de datos sin tocar la UI.

---

## State Machine

Reemplazar boolean hell con estados de status explícitos usando uniones discriminadas o `useReducer`.

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

**Cuándo usar:** Cualquier flujo con estados mutuamente excluyentes. Elimina estados imposibles en tiempo de compilación.

---

## Control Props

Componentes con modo controlado/no-controlado dual — el componente funciona de ambas formas.

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

**Cuándo usar:** Elementos de formulario, toggles, selects — cualquier cosa que deba funcionar de forma autónoma o controlada por el padre.

---

## Adapter Component

Envolver librerías de terceros para aislar el vendor lock-in.

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

**Cuándo usar:** Cualquier librería de UI de terceros (gráficas, mapas, editores, date pickers). Cambiar de vendor modificando un solo archivo.

---

## Strategy Pattern

Reemplazar bloques `if/else`/`switch` con objetos de estrategia intercambiables.

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

**Cuándo usar:** Renderizar variantes de UI diferentes basadas en un discriminador. Agregar una variante = agregar una entrada, sin condicionales.

---

## Observer Pattern

Bus de eventos para comunicación fuera del árbol de React (toasts, eventos WebSocket, micro-frontends).

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

**Cuándo usar:** Eventos transversales que no encajan en el árbol de componentes de React (toasts, analytics, enrutamiento WebSocket).

---

## Decorator Hooks

Envolver hooks existentes para agregar responsabilidades transversales (analytics, permisos, logging).

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

**Cuándo usar:** Agregar analytics, permisos, caché o logging a hooks existentes sin modificarlos.

---

## Factory Pattern

Generar hooks/componentes dinámicamente a partir de configuración.

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

**Cuándo usar:** Múltiples recursos con patrones idénticos de fetch/caché. Evitar copiar y pegar hooks.

---

## Relaciones Entre Patrones

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
