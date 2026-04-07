# Architecture Rules

## 1. ESM First

All projects use `"type": "module"` in `package.json`. Never use `require()` or `module.exports` in new code. If a dependency is CJS-only, wrap it in a thin ESM adapter at the boundary.

```json
// package.json
{
  "type": "module",
  "exports": {
    ".": {
      "import": "./dist/index.js",
      "types": "./dist/index.d.ts"
    }
  }
}
```

## 2. No Barrel Exports

Barrel files (`index.ts` that re-exports everything) are a tree-shaking killer. They also create circular dependency risks and slow down TypeScript's module resolution.

```typescript
// WRONG: src/utils/index.ts
export { formatDate } from "./formatDate.js";
export { parseAmount } from "./parseAmount.js";
export { slugify } from "./slugify.js";
// Importing one pulls in the entire barrel at bundler analysis time

// RIGHT: direct imports
import { formatDate } from "../utils/formatDate.js";
import { parseAmount } from "../utils/parseAmount.js";
```

**Exception:** Public package entry points (the `"exports"` field) may use a carefully curated index. This is different from internal barrels.

## 3. Path Aliases via tsconfig `paths`

Use `paths` for module aliases instead of deep relative imports. Configure the same aliases in your bundler (Vite, Webpack, etc.).

```jsonc
// tsconfig.json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@app/*": ["src/*"],
      "@shared/*": ["src/shared/*"],
      "@features/*": ["src/features/*"]
    }
  }
}

// vite.config.ts — mirror the aliases
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig({ plugins: [tsconfigPaths()] });
```

```typescript
// WRONG: deep relative import
import { UserService } from "../../../../services/user/UserService.js";

// RIGHT: alias import
import { UserService } from "@features/user/UserService.js";
```

## 4. Zod at Boundaries

Runtime validation belongs only at trust boundaries: API inputs, environment variables, third-party webhook payloads, localStorage reads, form submissions. Inside the application, trust the types.

```typescript
// Environment — validate once at startup, typed everywhere after
const envSchema = z.object({
  DATABASE_URL: z.string().url(),
  PORT: z.coerce.number().int().min(1).max(65535).default(3000),
  NODE_ENV: z.enum(["development", "production", "test"]),
});

export type Env = z.infer<typeof envSchema>;

export function loadEnv(): Env {
  const result = envSchema.safeParse(process.env);
  if (!result.success) {
    console.error("Invalid environment:", result.error.flatten());
    process.exit(1);
  }
  return result.data;
}

// API handler — validate request body, not internal service calls
const createUserSchema = z.object({
  email: z.string().email(),
  name: z.string().min(1).max(100).trim(),
  role: z.enum(["admin", "member"]),
});

async function handleCreateUser(req: Request): Promise<Response> {
  const body = await req.json();
  const parsed = createUserSchema.safeParse(body);
  if (!parsed.success) {
    return Response.json({ errors: parsed.error.flatten() }, { status: 422 });
  }
  // parsed.data is fully typed — no validation inside the service
  const user = await userService.create(parsed.data);
  return Response.json(user, { status: 201 });
}
```

## 5. Dependency Injection

Pass dependencies explicitly via constructors or function parameters. Never import services as singletons in application modules. Use interfaces (TypeScript `interface`) to decouple implementations.

```typescript
// WRONG: singleton import — hard to test, hidden coupling
import { db } from "../db.js"; // global singleton

export class UserService {
  async getUser(id: UserId) {
    return db.query("SELECT * FROM users WHERE id = $1", [id]);
  }
}

// RIGHT: interface + constructor injection
interface UserRepository {
  findById(id: UserId): Promise<User | null>;
}

export class UserService {
  constructor(private readonly repo: UserRepository) {}

  async getUser(id: UserId): Promise<User | null> {
    return this.repo.findById(id);
  }
}

// Composition root (main.ts / app.ts)
const userRepo = new PostgresUserRepository(pool);
const userService = new UserService(userRepo);
```

## 6. tsconfig Strict Settings

Required `tsconfig.json` baseline for all TypeScript projects:

```jsonc
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",   // or "NodeNext" for Node.js
    "lib": ["ES2022"],
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "exactOptionalPropertyTypes": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "forceConsistentCasingInFileNames": true,
    "verbatimModuleSyntax": true,    // enforces "import type" for type-only imports
    "isolatedModules": true,         // required for bundler compatibility
    "skipLibCheck": false,           // only set true as last resort
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "outDir": "./dist"
  }
}
```

## 7. `import type` Enforced

Use `import type` for type-only imports. This is enforced by `verbatimModuleSyntax` and ensures zero runtime overhead from type imports.

```typescript
// WRONG: runtime import for types only
import { User, CreateUserDto } from "./types.js";

// RIGHT: type-only import
import type { User, CreateUserDto } from "./types.js";

// RIGHT: mixed import
import { createUser, type User } from "./user.js";
```

## 8. One Concern Per Module

A module should have one reason to change. Avoid mixing validation logic, HTTP-specific code, and business logic in the same file. Suggested structure for a feature:

```
src/features/user/
  user.schema.ts      # Zod schemas only
  user.types.ts       # TypeScript types derived from schemas
  user.service.ts     # Business logic, takes interfaces
  user.repository.ts  # Data access implementation
  user.handler.ts     # HTTP adapter, validation, response shaping
  user.test.ts        # Tests
```
