---
name: git:commit
description: Analizar cambios staged y escribir un mensaje de commit convencional
tools: Bash, Read, Grep, Glob
---

Eres un experto en mensajes de commit Git. Este command es el **wrapper interactivo** para el usuario — toda la lógica de análisis de diff y ejecución del commit vive en la skill `git-commit`. Aquí solo se manejan las interacciones con el usuario (ticket opcional + confirmación final).

> **Idioma del mensaje:** el mensaje de commit generado va **siempre en inglés** (asunto, cuerpo y footer), sin importar el idioma del historial del repo ni el de esta conversación. Los identificadores de código, paths y flags se citan verbatim. La interacción contigo sigue en español.

> **No re-implementes la lógica de análisis aquí.** Si necesitas cambiar el formato del mensaje, los anti-patrones o la tabla de tipos de Conventional Commit, edita `skills/git-commit/SKILL.md`. Este command es deliberadamente delgado.

## Paso 1: Verificar cambios staged

Ejecutar `git diff --cached --stat`. Si no hay nada staged, decir al usuario "No se encontraron cambios staged. Primero agrega archivos con `git add`." y parar.

## Paso 2: Preparar el mensaje (sin commitear todavía)

Cargar la skill `git-commit` para analizar el diff y redactar el mensaje convencional, **pero NO ejecutar el commit en esta etapa** — el commit se hace en el Paso 4 después de la confirmación del usuario.

La skill se encarga de:
- Leer el diff staged (`git diff --cached --stat` + `git diff --cached`)
- Determinar tipo, scope, descripción, cuerpo y footer del Conventional Commit
- Detectar automáticamente referencias a ticket en el nombre de la rama
- Aplicar las reglas de formato (≤50 chars en asunto, imperativo, sin punto final, anti-patrones)

Para este flujo interactivo, **interceptar el output de la skill antes del Paso 5** (ejecución del commit). Quedarte solo con el `commit_message` que la skill compone y NO permitir que la skill ejecute `git commit` todavía.

> **Nota de implementación:** si el modelo no puede interceptar la ejecución de la skill, alternativamente puede aplicar los Pasos 1-4 de `skills/git-commit/SKILL.md` (análisis y redacción) sin invocar el Paso 5 de esa skill (ejecución). El objetivo es reusar las reglas y el formato sin duplicarlas en este command.

## Paso 3: Preguntar por referencia a ticket/issue

Antes de presentar el mensaje final, usar la herramienta `AskUserQuestion` para preguntar:

**Pregunta:** "¿Este commit pertenece a un ticket o issue?"
**Header:** "Ticket"
**Opciones:**
1. "Sin ticket" — No se necesita referencia a issue
2. "Sí, déjame escribirlo" — Proporcionaré el ID del ticket (ej. TECHADMIN-123, PROJ-456, #78)

Si el usuario proporciona una referencia a ticket:
- Agregarla al footer del mensaje como `Refs <TICKET-ID>` (ej. `Refs TECHADMIN-123`)
- Si el commit es un fix, usar `Fixes <TICKET-ID>` en su lugar
- Si la skill ya detectó una referencia desde el nombre de la rama, mostrar ambas y dejar al usuario confirmar cuál mantener

## Paso 4: Presentar al usuario y ejecutar el commit

Mostrar el mensaje de commit completo en un bloque de código, formateado exactamente como aparecerá en git.

Preguntar: **"¿Hago commit con este mensaje? (sí/editar/cancelar)"**

- Si el usuario dice **sí**: ejecutar `git commit -m "$(cat <<'EOF'\n<mensaje completo aquí>\nEOF\n)"` usando heredoc para preservar el formato
- Si el usuario dice **editar** o proporciona cambios: revisar el mensaje y preguntar de nuevo
- Si el usuario dice **cancelar**: parar sin hacer commit

## Para uso no-interactivo (sub-agentes)

Este command está pensado para el usuario humano. Los sub-agentes (ej. el `committer`) **NO deben invocar este command** — deben cargar directamente la skill `git-commit`, que es no-interactiva por diseño y termina con el commit ejecutado sin pedir confirmación.

Si un agente invoca este command por error, la skill quedará esperando input del usuario en los Pasos 3 y 4 y bloqueará la fase.
