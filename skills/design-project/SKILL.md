---
name: design-project
description: Quick entry point for resuming or starting design projects. Auto-detects the design tool (.pen → Pencil, Figma URL → Figma), opens the file, loads context (variables, components, screens), and prepares the workspace. Use when user says "open design", "resume design", "design project", "pencil project", "figma project", or wants to start designing.
---

# Design Project — Quick Start

> Entry point only. Opens the workspace and hands off to `/design-system` or direct design work.

## Step 1: Detect design files

Scan `~/projects/` for design files:

```bash
# Pencil files
find ~/projects -name "*.pen" -type f 2>/dev/null

# Figma references (check for .figma URLs in project docs or config)
grep -r "figma.com" ~/projects/*/README.md ~/projects/*/.env 2>/dev/null
```

## Step 2: Present options (in Spanish)

Show the user a numbered list:

```
Proyectos de diseno encontrados:

1. my-project/design/app.pen  (Pencil — repo: my-project)
2. https://figma.com/file/xxx  (Figma — repo: my-web-app)
N. [Crear nuevo] — crear un archivo de diseno en un repo existente

Cual quieres abrir?
```

For each file found, show:
- Path or URL
- Design tool detected (Pencil / Figma)
- Associated repo name

If NO design files exist, skip to Step 3 and ask which repo needs a new design.

## Step 3: Handle "Create new"

If the user picks an existing repo without a design file:

1. Ask: "Que herramienta de diseno? (Pencil / Figma)"
2. Ask: "Donde quieres guardar el archivo? (ej. `design/`, `src/design/`, raiz del repo?)"
3. Ask: "Que tipo de producto es? Web, mobile, ambos?"
4. Create the file:
   - **Pencil**: `open_document("new")`
   - **Figma**: guide user to create a Figma file and share the URL

## Step 4: Open and load context

### If Pencil (.pen file):

1. `open_document(filePath)`
2. `get_editor_state()` — current canvas, selection, active page
3. `get_variables()` — design tokens
4. `batch_get({ patterns: [{ reusable: true }] })` — reusable components
5. `batch_get({ patterns: [{ type: "frame", depth: 0 }] })` — top-level frames (screens)

### If Figma:

1. Load `/figma-use` skill (mandatory before any Figma tool call)
2. Read the Figma file structure
3. List pages, frames, and components

## Step 5: Present project status (in Spanish)

Summarize what you found:

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

If the file is empty (new), say so and suggest starting with `/design-system`.

## Step 6: Verify canvas organization (Pencil only)

After presenting the project status, check if the canvas is organized:

1. `snapshot_layout(maxDepth: 0)` — read all top-level frame positions
2. Check if frames follow the organization rules:
   - Row 1: Library + Component States
   - Row 2+: Screens in chronological order (v1, v2, etc.)
   - Last rows: Mobile screens
   - ~200px gaps between rows and between horizontal frames
3. If frames are disorganized, ask: "El canvas esta un poco desordenado, quieres que lo organice?"
4. Never reorganize without asking first

## Rules

- **Always speak Spanish** with the user
- **Never assume which project** — always ask, even if there's only one design file (confirm it)
- **Auto-detect the tool** — don't ask "Pencil or Figma?" if the file extension makes it obvious
- **Read-only on open** — this skill only opens and reads. No modifications to the design file
- **Hand off, don't overlap** — this skill does NOT create variables, components, or screens. That's `/design-system`'s job
- **Show repo name** — always show which repo the design belongs to so the user can correlate design ↔ code
