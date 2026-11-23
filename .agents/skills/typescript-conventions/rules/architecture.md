# Reglas de Arquitectura

## 1. ESM Primero

Todos los proyectos usan `"type": "module"` en `package.json`. Nunca usar `require()` o `module.exports` en código nuevo. Si una dependencia es solo CJS, envolverla en un adaptador ESM delgado en la frontera.

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

## 2. Sin Barrel Exports

Los barrel files (`index.ts` que re-exporta todo) matan el tree-shaking. También crean riesgos de dependencias circulares y ralentizan la resolución de módulos de TypeScript.

```typescript
// INCORRECTO: src/utils/index.ts
export { formatDate } from "./formatDate.js";
export { parseAmount } from "./parseAmount.js";
export { slugify } from "./slugify.js";
// Importar uno arrastra todo el barrel en el análisis del bundler

// CORRECTO: importaciones directas
import { formatDate } from "../utils/formatDate.js";
import { parseAmount } from "../utils/parseAmount.js";
```

**Excepción:** Los puntos de entrada de paquetes públicos (el campo `"exports"`) pueden usar un índice cuidadosamente curado. Esto es diferente de los barrels internos.

## 3. Path Aliases via `paths` de tsconfig

Usar `paths` para aliases de módulos en lugar de imports relativos profundos. Configurar los mismos aliases en el bundler (Vite, Webpack, etc.).

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

// vite.config.ts — reflejar los aliases
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig({ plugins: [tsconfigPaths()] });
```

```typescript
// INCORRECTO: import relativo profundo
import { UserService } from "../../../../services/user/UserService.js";

// CORRECTO: import con alias
import { UserService } from "@features/user/UserService.js";
```

## 4. Zod en las Fronteras

La validación en tiempo de ejecución pertenece solo a las fronteras de confianza: entradas de API, variables de entorno, payloads de webhooks de terceros, lecturas de localStorage, envíos de formularios. Dentro de la aplicación, confiar en los tipos.

```typescript
// Variables de entorno — validar una vez al inicio, tipadas en todas partes después
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

// Handler de API — validar el body del request, no las llamadas internas al servicio
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
  // parsed.data está completamente tipado — sin validación dentro del servicio
  const user = await userService.create(parsed.data);
  return Response.json(user, { status: 201 });
}
```

## 5. Inyección de Dependencias

Pasar dependencias explícitamente via constructores o parámetros de función. Nunca importar servicios como singletons en módulos de aplicación. Usar interfaces (TypeScript `interface`) para desacoplar implementaciones.

```typescript
// INCORRECTO: import de singleton — difícil de testear, acoplamiento oculto
import { db } from "../db.js"; // singleton global

export class UserService {
  async getUser(id: UserId) {
    return db.query("SELECT * FROM users WHERE id = $1", [id]);
  }
}

// CORRECTO: interface + inyección por constructor
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

## 6. Configuración Strict de tsconfig

Baseline requerida de `tsconfig.json` para todos los proyectos TypeScript:

```jsonc
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",   // o "NodeNext" para Node.js
    "lib": ["ES2022"],
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "exactOptionalPropertyTypes": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "forceConsistentCasingInFileNames": true,
    "verbatimModuleSyntax": true,    // fuerza "import type" para imports de solo tipo
    "isolatedModules": true,         // requerido para compatibilidad con bundlers
    "skipLibCheck": false,           // establecer en true solo como último recurso
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "outDir": "./dist"
  }
}
```

## 7. `import type` Forzado

Usar `import type` para imports de solo tipo. Esto es forzado por `verbatimModuleSyntax` y garantiza cero overhead en tiempo de ejecución por imports de tipos.

```typescript
// INCORRECTO: import en tiempo de ejecución para solo tipos
import { User, CreateUserDto } from "./types.js";

// CORRECTO: import de solo tipo
import type { User, CreateUserDto } from "./types.js";

// CORRECTO: import mixto
import { createUser, type User } from "./user.js";
```

## 8. Un Concern por Módulo

Un módulo debe tener un solo motivo para cambiar. Evitar mezclar lógica de validación, código específico de HTTP y lógica de negocio en el mismo archivo. Estructura sugerida para un feature:

```
src/features/user/
  user.schema.ts      # Solo esquemas Zod
  user.types.ts       # Tipos TypeScript derivados de los esquemas
  user.service.ts     # Lógica de negocio, toma interfaces
  user.repository.ts  # Implementación de acceso a datos
  user.handler.ts     # Adaptador HTTP, validación, formato de respuesta
  user.test.ts        # Tests
```
