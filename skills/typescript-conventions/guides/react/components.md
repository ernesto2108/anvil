# React Component Patterns Guide

## Function Components Only

Never use class components. All components are functions. All lifecycle logic lives in hooks.

```typescript
// WRONG: class component
class UserCard extends React.Component<Props, State> {
  state = { loading: false };
  render() { ... }
}

// RIGHT: function component
function UserCard({ userId, onSelect }: UserCardProps): React.ReactElement {
  const [loading, setLoading] = useState(false);
  ...
  return <div>...</div>;
}
```

## Props Typing with `interface`

Use `interface` for props (not `type`). This gives better error messages and allows declaration merging in libraries.

```typescript
// WRONG: type alias for props
type ButtonProps = {
  label: string;
  onClick: () => void;
};

// RIGHT: interface
interface ButtonProps {
  label: string;
  onClick: () => void;
  variant?: "primary" | "secondary" | "ghost";
  disabled?: boolean;
  className?: string;
  children?: React.ReactNode;
}

// RIGHT: extend HTML element attributes when wrapping native elements
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

## Hooks Rules

- Call hooks only at the top level — never inside conditions, loops, or nested functions
- Custom hooks start with `use` and encapsulate related stateful logic
- Extract complex `useEffect` into custom hooks — components should not contain data-fetching logic

```typescript
// WRONG: data fetching in component body
function UserProfile({ id }: { id: string }) {
  const [user, setUser] = useState<User | null>(null);
  useEffect(() => {
    fetch(`/api/users/${id}`).then(r => r.json()).then(setUser);
  }, [id]);
  ...
}

// RIGHT: extract to custom hook
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

Validate forms at the boundary with Zod. Never write manual validation in components.

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

Server Components fetch data directly without `useEffect`. They cannot use hooks, context, or browser APIs. Mark client components explicitly with `"use client"`.

```typescript
// app/users/[id]/page.tsx — Server Component (default)
// Can be async, fetches at render time
async function UserPage({ params }: { params: { id: string } }): Promise<React.ReactElement> {
  // Direct DB/service call — no useEffect, no loading state
  const user = await userService.findById(params.id);
  if (!user) notFound();

  return (
    <main>
      <UserProfile user={user} />
      <UserActions userId={user.id} /> {/* Client Component for interactivity */}
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

## Suspense and Error Boundaries

Wrap async Server Components and lazy-loaded components with `<Suspense>`. Always provide an error boundary.

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

// Lazy loading with explicit suspense boundary
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

// Error boundary for isolated sections
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
