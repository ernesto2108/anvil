# Gate Post-Implementación

Después de CUALQUIER cambio de código en archivos `.ts` o `.tsx`, invoca el skill `/lint` antes de considerar la tarea terminada. El skill de lint ejecuta `tsc --noEmit`, el linter del proyecto (ESLint o Biome) y la suite de tests.

## Verificación Manual

Además del gate de lint automatizado, verifica:

### Tipos

- [ ] `tsc --noEmit` pasa con cero errores — sin supresiones introducidas
- [ ] Cualquier nuevo `@ts-expect-error` tiene un comentario de razón clara que explica por qué es necesario
- [ ] No se introdujo `any` — verifica con `grep -r ": any" src/` y `grep -r "as any" src/`
- [ ] No se introdujo doble aserción `as unknown as X`

### Arquitectura

- [ ] No se agregaron nuevos barrel `index.ts` de re-exportación
- [ ] Todos los nuevos `import` usan extensión `.js`
- [ ] Los imports solo de tipos usan `import type`
- [ ] No se introdujeron llamadas `require()`

### Fronteras Zod

- [ ] Cada nueva fuente de datos externa (respuesta de API, variable de entorno, formulario, webhook) tiene un schema Zod
- [ ] Se usa `safeParse` (no `parse`) en las fronteras — quien llama maneja el branch de error
- [ ] Los errores de Zod se formatean y retornan (no se lanzan) desde los handlers de frontera

### Async

- [ ] Cada nueva llamada `fetch` dentro de un componente React o servicio de larga vida acepta o crea un `AbortController`
- [ ] Cada nuevo `useEffect` con lógica async tiene una función de cleanup que llama a `controller.abort()`
- [ ] Sin rechazos de Promise sin manejar (no hay `.then()` sin `.catch()`)

### React (cuando aplique)

- [ ] Sin nuevos class components
- [ ] Sin nuevo `dangerouslySetInnerHTML` sin `DOMPurify.sanitize()`
- [ ] Sin nuevo `key={index}` en listas dinámicas
- [ ] Los nuevos async Server Components tienen wrappers `<Suspense>` + `<ErrorBoundary>`

### Seguridad

- [ ] Sin secretos ni tokens almacenados en `localStorage`
- [ ] Sin uso de `eval()` o `new Function()`
- [ ] Sin `Math.random()` para valores relevantes de seguridad
- [ ] Sin interpolación de SQL/plantillas con input del usuario

### Tests

- [ ] El nuevo comportamiento está cubierto por tests de Vitest
- [ ] El comportamiento a nivel de tipos está cubierto por aserciones `expectTypeOf` donde aplique
- [ ] Todos los tests pasan: `vitest run`
- [ ] Sin llamadas `vi.mock` dejadas sin cleanup en `afterEach`
