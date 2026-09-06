---
name: typescript-conventions
description: Convenciones y estándares de código para TypeScript 5.x. Úsalo al escribir código TypeScript, revisar patrones TS, o cuando el usuario mencione "typescript conventions", "strict mode", "discriminated unions", "Zod validation", "Vitest", "type safety", "TS best practices", o al trabajar con archivos .ts/.tsx.
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->


# TypeScript Conventions

> **IMPORTANTE:** Este archivo es un dispatcher ligero. NO cargar todos los archivos referenciados a la vez. Leer la tabla de enrutamiento abajo, identificar qué archivos son relevantes para la tarea actual, y cargar SOLO esos usando la herramienta Read. Cada archivo tiene ~2-5KB. Cargar archivos innecesarios desperdicia tokens de contexto.

## Stack y Filosofía

- **Strict mode siempre** — `"strict": true` más `noUncheckedIndexedAccess`, `exactOptionalPropertyTypes`
- **Discriminated unions sobre enums** — los unions son componibles, serializables y verificables exhaustivamente
- **Zod en las fronteras** — validación en tiempo de ejecución en el ingreso de API, configuración de entorno, y datos externos
- **Vitest para pruebas** — ESM nativo, `expectTypeOf` para pruebas a nivel de tipos, configuración cero con Vite
- **ESM primero** — `"type": "module"` en `package.json`, sin CommonJS a menos que lo fuerce el tooling
- **Sin barrel exports** — solo imports directos; los archivos `index.ts` barrel rompen el tree-shaking y ralentizan los builds

## Señales de Alerta (siempre detener el trabajo)

- Tipo `any` sin comentario de supresión explícito → error
- `// @ts-ignore` sin razón → error (usar `@ts-expect-error` con comentario)
- Palabra clave `enum` → warning (usar unión `as const` en su lugar)
- Aserción non-null `!` en datos de usuario/API → warning
- Uso de `namespace` → error (ESM lo reemplaza)
- `innerHTML =` sin sanitización → error (vector XSS)
- Declaración `var` → error

## Detección de Anti-Patrones

**Detección pasiva:** Al revisar código TypeScript, cargar `detection/anti-patterns.md` y escanear para patrones `error` y `warning`. Reportar como `[file:line] [severity] [category] anti-pattern-name`.

**Detección activa:** Cuando el usuario pide "improve", "refactor", "optimize", o "clean" — también reportar patrones de nivel `suggestion` y proponer correcciones referenciando la regla o guía relevante.

## Qué Cargar

Cargar **solo** los archivos relevantes para la tarea actual:

### Reglas (referencia rápida, ~2-3KB cada uno)

| Trabajando en... | Cargar |
|---|---|
| Strict mode, operadores de tipos, manejo de errores, branded types | `rules/coding.md` |
| ESM, barrel exports, fronteras Zod, DI, tsconfig | `rules/architecture.md` |

### Guías (patrones detallados con código, ~3-5KB cada uno)

| Trabajando en... | Cargar |
|---|---|
| Discriminated unions, branded types, mapped/conditional types | `guides/patterns/types.md` |
| Configuración de Vitest, `expectTypeOf`, mocking, pruebas async | `guides/testing/vitest.md` |
| AbortController, Promise.allSettled, async iterators, timeouts | `guides/async/promises.md` |
| Componentes funcionales React, tipado de props, hooks, RSC, Zod+RHF | `guides/react/components.md` |
| XSS, sanitización de inputs, CSRF, CSP | `guides/security.md` |
| Configuración flat ESLint v8, reglas críticas, typescript-eslint | `guides/eslint.md` |
| Node.js ESM, verbatimModuleSyntax, subpath exports | `guides/node-esm.md` |

### Detección y Checklists

| Cuándo... | Cargar |
|---|---|
| Revisión de código | `detection/anti-patterns.md` |
| Antes de escribir código TypeScript | `checklists/pre.md` |
| Después de escribir código TypeScript | `checklists/post.md` |

## Puerta Post-Implementación

Después de CUALQUIER cambio de código en archivos `.ts` o `.tsx`, invocar la skill `/lint` antes de considerar la tarea como completada.
