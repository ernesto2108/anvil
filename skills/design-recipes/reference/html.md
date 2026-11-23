# Recetas — HTML (maqueta navegable)

Implementación de las recetas de `SKILL.md` para la herramienta HTML (default). Estructura de archivos, tokens y router: `reference/html-workflow.md` y carpeta `reference/html-starter/` de `/design-system`. Aquí solo lo específico de construir pantallas.

## Principios

- **Archivos separados** — tokens en `tokens.css`; componentes en `components.js` + `components.css`; pantallas solo en `screens.js` (rutas en `design_dir` / `components_path` / `tokens_path`, resueltas por el Gate de Ubicación de Artefactos).
- **Helpers, no markup repetido** — todo patrón que aparece 2+ veces es un helper JS en `components.js` (`btn`, `card`, `field`, `badge`, `tableRow`…). Equivale a "componente + instancias". Un componente nuevo se agrega a `components.js`/`components.css`, nunca inline en `screens.js`.
- **Tokens siempre** — `var(--token)`; nunca hex ni px sueltos fuera de `tokens.css`.
- **Mobile-first** — escribir el layout mobile en el CSS base de `components.css` y ampliar con `@media (min-width: …)`.
- **Una entrada de `GROUPS` por pantalla/estado** (en `screens.js`) — `id` kebab-case estable; label con convención `"N-Pantalla"` dentro del grupo (grupo = Journey).

## Construir una pantalla

1. Verifica que los helpers que necesitas existen en `components.js`; si falta uno, créalo primero en `components.js` (+ `components.css`) (tokens → componentes → pantallas).
2. En `screens.js`, escribe `r: () => page(titulo, contenido)` usando solo helpers y clases del sistema.
3. CTAs con `data-goto="<id destino>"` (helper `btn(..., { goto })`).
4. Completa `n: N(cambios, supuestos, pendientes)`.
5. Navega a `?clean#<id>` y captura en 375 / 768 / 1280 (Chrome MCP). Corrige antes de pasar a la siguiente.
6. Copia, no reconstruyas: variantes (estados, dark) = otra entrada que reutiliza el mismo helper con otros parámetros.

## Layouts responsive por receta

| Receta (SKILL.md) | Mobile (base) | Tablet `>= 600px` | Desktop `>= 1024px` |
|---|---|---|---|
| Auth (1) | Una columna: header de marca compacto + formulario | Formulario centrado, `max-width` | Dividido: panel de marca + panel de formulario (`grid-template-columns`) |
| App shell (2) | Barra superior + menú colapsable; contenido con padding `--sp-4` | Barra superior; padding `--sp-6` | Sidebar fija + contenido (`grid`) |
| Tabla de datos (3) | Lista de tarjetas (Receta 6) o tabla en `.twrap` con scroll horizontal | Tabla en `.twrap` | Tabla completa |
| Detalle (4) | Una columna apilada | Una columna ancha | Dos columnas (`grid`) |
| Wizard (5) | Stepper compacto (solo paso actual + n/N), botones a ancho completo | Stepper completo | Stepper completo + tarjeta centrada |
| Cards (6) | `grid` 1 col | 2 cols | 3 cols (`repeat(auto-fit, minmax(…))`) |

Para app nativa (iOS/Android), el HTML solo sirve como maqueta del flujo: usar las recetas M1–M4 y `mobile-patterns.md`; el frame de captura es 390 de ancho.

## Estados interactivos

- Dropdown/menú/modal/acordeón: construir el estado abierto como pantalla propia (o parámetro del `r()`), no solo el cerrado.
- Loading = skeleton con clase del sistema; empty/error con helpers `empty()`/`alertBox()` si el proyecto los define.

## Modo oscuro

Redefinir tokens semánticos en `body[data-mode="dark"]` (en `tokens.css`); verificar contraste en badges y texto con la fórmula de `color-craft.md`.

## Anti-patrones

| Anti-patrón | Corrección |
|---|---|
| Hex/px sueltos en CSS de componentes o inline | `var(--token)` |
| Markup repetido en varias pantallas | Helper JS en `components.js` |
| Componente o estilos inline en `screens.js` | Moverlo a `components.js` + `components.css` |
| Ancho fijo en px en contenedores | `max-width` + layout fluido |
| Pantalla sin verificar en 375/768/1280 | Capturar y revisar los tres anchos |
| CTA con `data-goto` a un id inexistente | Diseñar la pantalla o inventariarla como pendiente |
| Sin `<meta viewport>` | Restaurarla |
| Tokens paralelos cuando ya existe sistema | Mapear a los tokens existentes |
