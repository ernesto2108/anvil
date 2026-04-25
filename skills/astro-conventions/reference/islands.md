# Arquitectura de Islands

## Principio Fundamental

Las páginas se renderizan como HTML estático. Los componentes interactivos ("islands") se hidratan de forma independiente con su propio bundle de JS. El resto de la página es cero JS.

## Directivas de Cliente

| Directiva | Cuándo se hidrata | Usar para |
|---|---|---|
| `client:load` | Inmediatamente al cargar la página | Crítico: estado de auth, dropdowns de nav |
| `client:idle` | El navegador está inactivo | Medio: barra de búsqueda, formularios |
| `client:visible` | Desplazado al viewport | Debajo del fold: comentarios, carruseles, mapas |
| `client:media="(max-width: 768px)"` | La media query coincide | Solo móvil: menú hamburguesa |
| `client:only="react"` | Solo cliente, sin SSR | Componentes que no pueden renderizar en servidor |

```astro
---
import Counter from '../components/islands/Counter.tsx';
import Comments from '../components/islands/Comments.tsx';
import MobileNav from '../components/islands/MobileNav.tsx';
---
<!-- Critical — hydrate immediately -->
<Counter client:load initialCount={0} />

<!-- Below fold — hydrate when visible -->
<Comments client:visible postId={post.id} />

<!-- Mobile only -->
<MobileNav client:media="(max-width: 768px)" />
```

## Server Islands

`server:defer` aísla contenido de servidor lento/dinámico para que no bloquee el renderizado de la página:

```astro
---
import UserGreeting from '../components/UserGreeting.astro';
---
<UserGreeting server:defer>
  <p slot="fallback">Loading...</p>  <!-- shown until server responds -->
</UserGreeting>
```

Usar para: contenido personalizado, llamadas lentas a la API, UI dependiente de auth.

## Islands Multi-Framework

Diferentes frameworks coexisten porque cada island está aislada:

```astro
---
import ReactSearch from '../components/islands/Search.tsx';
import VueChart from '../components/islands/Chart.vue';
import SvelteToggle from '../components/islands/Toggle.svelte';
---
<ReactSearch client:idle />
<VueChart client:visible data={chartData} />
<SvelteToggle client:load />
```

Cada una envía solo su propio runtime de framework — las islands de React no cargan Vue, etc.

## Matriz de Decisión

```
¿Es interactivo? (necesita eventos del navegador, estado, efectos)
├── NO → Componente Astro (.astro) — cero JS
└── SÍ
    ├── ¿Está sobre el fold / es crítico para la primera interacción?
    │   ├── SÍ → client:load
    │   └── NO
    │       ├── ¿Es visible en el viewport inicial?
    │       │   ├── SÍ → client:idle
    │       │   └── NO → client:visible
    │       └── ¿Es específico del dispositivo?
    │           └── SÍ → client:media
    └── ¿Puede renderizar en el servidor?
        ├── SÍ → client:load/idle/visible (elige uno)
        └── NO → client:only="framework"
```

## Anti-Patrones

| Anti-Patrón | Por qué es malo | Solución |
|---|---|---|
| `client:load` en todo | Derrota el propósito de cero JS | Audita cada uno: ¿idle? ¿visible? |
| Wrapper de framework a nivel de layout | Envía el framework completo para layout estático | Rompe en islands pequeñas |
| `{items.map(() => <Island client:load />)}` | N islands = N bundles de JS | Lista estática + 1 controlador hidratado |
| Island para contenido estático | JS innecesario para contenido que no cambia | Usa componente .astro |
| Sin `client:*` en componente de framework | El componente renderiza pero no es interactivo | Agrega directiva o convierte a .astro |
