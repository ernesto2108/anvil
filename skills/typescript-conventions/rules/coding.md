# Coding Rules

## Strict Mode Settings

Always enable these in `tsconfig.json`. The `"strict": true` umbrella flag is not enough:

```jsonc
{
  "compilerOptions": {
    "strict": true,                        // enables 8 strict checks
    "noUncheckedIndexedAccess": true,      // arr[0] is T | undefined, not T
    "exactOptionalPropertyTypes": true,    // { a?: string } ≠ { a: string | undefined }
    "noImplicitReturns": true,             // every code path must return
    "noFallthroughCasesInSwitch": true,    // switch exhaustion
    "noUncheckedSideEffectImports": true   // TS 5.6+: explicit side-effect imports
  }
}
```

Never disable `strictNullChecks` per-file. If you must opt out of a single check, use a scoped `@ts-expect-error` with a reason.

## `satisfies` Operator

Use `satisfies` to validate a value against a type without widening the inferred type. This is especially useful for config objects and literal type narrowing.

```typescript
// WRONG: type assertion loses literal information
const palette = {
  red: [255, 0, 0],
  green: "#00ff00",
} as Record<string, string | number[]>;
// palette.red is now string | number[], not [number, number, number]

// RIGHT: satisfies validates without widening
const palette = {
  red: [255, 0, 0],
  green: "#00ff00",
} satisfies Record<string, string | number[]>;
// palette.red is still [number, number, number]
palette.red.at(0); // OK — tuple preserved
```

## Const Assertions

Use `as const` to freeze literal types for objects used as lookup tables or union sources.

```typescript
// WRONG: type is { statuses: string[] }
const CONFIG = { statuses: ["active", "inactive"] };

// RIGHT: type is { readonly statuses: readonly ["active", "inactive"] }
const CONFIG = { statuses: ["active", "inactive"] } as const;
type Status = typeof CONFIG.statuses[number]; // "active" | "inactive"
```

## Discriminated Unions (Not Enums)

Enums have poor ergonomics, are not tree-shakeable, and require an import to check values.

```typescript
// WRONG: enum
enum Direction { Up = "UP", Down = "DOWN" }

// RIGHT: const union — serializable, composable, no import needed at runtime
const DIRECTIONS = ["up", "down"] as const;
type Direction = typeof DIRECTIONS[number]; // "up" | "down"

// RIGHT: tagged union for algebraic types
type Result<T> =
  | { status: "ok"; data: T }
  | { status: "error"; code: string; message: string };

function handle(result: Result<User>) {
  switch (result.status) {
    case "ok":
      return result.data; // data: User — TypeScript narrows correctly
    case "error":
      return result.message; // message: string
  }
}
```

## Template Literal Types

Use template literal types to express string constraints at the type level.

```typescript
type EventName = `on${Capitalize<string>}`;
type CssUnit = `${number}${"px" | "rem" | "em" | "%"}`;
type RouteParam<T extends string> = T extends `${infer _}:${infer Param}/${infer Rest}`
  ? Param | RouteParam<Rest>
  : T extends `${infer _}:${infer Param}`
  ? Param
  : never;
```

## `using` Declarations (Explicit Resource Management — TS 5.2+)

Use `using` for any resource that has a `[Symbol.dispose]()` method. Never rely on `try/finally` for cleanup when `using` is available.

```typescript
// WRONG: manual cleanup, forgettable
const handle = await openFile("data.csv");
try {
  await processFile(handle);
} finally {
  handle.close(); // easy to forget or skip on early return
}

// RIGHT: using — cleanup is guaranteed regardless of throw/return
await using handle = await openFile("data.csv");
await processFile(handle); // handle.close() called automatically
```

## Branded / Nominal Types

Use branded types to prevent mixing up values of the same primitive type (e.g., `UserId` vs `OrderId`).

```typescript
// WRONG: all just strings — easy to pass wrong id
function getOrder(userId: string, orderId: string) { ... }

// RIGHT: branded types — compiler rejects wrong arguments
type UserId = string & { readonly __brand: unique symbol };
type OrderId = string & { readonly __brand: unique symbol };

function createUserId(raw: string): UserId {
  return raw as UserId; // single casting point
}

function getOrder(userId: UserId, orderId: OrderId) { ... }

const uid = createUserId("u_123");
const oid = createOrderId("o_456");
getOrder(uid, oid);   // OK
getOrder(oid, uid);   // TypeScript error
```

## `NoInfer<T>` Utility (TS 5.4+)

Use `NoInfer<T>` to block inference on a specific type parameter position, forcing callers to be explicit.

```typescript
// WRONG: TS infers T from both positions, accepting wrong fallbacks
function createStore<T>(initial: T, fallback: T): Store<T> { ... }
createStore({ a: 1 }, { b: 2 }); // T inferred as { a?: number; b?: number }

// RIGHT: NoInfer blocks inference from fallback
function createStore<T>(initial: T, fallback: NoInfer<T>): Store<T> { ... }
createStore({ a: 1 }, { b: 2 }); // Error: { b: number } not assignable to { a: number }
```

## Error Handling

Never throw raw strings. Never catch `any`. Libraries must not throw — return a `Result` type.

```typescript
// WRONG: throw string, catch any
throw "something went wrong";
catch (e: any) { console.log(e.message) }

// RIGHT: typed errors with discriminated result
type AppError =
  | { kind: "not-found"; resource: string }
  | { kind: "unauthorized" }
  | { kind: "validation"; fields: Record<string, string> };

type Result<T, E = AppError> = { ok: true; value: T } | { ok: false; error: E };

// Library code — never throws, always returns Result
function parseConfig(raw: unknown): Result<Config> {
  const parsed = configSchema.safeParse(raw);
  if (!parsed.success) {
    return { ok: false, error: { kind: "validation", fields: formatZodErrors(parsed.error) } };
  }
  return { ok: true, value: parsed.data };
}

// Catch with unknown — always narrow before using
try {
  await fetch(url);
} catch (e: unknown) {
  if (e instanceof TypeError) { ... }
  throw e; // re-throw what you can't handle
}
```
