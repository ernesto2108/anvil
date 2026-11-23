# Coding Standards — <ProjectName>

<!-- Naming, estructura de carpetas, idioma del código, reglas de linting y patrones prohibidos.
     Complementado con patrones de diseño detectados automáticamente en el código. -->

last_updated: <YYYY-MM-DD>

## Idioma del código

- **Código fuente:** `<inglés / español>`
- **Comentarios:** `<inglés / español>`
- **Commits:** `<español — Conventional Commits (tipos verbatim: feat, fix, chore…)>`
- **Documentación técnica:** `<español>`

## Naming

### General
- **Variables y funciones:** `<camelCase / snake_case / PascalCase>`
- **Constantes:** `<UPPER_SNAKE_CASE>`
- **Archivos:** `<kebab-case / snake_case>`
- **Tipos / Interfaces / Structs:** `PascalCase`

### Por dominio
- **Handlers:** `<convención — ej: Handle<Acción>>`
- **Servicios:** `<convención — ej: <Entidad>Service>`
- **Repositorios:** `<convención — ej: <Entidad>Repository>`

## Estructura de carpetas

```
<mostrar estructura del proyecto>
```

## Reglas de imports / dependencias

- <regla — ej: handlers NO importan directamente repos>
- <regla — ej: dominio no importa infraestructura>
- <regla — ej: no dependencias circulares entre dominios>

## Linting configurado

| Herramienta | Config | Reglas destacadas |
|---|---|---|
| `<golangci-lint / eslint / dart analyze>` | `<.golangci.yml / .eslintrc>` | <reglas clave> |

## Patrones prohibidos

<!-- Patrones que se consideraron y descartaron explícitamente -->

- **`<patrón>`:** <por qué está prohibido en este proyecto>
- **`<patrón>`:** <razón>

## Patrones de diseño detectados en el código

<!-- Inferidos automáticamente por context-init — verificar antes de usar -->

### <NombreInferido> — <archivo principal>
- **Archivo:** `<path>:<line>`
- **Qué hace:** <descripción>
- **Cuándo usar:** <contexto>
- **Anti-pattern:** <qué evitar>
