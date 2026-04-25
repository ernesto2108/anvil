# Plantilla de salida para project.md

Usa esta plantilla al escribir `{context_path}`. Incluye solo las secciones para los stacks detectados. Elimina las secciones de stack que no apliquen.

---

```markdown
# <Project Name> — Project Context

## Objetivo del producto (Canónico)
<resumen breve — qué es el producto y qué problema resuelve>

## Reglas no negociables
- <regla del usuario>

## Para qué debe optimizar la IA
- <punto del usuario>

## Qué NO sugerir
- <punto del usuario>

## Snapshot del repositorio (Contexto técnico)

### Stacks detectados
- <Go 1.23 | React 18 + Vite | Flutter 3.x | etc.>

### Árbol de directorios (real)
```text
<ÁRBOL REAL AQUÍ — profundidad 3, excluyendo <docs>/, .git/, vendor/, node_modules/, dist/, build/>
```

<!-- === SECCIÓN GO (incluir solo si se encuentra go.mod) === -->
### Go
- **Versión:** <de go.mod>
- **Módulo:** <module path>
- **Dependencias clave:** <solo top-level, ej., chi, sqlx, pgx, slog>
- **Archivos de test:** <cantidad y ejemplos de rutas>
- **Config de linter:** <.golangci.yml detectado / no encontrado>

<!-- === SECCIÓN REACT/NODE (incluir solo si se encuentra package.json) === -->
### React / Node
- **Framework:** <Next.js / Vite / CRA / none>
- **Dependencias clave:** <react, typescript, tailwind, etc.>
- **Test runner:** <vitest / jest / other>
- **Archivos de test:** <cantidad y ejemplos de rutas>
- **Config de lint:** <eslint config detectado / no encontrado>
- **TypeScript:** <tsconfig.json detectado / no encontrado>

<!-- === SECCIÓN FLUTTER (incluir solo si se encuentra pubspec.yaml) === -->
### Flutter
- **Versión de Dart:** <del environment en pubspec.yaml>
- **Dependencias clave:** <riverpod, bloc, dio, etc.>
- **Archivos de test:** <cantidad y ejemplos de rutas>
- **Config de análisis:** <analysis_options.yaml detectado / no encontrado>
- **Localización:** <l10n.yaml detectado / no encontrado>

<!-- === OTROS STACKS (agregar según sea necesario) === -->

### CI / Runtime detectado
- <Dockerfile: sí/no>
- <docker-compose: sí/no>
- <GitHub Actions: lista de archivos de workflow>

### Herramientas de build
- <Makefile / taskfile / scripts detectados>

### Archivos de configuración
- <cualquier otra configuración notable>

## Notas para agentes
- seguir la estructura existente exactamente
- no introducir nuevos patrones arquitectónicos
- colocar el nuevo código cerca de archivos similares existentes
```
