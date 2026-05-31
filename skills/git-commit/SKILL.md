---
name: git-commit
description: Analizar cambios staged, redactar un mensaje de Conventional Commits y ejecutar `git commit` sin interacción con el usuario. Usar cuando un sub-agente (típicamente el `committer`) necesite generar y persistir un commit no-interactivo después de hacer staging. NO usar en flujos interactivos del usuario — para eso existe el slash command `/git:commit`.
user-invocable: false
---

# Git Commit (no-interactivo)

## Filosofía

1. **Sin preguntas al usuario** — esta skill se ejecuta desde sub-agentes sin sesión humana. Cualquier ambigüedad se resuelve por defecto, no por pregunta.
2. **Una sola operación atómica** — leer diff, analizar, redactar y commitear son pasos de un único flujo que termina con un commit hash o con un error reportado al invocador.
3. **Mensaje fiel al diff** — el mensaje describe QUÉ cambió y POR QUÉ, basado en el diff staged, no en lo que el invocador "esperaba" cambiar.

## Precondición

Esta skill asume que el invocador ya hizo `git add` de los archivos correctos. La skill NO stagea archivos — solo opera sobre lo que ya esté staged.

Si al invocar esta skill no hay nada staged (`git diff --cached --stat` vacío), la skill DETIENE y reporta: "No hay cambios staged. Imposible commitear." No intenta stagear nada por su cuenta.

## Flujo de Trabajo

### Paso 1 — Leer el diff staged

Ejecutar dos comandos:

```bash
git diff --cached --stat
git diff --cached
```

El primero da una visión panorámica (qué archivos cambiaron y volumen). El segundo da el contenido textual para razonar sobre el cambio.

Si `git diff --cached --stat` está vacío → DETENER y reportar al invocador: "No hay cambios staged. No commiteo."

### Paso 2 — Analizar los cambios

Sobre el diff completo, determinar:

- **Qué** cambió (archivos, funciones, componentes específicos)
- **Por qué** cambió (corrección de bug, nueva funcionalidad, refactor, etc. — inferir desde la naturaleza del diff)
- **Impacto** (¿breaking? ¿cambio de API público? ¿cambio de comportamiento observable?)
- **Scope** (qué módulo/área del repo está afectada — usar como `scope` del Conventional Commit si es un cambio acotado a un área clara)

Ejecutar `git branch --show-current` y buscar en el nombre de la rama una referencia a ticket (patrones típicos: `feat/123-…`, `fix/PROJ-456`, `JIRA-789-…`). Si existe → guardarla para usar en el footer. Si no hay rama identificable o no contiene ticket → no agregar footer de referencia.

### Paso 3 — Seleccionar el tipo de Conventional Commit

Elegir el tipo más apropiado basado en el propósito **principal** del cambio:

| Tipo | Cuándo usarlo |
|------|---------------|
| `feat` | Nueva funcionalidad o capacidad para el usuario (bump MINOR) |
| `fix` | Corrección de bug (bump PATCH) |
| `docs` | Solo documentación (README, comentarios, JSDoc) |
| `style` | Formateo, espacios, punto y coma — sin cambio de lógica |
| `refactor` | Reestructuración de código — sin agregar feature, sin corregir bug |
| `test` | Agregar o actualizar tests únicamente |
| `chore` | Tareas de mantenimiento (deps, configs, tooling) |
| `perf` | Mejora de rendimiento sin cambio de comportamiento |
| `ci` | Cambios en pipeline CI/CD |
| `build` | Cambios en sistema de build o dependencias externas |

Si el diff mezcla varios tipos (ej: feat + tests para esa feat), elegir el tipo del cambio principal — el resto va al cuerpo del mensaje. Si el cambio es genuinamente mixto sin uno dominante, preferir `chore` con scope explícito.

### Paso 4 — Redactar el mensaje

**Formato de línea de asunto:**

```
<type>(<scope>): <description>
```

Reglas:
1. Tipo en minúsculas, de la tabla anterior.
2. Scope opcional — sustantivo corto que describe el área afectada (`auth`, `parser`, `api`, `ui`, `committer`). Omitir si el cambio es transversal.
3. Descripción empieza con letra minúscula.
4. Modo imperativo ("add" no "added", "fix" no "fixes").
5. NO terminar con punto.
6. Línea de asunto total ≤ 50 caracteres — límite duro. Si no cabe, acortar la descripción.

**Cuerpo (cuando el cambio no es trivial):**
- Separar del asunto con UNA línea en blanco.
- Ajustar cada línea a 72 caracteres.
- Explicar QUÉ y POR QUÉ, no CÓMO (el diff muestra cómo).
- Usar viñetas para múltiples items.

**Footer (cuando aplica):**
- Separar del cuerpo con UNA línea en blanco.
- Referencia a ticket (solo si el Paso 2 la detectó en la rama):
  - `Refs <TICKET-ID>` si el tipo no es `fix`
  - `Fixes <TICKET-ID>` si el tipo es `fix`
- Breaking changes: `BREAKING CHANGE: <descripción>`. Si se agrega `!` después de type/scope, incluir también el footer detallado.

**Anti-patrones — NUNCA escribir:**
- "fix bug" / "fix issue" — describir CUÁL bug
- "update code" / "update file" — describir QUÉ se actualizó
- "changes" / "misc" / "stuff" — siempre específico
- "WIP" — los commits deben ser atómicos
- Tiempo pasado ("added", "fixed") — usar imperativo
- Punto final en el asunto

### Paso 5 — Ejecutar el commit

Ejecutar el commit con heredoc para preservar saltos de línea y caracteres especiales del mensaje:

```bash
git commit -m "$(cat <<'EOF'
<mensaje completo aquí>
EOF
)"
```

**Si el commit tiene éxito:**
1. Capturar el hash con `git rev-parse HEAD`.
2. Capturar el subject con `git log -1 --format=%s`.
3. Devolver al invocador (ver Formato de Salida).

**Si el commit falla** (pre-commit hook, lint, build, formato, etc.):
1. NO reintentar automáticamente.
2. Capturar el output textual completo (stdout + stderr + exit code).
3. DETENER y reportar al invocador con el error textual exacto. El invocador (típicamente el `committer`) decide cómo enrutar (por ejemplo, escalando al humano para que invoque al developer del stack correspondiente: `developer-backend` / `developer-frontend` / `developer-mobile`).

## Formato de Salida

Al terminar exitosamente, devolver al invocador un bloque con esta estructura exacta:

```
commit_hash: <hash de git rev-parse HEAD>
commit_subject: <primera línea del mensaje>
commit_message: |
  <mensaje completo verbatim>
```

Al fallar, devolver:

```
commit_failed: true
exit_code: <código>
error_output: |
  <stdout + stderr completo del comando que falló>
intended_message: |
  <el mensaje que se intentó usar, para que el invocador lo registre>
```

## Checklist Pre-Ejecución

- [ ] Hay cambios staged (`git diff --cached --stat` no vacío)
- [ ] El tipo elegido corresponde al propósito principal del cambio
- [ ] La línea de asunto es ≤ 50 caracteres
- [ ] El asunto está en imperativo, sin punto final
- [ ] Si hay cuerpo, está separado del asunto por una línea en blanco
- [ ] El mensaje no contiene anti-patrones ("fix bug", "WIP", "update code")
- [ ] Si la rama contenía un ticket, el footer lo refleja con `Refs` o `Fixes`

## Restricciones

- Esta skill NO llama a `AskUserQuestion` bajo ninguna circunstancia.
- Esta skill NO ejecuta `git add` — solo opera sobre lo que ya está staged.
- Esta skill NO ejecuta `git push` ni operaciones sobre el remoto — termina en el commit local.
- Esta skill NO usa `git commit --amend`, `--no-verify` ni flags que reescriban historia o salten hooks.
- Si pre-commit hook falla, la skill reporta el error y se detiene — nunca lo bypassea.
