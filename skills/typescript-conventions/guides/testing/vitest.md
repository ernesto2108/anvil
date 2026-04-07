# Vitest Testing Guide

## Setup

Vitest is the standard test runner. It is native ESM, requires no Babel transform, and shares Vite configuration.

```typescript
// vitest.config.ts
import { defineConfig } from "vitest/config";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig({
  plugins: [tsconfigPaths()],
  test: {
    globals: false,          // prefer explicit imports — no global `describe`/`it`
    environment: "node",     // use "jsdom" for DOM tests
    coverage: {
      provider: "v8",
      reporter: ["text", "lcov"],
      include: ["src/**/*.ts"],
      exclude: ["src/**/*.test.ts", "src/**/*.d.ts"],
    },
    typecheck: {
      enabled: true,         // run tsc alongside tests
      tsconfig: "./tsconfig.json",
    },
  },
});
```

```json
// package.json scripts
{
  "test": "vitest run",
  "test:watch": "vitest",
  "test:ui": "vitest --ui",
  "test:coverage": "vitest run --coverage",
  "typecheck": "tsc --noEmit"
}
```

## Test File Structure

```typescript
// user.service.test.ts
import { describe, it, expect, beforeEach, vi } from "vitest";
import { UserService } from "./user.service.js";
import type { UserRepository } from "./user.repository.js";

describe("UserService", () => {
  let service: UserService;
  let repo: UserRepository;

  beforeEach(() => {
    // Fresh mock per test — never share state between tests
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

## `expectTypeOf` — Type-Level Tests

Use `expectTypeOf` to assert TypeScript types in tests. This runs through `tsc`, not at runtime.

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

## Mocking Strategies

### Mock functions with `vi.fn()`

```typescript
import { vi, expect } from "vitest";

const mockFetch = vi.fn<typeof fetch>();
mockFetch.mockResolvedValueOnce(new Response(JSON.stringify({ id: 1 })));

// Type-safe mock return values
const mockSend = vi.fn<(msg: string) => Promise<void>>();
mockSend.mockResolvedValue(undefined);
```

### Mock modules with `vi.mock()`

```typescript
// Must be at the top level, hoisted automatically
vi.mock("../db.js", () => ({
  db: {
    query: vi.fn().mockResolvedValue({ rows: [] }),
  },
}));

// Access the mock in tests
import { db } from "../db.js";
vi.mocked(db.query).mockResolvedValueOnce({ rows: [{ id: 1 }] });
```

### Mock timers

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

## Testing Async Code

```typescript
// WRONG: testing promise without await — test passes regardless
it("saves user", () => {
  service.save(user); // fire and forget — assertions never run
  expect(repo.save).toHaveBeenCalled();
});

// RIGHT: always await async operations
it("saves user", async () => {
  await service.save(user);
  expect(repo.save).toHaveBeenCalledOnce();
});

// Testing rejections — use rejects matcher
it("throws on duplicate email", async () => {
  vi.mocked(repo.findByEmail).mockResolvedValueOnce(existingUser);

  await expect(service.createUser(dto)).rejects.toMatchObject({
    kind: "validation",
    fields: { email: expect.stringContaining("already taken") },
  });
});
```

## Snapshot Testing

Use snapshots sparingly — only for serialized output (API responses, rendered HTML strings). Avoid snapshots for internal data structures.

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
// Prefer inline snapshots (toMatchInlineSnapshot) over external .snap files
// — they are visible in the test file and reviewed in PRs
```
