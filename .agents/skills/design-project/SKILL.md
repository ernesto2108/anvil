---
name: design-project
description: Punto de entrada rápido para retomar o iniciar proyectos de diseño. Auto-detecta la herramienta de diseño (.pen → Pencil, URL de Figma → Figma), abre el archivo, carga el contexto (variables, componentes, pantallas) y prepara el workspace. Usar cuando el usuario diga "open design", "resume design", "design project", "pencil project", "figma project", o quiera comenzar a diseñar.
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->


# Design Project — Quick Start

> Solo punto de entrada. Abre el workspace y deriva a `/design-system` o trabajo de diseño directo.

## Paso 1: Detectar archivos de diseño

Escanear `~/projects/` en busca de archivos de diseño:

```bash
# Archivos Pencil
find ~/projects -name "*.pen" -type f 2>/dev/null

# Referencias de Figma (verificar URLs .figma en docs del proyecto o config)
grep -r "figma.com" ~/projects/*/README.md ~/projects/*/.env 2>/dev/null
```

## Paso 2: Presentar opciones (en español)

Mostrar al usuario una lista numerada:

```
Proyectos de diseno encontrados:

1. my-project/design/app.pen  (Pencil — repo: my-project)
2. https://figma.com/file/xxx  (Figma — repo: my-web-app)
N. [Crear nuevo] — crear un archivo de diseno en un repo existente

Cual quieres abrir?
```

Para cada archivo encontrado, mostrar:
- Ruta o URL
- Herramienta de diseño detectada (Pencil / Figma)
- Nombre del repo asociado

Si NO existen archivos de diseño, saltar al Paso 3 y preguntar qué repo necesita un nuevo diseño.

## Paso 3: Manejar "Crear nuevo"

Si el usuario elige un repo existente sin archivo de diseño:

1. Preguntar: "Que herramienta de diseno? (Pencil / Figma)"
2. Preguntar: "Donde quieres guardar el archivo? (ej. `design/`, `src/design/`, raiz del repo?)"
3. Preguntar: "Que tipo de producto es? Web, mobile, ambos?"
4. Crear el archivo:
   - **Pencil**: El MCP de Pencil no puede crear archivos — requiere acción manual del usuario. Indicar al usuario: "Crea un archivo `.pen` vacío desde el editor Pencil (File → New), ábrelo, y pásame el path para continuar."
   - **Figma**: guiar al usuario a crear un archivo de Figma y compartir la URL

## Paso 4: Abrir y cargar contexto

### Si es Pencil (archivo .pen):

1. `get_editor_state({ include_schema: true })` — canvas actual, schema del .pen, página activa
2. `snapshot_layout({ maxDepth: 0 })` — frames de nivel superior (pantallas) sin bajar al árbol
3. `get_variables()` — design tokens
4. `batch_get({ patterns: [{ reusable: true }], searchDepth: 2, readDepth: 1 })` — componentes reutilizables

### Si es Figma:

1. Cargar `reference/figma-workflow.md` desde el skill `/design-system` para instrucciones de trabajo con Figma.
2. Leer la estructura del archivo Figma
3. Listar páginas, frames y componentes

## Paso 5: Presentar el estado del proyecto (en español)

Resumir lo que encontraste:

```
Proyecto: my-project
Herramienta: Pencil
Archivo: design/app.pen
Repo: ~/projects/my-project

Estado actual:
- Variables: 45 definidas (colores, tipografia, spacing)
- Componentes: 12 reusables (Button, Card, Badge, ...)
- Pantallas: 2 web (dark/light), 2 mobile (dark/light)
- Pendiente: version mobile del blog

Listo para trabajar. Usa /design-system para modificar tokens/componentes
o dime que pantalla quieres disenar.
```

Si el archivo está vacío (nuevo), indicarlo y sugerir comenzar con `/design-system`.

## Paso 6: Verificar organización del canvas (solo Pencil)

Después de presentar el estado del proyecto, verificar si el canvas está organizado:

1. `snapshot_layout(maxDepth: 0)` — leer posiciones de todos los frames de nivel superior
2. Verificar si los frames siguen las reglas de organización:
   - Fila 1: Library + Component States
   - Fila 2+: Pantallas en orden cronológico (v1, v2, etc.)
   - Últimas filas: Pantallas mobile
   - ~200px de separación entre filas y entre frames horizontales
3. Si los frames están desorganizados, preguntar: "El canvas esta un poco desordenado, quieres que lo organice?"
4. Nunca reorganizar sin preguntar primero

## Reglas

- **Hablar siempre en español** con el usuario
- **Nunca asumir qué proyecto** — siempre preguntar, incluso si solo hay un archivo de diseño (confirmar)
- **Auto-detectar la herramienta** — no preguntar "¿Pencil o Figma?" si la extensión del archivo lo hace obvio
- **Solo lectura al abrir** — este skill solo abre y lee. Sin modificaciones al archivo de diseño
- **Delegar, no solapar** — este skill NO crea variables, componentes ni pantallas. Eso es responsabilidad de `/design-system`
- **Mostrar el nombre del repo** — siempre mostrar a qué repo pertenece el diseño para que el usuario pueda correlacionar diseño ↔ código
