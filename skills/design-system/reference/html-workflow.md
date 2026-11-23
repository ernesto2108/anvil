# HTML Design Tool — Flujo de Trabajo del Design System

Cómo construir diseño como **maqueta HTML navegable** (herramienta por defecto). No requiere servidor, build ni licencia: se abre con doble clic (`file://`). El diseño se separa en capas — tokens, componentes y pantallas — en archivos distintos; las pantallas se registran como datos, así que son fáciles de revisar, versionar y capturar.

## Contenido

- Ubicación y starter
- Estructura de archivos
- Mapeo desde el Design Spec
- Reglas por archivo
- Responsive (obligatorio)
- Reglas de tokens
- Validación y capturas
- Auditoría de CTAs
- Modo claro/oscuro y tema
- Limitaciones

## Ubicación y starter

- **Las rutas las decide el humano** vía el **Gate de Ubicación de Artefactos** de `design-project` (único lugar donde se define). No crear ni escribir nada antes de que ese gate se resuelva. Campos resultantes: `design_dir`, `components_path`, `tokens_path`.
- Default sugerido (compartir el sistema entre features): `design/<feature>/` para `index.html` + `screens.js`, y `design/_system/` para `tokens.css`, `components.css` y `components.js`. Opcional: `README.md` corto en `design_dir` (cómo verla y origen de valores).
- Punto de partida: la carpeta `reference/html-starter/` de esta skill (`index.html`, `tokens.css`, `components.css`, `components.js`, `screens.js`). Genérico: sin dominio ni framework.
  - `index.html` y `screens.js` → `design_dir`.
  - `tokens.css`, `components.css`, `components.js` → `tokens_path` / `components_path`. **Si ya existen ahí, se reutilizan y se extienden; nunca se sobrescriben** (otra feature ya depende de ellos).
  - Ajustar solo las rutas relativas de `index.html` (`<link>` y `<script src>`) si `design_dir` y el sistema no son hermanos como en el default.
- Si el proyecto ya tiene sistema de diseño (Gate de Discovery de Sistema Existente de `design-project`, `existing_design_system: yes`): `tokens.css` pasa a ser un **mapa de alias** a los tokens existentes (o se enlaza el CSS existente desde `index.html`), y `components.js`/`components.css` **referencian/mapean** la librería existente en vez de duplicarla. No inventar paralelos.

## Estructura de archivos

| Archivo | Capa | Qué contiene | Se edita al... |
|---|---|---|---|
| `tokens.css` | Tokens | Única fuente de valores visuales en `:root` (color, tipografía, espaciado, radio, sombra, touch-min); tema (`body[data-theme]`) y modo oscuro (`body[data-mode="dark"]`). Solo custom properties | definir/mapear tokens |
| `components.css` | Componentes | Base + clases de componentes mobile-first que solo consumen `var(--token)` + media queries `min-width` | agregar/variar componentes |
| `components.js` | Componentes | Helpers (`btn`, `field`, `card`, `badge`, `page`, `ic`/`ICONS`…): una función por componente, devuelve HTML | agregar/variar componentes |
| `screens.js` | Pantallas | `MOCK` (datos), `N` (notas), `GROUPS` (pantallas por journey: `{ id, label, r: () => html, n }`), `ALL` (índice plano), router, tema/modo, `auditCtas()` | agregar pantallas |
| `index.html` | Shell | `<meta viewport>`, `<link>` a los CSS, estilos de la herramienta (selector, barra, notas, `?clean`), `<script src>` en orden | rutas relativas (casi nunca) |

**Orden de carga (obligatorio):** `tokens.css` → `components.css` → DOM → `components.js` → `screens.js`. Son scripts clásicos (no módulos ES): los módulos ES no cargan por `file://` por restricción de CORS del navegador. Las declaraciones top-level de `components.js` quedan visibles para `screens.js`.

Dentro de `screens.js`, el router y `auditCtas()` están en la sección "NO EDITAR"; solo se tocan `MOCK` y `GROUPS`.

## Mapeo desde el Design Spec

| Design Spec | En HTML |
|---|---|
| Tokens / variables (`$token`) | Custom properties en `tokens.css` (`--color-primary`) |
| Modos (light/dark) | Redefinir tokens semánticos bajo `body[data-mode="dark"]` en `tokens.css` |
| Componentes | Helper en `components.js` + clase en `components.css`; variantes = parámetro/modificador de clase |
| Pantallas (`"Journey / N-Pantalla"`) | Entrada en `GROUPS` de `screens.js` (grupo = Journey; `label` = `N-Pantalla`; `id` estable en kebab-case) |
| Estados (loading/empty/error/expandido) | Una entrada de `GROUPS` por estado, o parámetro del `r()` |
| Datos de ejemplo | `MOCK` en `screens.js` (contenido real del Design Spec) |
| Notas, supuestos, pendientes | Campo `n` de la pantalla |
| Navegación entre pantallas | `data-goto="<id>"` en el CTA (helper `btn(..., { goto })`) |

## Reglas por archivo

- **Componente nuevo** → se agrega a `components.js` (+ `components.css`). **Nunca inline en `screens.js`**: ni markup repetido ni `<style>`/estilos de componente.
- **Pantalla nueva** → solo se toca `screens.js`. Si necesita un componente que no existe: primero el componente, luego la pantalla (tokens → componentes → pantallas).
- **Token nuevo** → solo `tokens.css`; `components.css` y `screens.js` jamás definen valores.
- Si `existing_design_system: yes`: no se crea componente ni token que ya exista; se propone la adición faltante con aprobación.
- `tokens.css` y los componentes son compartidos entre features: cambiar un token/componente existente afecta a las demás maquetas; explicar el alcance antes de tocarlos (Reglas de Iteración).

## Responsive (obligatorio)

Todo diseño es responsive, sin excepción.

- **Breakpoints estándar** (si el sistema existente declara otros, usar esos y declararlos en el README): mobile `< 600px`, tablet `600–1023px`, desktop `>= 1024px`. Anchos de verificación: **375 / 768 / 1280**.
- **Mobile-first:** el CSS base es el layout mobile; las media queries solo usan `min-width` para ampliar.
- **Viewport meta** siempre presente (en `index.html`).
- **Layout fluido:** `grid`/`flex` con `minmax`, `auto-fit`, `max-width`; sin anchos fijos en px para contenedores. Tablas anchas en contenedor con `overflow-x: auto`.
- **Touch targets** `>= 44px` (`--touch-min`) en controles interactivos.
- Una pantalla no está terminada hasta verificarla en los tres anchos. Si el comportamiento cambia por breakpoint (p. ej. sidebar → menú, tabla → tarjetas), documentarlo en el campo `n` de la pantalla.
- Los CSS variables no funcionan dentro de `@media`: los valores de breakpoint se repiten literalmente en las media queries de `components.css` (y de `index.html` para la herramienta), declarados en un comentario junto a ellas (única excepción permitida a "sin valores sueltos"). Si cambian, cambiarlos en ambos sitios.

## Reglas de tokens

- Cada propiedad visual usa `var(--token)`. El equivalente de la regla "`$variable` en Pencil".
- Nombrar semánticamente (`--color-primary`), no por valor.
- Prohibido: hex, `px` de color/tipografía/espaciado/radio fuera de `tokens.css` (los estilos de herramienta de `index.html` con prefijo `--tool-` no son parte del diseño).
- Sin dominio inventado: usar contenido real del Design Spec; si falta, pedirlo.

## Validación y capturas

Con Chrome MCP (o el navegador disponible):

1. Abrir `file://<ruta>/index.html?clean#<id>` (el `index.html` de `design_dir`) — `?clean` oculta selector, barra y notas; deja solo la pantalla.
2. Redimensionar la ventana a 375, 768 y 1280 de ancho y capturar cada pantalla.
3. Revisar: nada desbordado, sin scroll horizontal no intencional, jerarquía legible, contraste, estados. Comprobar también que no hay 404 de `tokens.css`/`components.*` (rutas relativas correctas).
4. Si no hay navegador disponible: reportarlo explícitamente como validación pendiente; no declarar la pantalla verificada.

## Auditoría de CTAs

1. Enumerar todos los `data-goto` de las pantallas construidas.
2. Cruzar cada destino contra los `id` de `ALL`. `screens.js` expone `auditCtas()` en consola, que devuelve `origen → destino` de los destinos inexistentes.
3. Todo CTA cuyo destino no existe ni está inventariado como pendiente explícito es un **CTA huérfano** y va al reporte de cierre.

## Modo claro/oscuro y tema

- Modo oscuro: redefinir solo tokens semánticos bajo `body[data-mode="dark"]` en `tokens.css` (superficies elevadas más claras, no inversión).
- Tema (color primario alterno): `body[data-theme="..."]` sobreescribiendo `--color-primary*` en `tokens.css`.
- Ambos se alternan con los botones de la barra (router de `screens.js`).

## Limitaciones

- No hay componentes "madre/instancia" nativos: la disciplina es un helper por componente; duplicar markup en pantallas es el anti-patrón equivalente.
- Los íconos del starter son un set mínimo en `ICONS` (`components.js`); ampliar con el set del proyecto (documentar su nombre para los devs).
- El router vive en `screens.js`: cada feature lleva su copia (no editable). Si cambia el router del starter, hay que propagarlo a las features existentes.
- El HTML es referencia de diseño, no código de producto.
