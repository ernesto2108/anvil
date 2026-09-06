# Reglas de Codificación

## Configuración de Strict Mode

Siempre habilitar estas opciones en `tsconfig.json`. El flag umbrella `"strict": true` no es suficiente:

```jsonc
{
  "compilerOptions": {
    "strict": true,                        // habilita 8 verificaciones estrictas
    "noUncheckedIndexedAccess": true,      // arr[0] es T | undefined, no T
    "exactOptionalPropertyTypes": true,    // { a?: string } ≠ { a: string | undefined }
    "noImplicitReturns": true,             // cada camino de código debe retornar
    "noFallthroughCasesInSwitch": true,    // exhaustividad de switch
    "noUncheckedSideEffectImports": true   // TS 5.6+: imports de efectos secundarios explícitos
  }
}
```

Nunca deshabilitar `strictNullChecks` por archivo. Si es necesario omitir una sola verificación, usar un `@ts-expect-error` con alcance específico y motivo.

## Operador `satisfies`

Usar `satisfies` para validar un valor contra un tipo sin ampliar el tipo inferido. Es especialmente útil para objetos de configuración y estrechamiento de tipos literales.

```typescript
// INCORRECTO: la aserción de tipo pierde información de literales
const palette = {
  red: [255, 0, 0],
  green: "#00ff00",
} as Record<string, string | number[]>;
// palette.red ahora es string | number[], no [number, number, number]

// CORRECTO: satisfies valida sin ampliar
const palette = {
  red: [255, 0, 0],
  green: "#00ff00",
} satisfies Record<string, string | number[]>;
// palette.red sigue siendo [number, number, number]
palette.red.at(0); // OK — tupla preservada
```

## Const Assertions

Usar `as const` para congelar tipos literales de objetos usados como tablas de lookup o fuentes de uniones.

```typescript
// INCORRECTO: el tipo es { statuses: string[] }
const CONFIG = { statuses: ["active", "inactive"] };

// CORRECTO: el tipo es { readonly statuses: readonly ["active", "inactive"] }
const CONFIG = { statuses: ["active", "inactive"] } as const;
type Status = typeof CONFIG.statuses[number]; // "active" | "inactive"
```

## Discriminated Unions (No Enums)

Los enums tienen ergonomía pobre, no son tree-shakeable y requieren un import para verificar valores.

```typescript
// INCORRECTO: enum
enum Direction { Up = "UP", Down = "DOWN" }

// CORRECTO: unión const — serializable, composable, sin import necesario en tiempo de ejecución
const DIRECTIONS = ["up", "down"] as const;
type Direction = typeof DIRECTIONS[number]; // "up" | "down"

// CORRECTO: tagged union para tipos algebraicos
type Result<T> =
  | { status: "ok"; data: T }
  | { status: "error"; code: string; message: string };

function handle(result: Result<User>) {
  switch (result.status) {
    case "ok":
      return result.data; // data: User — TypeScript estrecha correctamente
    case "error":
      return result.message; // message: string
  }
}
```

## Template Literal Types

Usar template literal types para expresar restricciones de string a nivel de tipos.

```typescript
type EventName = `on${Capitalize<string>}`;
type CssUnit = `${number}${"px" | "rem" | "em" | "%"}`;
type RouteParam<T extends string> = T extends `${infer _}:${infer Param}/${infer Rest}`
  ? Param | RouteParam<Rest>
  : T extends `${infer _}:${infer Param}`
  ? Param
  : never;
```

## Declaraciones `using` (Gestión Explícita de Recursos — TS 5.2+)

Usar `using` para cualquier recurso que tenga un método `[Symbol.dispose]()`. Nunca depender de `try/finally` para limpieza cuando `using` esté disponible.

```typescript
// INCORRECTO: limpieza manual, olvidable
const handle = await openFile("data.csv");
try {
  await processFile(handle);
} finally {
  handle.close(); // fácil de olvidar o saltarse en un return temprano
}

// CORRECTO: using — la limpieza está garantizada independientemente de throw/return
await using handle = await openFile("data.csv");
await processFile(handle); // handle.close() se llama automáticamente
```

## Branded / Nominal Types

Usar branded types para prevenir mezclar valores del mismo tipo primitivo (e.g., `UserId` vs `OrderId`).

```typescript
// INCORRECTO: todos son solo strings — fácil pasar el id equivocado
function getOrder(userId: string, orderId: string) { ... }

// CORRECTO: branded types — el compilador rechaza argumentos incorrectos
type UserId = string & { readonly __brand: unique symbol };
type OrderId = string & { readonly __brand: unique symbol };

function createUserId(raw: string): UserId {
  return raw as UserId; // único punto de casting
}

function getOrder(userId: UserId, orderId: OrderId) { ... }

const uid = createUserId("u_123");
const oid = createOrderId("o_456");
getOrder(uid, oid);   // OK
getOrder(oid, uid);   // Error de TypeScript
```

## Utilitario `NoInfer<T>` (TS 5.4+)

Usar `NoInfer<T>` para bloquear la inferencia en una posición específica de parámetro de tipo, forzando a los callers a ser explícitos.

```typescript
// INCORRECTO: TS infiere T de ambas posiciones, aceptando fallbacks incorrectos
function createStore<T>(initial: T, fallback: T): Store<T> { ... }
createStore({ a: 1 }, { b: 2 }); // T inferido como { a?: number; b?: number }

// CORRECTO: NoInfer bloquea la inferencia desde fallback
function createStore<T>(initial: T, fallback: NoInfer<T>): Store<T> { ... }
createStore({ a: 1 }, { b: 2 }); // Error: { b: number } no asignable a { a: number }
```

## Manejo de Errores

Nunca lanzar strings crudos. Nunca capturar `any`. Las librerías no deben lanzar excepciones — retornar un tipo `Result`.

```typescript
// INCORRECTO: lanzar string, capturar any
throw "something went wrong";
catch (e: any) { console.log(e.message) }

// CORRECTO: errores tipados con resultado discriminado
type AppError =
  | { kind: "not-found"; resource: string }
  | { kind: "unauthorized" }
  | { kind: "validation"; fields: Record<string, string> };

type Result<T, E = AppError> = { ok: true; value: T } | { ok: false; error: E };

// Código de librería — nunca lanza, siempre retorna Result
function parseConfig(raw: unknown): Result<Config> {
  const parsed = configSchema.safeParse(raw);
  if (!parsed.success) {
    return { ok: false, error: { kind: "validation", fields: formatZodErrors(parsed.error) } };
  }
  return { ok: true, value: parsed.data };
}

// Capturar con unknown — siempre estrechar antes de usar
try {
  await fetch(url);
} catch (e: unknown) {
  if (e instanceof TypeError) { ... }
  throw e; // re-lanzar lo que no puedes manejar
}
```
