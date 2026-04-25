---
name: react-conventions
description: Convenciones y estándares de código frontend React/TypeScript. Usar al escribir componentes React, revisar código frontend, o cuando el usuario mencione "React patterns", "component structure", "hooks best practices", "accessibility", "state management", o al trabajar con archivos .tsx/.jsx.
---

# React Conventions

## Filosofía

- **La UI es una función del estado** — los componentes renderizan de forma predecible a partir de props y estado, nada más
- **Composición sobre configuración** — construir UIs complejas a partir de piezas simples y enfocadas
- **La accesibilidad no es opcional** — si no es navegable por teclado y legible por lector de pantalla, no está terminado
- **Server-first** — usar Server Components por defecto; usar Client Components solo cuando sea necesario

## Stack

- React 19+ con TypeScript (modo strict)
- Next.js App Router preferido para nuevos proyectos
- Vitest + React Testing Library para testing

## Reglas de Código

- Solo componentes funcionales — sin componentes de clase
- Composición sobre herencia — pasar children, no extender
- Componentes pequeños y reutilizables — una responsabilidad por componente
- Accesibilidad primero (atributos ARIA, HTML semántico, navegación por teclado)
- Estado predecible (evitar estado derivado, fuente única de verdad)
- Sin lógica de negocio en componentes UI — extraer a custom hooks
- Sin lógica de formulario (tipos, validación, estado) en componentes — extraer al hook `useXxxForm`
- Antes de crear un helper/util, hacer grep — si existe en otra feature, extraer a `shared/utils/`
- Un canal de error por acción — toast para errores de API, inline para validación de campos. Nunca ambos
- Exports nombrados preferidos sobre exports por defecto (excepto pages/layouts en Next.js)

## Reglas de TypeScript

- `strict: true` con `strictNullChecks`, `noImplicitAny`, `noUncheckedIndexedAccess`
- **`unknown` sobre `any`** — siempre. Usar type guards para narrowing
- **Discriminated unions** para estados mutuamente excluyentes (loading/error/success)
- **Verificación exhaustiva** con `never` en sentencias switch
- **Zod** para validación en runtime con tipos inferidos (`z.infer<typeof schema>`)
- Props definidas como `interface`, no `type`
- `as const` para retornos de tupla desde hooks

## Reglas de Arquitectura

1. **Estructura de carpetas basada en features** — `src/features/{name}/{api,components,hooks,stores,types}`
2. **Imports unidireccionales** — `shared → features → app`. Nunca importar entre features
3. **El hook ES el container** — los custom hooks encapsulan toda la lógica, los componentes son UI pura
4. **Cliente API centralizado** — nunca hacer fetch directamente en componentes
5. **Error boundaries a nivel de ruta** — usar `react-error-boundary`
6. **Lazy loading / code splitting** para rutas (`React.lazy` + `<Suspense>`)
7. **Server Components por defecto** (Next.js) — `'use client'` solo para estado/eventos/browser APIs

## Reglas de Gestión de Estado

| Tipo de Estado | Herramienta |
|---|---|
| Datos server/remotos | **TanStack Query** (o SWR) |
| Estado de URL | **nuqs** |
| Estado compartido del cliente | **Zustand** |
| Estado de formulario | **React Hook Form + Zod** |
| Global de baja velocidad (tema, auth) | **React Context** |

Escalado: Context → Zustand → Redux. Ver `state-management-guide.md` para patrones completos.

## Convenciones de Nomenclatura (Estándar Airbnb)

- **Archivos**: PascalCase para componentes (`ReservationCard.tsx`), camelCase para hooks/utils
- **Un componente por archivo** (se permiten múltiples componentes sin estado)
- **Props**: nombres camelCase. Omitir valores booleanos `true` (`<Foo hidden />`)
- **Prop key**: siempre usar IDs estables, nunca índices de array

## Reglas de Iconos

- **Usar `lucide-react`** — nunca inline SVG icons en componentes. Lucide proporciona componentes de iconos tree-shakable, tipados y configurables
- **Importar individualmente** — `import { User, Moon } from 'lucide-react'` (el tree-shaking solo incluye los iconos importados)
- **Tamaño vía prop** — `<User size={18} />` no `className="w-[18px] h-[18px]"`
- **Color vía `currentColor`** — los iconos heredan el color del texto del padre. Usar Tailwind `text-*` en el padre, no la prop `color`
- **Nunca importar todo** — `import * as icons from 'lucide-react'` destruye el tree-shaking
- **Tipo para props** — usar el tipo `LucideIcon` al aceptar componentes de iconos como props: `icon: LucideIcon`
- **Iconos personalizados** — si lucide no tiene lo que se necesita, crear un componente en `shared/components/icons/` siguiendo el mismo patrón `size` + `currentColor`

```tsx
// BIEN — tree-shakable, tipado, configurable
import { User, Settings, LogOut } from 'lucide-react'
<User size={18} />

// MAL — inline SVG, no configurable, duplicado
function UserIcon() {
  return <svg width="18" height="18" ...>...</svg>
}
```

## Reglas de Tailwind v4

Tailwind CSS v4 cambió cómo se referencian las CSS custom properties en las clases utilitarias.

- **Las variables CSS usan paréntesis, no corchetes:**
  - CORRECTO (v4): `bg-(--bg-surface)`, `text-(--text-primary)`, `border-(--border-default)`
  - INCORRECTO (v3): `bg-[var(--bg-surface)]`, `text-[var(--text-primary)]`
- **Preferir clases estándar de Tailwind** sobre valores arbitrarios cuando existe un equivalente:
  - `size-4` no `w-[16px] h-[16px]`
  - `translate-x-4` no `translate-x-[16px]`
  - `gap-3` no `gap-[12px]`
  - `h-5` no `h-[20px]`
- **Los valores arbitrarios son aceptables** para medidas específicas del diseño que no tienen equivalente en Tailwind: `max-w-[400px]`, `w-[560px]`, `text-[42px]`

## Patrones de Validación de Input

Los navegadores NO aplican restricciones de input para `type="tel"` o `type="number"` de forma consistente. Siempre filtrar en `onChange`.

| Tipo de input | `inputMode` | Filtrar en onChange | Validación |
|-----------|-------------|-------------------|------------|
| Teléfono | `tel` | Strip chars not in `[0-9+\-\s()]` | 10-15 dígitos |
| Email | `email` | Ninguno (validar en blur) | Regex RFC 5322 |
| Código/OTP | `numeric` | Solo dígitos, auto-avanzar foco | Coincidencia de longitud exacta |
| Moneda | `decimal` | Strip chars not in `[0-9.]` | Número positivo, máx 2 decimales |
| Contraseña | — | Ninguno | Longitud mínima + reglas de complejidad |

```tsx
// Input de teléfono con filtrado
function handlePhoneChange(value: string) {
  const filtered = value.replace(/[^0-9+\-\s()]/g, '')
  setPhone(filtered)
}
<Input type="tel" inputMode="tel" value={phone} onChange={e => handlePhoneChange(e.target.value)} />
```

## Checklist Pre-Implementación

- [ ] No existe ya un componente/hook similar
- [ ] El componente tiene una sola responsabilidad
- [ ] El estado está categorizado correctamente (server vs client vs URL vs form)
- [ ] Sin lógica de negocio en el cuerpo del componente
- [ ] Accesible: HTML semántico, etiquetas ARIA, manejadores de teclado
- [ ] Los tipos TypeScript son explícitos (sin `any`)
- [ ] La decisión Server vs Client Component es intencional

## Detección de Anti-Patrones

Ver `anti-patterns.md` para la referencia completa de detección con niveles de severidad.

**Detección pasiva:** Al revisar código React, escanear automáticamente en busca de patrones `error` y `warning`. Reportar como `[file:line] [severity] [category] anti-pattern-name`.

**Detección activa:** Cuando el usuario pide "mejorar", "refactorizar", "optimizar" — también reportar nivel `suggestion` y proponer correcciones.

Red flags que siempre deben detener el trabajo:
- `useEffect` con dependencias faltantes/incorrectas → stale-closure (error)
- `setInterval`/`setTimeout` sin cleanup → memory-leak (error)
- `addEventListener` sin remoción → event-leak (error)
- Mutación directa del DOM → dom-bypass (error)
- Imports circulares → circular-deps (error)

## Archivos de Soporte

- `patterns-guide.md` — Patrones React (custom hooks, compound components, facade hooks, state machine, control props, adapter, strategy, observer, decorator, factory)
- `state-management-guide.md` — Gestión de estado (TanStack Query, Zustand, Context, React Hook Form + Zod, Redux)
- `testing-guide.md` — Estrategia de testing (Vitest + RTL, MSW, Playwright, axe-core)
- `performance-guide.md` — Rendimiento (React Compiler, code splitting, reglas de memoización, patrones Netflix/Spotify)
- `accessibility-guide.md` — Accesibilidad (WCAG 2.2 AA, HTML semántico, ARIA, teclado, gestión de foco, testing)
- `anti-patterns.md` — Tabla de detección de anti-patrones con niveles de severidad y mapeo de correcciones
