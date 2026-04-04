# React State Management Guide

## State Categorization

> ~90% of traditional state management concerns disappear when you properly categorize state.

| State Type | Tool | Rationale |
|---|---|---|
| Server/remote data | **TanStack Query** (or SWR) | Caching, deduplication, retries, pagination, optimistic updates |
| URL state | **nuqs** | Type-safe URL query parameter sync |
| Client shared state | **Zustand** | Minimal API, no provider needed, granular subscriptions |
| Complex interdependent state | **Jotai** | Atomic model, computed state graphs |
| Form state | **React Hook Form + Zod** | Uncontrolled inputs for performance, schema validation |
| Low-velocity global (theme, auth) | **React Context** | Built-in, sufficient for rarely-changing data |

---

## Escalation Path

```
Context API → (need more features?) → Zustand → (need time-travel / complex middleware?) → Redux
```

| Scope | Simple | Medium | Complex |
|---|---|---|---|
| Single Component | Context | Context | Zustand |
| Few Components | Context | Zustand | Zustand |
| Many Components | Zustand | Zustand | Redux |
| Global (App-wide) | Zustand | Zustand | Redux |

---

## TanStack Query — Server State (Preferred)

Handles caching, deduplication, retries, pagination, and optimistic updates. Eliminates most manual loading/error state.

### Basic Query

```tsx
function useUser(id: string) {
  return useQuery({
    queryKey: ['user', id],
    queryFn: () => api.getUser(id),
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}

function UserProfile({ id }: { id: string }) {
  const { data: user, isLoading, error } = useUser(id)

  if (isLoading) return <Skeleton />
  if (error) return <ErrorMessage error={error} />
  return <ProfileCard user={user} />
}
```

### Mutations with Optimistic Updates

```tsx
function useUpdateUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: UpdateUserInput) => api.updateUser(data),
    onMutate: async (newData) => {
      await queryClient.cancelQueries({ queryKey: ['user', newData.id] })
      const previous = queryClient.getQueryData(['user', newData.id])
      queryClient.setQueryData(['user', newData.id], (old: User) => ({
        ...old,
        ...newData,
      }))
      return { previous }
    },
    onError: (_err, _vars, context) => {
      queryClient.setQueryData(['user', context?.previous], context?.previous)
    },
    onSettled: (_data, _err, vars) => {
      queryClient.invalidateQueries({ queryKey: ['user', vars.id] })
    },
  })
}
```

### Infinite Queries (Pagination)

```tsx
function useInfiniteUsers() {
  return useInfiniteQuery({
    queryKey: ['users'],
    queryFn: ({ pageParam = 1 }) => api.getUsers({ page: pageParam }),
    getNextPageParam: (lastPage) => lastPage.nextPage ?? undefined,
    initialPageParam: 1,
  })
}
```

### Rules

- Query keys must be deterministic and unique per resource
- Use `staleTime` to control refetch frequency (default 0 = always stale)
- Prefetch on hover for instant navigation: `queryClient.prefetchQuery()`
- Keep mutations close to the component that triggers them

---

## Zustand — Client Shared State (Preferred)

Minimal API, no provider needed, granular subscriptions via selectors.

### Basic Store

```tsx
import { create } from 'zustand'

interface CartStore {
  items: CartItem[]
  addItem: (item: CartItem) => void
  removeItem: (id: string) => void
  clearCart: () => void
  total: () => number
}

const useCartStore = create<CartStore>((set, get) => ({
  items: [],
  addItem: (item) => set((state) => ({ items: [...state.items, item] })),
  removeItem: (id) => set((state) => ({ items: state.items.filter(i => i.id !== id) })),
  clearCart: () => set({ items: [] }),
  total: () => get().items.reduce((sum, item) => sum + item.price * item.quantity, 0),
}))
```

### Consuming with Selectors

```tsx
// good: only re-renders when items change
const items = useCartStore(state => state.items)
const addItem = useCartStore(state => state.addItem)

// bad: re-renders on ANY store change
const store = useCartStore()
```

### Middleware (Persist + DevTools)

```tsx
import { create } from 'zustand'
import { persist, devtools } from 'zustand/middleware'

const useCartStore = create<CartStore>()(
  devtools(
    persist(
      (set, get) => ({
        items: [],
        addItem: (item) => set((state) => ({ items: [...state.items, item] }), false, 'addItem'),
        // ...
      }),
      { name: 'cart-storage' }
    )
  )
)
```

### Slices Pattern (Large Stores)

```tsx
interface AuthSlice {
  user: User | null
  login: (creds: Credentials) => Promise<void>
  logout: () => void
}

interface CartSlice {
  items: CartItem[]
  addItem: (item: CartItem) => void
}

const createAuthSlice: StateCreator<AuthSlice & CartSlice, [], [], AuthSlice> = (set) => ({
  user: null,
  login: async (creds) => {
    const user = await api.login(creds)
    set({ user })
  },
  logout: () => set({ user: null }),
})

const createCartSlice: StateCreator<AuthSlice & CartSlice, [], [], CartSlice> = (set) => ({
  items: [],
  addItem: (item) => set((state) => ({ items: [...state.items, item] })),
})

const useStore = create<AuthSlice & CartSlice>()((...a) => ({
  ...createAuthSlice(...a),
  ...createCartSlice(...a),
}))
```

### Rules

- Always use selectors — never destructure the entire store
- Name actions in devtools for debugging (`set({...}, false, 'actionName')`)
- Use `persist` middleware for data that survives refresh (cart, preferences)
- Split into slices when store exceeds ~10 properties

---

## React Context — Low-Velocity Global State

Only for data that changes infrequently (theme, locale, auth status).

```tsx
interface ThemeContextValue {
  theme: 'light' | 'dark'
  toggleTheme: () => void
}

const ThemeContext = createContext<ThemeContextValue | null>(null)

function useTheme() {
  const ctx = useContext(ThemeContext)
  if (!ctx) throw new Error('useTheme must be used within ThemeProvider')
  return ctx
}

function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setTheme] = useState<'light' | 'dark'>('light')
  const toggleTheme = useCallback(() => {
    setTheme(prev => prev === 'light' ? 'dark' : 'light')
  }, [])

  const value = useMemo(() => ({ theme, toggleTheme }), [theme, toggleTheme])

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}
```

### Rules

- Context created with `createContext<Type | null>(null)` — always null check in hook
- `useMemo` the value object to prevent unnecessary re-renders
- Max 3 values in a single context — split if more
- Never use Context for frequently-changing data (use Zustand instead)
- Render providers as deep as possible in the tree

---

## React Hook Form + Zod — Form State

```tsx
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

const loginSchema = z.object({
  email: z.string().email('Invalid email'),
  password: z.string().min(8, 'Must be at least 8 characters'),
})

type LoginFormData = z.infer<typeof loginSchema>

function LoginForm() {
  const { register, handleSubmit, formState: { errors } } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  })

  const onSubmit = (data: LoginFormData) => {
    // data is fully typed and validated
    login(data)
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <input {...register('email')} />
      {errors.email && <span role="alert">{errors.email.message}</span>}

      <input type="password" {...register('password')} />
      {errors.password && <span role="alert">{errors.password.message}</span>}

      <button type="submit">Login</button>
    </form>
  )
}
```

### Rules

- Zod schema is the single source of truth for validation
- `z.infer<typeof schema>` generates the TypeScript type
- Uncontrolled inputs by default (performance) — use `watch()` only when needed
- Share schemas between frontend and backend

---

## Redux Toolkit — Complex Enterprise State

Only when you need time-travel debugging, complex middleware, or audit trails.

```tsx
import { createSlice, createAsyncThunk, configureStore } from '@reduxjs/toolkit'

const fetchUsers = createAsyncThunk('users/fetch', async () => {
  return await api.getUsers()
})

const usersSlice = createSlice({
  name: 'users',
  initialState: { items: [] as User[], status: 'idle' as 'idle' | 'loading' | 'error' },
  reducers: {},
  extraReducers: (builder) => {
    builder
      .addCase(fetchUsers.pending, (state) => { state.status = 'loading' })
      .addCase(fetchUsers.fulfilled, (state, action) => {
        state.status = 'idle'
        state.items = action.payload
      })
      .addCase(fetchUsers.rejected, (state) => { state.status = 'error' })
  },
})

// typed hooks
const useAppSelector: TypedUseSelectorHook<RootState> = useSelector
const useAppDispatch: () => AppDispatch = useDispatch
```

### Rules

- Always use RTK (Redux Toolkit) — never raw Redux
- Always use typed hooks (`useAppSelector`, `useAppDispatch`)
- RTK Query for server state (similar to TanStack Query)
- Wrap with facade hooks — components never import `useAppSelector` directly
