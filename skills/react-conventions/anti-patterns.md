# Referencia de Detección de Anti-Patrones React

## Modos de Detección

**Detección pasiva:** Al revisar código React, buscar automáticamente patrones `error` y `warning`. Reportar como `[file:line] [severity] [category] anti-pattern-name`.

**Detección activa:** Cuando el usuario pide "mejorar", "refactorizar", "optimizar" — también reportar nivel `suggestion` y proponer correcciones referenciando el patrón/guía relevante.

---

## Error (Debe Corregirse)

| Patrón de código | Anti-Patrón | Categoría | Corrección |
|---|---|---|---|
| `useEffect` con deps faltantes/incorrectas | stale-closure | bugs | Agregar deps correctas o refactorizar para evitar el effect |
| `setInterval`/`setTimeout` sin return de limpieza | memory-leak | memory | Retornar limpieza en `useEffect` |
| `addEventListener` sin `removeEventListener` | event-leak | memory | Limpieza en el return de `useEffect` |
| Mutación directa del DOM (`document.getElementById`) | dom-bypass | architecture | Usar refs (`useRef`) o estado |
| Imports circulares entre módulos | circular-deps | modules | Extraer interfaz compartida a un tercer módulo |
| Íconos SVG inline en componentes (`<svg>...</svg>`) | inline-svg-icons | maintenance | Usar componentes de ícono `lucide-react` con prop `size` |
| `showToast(error)` + `{isError && <p>error</p>}` para el mismo error | dual-error-feedback | ux | Elegir un canal: toast para errores de API, inline para validación de campo |
| Helper/constante duplicado entre features | duplicated-util | maintenance | Extraer a `shared/utils/` antes de crear, buscar primero con grep |
| Tipos de formulario + validación + lógica de estado dentro del componente | form-logic-in-view | architecture | Extraer a hook `useXxxForm`; el componente es UI pura |
| `import * as icons from 'lucide-react'` | icon-barrel-import | performance | Importar íconos individuales: `import { User } from 'lucide-react'` |
| `[var(--X)]` en className de Tailwind | tailwind-v3-var-syntax | style | Usar sintaxis de paréntesis `(--X)` para Tailwind v4 |
| `w-[16px]` cuando existe `w-4` | unnecessary-arbitrary-value | style | Usar clase estándar de Tailwind en lugar de valor arbitrario |
| `type="tel"` sin filtro en onChange | unfiltered-phone-input | validation | Filtrar caracteres no telefónicos en el manejador onChange |

---

## Warning (Debería Corregirse)

| Patrón de código | Anti-Patrón | Categoría | Corrección |
|---|---|---|---|
| Props pasando por más de 3 niveles de componentes | prop-drilling | state | Context, hook personalizado, Zustand o composición |
| Múltiples booleanos `isLoading`/`isError`/`isSuccess` | boolean-hell | state | Unión discriminada o máquina de estados (ver `patterns-guide.md`) |
| `useEffect` para estado derivado | unnecessary-effect | performance | Calcular durante el render o `useMemo` |
| Cadenas en cascada de `useEffect` (A→B→C) | effect-cascade | architecture | Consolidar en un solo effect o manejador de evento |
| Lógica de negocio dentro del cuerpo del componente | logic-in-view | architecture | Extraer a hook personalizado (patrón facade) |
| `useSelector`/`useDispatch` directamente en componentes | store-coupling | architecture | Hook facade (`useAuth()`, `useCart()`) |
| Context con más de 3 valores que cambian frecuentemente | context-misuse | state | Separar contexts o usar Zustand |
| `useState` para valores derivados de otro estado | state-overuse | state | Calcular durante el render |
| Context default vacío (`createContext({})`) | context-default | bugs | Usar `createContext<T | null>(null)` + verificación de null |

---

## Suggestion (Considerar Corregir)

| Patrón de código | Anti-Patrón | Categoría | Corrección |
|---|---|---|---|
| Función inline en JSX que causa re-renders | inline-callback | performance | `useCallback` o extraer manejador |
| Literal de objeto/array en props de JSX | object-recreation | performance | Extraer a constante o `useMemo` |
| `export * from './...'` en archivos index | barrel-export | modules | Imports directos para mejor tree-shaking |
| Tipo `any` en TypeScript | untyped-code | types | Tipos concretos, genéricos o `unknown` |
| Componente de más de 200 líneas | large-component | readability | Dividir en componentes más pequeños y compuestos |
| Ternarios profundamente anidados en JSX | ternary-hell | readability | Returns tempranos, guard clauses o componentes extraídos |
| Exports por defecto en utilities/hooks | default-export | modules | Exports nombrados para refactorización y tree-shaking |
| `useMemo`/`useCallback` manual en todos lados | over-memoization | performance | Confiar en React Compiler; perfilar primero |
| `fetch` directamente en el componente | direct-fetch | architecture | TanStack Query o cliente API centralizado |
| `console.log` dejado en el código | debug-artifact | quality | Eliminar o usar logging estructurado |
| Estilos inline en componentes reutilizables | inline-styles | style | CSS modules, Tailwind o styled-components |

---

## Patrones de Detección (tipo Regex)

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

## Mapa de Correcciones

| Anti-Patrón | Patrón Recomendado | Guía |
|---|---|---|
| dual-error-feedback | Toast solo para errores de API, inline para validación de campo | — |
| duplicated-util | Extracción a `shared/utils/` | — |
| form-logic-in-view | Patrón hook `useXxxForm` | `patterns-guide.md` |
| boolean-hell | State Machine | `patterns-guide.md` |
| prop-drilling | Facade Hook / Context | `patterns-guide.md`, `state-management-guide.md` |
| store-coupling | Facade Hook | `patterns-guide.md` |
| logic-in-view | Custom Hook (container) | `patterns-guide.md` |
| effect-cascade | Manejador de evento o effect único | `patterns-guide.md` |
| context-misuse | Zustand | `state-management-guide.md` |
| state-overuse | Computado durante el render | `state-management-guide.md` |
| direct-fetch | TanStack Query | `state-management-guide.md` |
| over-memoization | React Compiler | `performance-guide.md` |
| inline-styles | CSS modules/Tailwind | preferencia del proyecto |
