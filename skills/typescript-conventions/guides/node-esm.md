# Node.js ESM Patterns (2025)

## Package Setup

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

## Import Rules

```typescript
// RIGHT — node: protocol for builtins
import { readFile } from "node:fs/promises";
import { join } from "node:path";
import { setTimeout } from "node:timers/promises";

// RIGHT — .js extension required in NodeNext (even for .ts source)
import { helper } from "./utils.js";
import type { Config } from "./types.js";

// RIGHT — top-level await (ESM only)
const config = JSON.parse(await readFile("./config.json", "utf-8"));
```

## `verbatimModuleSyntax`

```typescript
// RIGHT — explicit type vs value imports
import type { User } from "./models.js";   // erased at runtime
import { validateUser } from "./models.js"; // kept at runtime

// WRONG — ambiguous without verbatimModuleSyntax
import { User, validateUser } from "./models"; // which is type?
```

## `Promise.withResolvers()` (ES2024)

```typescript
// WRONG — manual deferred pattern
let resolve: (value: string) => void;
let reject: (error: Error) => void;
const promise = new Promise<string>((res, rej) => {
  resolve = res;
  reject = rej;
});

// RIGHT — Promise.withResolvers (TS 5.4+)
const { promise, resolve, reject } = Promise.withResolvers<string>();
```

## Web Standard APIs (Portable)

```typescript
// Use these — they work in Node 18+, Bun, Deno, and browsers
const response = await fetch("https://api.example.com/data");
const url = new URL("/path", "https://example.com");
const encoded = new TextEncoder().encode("hello");
const id = crypto.randomUUID();
const bytes = crypto.getRandomValues(new Uint8Array(32));
```

## Subpath Exports (Replaces Path Aliases in Libraries)

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
// Consumer imports
import { formatCurrency } from "@myapp/shared/utils";
import { emailSchema } from "@myapp/shared/validators";
```
