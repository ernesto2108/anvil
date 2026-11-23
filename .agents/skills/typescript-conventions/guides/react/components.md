# Guía de Patrones de Componentes React

## Solo Componentes Función

Nunca usar componentes de clase. Todos los componentes son funciones. Toda la lógica de ciclo de vida vive en hooks.

```typescript
// INCORRECTO: componente de clase
class UserCard extends React.Component<Props, State> {
  state = { loading: false };
  render() { ... }
}

// CORRECTO: componente función
function UserCard({ userId, onSelect }: UserCardProps): React.ReactElement {
  const [loading, setLoading] = useState(false);
  ...
  return <div>...</div>;
}
```

## Tipado de Props con `interface`

Usar `interface` para props (no `type`). Esto produce mejores mensajes de error y permite declaration merging en librerías.

```typescript
// INCORRECTO: type alias para props
type ButtonProps = {
  label: string;
  onClick: () => void;
};

// CORRECTO: interface
interface ButtonProps {
  label: string;
  onClick: () => void;
  variant?: "primary" | "secondary" | "ghost";
  disabled?: boolean;
  className?: string;
  children?: React.ReactNode;
}

// CORRECTO: extender atributos de elementos HTML al envolver elementos nativos
interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
  hint?: string;
}

function Input({ label, error, hint, ...inputProps }: InputProps): React.ReactElement {
  return (
    <div>
      <label>{label}</label>
      <input {...inputProps} aria-invalid={!!error} />
      {error && <p role="alert">{error}</p>}
      {hint && <p>{hint}</p>}
    </div>
  );
}
```

## Reglas de Hooks

- Llamar hooks solo en el nivel superior — nunca dentro de condiciones, loops o funciones anidadas
- Los custom hooks comienzan con `use` y encapsulan lógica stateful relacionada
- Extraer `useEffect` complejos a custom hooks — los componentes no deben contener lógica de data-fetching

```typescript
// INCORRECTO: data fetching en el cuerpo del componente
function UserProfile({ id }: { id: string }) {
  const [user, setUser] = useState<User | null>(null);
  useEffect(() => {
    fetch(`/api/users/${id}`).then(r => r.json()).then(setUser);
  }, [id]);
  ...
}

// CORRECTO: extraer a custom hook
function useUser(id: string): { user: User | null; loading: boolean; error: Error | null } {
  const [state, setState] = useState<{
    user: User | null;
    loading: boolean;
    error: Error | null;
  }>({ user: null, loading: true, error: null });

  useEffect(() => {
    const controller = new AbortController();
    setState(prev => ({ ...prev, loading: true, error: null }));

    fetch(`/api/users/${id}`, { signal: controller.signal })
      .then(r => {
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        return r.json() as Promise<User>;
      })
      .then(user => setState({ user, loading: false, error: null }))
      .catch(e => {
        if (e.name !== "AbortError") {
          setState({ user: null, loading: false, error: e });
        }
      });

    return () => controller.abort();
  }, [id]);

  return state;
}

function UserProfile({ id }: { id: string }) {
  const { user, loading, error } = useUser(id);
  if (loading) return <Spinner />;
  if (error) return <ErrorMessage error={error} />;
  return <div>{user?.name}</div>;
}
```

## Zod + react-hook-form

Validar formularios en la frontera con Zod. Nunca escribir validación manual en los componentes.

```typescript
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";

const loginSchema = z.object({
  email: z.string().email("Enter a valid email"),
  password: z.string().min(8, "Password must be at least 8 characters"),
});

type LoginForm = z.infer<typeof loginSchema>;

interface LoginFormProps {
  onSuccess: (email: string) => void;
}

function LoginForm({ onSuccess }: LoginFormProps): React.ReactElement {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>({ resolver: zodResolver(loginSchema) });

  async function onSubmit(data: LoginForm) {
    await authService.login(data.email, data.password);
    onSuccess(data.email);
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <Input label="Email" type="email" error={errors.email?.message} {...register("email")} />
      <Input label="Password" type="password" error={errors.password?.message} {...register("password")} />
      <button type="submit" disabled={isSubmitting}>
        {isSubmitting ? "Signing in..." : "Sign in"}
      </button>
    </form>
  );
}
```

## Server Components (RSC)

Los Server Components obtienen datos directamente sin `useEffect`. No pueden usar hooks, context ni APIs del navegador. Marcar los componentes cliente explícitamente con `"use client"`.

```typescript
// app/users/[id]/page.tsx — Server Component (por defecto)
// Puede ser async, hace fetch en tiempo de render
async function UserPage({ params }: { params: { id: string } }): Promise<React.ReactElement> {
  // Llamada directa a DB/servicio — sin useEffect, sin estado de carga
  const user = await userService.findById(params.id);
  if (!user) notFound();

  return (
    <main>
      <UserProfile user={user} />
      <UserActions userId={user.id} /> {/* Client Component para interactividad */}
    </main>
  );
}

// components/UserActions.tsx — Client Component
"use client";

interface UserActionsProps {
  userId: string;
}

function UserActions({ userId }: UserActionsProps): React.ReactElement {
  function handleDelete() { ... }
  return <button onClick={handleDelete}>Delete</button>;
}
```

## Suspense y Error Boundaries

Envolver Server Components asíncronos y componentes cargados de forma lazy con `<Suspense>`. Siempre proveer un error boundary.

```typescript
// app/layout.tsx
import { Suspense } from "react";
import { ErrorBoundary } from "react-error-boundary";

function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <ErrorBoundary fallback={<GlobalError />}>
      <Suspense fallback={<PageSkeleton />}>
        {children}
      </Suspense>
    </ErrorBoundary>
  );
}

// Lazy loading con boundary de suspense explícito
const HeavyChart = lazy(() => import("./HeavyChart.js"));

function Dashboard() {
  return (
    <section>
      <Suspense fallback={<ChartSkeleton />}>
        <HeavyChart data={data} />
      </Suspense>
    </section>
  );
}

// Error boundary para secciones aisladas
function UserSection({ userId }: { userId: string }) {
  return (
    <ErrorBoundary
      fallbackRender={({ error }) => <SectionError message={error.message} />}
      onError={(error) => logger.error("UserSection failed", { error })}
    >
      <Suspense fallback={<UserSkeleton />}>
        <UserProfile userId={userId} />
      </Suspense>
    </ErrorBoundary>
  );
}
```
