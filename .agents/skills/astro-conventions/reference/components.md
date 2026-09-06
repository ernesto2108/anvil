# Modelo de Componentes Astro

## Estructura del Archivo .astro

Estructura de dos partes separadas por delimitadores `---`:

```astro
---
// Server-only script (runs at build time, never ships to browser)
interface Props {
  title: string;
  description?: string;
  class?: string;
}
const { title, description = "Default", class: className } = Astro.props;
---
<article class={className}>
  <h2>{title}</h2>
  <p>{description}</p>
  <slot />
</article>

<style>
  /* Scoped to this component automatically */
  article { padding: var(--sp-6); }
</style>
```

## Props

- Siempre define `interface Props` para seguridad de tipos
- Desestructura con valores por defecto: `const { title, show = true } = Astro.props`
- Acepta prop `class` para estilos: `const { class: className } = Astro.props`
- Accede a todas las props: objeto `Astro.props`

## Slots (Proyección de Contenido)

```astro
<!-- Component definition -->
<div class="card">
  <header><slot name="header" /></header>
  <main><slot /></main>           <!-- default slot -->
  <footer><slot name="footer">Default footer</slot></footer>
</div>

<!-- Usage -->
<Card>
  <h2 slot="header">Title</h2>
  <p>Body content goes in default slot</p>
  <Fragment slot="footer">
    <a href="/more">Read more</a>
  </Fragment>
</Card>
```

- Slot por defecto: `<slot />` acepta hijos sin nombre
- Slots nombrados: `<slot name="x" />`, inyectar con atributo `slot="x"`
- Fallback: coloca markup por defecto dentro de `<slot>...</slot>`
- `<Fragment slot="name">` pasa múltiples elementos sin wrapper

## Expresiones en la Plantilla

```astro
---
const items = ['Go', 'Astro', 'TypeScript'];
const isActive = true;
---
<!-- Dynamic content -->
<h1>{title}</h1>

<!-- Lists -->
<ul>
  {items.map(item => <li>{item}</li>)}
</ul>

<!-- Conditional classes -->
<div class:list={["base", { active: isActive, hidden: !isActive }]}>

<!-- Conditional rendering -->
{showBanner && <Banner />}
{isLoggedIn ? <Dashboard /> : <Login />}

<!-- Dynamic HTML (use with caution) -->
<div set:html={rawHtml} />
```

## Composición de Componentes

```astro
---
// Import other Astro components
import Header from '../components/Header.astro';
import Card from '../components/Card.astro';

// Import framework components (for islands)
import SearchBar from '../components/islands/SearchBar.tsx';
---
<Header />
<main>
  <Card title="Static card" />
  <SearchBar client:idle />  <!-- Only this ships JS -->
</main>
```

## Diferencias Clave con React/JSX

| Astro | React |
|---|---|
| Atributo `class` | `className` |
| `class:list={[...]}` | Librería `classnames(...)` |
| `<slot />` | `{children}` |
| `<slot name="x" />` | Named prop o render prop |
| `set:html={raw}` | `dangerouslySetInnerHTML` |
| Sin virtual DOM | Virtual DOM diffing |
| Solo servidor por defecto | Client-side por defecto |
| `Astro.props` | Parámetros de función / `props` |
