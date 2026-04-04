# Astro Component Model

## .astro File Structure

Two-part structure separated by `---` fences:

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

- Always define `interface Props` for type safety
- Destructure with defaults: `const { title, show = true } = Astro.props`
- Accept `class` prop for styling: `const { class: className } = Astro.props`
- Access all props: `Astro.props` object

## Slots (Content Projection)

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

- Default slot: `<slot />` accepts unnamed children
- Named slots: `<slot name="x" />`, inject with `slot="x"` attribute
- Fallback: place default markup inside `<slot>...</slot>`
- `<Fragment slot="name">` passes multiple elements without wrapper

## Template Expressions

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

## Component Composition

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

## Key Differences from React/JSX

| Astro | React |
|---|---|
| `class` attribute | `className` |
| `class:list={[...]}` | `classnames(...)` library |
| `<slot />` | `{children}` |
| `<slot name="x" />` | Named prop or render prop |
| `set:html={raw}` | `dangerouslySetInnerHTML` |
| No virtual DOM | Virtual DOM diffing |
| Server-only by default | Client-side by default |
| `Astro.props` | Function params / `props` |
