# Guía de Patrones de Tipos

## Discriminated Unions (Tagged Unions)

El campo discriminante (tag) debe ser un tipo literal. Usar `switch` con verificación exhaustiva.

```typescript
// INCORRECTO: flags booleanos — explosión combinatoria, sin exhaustividad
type Shape = {
  isCircle: boolean;
  isRect: boolean;
  radius?: number;
  width?: number;
  height?: number;
};

// CORRECTO: discriminated union — cada variante es autocontenida
type Shape =
  | { kind: "circle"; radius: number }
  | { kind: "rect"; width: number; height: number }
  | { kind: "triangle"; base: number; height: number };

function area(shape: Shape): number {
  switch (shape.kind) {
    case "circle":
      return Math.PI * shape.radius ** 2; // shape es { kind: "circle"; radius: number }
    case "rect":
      return shape.width * shape.height;
    case "triangle":
      return 0.5 * shape.base * shape.height;
    default:
      // Verificación exhaustiva — TypeScript da error si falta un case
      const _exhaustive: never = shape;
      throw new Error(`Unhandled shape: ${JSON.stringify(_exhaustive)}`);
  }
}
```

## Switch Exhaustivo con `never`

Siempre agregar una verificación exhaustiva `never` en switches sobre discriminated unions. Esto convierte agregar nuevas variantes en un error de compilación.

```typescript
function assertNever(value: never, message?: string): never {
  throw new Error(message ?? `Unexpected value: ${JSON.stringify(value)}`);
}

// Uso
function handleStatus(status: "pending" | "active" | "closed"): string {
  switch (status) {
    case "pending": return "Awaiting review";
    case "active":  return "In progress";
    case "closed":  return "Done";
    default:        return assertNever(status); // error de compilación si status se expande
  }
}
```

## Branded Types

Previenen mezclar valores de tipos primitivos idénticos. La marca se borra en tiempo de ejecución — cero overhead.

```typescript
// Patrón de declaración
declare const __brand: unique symbol;
type Brand<T, B> = T & { readonly [__brand]: B };

// Tipos branded específicos
type UserId  = Brand<string, "UserId">;
type OrderId = Brand<string, "OrderId">;
type Cents   = Brand<number, "Cents">;

// Funciones constructoras — único punto de aserción
const UserId  = (raw: string): UserId   => raw as UserId;
const OrderId = (raw: string): OrderId  => raw as OrderId;
const Cents   = (raw: number): Cents    => raw as Cents;

// Uso — el compilador rechaza mezclas
function processOrder(userId: UserId, orderId: OrderId, amount: Cents) { ... }

const uid = UserId("u_123");
const oid = OrderId("o_456");
const amt = Cents(2500);

processOrder(uid, oid, amt);   // OK
processOrder(oid, uid, amt);   // Error: OrderId no asignable a UserId
processOrder(uid, oid, 2500);  // Error: number no asignable a Cents
```

## Mapped Types

Transformar propiedades de un tipo existente de forma sistemática.

```typescript
// Hacer todas las propiedades nullable
type Nullable<T> = { [K in keyof T]: T[K] | null };

// Hacer claves específicas requeridas
type RequireFields<T, K extends keyof T> = T & Required<Pick<T, K>>;

// Deep readonly
type DeepReadonly<T> = {
  readonly [K in keyof T]: T[K] extends object ? DeepReadonly<T[K]> : T[K];
};

// Renombrar claves con prefijo
type Prefixed<T, P extends string> = {
  [K in keyof T as `${P}${Capitalize<string & K>}`]: T[K];
};

// Ejemplo: Prefixed<User, "user"> → { userId: string; userName: string }
```

## Conditional Types

Usar conditional types para lógica a nivel de tipos. Combinar con `infer` para extracción.

```typescript
// Desenvolver Promise
type Awaited<T> = T extends Promise<infer U> ? U : T;

// Desenvolver elemento de array
type ElementOf<T> = T extends (infer E)[] ? E : never;

// Extraer tipo de retorno de función (equivalente a ReturnType<>)
type Return<T> = T extends (...args: any[]) => infer R ? R : never;

// Non-nullable — eliminar null y undefined de la unión
type NonNullable<T> = T extends null | undefined ? never : T;

// Distribuir sobre una unión
type ToArray<T> = T extends any ? T[] : never;
// ToArray<string | number> → string[] | number[]  (distribuido, no (string | number)[])
```

## Template Literal Types

Expresar restricciones de forma de string a nivel de tipos.

```typescript
// Nombres de event handlers
type EventName = `on${Capitalize<string>}`;

// Métodos HTTP
type HttpMethod = "GET" | "POST" | "PUT" | "DELETE" | "PATCH";
type Endpoint = `/${string}`;
type Route = `${HttpMethod} ${Endpoint}`;

// Custom properties de CSS
type CssVar = `--${string}`;
type CssVarValue = `var(${CssVar})`;

// Extraer parámetros de ruta
type ExtractParams<T extends string> =
  T extends `${string}:${infer Param}/${infer Rest}`
    ? Param | ExtractParams<`/${Rest}`>
    : T extends `${string}:${infer Param}`
    ? Param
    : never;

// ExtractParams<"/users/:userId/orders/:orderId"> → "userId" | "orderId"
type UserOrderParams = ExtractParams<"/users/:userId/orders/:orderId">;
```

## Patrones con el Operador `satisfies`

Usar `satisfies` cuando se quiere validar el tipo sin perder el tipo literal/estrechado.

```typescript
// Patrón 1: objeto de config — conservar literales, validar forma
type ThemeColors = Record<string, string | [number, number, number]>;

const THEME = {
  primary: "#3b82f6",
  accent: [59, 130, 246],
  danger: "#ef4444",
} satisfies ThemeColors;

THEME.primary;      // string (no string | [number, number, number])
THEME.accent;       // [number, number, number] (no string | [...])
THEME.unknown;      // Error: la propiedad no existe en ThemeColors

// Patrón 2: mapa de rutas — validar handlers, conservar claves de ruta literales
const routes = {
  "/users": handleUsers,
  "/orders": handleOrders,
} satisfies Record<`/${string}`, Handler>;
// routes["/users"] aún tiene el tipo del handler estrechado

// Patrón 3: config de discriminated union
type Action =
  | { type: "redirect"; to: string }
  | { type: "render"; component: string };

const actions = {
  login: { type: "redirect", to: "/dashboard" },
  signup: { type: "render", component: "SignupForm" },
} satisfies Record<string, Action>;
```

## Const Type Parameters (TS 5.0+)

Usar el modificador `const` en parámetros de tipo para inferir el tipo más estrecho posible de un argumento literal.

```typescript
// INCORRECTO: T inferido como string[], perdiendo los literales
function createTags<T extends string[]>(tags: T): T { return tags; }
const tags = createTags(["a", "b"]); // string[]

// CORRECTO: const T — preserva los literales
function createTags<const T extends string[]>(tags: T): T { return tags; }
const tags = createTags(["a", "b"]); // readonly ["a", "b"]
type Tag = typeof tags[number]; // "a" | "b"
```
