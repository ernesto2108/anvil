# Guía de Gestión de Estado en React

## Categorización del Estado

> ~90% de las preocupaciones tradicionales de gestión de estado desaparecen cuando categorizas correctamente el estado.

| Tipo de estado | Herramienta | Justificación |
|---|---|---|
| Datos del servidor/remotos | **TanStack Query** (o SWR) | Caché, deduplicación, reintentos, paginación, actualizaciones optimistas |
| Estado de URL | **nuqs** | Sincronización type-safe de query params en URL |
| Estado compartido del cliente | **Zustand** | API mínima, sin provider necesario, suscripciones granulares |
| Estado interdependiente complejo | **Jotai** | Modelo atómico, grafos de estado computado |
| Estado de formulario | **React Hook Form + Zod** | Inputs no controlados para performance, validación por esquema |
| Global de baja velocidad (theme, auth) | **React Context** | Incorporado, suficiente para datos que cambian raramente |

---

## Ruta de Escalamiento

```
Context API → (need more features?) → Zustand → (need time-travel / complex middleware?) → Redux
```

| Alcance | Simple | Medio | Complejo |
|---|---|---|---|
| Componente único | Context | Context | Zustand |
| Pocos componentes | Context | Zustand | Zustand |
| Muchos componentes | Zustand | Zustand | Redux |
| Global (toda la app) | Zustand | Zustand | Redux |

---

## TanStack Query — Server State (Preferido)

Maneja caché, deduplicación, reintentos, paginación y actualizaciones optimistas. Elimina la mayor parte del estado manual de loading/error.

### Query Básico

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

### Mutations con Actualizaciones Optimistas

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

### Infinite Queries (Paginación)

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

### Reglas

- Las query keys deben ser deterministas y únicas por recurso
- Usar `staleTime` para controlar la frecuencia de refetch (por defecto 0 = siempre stale)
- Prefetch al hover para navegación instantánea: `queryClient.prefetchQuery()`
- Mantener las mutations cerca del componente que las dispara

---

## Zustand — Estado Compartido del Cliente (Preferido)

API mínima, sin provider necesario, suscripciones granulares vía selectores.

### Store Básico

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

### Consumo con Selectores

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

### Patrón de Slices (Stores Grandes)

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

### Reglas

- Usar selectores siempre — nunca desestructurar el store completo
- Nombrar acciones en devtools para depuración (`set({...}, false, 'actionName')`)
- Usar middleware `persist` para datos que sobreviven al refresh (carrito, preferencias)
- Dividir en slices cuando el store supere ~10 propiedades

---

## React Context — Estado Global de Baja Velocidad

Solo para datos que cambian con poca frecuencia (theme, locale, estado de auth).

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

### Reglas

- Context creado con `createContext<Type | null>(null)` — siempre verificar null en el hook
- `useMemo` en el objeto value para prevenir re-renders innecesarios
- Máximo 3 valores en un solo context — separar si hay más
- Nunca usar Context para datos que cambian frecuentemente (usar Zustand en su lugar)
- Renderizar providers lo más profundo posible en el árbol

---

## React Hook Form + Zod — Estado de Formulario

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

### Reglas

- El esquema Zod es la única fuente de verdad para validación
- `z.infer<typeof schema>` genera el tipo TypeScript
- Inputs no controlados por defecto (performance) — usar `watch()` solo cuando sea necesario
- Compartir esquemas entre frontend y backend

---

## Redux Toolkit — Estado Empresarial Complejo

Solo cuando necesitas depuración con viaje en el tiempo, middleware complejo o auditorías.

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

### Reglas

- Siempre usar RTK (Redux Toolkit) — nunca Redux plano
- Siempre usar hooks tipados (`useAppSelector`, `useAppDispatch`)
- RTK Query para server state (similar a TanStack Query)
- Envolver con facade hooks — los componentes nunca importan `useAppSelector` directamente
