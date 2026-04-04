# React Anti-Pattern Detection Reference

## Detection Modes

**Passive detection:** When reviewing React code, automatically scan for `error` and `warning` patterns. Report as `[file:line] [severity] [category] anti-pattern-name`.

**Active detection:** When user asks to "improve", "refactor", "optimize" — also report `suggestion` level and propose fixes referencing the relevant pattern/guide.

---

## Error (Must Fix)

| Code Pattern | Anti-Pattern | Category | Fix |
|---|---|---|---|
| `useEffect` with missing/wrong deps | stale-closure | bugs | Add correct deps or refactor to avoid effect |
| `setInterval`/`setTimeout` without cleanup return | memory-leak | memory | Return cleanup in `useEffect` |
| `addEventListener` without `removeEventListener` | event-leak | memory | Cleanup in `useEffect` return |
| Direct DOM mutation (`document.getElementById`) | dom-bypass | architecture | Use refs (`useRef`) or state |
| Circular imports between modules | circular-deps | modules | Extract shared interface to third module |
| Inline SVG icons in components (`<svg>...</svg>`) | inline-svg-icons | maintenance | Use `lucide-react` icon components with `size` prop |
| `showToast(error)` + `{isError && <p>error</p>}` for same error | dual-error-feedback | ux | Choose one channel: toast for API errors, inline for field validation |
| Helper/constant duplicated across features | duplicated-util | maintenance | Extract to `shared/utils/` before creating, grep first |
| Form types + validation + state logic inside component | form-logic-in-view | architecture | Extract to `useXxxForm` hook; component is pure UI |
| `import * as icons from 'lucide-react'` | icon-barrel-import | performance | Import individual icons: `import { User } from 'lucide-react'` |
| `[var(--X)]` in Tailwind className | tailwind-v3-var-syntax | style | Use `(--X)` parenthesis syntax for Tailwind v4 |
| `w-[16px]` when `w-4` exists | unnecessary-arbitrary-value | style | Use standard Tailwind class instead of arbitrary value |
| `type="tel"` without onChange filter | unfiltered-phone-input | validation | Filter non-phone characters in onChange handler |

---

## Warning (Should Fix)

| Code Pattern | Anti-Pattern | Category | Fix |
|---|---|---|---|
| Props passing through >3 component levels | prop-drilling | state | Context, custom hook, Zustand, or composition |
| Multiple `isLoading`/`isError`/`isSuccess` booleans | boolean-hell | state | Discriminated union or state machine (see `patterns-guide.md`) |
| `useEffect` for derived state | unnecessary-effect | performance | Compute during render or `useMemo` |
| Cascading `useEffect` chains (A→B→C) | effect-cascade | architecture | Consolidate into single effect or event handler |
| Business logic inside component body | logic-in-view | architecture | Extract to custom hook (facade pattern) |
| `useSelector`/`useDispatch` directly in components | store-coupling | architecture | Facade hook (`useAuth()`, `useCart()`) |
| Context with >3 values that change frequently | context-misuse | state | Split contexts or use Zustand |
| `useState` for values derived from other state | state-overuse | state | Compute during render |
| Empty context default (`createContext({})`) | context-default | bugs | Use `createContext<T | null>(null)` + null check |

---

## Suggestion (Consider Fixing)

| Code Pattern | Anti-Pattern | Category | Fix |
|---|---|---|---|
| Inline function in JSX causing re-renders | inline-callback | performance | `useCallback` or extract handler |
| Object/array literal in JSX props | object-recreation | performance | Extract to constant or `useMemo` |
| `export * from './...'` in index files | barrel-export | modules | Direct imports for better tree-shaking |
| `any` type in TypeScript | untyped-code | types | Concrete types, generics, or `unknown` |
| Component >200 lines | large-component | readability | Split into smaller composed components |
| Deeply nested ternaries in JSX | ternary-hell | readability | Early returns, guard clauses, or extracted components |
| Default exports in utilities/hooks | default-export | modules | Named exports for refactoring and tree-shaking |
| Manual `useMemo`/`useCallback` everywhere | over-memoization | performance | Trust React Compiler; profile first |
| `fetch` directly in component | direct-fetch | architecture | TanStack Query or centralized API client |
| `console.log` left in code | debug-artifact | quality | Remove or use structured logging |
| Inline styles on reusable components | inline-styles | style | CSS modules, Tailwind, or styled-components |

---

## Detection Patterns (Regex-like)

```
# stale-closure
useEffect\(\s*\(\)\s*=>\s*\{[^}]*\}\s*\) → missing deps array

# memory-leak
useEffect.*setInterval|setTimeout → no return cleanup

# boolean-hell
const \[is\w+, setIs\w+\] = useState.*\n.*const \[is\w+, setIs\w+ → multiple boolean states

# prop-drilling
\.props\.\w+\.props\.\w+ → deep prop passing

# store-coupling
import.*useSelector|useDispatch.*from.*react-redux → direct store access in component

# barrel-export
export \* from → barrel re-export

# dual-error-feedback
showToast.*error.*\n.*isError.*&& → toast + inline for same error

# form-logic-in-view (in component files)
^interface Form\w+|^function validate\w+|^type Form\w+ → form types/logic in component file

# duplicated-util
function (formatDate|truncateId|timeAgo|isValidEmail) → check if already exists in shared/utils
```

---

## Fix Mapping

| Anti-Pattern | Recommended Pattern | Guide |
|---|---|---|
| dual-error-feedback | Toast only for API errors, inline for field validation | — |
| duplicated-util | `shared/utils/` extraction | — |
| form-logic-in-view | `useXxxForm` hook pattern | `patterns-guide.md` |
| boolean-hell | State Machine | `patterns-guide.md` |
| prop-drilling | Facade Hook / Context | `patterns-guide.md`, `state-management-guide.md` |
| store-coupling | Facade Hook | `patterns-guide.md` |
| logic-in-view | Custom Hook (container) | `patterns-guide.md` |
| effect-cascade | Event handler or single effect | `patterns-guide.md` |
| context-misuse | Zustand | `state-management-guide.md` |
| state-overuse | Computed during render | `state-management-guide.md` |
| direct-fetch | TanStack Query | `state-management-guide.md` |
| over-memoization | React Compiler | `performance-guide.md` |
| inline-styles | CSS modules/Tailwind | project preference |
