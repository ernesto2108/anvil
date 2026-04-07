# Post-Implementation Gate

After ANY code change to `.ts` or `.tsx` files, invoke the `/lint` skill before considering the task done. The lint skill runs `tsc --noEmit`, the project linter (ESLint or Biome), and the test suite.

## Manual Verification

In addition to the automated lint gate, verify:

### Types

- [ ] `tsc --noEmit` passes with zero errors — no suppressions introduced
- [ ] Any new `@ts-expect-error` has a clear reason comment explaining why it is necessary
- [ ] No `any` introduced — check with `grep -r ": any" src/` and `grep -r "as any" src/`
- [ ] No double assertion `as unknown as X` introduced

### Architecture

- [ ] No new barrel `index.ts` re-exports added
- [ ] All new `import` statements use `.js` extension
- [ ] Type-only imports use `import type`
- [ ] No `require()` calls introduced

### Zod Boundaries

- [ ] Every new external data source (API response, env var, form, webhook) has a Zod schema
- [ ] `safeParse` used (not `parse`) at boundaries — caller handles the error branch
- [ ] Zod errors formatted and returned (not thrown) from boundary handlers

### Async

- [ ] Every new `fetch` call inside a React component or long-lived service accepts or creates an `AbortController`
- [ ] Every new `useEffect` with async logic has a cleanup function that calls `controller.abort()`
- [ ] No unhandled Promise rejections (no `.then()` without `.catch()`)

### React (when applicable)

- [ ] No new class components
- [ ] No new `dangerouslySetInnerHTML` without `DOMPurify.sanitize()`
- [ ] No new `key={index}` in dynamic lists
- [ ] New async Server Components have `<Suspense>` + `<ErrorBoundary>` wrappers

### Security

- [ ] No secrets or tokens stored in `localStorage`
- [ ] No `eval()` or `new Function()` usage
- [ ] No `Math.random()` for security-relevant values
- [ ] No SQL/template interpolation with user input

### Tests

- [ ] New behavior covered by Vitest tests
- [ ] Type-level behavior covered by `expectTypeOf` assertions where applicable
- [ ] All tests pass: `vitest run`
- [ ] No `vi.mock` calls left without cleanup in `afterEach`
