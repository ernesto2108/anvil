# Checklist Pre-Implementación

Antes de escribir código TypeScript, verifica:

## Tipos y Modo Estricto

- [ ] `tsconfig.json` tiene `"strict": true`, `noUncheckedIndexedAccess` y `exactOptionalPropertyTypes`
- [ ] `verbatimModuleSyntax` está habilitado — usa `import type` para imports solo de tipos
- [ ] Sin uso planificado de `any` — usa `unknown` y nárrowea
- [ ] Sin uso planificado de `enum` — usa `as const` + union type
- [ ] Discriminated unions planificadas para tipos variantes (no flags booleanos)
- [ ] Branded types planificados para IDs y primitivos específicos del dominio

## Arquitectura

- [ ] El proyecto tiene `"type": "module"` en `package.json`
- [ ] Los nuevos imports usan extensiones `.js` (requerido para ESM)
- [ ] Sin archivos barrel `index.ts` planificados — solo imports directos
- [ ] Path aliases (`@app/*`) configurados tanto en tsconfig como en la config del bundler
- [ ] Schemas Zod definidos en todas las fronteras externas (API, env, input de formularios)
- [ ] Dependencias inyectadas vía constructor/parámetro — sin imports de singleton en servicios

## Manejo de Errores

- [ ] El código de librería/servicio retorna `Result<T>` — no lanza excepciones
- [ ] El código de aplicación captura `unknown` y lo nárrowea antes de usarlo
- [ ] Sin `throw "string"` planificado — solo `throw new Error(...)`

## Async

- [ ] Las operaciones de larga duración aceptan `AbortSignal`
- [ ] Las operaciones concurrentes sobre arrays grandes usan un helper con límite de concurrencia (no `Promise.all` ilimitado)
- [ ] Se usa `Promise.allSettled` cuando el éxito parcial es aceptable

## React (cuando aplique)

- [ ] Todos los componentes son function components
- [ ] Props tipadas con `interface` (no `type`)
- [ ] El data-fetching está extraído en un custom hook con cleanup de AbortController
- [ ] Los formularios usan `react-hook-form` + `zodResolver` — sin validación manual
- [ ] Los async Server Components envueltos en `<Suspense>` con fallback
- [ ] Error boundaries presentes alrededor de secciones async

## Seguridad

- [ ] Sin `innerHTML` / `dangerouslySetInnerHTML` con contenido no confiable planificado
- [ ] El output de strings orientado al usuario usa `textContent` o React JSX (auto-escapado)
- [ ] `crypto.getRandomValues()` planificado para cualquier generación de token/secreto (no `Math.random()`)
- [ ] Los tokens de autenticación planificados para cookies httpOnly (no localStorage)
