# TypeScript Anti-Patterns — Referencia de Detección

## Detección Pasiva

Al revisar código TypeScript, escanear estos patrones y reportar con el formato:
`[file:line] [severidad] [categoría] nombre-del-anti-patrón`

Solo reportar `error` y `warning` por defecto. Reportar `suggestion` únicamente cuando el usuario pida mejorar/refactorizar/optimizar.

## Tabla de Anti-Patrones

| Patrón en el código | Anti-patrón | Severidad | Categoría | Corrección → Patrón |
|---|---|---|---|---|
| `any` sin supresión explícita | untyped-any | error | types | Usar `unknown` y estrechar, o definir el tipo — ver `rules/coding.md` |
| `// @ts-ignore` sin motivo | ts-ignore-no-reason | error | types | Reemplazar con `// @ts-expect-error: reason` — debe incluir el motivo |
| Palabra clave `enum` | enum-usage | warning | types | Unión `as const`: `const X = ["a","b"] as const; type X = typeof X[number]` — ver `guides/patterns/types.md` |
| Aserción de tipo `x as SomeType` | unsafe-assertion | warning | types | Usar type guard (`instanceof`, `in`, comprobación discriminante) — ver `guides/patterns/types.md` |
| Aserción non-null `!` sobre datos de API/usuario | non-null-assertion | warning | reliability | Estrechar con verificación `if (x == null)` en su lugar |
| Palabra clave `namespace` | namespace-usage | error | types | ESM reemplaza a los namespaces — usar módulos |
| Declaración `var` | var-usage | error | style | Usar `const` o `let` |
| Index signature `obj[key]` sin `noUncheckedIndexedAccess` | unsafe-index-access | warning | types | Habilitar `noUncheckedIndexedAccess` en tsconfig — el resultado es `T \| undefined` |
| Barrel export `index.ts` que re-exporta todo | barrel-export | warning | architecture | Importaciones directas — ver `rules/architecture.md` |
| `require()` en proyecto ESM | commonjs-require | error | architecture | Usar `import` — ver `rules/architecture.md` |
| `import X from "module"` para importación de solo tipo | missing-import-type | warning | performance | `import type X from "module"` — habilita `verbatimModuleSyntax` |
| `innerHTML =` con contenido no confiable | xss-innerHTML | error | security | Usar `textContent` o DOMPurify — ver `guides/security.md` |
| `dangerouslySetInnerHTML` sin sanitización | xss-dangerously-set | error | security | Sanitizar con DOMPurify antes de renderizar — ver `guides/security.md` |
| `document.write(...)` | xss-document-write | error | security | Nunca usar `document.write` |
| `localStorage.setItem("token", ...)` | token-in-localstorage | warning | security | Usar cookies httpOnly — ver `guides/security.md` |
| `fetch(url)` sin `signal` en contexto de larga duración | missing-abort-signal | suggestion | reliability | Agregar `AbortController` — ver `guides/async/promises.md` |
| `new XMLHttpRequest()` / XHR sincrónico | sync-xhr | error | performance | Usar `fetch()` con `AbortSignal` |
| `Promise.all(arr.map(...))` sobre arrays grandes | unbounded-concurrency | warning | performance | Usar helper con límite de concurrencia — ver `guides/async/promises.md` |
| `catch (e: any)` | catch-any | warning | types | `catch (e: unknown)` y luego estrechar — ver `rules/coding.md` |
| `throw "string message"` | throw-string | warning | errors | Lanzar instancias de `Error`: `throw new Error("...")` |
| Componente de clase extendiendo `React.Component` | class-component | warning | react | Componente función con hooks — ver `guides/react/components.md` |
| `type FooProps = {...}` para props de React | props-type-alias | suggestion | react | Usar `interface FooProps` — ver `guides/react/components.md` |
| `useEffect` con data-fetching dentro del componente | effect-data-fetch | warning | react | Extraer a custom hook con AbortController — ver `guides/react/components.md` |
| Prop `key` faltante en render de lista | missing-key-prop | error | react | Agregar `key` estable y único a cada elemento de la lista |
| `key={index}` en lista dinámica | index-as-key | warning | react | Usar identificador estable (id, slug) como key |
| `process.env.X` sin validación con Zod | unvalidated-env | warning | architecture | Validar todas las variables de entorno al inicio — ver `rules/architecture.md` |
| `import X from "./module"` sin extensión `.js` | missing-js-extension | warning | architecture | Agregar extensión `.js` para compatibilidad ESM: `"./module.js"` |
| Object spread para clonar en profundidad `{ ...obj }` | shallow-clone | warning | reliability | Usar `structuredClone(obj)` para clon profundo |
| `Math.random()` para tokens de seguridad | insecure-random | error | security | Usar `crypto.getRandomValues()` o `crypto.randomUUID()` |
| `eval(...)` o `new Function(...)` | eval-usage | error | security | Nunca usar `eval` — reescribir con alternativas seguras |
| Falta error boundary alrededor de componente React asíncrono | missing-error-boundary | warning | react | Envolver con `<ErrorBoundary>` — ver `guides/react/components.md` |
| Mutación del objeto de parámetro de función | param-mutation | warning | reliability | Clonar con spread o `structuredClone`, retornar nuevo objeto |
| `console.log` en código que no es de depuración | console-log | suggestion | observability | Usar logger estructurado con niveles de log |
| Tipo `object` o `{}` en código de dominio | vague-object-type | warning | types | Definir una `interface` concreta o usar `Record<string, unknown>` |
| Doble aserción `as unknown as TargetType` | double-assertion | warning | types | Encontrar el tipo correcto — las dobles aserciones ocultan errores de tipos |
