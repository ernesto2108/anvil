# TypeScript Anti-Patterns — Detection Reference

## Passive Detection

When reviewing TypeScript code, scan for these patterns and report using the format:
`[file:line] [severity] [category] anti-pattern-name`

Only report `error` and `warning` by default. Report `suggestion` only when user asks to improve/refactor/optimize.

## Anti-Pattern Table

| Code Pattern | Anti-Pattern | Severity | Category | Fix → Pattern |
|---|---|---|---|---|
| `any` without explicit suppression | untyped-any | error | types | Use `unknown` and narrow, or define the type — see `rules/coding.md` |
| `// @ts-ignore` without reason | ts-ignore-no-reason | error | types | Replace with `// @ts-expect-error: reason` — must include why |
| `enum` keyword | enum-usage | warning | types | `as const` union: `const X = ["a","b"] as const; type X = typeof X[number]` — see `guides/patterns/types.md` |
| `x as SomeType` type assertion | unsafe-assertion | warning | types | Use type guard (`instanceof`, `in`, discriminant check) — see `guides/patterns/types.md` |
| Non-null assertion `!` on API/user data | non-null-assertion | warning | reliability | Narrow with `if (x == null)` check instead |
| `namespace` keyword | namespace-usage | error | types | ESM supersedes namespaces — use modules |
| `var` declaration | var-usage | error | style | Use `const` or `let` |
| Index signature `obj[key]` without `noUncheckedIndexedAccess` | unsafe-index-access | warning | types | Enable `noUncheckedIndexedAccess` in tsconfig — result is `T \| undefined` |
| Barrel export `index.ts` re-exporting everything | barrel-export | warning | architecture | Direct imports — see `rules/architecture.md` |
| `require()` in ESM project | commonjs-require | error | architecture | Use `import` — see `rules/architecture.md` |
| `import X from "module"` for type-only import | missing-import-type | warning | performance | `import type X from "module"` — enables `verbatimModuleSyntax` |
| `innerHTML =` with untrusted content | xss-innerHTML | error | security | Use `textContent` or DOMPurify — see `guides/security.md` |
| `dangerouslySetInnerHTML` without sanitization | xss-dangerously-set | error | security | Sanitize with DOMPurify before rendering — see `guides/security.md` |
| `document.write(...)` | xss-document-write | error | security | Never use `document.write` |
| `localStorage.setItem("token", ...)` | token-in-localstorage | warning | security | Use httpOnly cookies — see `guides/security.md` |
| `fetch(url)` without `signal` in long-lived context | missing-abort-signal | suggestion | reliability | Add `AbortController` — see `guides/async/promises.md` |
| `new XMLHttpRequest()` / synchronous XHR | sync-xhr | error | performance | Use `fetch()` with `AbortSignal` |
| `Promise.all(arr.map(...))` on large arrays | unbounded-concurrency | warning | performance | Use concurrency-limited helper — see `guides/async/promises.md` |
| `catch (e: any)` | catch-any | warning | types | `catch (e: unknown)` then narrow — see `rules/coding.md` |
| `throw "string message"` | throw-string | warning | errors | Throw `Error` instances: `throw new Error("...")` |
| Class component extending `React.Component` | class-component | warning | react | Function component with hooks — see `guides/react/components.md` |
| `type FooProps = {...}` for React props | props-type-alias | suggestion | react | Use `interface FooProps` — see `guides/react/components.md` |
| `useEffect` with data-fetching inside component | effect-data-fetch | warning | react | Extract to custom hook with AbortController — see `guides/react/components.md` |
| Missing `key` prop in list render | missing-key-prop | error | react | Add stable, unique `key` to each list element |
| `key={index}` in dynamic list | index-as-key | warning | react | Use stable identifier (id, slug) as key |
| `process.env.X` without Zod validation | unvalidated-env | warning | architecture | Validate all env vars at startup — see `rules/architecture.md` |
| `import X from "./module"` without `.js` extension | missing-js-extension | warning | architecture | Add `.js` extension for ESM compatibility: `"./module.js"` |
| Object spread to deep clone `{ ...obj }` | shallow-clone | warning | reliability | Use `structuredClone(obj)` for deep clone |
| `Math.random()` for security tokens | insecure-random | error | security | Use `crypto.getRandomValues()` or `crypto.randomUUID()` |
| `eval(...)` or `new Function(...)` | eval-usage | error | security | Never use `eval` — rewrite with safe alternatives |
| Missing error boundary around async React component | missing-error-boundary | warning | react | Wrap with `<ErrorBoundary>` — see `guides/react/components.md` |
| Mutation of function parameter object | param-mutation | warning | reliability | Clone with spread or `structuredClone`, return new object |
| `console.log` in non-debug code | console-log | suggestion | observability | Use structured logger with log levels |
| Type `object` or `{}` in domain code | vague-object-type | warning | types | Define a concrete `interface` or use `Record<string, unknown>` |
| `as unknown as TargetType` double assertion | double-assertion | warning | types | Find the correct type — double assertions hide type errors |
