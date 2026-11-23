---
name: design-project
description: Punto de entrada para retomar o iniciar proyectos de diseño. Define el Gate de Selección de Herramienta (HTML por defecto; Pencil y Figma opcionales), el Gate de Discovery de Sistema Existente (detectar/preguntar si el proyecto ya tiene tokens, variables o componentes), el Gate de Ubicación de Artefactos (preguntar en qué ruta guardar pantallas, componentes y tokens HTML) y el Gate de Responsive. Carga el contexto del workspace. Usar cuando el usuario diga "open design", "resume design", "design project", "pencil project", "figma project", "maqueta html", "design system existente", o quiera comenzar a diseñar.
---

# Design Project — Quick Start

> Solo punto de entrada. Resuelve herramienta y sistema existente, abre el workspace y delega la construcción a `/design-system` o a trabajo de diseño directo. Es el **único lugar** donde se definen los gates de selección, discovery y ubicación; las demás skills y agentes los referencian por nombre.

## Gate de Selección de Herramienta (criterio único)

Evaluar en este orden; la primera regla que aplica decide:

1. **Herramienta fijada** por prompt, spec o humano → se respeta (`design_tool: pencil | html | figma`).
2. **HTML (default)** — existe `design/**/*.html` en el repo, o no hay ningún artefacto de diseño.
3. **Pencil (opcional)** — solo si existe un `.pen` **y** `get_editor_state` (MCP `mcp__pencil__*`) responde sin error. Si el MCP falla o no está disponible → no usar Pencil; caer a HTML e informarlo.
4. **Figma (opcional)** — solo con URL de Figma provista.
5. **Ambigüedad** (p. ej. hay `.pen` y también HTML, sin instrucción) → preguntar al humano en una línea.

Registrar la herramienta resuelta en el resumen de estado (Paso 5); el resto del flujo la hereda como `design_tool`.

## Gate de Discovery de Sistema Existente (BLOQUEANTE antes de crear tokens/variables/componentes)

Hay proyectos que ya tienen sistema de diseño. Crear uno paralelo genera deriva visual y retrabajo. Antes de definir CUALQUIER token, variable o componente:

**1. Detectar** (solo lectura, en el repo del proyecto):

| Señal | Dónde buscar |
|---|---|
| CSS custom properties | `:root { --… }` en CSS/SCSS globales |
| Tailwind | `tailwind.config.*`, `@theme` en CSS |
| Theme de librería de UI | theme de MUI / Chakra / Mantine / similares (`createTheme`, `extendTheme`…) |
| Tokens exportados | JSON de tokens (`tokens.json`, Style Dictionary) |
| Librería de componentes interna | paquete workspace o dependencia `@<org>/…-ui` en `package.json` |
| Documento de diseño | `DESIGN.md`, `design-system.md`, `design_system_path` inyectado |
| Sistema previo en herramienta | variables/componentes reutilizables en `.pen` (vía `get_variables`, `batch_get` reusable) o tokens/componentes HTML bajo `design/` (`tokens.css`, `components.js`, o `:root` en un `.html`) |

**2. Preguntar** (si la detección es inconclusa, o para confirmar lo detectado; máx 3 preguntas por turno, en español, una línea cada una):
- "¿El proyecto ya tiene sistema de diseño, tokens o librería de componentes? ¿Dónde vive (path o paquete)?"
- "¿Lo reutilizo como fuente de verdad, o es un diseño aparte?"

**3. Resolver** y registrar el campo `existing_design_system`:

| Valor | Significado | Consecuencia |
|---|---|---|
| `yes` + paths | Existe y se reutiliza | **Reutilizar/mapear** a lo existente (mismos nombres y valores, mismos breakpoints, mismos componentes). Prohibido inventar tokens/componentes paralelos. Registrar la fuente de verdad (paths). Solo proponer adiciones que falten, con aprobación |
| `no` | No existe | Crear desde cero con `/design-system` (variables → componentes → pantallas) |
| `unknown` | No se pudo determinar | **DETENER** antes de crear tokens y preguntar; no avanzar con suposiciones |

El valor y los paths se declaran en el bloque de arranque/inputs de `designer-spec` y `designer-visual`, y se muestran en el resumen de estado (Paso 5).

## Gate de Ubicación de Artefactos (BLOQUEANTE antes de crear o escribir cualquier archivo HTML)

En HTML no hay un archivo que el humano abra de antemano (como el `.pen`): las rutas se deciden explícitamente. **Sin respuesta del humano no se crea ni se escribe nada.** Aplica a la herramienta HTML; en Pencil/Figma el artefacto lo provee el humano (`.pen`/URL).

**1. Detectar** (solo lectura): archivos de diseño HTML existentes bajo `design/` (`index.html`/`*-mockup.html`, `screens.js`, `components.js`, `tokens.css`). Si `existing_design_system: yes` apunta a tokens/componentes en el repo, esa ubicación es la candidata a reutilizar.

**2. Proponer un default y preguntar** (una pregunta de una línea, en español, con el default dentro):
- Sin archivos previos → "Propongo `design/<feature>/` para pantallas (`index.html`, `screens.js`) y `design/_system/` compartido para componentes y tokens (`components.js`, `components.css`, `tokens.css`), así varias features reutilizan el mismo sistema. ¿Confirmas o prefieres otra ruta?"
- Con archivos previos → proponer **reutilizar** esa ubicación (componentes/tokens existentes) y solo ubicar las pantallas nuevas junto a ellas.

**3. Esperar confirmación o ruta alternativa** y registrar los campos:

| Campo | Significado | Default sugerido |
|---|---|---|
| `design_dir` | Carpeta de pantallas de la feature (`index.html`, `screens.js`) | `design/<feature>/` |
| `components_path` | Carpeta de componentes (`components.js`, `components.css`) | `design/_system/` |
| `tokens_path` | Carpeta de tokens (`tokens.css`) | `design/_system/` |

Si `existing_design_system: yes`, `components_path`/`tokens_path` apuntan al mapa de alias/referencias a lo existente (ver `reference/html-workflow.md` de `/design-system`), no a una copia.

**Orden y límite de preguntas:** los gates se resuelven en este orden — Selección de Herramienta → Discovery de Sistema Existente → Ubicación de Artefactos (depende del discovery). Máx 3 preguntas por turno sumando todos los gates; si no caben, las restantes van al turno siguiente en ese mismo orden. Los campos resueltos se declaran en el bloque de arranque de `designer-spec` y `designer-visual` y en el resumen de estado (Paso 5).

## Gate de Responsive (aplica a toda herramienta)

Todo diseño es responsive, sin excepción. Esta skill lo verifica al presentar el estado; la definición operativa vive en `/design-system` (breakpoints, mobile-first) y `/design-recipes`:

- Breakpoints declarados (los del sistema existente si los hay; si no, mobile `< 600`, tablet `600–1023`, desktop `>= 1024`; verificación a 375 / 768 / 1280).
- Una pantalla sin variantes/verificación en mobile, tablet y desktop no se considera terminada.

## Paso 1: Detectar archivos de diseño

Escanear el repo (o los repos indicados por el humano) en busca de artefactos de diseño:

```bash
# Maquetas HTML (default): entradas, pantallas, componentes y tokens
find <repo> -path '*/design/*' \( -name '*.html' -o -name 'screens.js' -o -name 'components.js' -o -name 'tokens.css' \) -not -path '*/node_modules/*' 2>/dev/null

# Archivos Pencil (opcional)
find <repo> -name "*.pen" -type f -not -path '*/node_modules/*' 2>/dev/null

# Referencias de Figma (URLs en docs del proyecto o config)
grep -r "figma.com" <repo>/README.md <repo>/.env 2>/dev/null
```

Si el humano no indicó el repo, preguntar cuál (no escanear todo el home).

## Paso 2: Presentar opciones (en español)

Mostrar una lista numerada:

```
Proyectos de diseno encontrados:

1. design/checkout/index.html            (HTML — repo: my-project)
2. design/app.pen                         (Pencil — repo: my-project)
3. https://figma.com/file/xxx            (Figma — repo: my-web-app)
N. [Crear nuevo] — crear una maqueta de diseno en un repo existente

Cual quieres abrir?
```

Para cada artefacto: ruta o URL, herramienta detectada y repo asociado. Si NO existen artefactos de diseño, saltar al Paso 3.

## Paso 3: Manejar "Crear nuevo"

Si el usuario elige un repo sin artefacto de diseño:

1. Aplicar el **Gate de Selección de Herramienta** (default HTML; no preguntar la herramienta salvo que el humano ya haya pedido Pencil/Figma).
2. Aplicar el **Gate de Discovery de Sistema Existente** antes de crear nada.
3. Si la herramienta es HTML, aplicar el **Gate de Ubicación de Artefactos** (preguntar dónde guardar pantallas, componentes y tokens; sin respuesta no se crea nada). Preguntar además qué tipo de producto es (web / mobile / ambos). Máx 3 preguntas por turno en total, una línea cada una.
4. Crear el artefacto:
   - **HTML**: copiar la carpeta `reference/html-starter/` de `/design-system` según las rutas confirmadas: `index.html` + `screens.js` → `design_dir`; `tokens.css` → `tokens_path`; `components.css` + `components.js` → `components_path` (si ya existen, reutilizarlos, no sobrescribir); ajustar las rutas relativas de `index.html`. Mapear los tokens (existentes o nuevos). Seguir `reference/html-workflow.md`.
   - **Pencil**: el MCP no puede crear archivos — indicar: "Crea un archivo `.pen` vacío desde el editor Pencil (File → New), ábrelo, y pásame el path."
   - **Figma**: guiar al usuario a crear el archivo y compartir la URL.

## Paso 4: Abrir y cargar contexto

### Si es HTML:

1. Leer los archivos: `tokens.css` (tokens y breakpoints), `components.js` (helpers) y `screens.js` (`GROUPS`, pantallas por grupo), más el `index.html` de entrada.
2. Listar pantallas (`id`, `label`, grupo) y helpers/componentes existentes.
3. Si el navegador está disponible (Chrome MCP), abrir `index.html?clean#<id>` para ver el estado actual de la pantalla pedida.

### Si es Pencil (archivo .pen, MCP disponible):

1. `get_editor_state({ include_schema: true })` — canvas actual, schema del .pen, página activa
2. `snapshot_layout({ maxDepth: 0 })` — frames de nivel superior (pantallas) sin bajar al árbol
3. `get_variables()` — design tokens
4. `batch_get({ patterns: [{ reusable: true }], searchDepth: 2, readDepth: 1 })` — componentes reutilizables

### Si es Figma:

1. Cargar `reference/figma-workflow.md` desde el skill `/design-system` para instrucciones de trabajo con Figma.
2. Leer la estructura del archivo Figma
3. Listar páginas, frames y componentes

## Paso 5: Presentar el estado del proyecto (en español)

```
Proyecto: my-project
Herramienta: HTML (resuelta por Gate de Selección)
Ubicación: pantallas design/checkout/ · componentes design/_system/ · tokens design/_system/   (design_dir / components_path / tokens_path)
Repo: my-project
Sistema existente: yes — tokens en src/theme.ts, librería @org/ui   (existing_design_system: yes)
Breakpoints: 600 / 1024 (del sistema existente)

Estado actual:
- Tokens: 45 definidos (colores, tipografia, spacing)
- Componentes/helpers: 12 (btn, card, field, ...)
- Pantallas: 6 (verificadas responsive: 4/6)
- Pendiente: variantes tablet de pantallas 5 y 6

Listo para trabajar. Usa /design-system para modificar tokens/componentes
o dime que pantalla quieres disenar.
```

Si el artefacto está vacío (nuevo), indicarlo y sugerir comenzar con `/design-system`.

## Paso 6: Verificar organización (solo Pencil)

Solo si la herramienta resuelta es Pencil, después de presentar el estado verificar si el canvas está organizado:

1. `snapshot_layout(maxDepth: 0)` — leer posiciones de todos los frames de nivel superior
2. Verificar si los frames siguen las reglas de organización:
   - Fila 1: Library + Component States
   - Fila 2+: Pantallas por iteración en orden cronológico (v1, v2, etc.); dentro de cada iteración, agrupadas por journey en orden de flujo con la nomenclatura `"Journey / N-Pantalla"` (ver `reference/pencil-workflow.md` de `/design-system`)
   - Últimas filas: Pantallas mobile
   - ~200px de separación entre filas y entre frames horizontales
3. Si los frames están desorganizados, preguntar: "El canvas esta un poco desordenado, quieres que lo organice?"
4. Nunca reorganizar sin preguntar primero

En HTML la organización equivalente es el orden de `GROUPS` (journey en orden de flujo).

## Reglas

- **Hablar siempre en español** con el usuario
- **Nunca asumir qué proyecto** — siempre preguntar, incluso si solo hay un artefacto de diseño (confirmar)
- **Herramienta por el Gate de Selección** — HTML por defecto; no preguntar la herramienta si el gate la resuelve
- **Discovery antes de crear** — ningún token, variable o componente se crea sin resolver `existing_design_system`
- **Ubicación antes de escribir** — en HTML, ningún archivo se crea ni escribe sin resolver el Gate de Ubicación de Artefactos (`design_dir`, `components_path`, `tokens_path`)
- **Solo lectura al abrir** — este skill solo abre y lee. Sin modificaciones al artefacto de diseño (salvo crear la maqueta nueva desde el starter en el Paso 3, con rutas confirmadas)
- **Delegar, no solapar** — este skill NO crea variables, componentes ni pantallas. Eso es responsabilidad de `/design-system`
- **Mostrar el nombre del repo** — siempre mostrar a qué repo pertenece el diseño para correlacionar diseño ↔ código
