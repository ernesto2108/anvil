---
name: typescript-conventions
description: TypeScript 5.x conventions and coding standards. Use when writing TypeScript code, reviewing TS patterns, or user mentions "typescript conventions", "strict mode", "discriminated unions", "Zod validation", "Vitest", "type safety", "TS best practices", or working with .ts/.tsx files.
---

# TypeScript Conventions

> **IMPORTANT:** This file is a lightweight dispatcher. Do NOT load all referenced files at once. Read the routing table below, identify which files are relevant to the current task, and load ONLY those using the Read tool. Each file is ~2-5KB. Loading unnecessary files wastes context tokens.

## Stack & Philosophy

- **Strict mode always** — `"strict": true` plus `noUncheckedIndexedAccess`, `exactOptionalPropertyTypes`
- **Discriminated unions over enums** — unions are composable, serializable, and exhaustively checkable
- **Zod at boundaries** — runtime validation at API ingress, environment config, and external data
- **Vitest for testing** — native ESM, `expectTypeOf` for type-level tests, zero-config with Vite
- **ESM first** — `"type": "module"` in `package.json`, no CommonJS unless forced by tooling
- **No barrel exports** — direct imports only; barrel `index.ts` files break tree-shaking and slow builds

## Red Flags (always stop work)

- `any` type without explicit suppression comment → error
- `// @ts-ignore` without a reason → error (use `@ts-expect-error` with comment)
- `enum` keyword → warning (use `as const` union instead)
- Non-null assertion `!` on user/API data → warning
- `namespace` usage → error (ESM supersedes it)
- `innerHTML =` without sanitization → error (XSS vector)
- `var` declaration → error

## Anti-Pattern Detection

**Passive detection:** When reviewing TypeScript code, load `detection/anti-patterns.md` and scan for `error` and `warning` patterns. Report as `[file:line] [severity] [category] anti-pattern-name`.

**Active detection:** When user asks to "improve", "refactor", "optimize", or "clean" — also report `suggestion` level patterns and propose fixes referencing the relevant rule or guide.

## What to Load

Load **only** the files relevant to the current task:

### Rules (quick reference, ~2-3KB each)

| Working on... | Load |
|---|---|
| Strict mode, type operators, error handling, branded types | `rules/coding.md` |
| ESM, barrel exports, Zod boundaries, DI, tsconfig | `rules/architecture.md` |

### Guides (detailed patterns with code, ~3-5KB each)

| Working on... | Load |
|---|---|
| Discriminated unions, branded types, mapped/conditional types | `guides/patterns/types.md` |
| Vitest setup, `expectTypeOf`, mocking, async tests | `guides/testing/vitest.md` |
| AbortController, Promise.allSettled, async iterators, timeouts | `guides/async/promises.md` |
| React function components, props typing, hooks, RSC, Zod+RHF | `guides/react/components.md` |
| XSS, input sanitization, CSRF, CSP | `guides/security.md` |

### Detection & Checklists

| When... | Load |
|---|---|
| Code review | `detection/anti-patterns.md` |
| Before writing TypeScript code | `checklists/pre.md` |
| After writing TypeScript code | `checklists/post.md` |

## Post-Implementation Gate

After ANY code change to `.ts` or `.tsx` files, invoke the `/lint` skill before considering the task done.
