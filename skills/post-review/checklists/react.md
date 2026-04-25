# Checklist de Revisión React (JS/TS)

## Correctness

- [ ] useEffect tiene dependency array correcta (no faltan deps, no sobran)
- [ ] useEffect con subscriptions/timers tiene cleanup function
- [ ] Estado derivado no se duplica en useState (calcular en render)
- [ ] Keys en listas son estables y unicas (no usar index como key en listas dinamicas)
- [ ] Conditional hooks no existen (hooks siempre en el mismo orden)
- [ ] Async operations en useEffect manejan unmount (abort controller o flag)
- [ ] Forms controlados tienen value + onChange en sync
- [ ] No hay mutacion directa de estado (spread o structuredClone)

## Security

- [ ] No hay `dangerouslySetInnerHTML` sin sanitizacion
- [ ] No hay secrets/API keys en codigo frontend (usar env vars server-side)
- [ ] URLs de usuario se validan antes de renderizar en `href` o `src`
- [ ] No hay `eval()` ni `new Function()` con input de usuario
- [ ] Tokens se guardan en httpOnly cookies, no en localStorage
- [ ] No se exponen endpoints internos en el bundle del cliente

## Conventions

- [ ] Componentes en PascalCase, hooks en camelCase con prefijo `use`
- [ ] Un componente por archivo (excepto subcomponentes internos pequenos)
- [ ] Props tipadas con interface/type (no `any`)
- [ ] Custom hooks extraen logica reutilizable, no solo por "limpieza"
- [ ] Archivos organizados por feature, no por tipo
- [ ] No hay console.log residuales (excepto en dev/debug intencional)
- [ ] Imports ordenados: React, third-party, internos, estilos

## Tests

- [ ] Componentes nuevos tienen test de render basico
- [ ] Tests verifican comportamiento, no implementacion (no testear state interno)
- [ ] Interacciones de usuario testeadas con user-event, no fireEvent
- [ ] Tests de error states y loading states
- [ ] No hay tests con `waitFor` innecesarios o timeouts arbitrarios

## Performance

- [ ] No hay renders innecesarios (verificar con React DevTools si hay duda)
- [ ] useMemo/useCallback solo donde hay costo real de re-calculo
- [ ] No hay objetos/arrays creados inline en props (`style={{}}`, `options={[]}`)
- [ ] Lazy loading para rutas/componentes pesados
- [ ] Imagenes optimizadas (next/image, srcset, o compression)
- [ ] No hay fetching redundante (deduplicar con React Query, SWR, o similar)
- [ ] Bundle no incluye deps pesadas sin justificacion (moment.js, lodash completo)
