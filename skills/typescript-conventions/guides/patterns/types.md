# Type Patterns Guide

## Discriminated Unions (Tagged Unions)

The discriminant field (tag) must be a literal type. Use `switch` with exhaustive checking.

```typescript
// WRONG: boolean flags — combinatorial explosion, no exhaustion
type Shape = {
  isCircle: boolean;
  isRect: boolean;
  radius?: number;
  width?: number;
  height?: number;
};

// RIGHT: discriminated union — each variant is self-contained
type Shape =
  | { kind: "circle"; radius: number }
  | { kind: "rect"; width: number; height: number }
  | { kind: "triangle"; base: number; height: number };

function area(shape: Shape): number {
  switch (shape.kind) {
    case "circle":
      return Math.PI * shape.radius ** 2; // shape is { kind: "circle"; radius: number }
    case "rect":
      return shape.width * shape.height;
    case "triangle":
      return 0.5 * shape.base * shape.height;
    default:
      // Exhaustive check — TypeScript errors if a case is missing
      const _exhaustive: never = shape;
      throw new Error(`Unhandled shape: ${JSON.stringify(_exhaustive)}`);
  }
}
```

## Exhaustive Switch with `never`

Always add an exhaustive `never` check in switches over discriminated unions. This makes adding new variants a compile-time error.

```typescript
function assertNever(value: never, message?: string): never {
  throw new Error(message ?? `Unexpected value: ${JSON.stringify(value)}`);
}

// Usage
function handleStatus(status: "pending" | "active" | "closed"): string {
  switch (status) {
    case "pending": return "Awaiting review";
    case "active":  return "In progress";
    case "closed":  return "Done";
    default:        return assertNever(status); // compile error if status expands
  }
}
```

## Branded Types

Prevent mixing values of identical primitive types. The brand is erased at runtime — zero overhead.

```typescript
// Declaration pattern
declare const __brand: unique symbol;
type Brand<T, B> = T & { readonly [__brand]: B };

// Specific branded types
type UserId  = Brand<string, "UserId">;
type OrderId = Brand<string, "OrderId">;
type Cents   = Brand<number, "Cents">;

// Constructor functions — single assertion point
const UserId  = (raw: string): UserId   => raw as UserId;
const OrderId = (raw: string): OrderId  => raw as OrderId;
const Cents   = (raw: number): Cents    => raw as Cents;

// Usage — compiler rejects mixing
function processOrder(userId: UserId, orderId: OrderId, amount: Cents) { ... }

const uid = UserId("u_123");
const oid = OrderId("o_456");
const amt = Cents(2500);

processOrder(uid, oid, amt);   // OK
processOrder(oid, uid, amt);   // Error: OrderId not assignable to UserId
processOrder(uid, oid, 2500);  // Error: number not assignable to Cents
```

## Mapped Types

Transform properties of an existing type systematically.

```typescript
// Make all properties nullable
type Nullable<T> = { [K in keyof T]: T[K] | null };

// Make specific keys required
type RequireFields<T, K extends keyof T> = T & Required<Pick<T, K>>;

// Deep readonly
type DeepReadonly<T> = {
  readonly [K in keyof T]: T[K] extends object ? DeepReadonly<T[K]> : T[K];
};

// Rename keys with prefix
type Prefixed<T, P extends string> = {
  [K in keyof T as `${P}${Capitalize<string & K>}`]: T[K];
};

// Example: Prefixed<User, "user"> → { userId: string; userName: string }
```

## Conditional Types

Use conditional types for type-level logic. Combine with `infer` for extraction.

```typescript
// Unwrap Promise
type Awaited<T> = T extends Promise<infer U> ? U : T;

// Unwrap array element
type ElementOf<T> = T extends (infer E)[] ? E : never;

// Extract function return type (equivalent to ReturnType<>)
type Return<T> = T extends (...args: any[]) => infer R ? R : never;

// Non-nullable — remove null and undefined from union
type NonNullable<T> = T extends null | undefined ? never : T;

// Distribute over union
type ToArray<T> = T extends any ? T[] : never;
// ToArray<string | number> → string[] | number[]  (distributed, not (string | number)[])
```

## Template Literal Types

Express string shape constraints at the type level.

```typescript
// Event handler names
type EventName = `on${Capitalize<string>}`;

// HTTP methods
type HttpMethod = "GET" | "POST" | "PUT" | "DELETE" | "PATCH";
type Endpoint = `/${string}`;
type Route = `${HttpMethod} ${Endpoint}`;

// CSS custom properties
type CssVar = `--${string}`;
type CssVarValue = `var(${CssVar})`;

// Extract route params
type ExtractParams<T extends string> =
  T extends `${string}:${infer Param}/${infer Rest}`
    ? Param | ExtractParams<`/${Rest}`>
    : T extends `${string}:${infer Param}`
    ? Param
    : never;

// ExtractParams<"/users/:userId/orders/:orderId"> → "userId" | "orderId"
type UserOrderParams = ExtractParams<"/users/:userId/orders/:orderId">;
```

## `satisfies` Operator Patterns

Use `satisfies` when you want type validation without losing the literal/narrowed type.

```typescript
// Pattern 1: config object — keep literals, validate shape
type ThemeColors = Record<string, string | [number, number, number]>;

const THEME = {
  primary: "#3b82f6",
  accent: [59, 130, 246],
  danger: "#ef4444",
} satisfies ThemeColors;

THEME.primary;      // string (not string | [number, number, number])
THEME.accent;       // [number, number, number] (not string | [...])
THEME.unknown;      // Error: property does not exist on ThemeColors

// Pattern 2: route map — validate handlers, keep literal route keys
const routes = {
  "/users": handleUsers,
  "/orders": handleOrders,
} satisfies Record<`/${string}`, Handler>;
// routes["/users"] still has the narrowed handler type

// Pattern 3: discriminated union config
type Action =
  | { type: "redirect"; to: string }
  | { type: "render"; component: string };

const actions = {
  login: { type: "redirect", to: "/dashboard" },
  signup: { type: "render", component: "SignupForm" },
} satisfies Record<string, Action>;
```

## Const Type Parameters (TS 5.0+)

Use `const` modifier on type parameters to infer the narrowest possible type from a literal argument.

```typescript
// WRONG: T inferred as string[], losing literals
function createTags<T extends string[]>(tags: T): T { return tags; }
const tags = createTags(["a", "b"]); // string[]

// RIGHT: const T — preserves literals
function createTags<const T extends string[]>(tags: T): T { return tags; }
const tags = createTags(["a", "b"]); // readonly ["a", "b"]
type Tag = typeof tags[number]; // "a" | "b"
```
