# Patrones de Node.js ESM (2025)

## Configuración del Paquete

```jsonc
// package.json
{
  "type": "module",
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "import": "./dist/index.js",
      "require": "./dist/index.cjs"
    }
  }
}
```

```jsonc
// tsconfig.json
{
  "compilerOptions": {
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "verbatimModuleSyntax": true
  }
}
```

## Reglas de Import

```typescript
// CORRECTO — protocolo node: para módulos built-in
import { readFile } from "node:fs/promises";
import { join } from "node:path";
import { setTimeout } from "node:timers/promises";

// CORRECTO — extensión .js requerida en NodeNext (incluso para fuente .ts)
import { helper } from "./utils.js";
import type { Config } from "./types.js";

// CORRECTO — top-level await (solo ESM)
const config = JSON.parse(await readFile("./config.json", "utf-8"));
```

## `verbatimModuleSyntax`

```typescript
// CORRECTO — imports explícitos de tipo vs valor
import type { User } from "./models.js";   // borrado en tiempo de ejecución
import { validateUser } from "./models.js"; // conservado en tiempo de ejecución

// INCORRECTO — ambiguo sin verbatimModuleSyntax
import { User, validateUser } from "./models"; // ¿cuál es el tipo?
```

## `Promise.withResolvers()` (ES2024)

```typescript
// INCORRECTO — patrón deferred manual
let resolve: (value: string) => void;
let reject: (error: Error) => void;
const promise = new Promise<string>((res, rej) => {
  resolve = res;
  reject = rej;
});

// CORRECTO — Promise.withResolvers (TS 5.4+)
const { promise, resolve, reject } = Promise.withResolvers<string>();
```

## APIs Web Estándar (Portables)

```typescript
// Usar estas — funcionan en Node 18+, Bun, Deno y navegadores
const response = await fetch("https://api.example.com/data");
const url = new URL("/path", "https://example.com");
const encoded = new TextEncoder().encode("hello");
const id = crypto.randomUUID();
const bytes = crypto.getRandomValues(new Uint8Array(32));
```

## Subpath Exports (Reemplaza los Path Aliases en Librerías)

```jsonc
// package.json
{
  "exports": {
    ".": "./dist/index.js",
    "./utils": "./dist/utils/index.js",
    "./validators": "./dist/validators/index.js"
  }
}
```

```typescript
// Imports del consumidor
import { formatCurrency } from "@myapp/shared/utils";
import { emailSchema } from "@myapp/shared/validators";
```
