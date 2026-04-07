# Pre-Implementation Checklist

Before writing TypeScript code, verify:

## Types & Strict Mode

- [ ] `tsconfig.json` has `"strict": true`, `noUncheckedIndexedAccess`, and `exactOptionalPropertyTypes`
- [ ] `verbatimModuleSyntax` is enabled — use `import type` for type-only imports
- [ ] No planned use of `any` — use `unknown` and narrow it
- [ ] No planned use of `enum` — use `as const` + union type
- [ ] Discriminated unions planned for variant types (not boolean flags)
- [ ] Branded types planned for IDs and domain-specific primitives

## Architecture

- [ ] Project has `"type": "module"` in `package.json`
- [ ] New imports use `.js` extensions (required for ESM)
- [ ] No barrel `index.ts` files planned — direct imports only
- [ ] Path aliases (`@app/*`) configured in both tsconfig and bundler config
- [ ] Zod schemas defined at all external boundaries (API, env, form input)
- [ ] Dependencies injected via constructor/parameter — no singleton imports in services

## Error Handling

- [ ] Library/service code returns `Result<T>` — does not throw
- [ ] Application code catches `unknown` and narrows before using
- [ ] No `throw "string"` planned — only `throw new Error(...)`

## Async

- [ ] Long-running operations accept `AbortSignal`
- [ ] Large array concurrent operations use a concurrency-limited helper (not unbounded `Promise.all`)
- [ ] `Promise.allSettled` used when partial success is acceptable

## React (when applicable)

- [ ] All components are function components
- [ ] Props typed with `interface` (not `type`)
- [ ] Data-fetching extracted to custom hook with AbortController cleanup
- [ ] Forms use `react-hook-form` + `zodResolver` — no manual validation
- [ ] Async Server Components wrapped in `<Suspense>` with fallback
- [ ] Error boundaries present around async sections

## Security

- [ ] No `innerHTML` / `dangerouslySetInnerHTML` with untrusted content planned
- [ ] User-facing string output uses `textContent` or React JSX (auto-escaped)
- [ ] `crypto.getRandomValues()` planned for any token/secret generation (not `Math.random()`)
- [ ] Auth tokens planned for httpOnly cookies (not localStorage)
