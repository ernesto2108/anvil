---
name: git:commit
description: Analizar cambios staged y escribir un mensaje de commit convencional
---

Eres un experto en mensajes de commit Git. Analiza los cambios staged y escribe un mensaje de commit convencional de alta calidad.

## Paso 1: Verificar cambios staged

Ejecutar `git diff --cached --stat` para ver qué archivos están staged. Si no hay nada staged, decir al usuario "No se encontraron cambios staged. Primero agrega archivos con `git add`." y parar.

Luego ejecutar `git diff --cached` para ver el diff completo de los cambios staged.

## Paso 2: Analizar los cambios

Leer cuidadosamente el diff y determinar:
- **Qué** cambió (archivos, funciones, componentes)
- **Por qué** cambió (corrección de bug, nueva funcionalidad, refactor, etc.)
- **Impacto** (breaking changes, cambios de API, cambios de comportamiento)
- **Scope** (qué módulo/componente/área está afectado)
- **Issues relacionados** (buscar referencias a issues en nombre de rama, comentarios del diff, o marcadores TODO)

Ejecutar `git branch --show-current` para verificar el nombre de la rama por referencias a issues (ej. `feat/123-add-login`, `fix/PROJ-456`).

## Paso 3: Seleccionar el tipo de commit

Elegir el tipo más apropiado basado en el propósito PRINCIPAL del cambio:

| Tipo | Cuándo usarlo |
|------|---------------|
| `feat` | Nueva funcionalidad o capacidad para el usuario (dispara bump de versión MINOR) |
| `fix` | Corrección de bug (dispara bump de versión PATCH) |
| `docs` | Solo documentación (README, comentarios, JSDoc, etc.) |
| `style` | Formateo, espacios, punto y coma — sin cambio de lógica |
| `refactor` | Reestructuración de código — sin agregar feature, sin corregir bug |
| `test` | Agregar o actualizar tests únicamente |
| `chore` | Tareas de mantenimiento (deps, configs, tooling) |
| `perf` | Mejora de rendimiento sin cambio de comportamiento |
| `ci` | Cambios en pipeline CI/CD (GitHub Actions, Jenkins, etc.) |
| `build` | Cambios en sistema de build o dependencias externas |

## Paso 4: Escribir el mensaje de commit

Seguir estas reglas estrictamente:

### Formato de línea de asunto
```
<type>(<scope>): <description>
```

**Reglas:**
1. El tipo es en minúsculas, de la tabla anterior
2. El scope es opcional — un sustantivo que describe el área afectada (ej. `auth`, `parser`, `api`, `ui`)
3. La descripción empieza con letra minúscula
4. Usar modo imperativo ("add" no "added", "fix" no "fixes")
5. NO terminar con punto
6. La línea de asunto total DEBE ser 50 caracteres o menos — este es un límite duro
7. Si 50 chars es muy ajustado, acortar la descripción — nunca excederlo

### Cuerpo (cuando sea necesario)
- Separar del asunto con UNA línea en blanco
- Ajustar cada línea a 72 caracteres
- Explicar QUÉ cambió y POR QUÉ, no CÓMO (el diff muestra cómo)
- Usar viñetas para múltiples items
- Incluir cuerpo para cualquier cambio no trivial

### Footer (cuando sea necesario)
- Separar del cuerpo con UNA línea en blanco
- Referencias a issues: `Closes #123`, `Fixes #456`, `Refs PROJ-789`
- Breaking changes: `BREAKING CHANGE: <descripción de qué rompe>`
- Si se agrega `!` después de type/scope, TAMBIÉN incluir footer BREAKING CHANGE con detalles
- Co-autores: `Co-authored-by: Name <email>`

### Anti-patrones — NUNCA escribir estos:
- "fix bug" / "fix issue" — describir CUÁL bug
- "update code" / "update file" — describir QUÉ se actualizó
- "changes" / "misc" / "stuff" — siempre ser específico
- "WIP" — los commits deben ser atómicos y completos
- Tiempo pasado ("added", "fixed", "removed") — usar imperativo
- Terminar el asunto con punto

## Paso 5: Preguntar por referencia a ticket/issue

Antes de presentar el mensaje final, usar la herramienta AskUserQuestion para preguntar:

**Pregunta:** "¿Este commit pertenece a un ticket o issue?"
**Header:** "Ticket"
**Opciones:**
1. "Sin ticket" — No se necesita referencia a issue
2. "Sí, déjame escribirlo" — Proporcionaré el ID del ticket (ej. TECHADMIN-123, PROJ-456, #78)

Si el usuario proporciona una referencia a ticket:
- Agregarla al footer del mensaje de commit como `Refs <TICKET-ID>` (ej. `Refs TECHADMIN-123`)
- Si el commit es un fix, usar `Fixes <TICKET-ID>` en su lugar
- Si ya se detectó una referencia a issue del nombre de la rama, mostrar ambas y dejar al usuario confirmar cuál mantener

## Paso 6: Presentar al usuario

Mostrar el mensaje de commit completo en un bloque de código. Formatearlo exactamente como aparecerá en git.

Luego preguntar: **"¿Hago commit con este mensaje? (sí/editar/cancelar)"**

- Si el usuario dice **sí**: ejecutar `git commit -m "$(cat <<'EOF'\n<mensaje completo aquí>\nEOF\n)"` usando un heredoc para formato correcto
- Si el usuario dice **editar** o proporciona cambios: revisar el mensaje y preguntar de nuevo
- Si el usuario dice **cancelar**: parar sin hacer commit

## Ejemplos de buenos mensajes de commit

```
feat(auth): add OAuth2 login with Google provider
```

```
fix(parser): handle empty input without crashing

Previously, passing an empty string to parse() would throw an
unhandled TypeError. Now returns an empty result object.

Closes #342
```

```
refactor(api)!: rename user endpoints to follow REST conventions

Rename /getUser to GET /users/:id and /createUser to POST /users
to align with REST standards.

BREAKING CHANGE: all /getUser and /createUser endpoints have been
removed. Clients must migrate to /users/:id (GET) and /users (POST).

Refs PROJ-891
```

```
perf(db): add index on orders.customer_id for faster lookups

Query time for customer order history reduced from ~800ms to ~15ms
on production dataset (12M rows).
```
