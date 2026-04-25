---
name: astro-conventions
description: Convenciones del framework Astro y estándares de código para sitios estáticos y orientados a contenido. Usar al escribir componentes Astro, revisar código Astro, o cuando el usuario mencione "Astro patterns", "islands architecture", "content collections", "static site", ".astro files", "astro components", "client directives", o al trabajar con archivos .astro.
---

# Convenciones de Astro

> **IMPORTANTE:** Solo dispatcher. Cargar archivos de referencia bajo demanda. Ver tabla de enrutamiento abajo.

## Filosofía

- **Content-first, JS-last** — las páginas se renderizan como HTML estático con cero JS por defecto. Agregar JS solo donde se necesita interactividad genuina
- **Islands sobre SPAs** — los componentes interactivos son islas aisladas que hidratan de forma independiente. La página es un documento con widgets embebidos, no una app
- **Estático hasta que se demuestre lo contrario** — comenzar con SSG. Solo agregar SSR para páginas que realmente necesiten datos en tiempo de solicitud
- **Enviar solo lo que se usa** — sin runtime de framework, sin CSS sin usar, sin JS especulativo. Cada byte debe ganarse su lugar

## Stack

- Astro 5+ con TypeScript (strict mode)
- Tailwind CSS v4 via plugin `@tailwindcss/vite`
- Content Collections con schemas Zod
- Vitest + Playwright para pruebas
- Deploy: Vercel, Netlify, o Cloudflare Pages

## Cuándo usar Astro (vs Next.js)

| Usar Astro | Usar Next.js |
|---|---|
| Blogs, docs, portafolios, marketing | SPAs completas, routing pesado del lado cliente |
| Orientado a contenido, interactividad mínima | Dashboards en tiempo real, auth compleja |
| SEO y rendimiento no negociables | Se necesitan React Server Components |
| Hosting estático, bajo costo a escala | Server-heavy con streaming SSR |

## Estructura del proyecto

```
src/
├── components/
│   ├── common/           # Compartidos: Button, Card, Badge
│   ├── features/         # De dominio: BlogCard, ProjectGrid
│   └── islands/          # Interactivos con directivas client:*
├── content/
│   ├── blog/             # Colección Markdown/MDX
│   ├── projects/         # Content collection
│   └── config.ts         # Schemas Zod
├── layouts/              # BaseLayout.astro, BlogLayout.astro
├── pages/                # Routing basado en archivos (único directorio reservado)
├── styles/               # global.css (@import "tailwindcss")
├── lib/                  # Utilidades, helpers
└── env.d.ts              # Extensiones de tipos globales
public/                   # Assets estáticos (favicon, robots.txt)
```

**Reglas:**
- Componentes interactivos con `client:*` → `src/components/islands/` (hace explícitos los límites de JS)
- Contenido con schemas Zod → `src/content/` (nunca markdown crudo sin validar)
- Path aliases: `@components/*`, `@layouts/*`, `@lib/*`
- `public/` es SOLO para assets estáticos — nunca CSS/JS (esos van en `src/`)

## Señales de alerta (detener el trabajo siempre)

- `client:load` en contenido debajo del fold → usar `client:visible`
- Layout completo envuelto en React/Vue → extraer islas granulares
- Array grande mapeado para generar islas → renderizar estáticamente, hidratar un controlador
- `<img>` crudo para assets locales → usar `<Image />` de `astro:assets`
- Contenido sin schema Zod → definir schema en `config.ts`
- Fetching a la propia API vía HTTP en build → importar la función directamente
- Usar Astro para una SPA completa → herramienta incorrecta, usar Next.js

## Lista de verificación pre-implementación

- [ ] Astro es la elección correcta (orientado a contenido, no SPA)
- [ ] TypeScript strict mode (`extends: "astro/tsconfigs/strict"`)
- [ ] Content Collections con schemas Zod definidos
- [ ] Tailwind v4 via `@tailwindcss/vite`
- [ ] Layout base con `<ViewTransitions />`
- [ ] Componentes interactivos aislados en `islands/`
- [ ] Todas las imágenes usan `<Image />` de `astro:assets`
- [ ] Path aliases en tsconfig

## Validación de nuevos patrones

Al introducir un patrón que aún no existe en el proyecto (sistema de iconos, organización de scripts, arquitectura CSS, gestión de estado, enfoque i18n):

1. **Verificar si el proyecto ya tiene una convención** — leer el código existente primero
2. **Si no existe convención**, buscar en la documentación oficial del framework y las mejores prácticas de la comunidad antes de implementar. No asumir — verificar
3. **Si existe una convención**, seguirla. No introducir un patrón competidor

Esto NO significa buscar en la web para cada cambio. Solo para **nuevos patrones arquitectónicos** que se usarán en todo el proyecto.

## Detección de anti-patrones

| Anti-Patrón | Severidad | Corrección |
|---|---|---|
| `client:load` en widget no crítico | warning | Usar `client:idle` o `client:visible` |
| Layout envuelto en componente de framework | error | Extraer islas granulares |
| Array mapeado para generar islas | error | Lista estática + un controlador hidratado |
| `<img>` crudo para asset local | warning | `<Image />` de `astro:assets` |
| Contenido sin schema Zod | error | Definir en `content/config.ts` |
| HTTP fetch a propia API en build | error | Importar función directamente |
| `public/` usado para CSS/JS | error | Mover a `src/` |
| Sin interfaz `Props` en archivo .astro | warning | Agregar `interface Props` |
| Usar Astro para SPA completa | error | Herramienta incorrecta — usar Next.js |
| Componente interactivo no en `islands/` | warning | Mover a `src/components/islands/` |

## Archivos de referencia

Cargar SOLO cuando sea necesario:

| Trabajando en... | Cargar |
|---|---|
| Modelo de componente .astro, Props, slots, expresiones de template | `reference/components.md` |
| Directivas client:*, estrategia de hidratación, server islands | `reference/islands.md` |
| Content Collections, schemas Zod, loaders, MDX | `reference/content-collections.md` |
| CSS con scope, Tailwind v4, define:vars, preprocesadores | `reference/styling.md` |
| TypeScript, imágenes, View Transitions, SSR, middleware, testing | `reference/patterns.md` |
| Arquitectura Pods, mappers, fetching defensivo, consumo de API | `reference/api-patterns.md` |

## Gate post-implementación

Después de CUALQUIER cambio a archivos `.astro`:
1. Ejecutar `astro check` para errores de tipos
2. Ejecutar `astro build` para verificar que la generación tenga éxito
3. Invocar el skill `/lint`
