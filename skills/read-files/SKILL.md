---
name: read-files
description: Leer archivos del proyecto de forma segura para recopilar contexto antes de tomar decisiones. Usar cuando se necesite entender el código existente antes de modificarlo, recopilar contexto de múltiples archivos, o resumir codebases grandes.
user-invocable: false
---

# Skill — read-files

Convenciones de lectura segura y eficiente para cualquier agente que necesite recopilar contexto del repo antes de decidir o escribir. La filosofía es: **leer solo lo necesario, leerlo bien, citar la fuente exacta**.

## Capacidades

- Leer uno o múltiples archivos.
- Leer carpetas de forma recursiva (con `Glob`/`LS`).
- Resumir archivos grandes extrayendo solo las secciones relevantes.
- Buscar por keyword o símbolo con `Grep` antes de `Read` cuando el archivo es grande.

## Paths prohibidos

Nunca leer, buscar (`Grep`), ni listar (`Glob`) sobre estos paths — son artefactos generados o dependencias externas que contaminan el contexto sin aportar valor:

- `node_modules/**`
- `.pnpm-store/**`, `.yarn/cache/**`
- `dist/**`, `build/**`, `out/**`
- `.next/**`, `.nuxt/**`, `.svelte-kit/**`, `.astro/**`
- `coverage/**`

Si necesitas verificar una librería de terceros, leer el código fuente publicado (docs/web) — no `node_modules/`.

## Reglas de lectura

1. **Paths absolutos siempre.** Nunca usar paths relativos en herramientas de lectura — el cwd puede cambiar entre tool calls. Si recibes un path relativo, resuélvelo a absoluto antes de leerlo.

2. **No releer.** No volver a leer un archivo que ya fue leído en esta sesión, ni un archivo cuyo contenido ya fue pasado inline en el prompt. Si necesitas refrescar memoria, citar el path y línea que ya viste — no re-disparar `Read`.

3. **Lectura progresiva para archivos grandes.** Para archivos de >500 líneas donde el objetivo es específico (una función, un tipo, un campo concreto): usar `Grep` primero para localizar la línea relevante, luego `Read` con `offset`/`limit` solo de esa sección. No leer 2000 líneas para extraer 20.

4. **Lectura completa para archivos medianos.** Para archivos <500 líneas, leer completo antes de concluir que no responden — las primeras N líneas pueden no contener la respuesta aunque el archivo sí la tenga más abajo.

5. **Prioridad: contratos antes que implementaciones.** Leer primero archivos de tipos, interfaces, schemas y contratos (`*.proto`, `*.d.ts`, archivos de `types/`, `interfaces/`, `domain/`, schemas SQL) antes que las implementaciones que los usan. Son más densos en información por línea y orientan la lectura del resto.

6. **Citar fuente en cada hallazgo.** Cada afirmación derivada de un archivo debe llevar cita exacta: `path/absoluto:línea` o rango `path:línea-línea`. Para fuentes web: URL completa + fecha de acceso. Sin cita, no es un hallazgo — es una asunción.

7. **No asumir contenido.** Nunca alucinar ni inferir lo que dice un archivo sin haberlo leído. Si no lo leíste, dilo explícitamente ("no consultado") o léelo. No hay punto medio.

## Formato de salida

Cuando reportes hallazgos basados en lecturas:

- **Path** absoluto del archivo (o URL para fuentes web).
- **Resumen breve** de lo relevante (1-2 líneas).
- **Snippets** solo de las secciones relevantes — no copiar archivos enteros.
- **Cita exacta** (`path:línea`) por cada afirmación.

## Nunca

- Alucinar o inferir contenido sin haber leído.
- Asumir código sin fuente citada.
- Re-leer archivos ya cargados en la sesión.
- Tocar paths prohibidos.
- Usar paths relativos.
