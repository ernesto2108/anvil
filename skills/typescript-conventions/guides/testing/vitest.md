# Guía de Testing con Vitest

## Configuración

Vitest es el test runner estándar. Es ESM nativo, no requiere transformación con Babel y comparte la configuración de Vite.

```typescript
// vitest.config.ts
import { defineConfig } from "vitest/config";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig({
  plugins: [tsconfigPaths()],
  test: {
    globals: false,          // preferir imports explícitos — sin `describe`/`it` globales
    environment: "node",     // usar "jsdom" para tests de DOM
    coverage: {
      provider: "v8",
      reporter: ["text", "lcov"],
      include: ["src/**/*.ts"],
      exclude: ["src/**/*.test.ts", "src/**/*.d.ts"],
    },
    typecheck: {
      enabled: true,         // ejecutar tsc junto con los tests
      tsconfig: "./tsconfig.json",
    },
  },
});
```

```json
// scripts de package.json
{
  "test": "vitest run",
  "test:watch": "vitest",
  "test:ui": "vitest --ui",
  "test:coverage": "vitest run --coverage",
  "typecheck": "tsc --noEmit"
}
```

## Estructura de Archivos de Test

```typescript
// user.service.test.ts
import { describe, it, expect, beforeEach, vi } from "vitest";
import { UserService } from "./user.service.js";
import type { UserRepository } from "./user.repository.js";

describe("UserService", () => {
  let service: UserService;
  let repo: UserRepository;

  beforeEach(() => {
    // Mock fresco por test — nunca compartir estado entre tests
    repo = {
      findById: vi.fn(),
      save: vi.fn(),
    };
    service = new UserService(repo);
  });

  describe("getUser", () => {
    it("returns user when found", async () => {
      const user = { id: "u_1", name: "Alice", email: "alice@example.com" };
      vi.mocked(repo.findById).mockResolvedValueOnce(user);

      const result = await service.getUser("u_1" as UserId);

      expect(result).toEqual(user);
      expect(repo.findById).toHaveBeenCalledWith("u_1");
      expect(repo.findById).toHaveBeenCalledOnce();
    });

    it("returns null when not found", async () => {
      vi.mocked(repo.findById).mockResolvedValueOnce(null);

      const result = await service.getUser("u_missing" as UserId);

      expect(result).toBeNull();
    });
  });
});
```

## `expectTypeOf` — Tests a Nivel de Tipos

Usar `expectTypeOf` para afirmar tipos de TypeScript en los tests. Esto se ejecuta a través de `tsc`, no en tiempo de ejecución.

```typescript
import { expectTypeOf, describe, it } from "vitest";
import type { Result } from "./result.js";
import { parseConfig } from "./config.js";

describe("type assertions", () => {
  it("Result<T> ok branch has data", () => {
    type OkResult = Extract<Result<number>, { ok: true }>;
    expectTypeOf<OkResult>().toHaveProperty("value");
    expectTypeOf<OkResult["value"]>().toEqualTypeOf<number>();
  });

  it("parseConfig returns Result<Config>", () => {
    expectTypeOf(parseConfig).returns.toMatchTypeOf<Result<Config>>();
  });

  it("UserId is not assignable to OrderId", () => {
    expectTypeOf<UserId>().not.toMatchTypeOf<OrderId>();
  });
});
```

## Estrategias de Mocking

### Mocking de interfaces con jest-mock-extended (OBLIGATORIO)

**Nunca crear objetos mock manuales para interfaces.** Los mocks manuales (`{ findById: vi.fn(), save: vi.fn() }`) pueden divergir de la interfaz real — el test pasa pero el código está roto. Usar `jest-mock-extended` que genera mocks type-safe desde interfaces TypeScript.

#### Setup

```bash
pnpm add -D jest-mock-extended
```

#### Uso

```typescript
import { mock, MockProxy } from "jest-mock-extended";
import type { UserRepository } from "./user.repository.js";

describe("UserService", () => {
  let repo: MockProxy<UserRepository>;
  let service: UserService;

  beforeEach(() => {
    repo = mock<UserRepository>();
    service = new UserService(repo);
  });

  it("returns user when found", async () => {
    const user = { id: "u_1", name: "Alice", email: "alice@example.com" };
    repo.findById.calledWith("u_1").mockResolvedValue(user);

    const result = await service.getUser("u_1" as UserId);

    expect(result).toEqual(user);
    expect(repo.findById).toHaveBeenCalledWith("u_1");
  });

  it("returns null when not found", async () => {
    repo.findById.mockResolvedValue(null);

    const result = await service.getUser("u_missing" as UserId);

    expect(result).toBeNull();
  });
});
```

Si `UserRepository` gana un nuevo método, `MockProxy<UserRepository>` lo incluye automáticamente — TypeScript falla al compilar si el mock no coincide.

#### Reglas

- `mock<Interface>()` para toda interfaz de servicio/repositorio — nunca objetos literales con `vi.fn()`
- `.calledWith(args).mockReturnValue(val)` para respuestas condicionales
- Funciona con Vitest sin configuración extra (compatible con la API de jest)
- **Si jest-mock-extended no está instalado** — NO recurrir a mocks manuales. Reportar al orquestador

### Funciones mock con `vi.fn()` (solo para callbacks y handlers)

`vi.fn()` sigue siendo válido para callbacks simples (props de componentes, event handlers) donde no hay interfaz:

```typescript
const onSubmit = vi.fn();
render(<LoginForm onSubmit={onSubmit} />);
// ...
expect(onSubmit).toHaveBeenCalledWith({ email: "test@test.com" });
```

### Mock de módulos con `vi.mock()`

```typescript
// Debe estar en el nivel superior, se eleva automáticamente
vi.mock("../db.js", () => ({
  db: {
    query: vi.fn().mockResolvedValue({ rows: [] }),
  },
}));

// Acceder al mock en los tests
import { db } from "../db.js";
vi.mocked(db.query).mockResolvedValueOnce({ rows: [{ id: 1 }] });
```

### Mock de timers

```typescript
import { vi, describe, it, expect, beforeEach, afterEach } from "vitest";

describe("debounce", () => {
  beforeEach(() => vi.useFakeTimers());
  afterEach(() => vi.useRealTimers());

  it("fires after delay", async () => {
    const fn = vi.fn();
    const debounced = debounce(fn, 300);

    debounced();
    expect(fn).not.toHaveBeenCalled();

    vi.advanceTimersByTime(300);
    expect(fn).toHaveBeenCalledOnce();
  });
});
```

## Testing de Código Asíncrono

```typescript
// INCORRECTO: testear promise sin await — el test pasa independientemente
it("saves user", () => {
  service.save(user); // fire and forget — las aserciones nunca se ejecutan
  expect(repo.save).toHaveBeenCalled();
});

// CORRECTO: siempre await las operaciones asíncronas
it("saves user", async () => {
  await service.save(user);
  expect(repo.save).toHaveBeenCalledOnce();
});

// Testear rechazos — usar el matcher rejects
it("throws on duplicate email", async () => {
  vi.mocked(repo.findByEmail).mockResolvedValueOnce(existingUser);

  await expect(service.createUser(dto)).rejects.toMatchObject({
    kind: "validation",
    fields: { email: expect.stringContaining("already taken") },
  });
});
```

## Snapshot Testing

Usar snapshots con moderación — solo para output serializado (respuestas de API, strings HTML renderizados). Evitar snapshots para estructuras de datos internas.

```typescript
it("serializes error response correctly", async () => {
  const response = formatError({ kind: "not-found", resource: "User" });
  expect(response).toMatchInlineSnapshot(`
    {
      "error": {
        "code": "NOT_FOUND",
        "message": "User not found",
      },
      "status": 404,
    }
  `);
});
// Preferir inline snapshots (toMatchInlineSnapshot) sobre archivos .snap externos
// — son visibles en el archivo de test y se revisan en PRs
```
