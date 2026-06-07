---
name: git-commit
description: Analizar diff staged, redactar mensaje Conventional Commit y ejecutar el commit. Úsalo cuando necesites commitear cambios ya staged con un mensaje convencional ("commit", "conventional commit", "git commit", "redactar mensaje de commit"). No-interactiva por diseño — el command `/git:commit` la usa interceptando el Paso 5 para pedir confirmación al usuario antes de ejecutar.
user-invocable: false
---

# Git Commit

Procedimiento no-interactivo para analizar el diff staged, redactar un mensaje Conventional Commit y ejecutarlo. Pensada para que la consuman agentes (ej. `committer`) o el wrapper interactivo `commands/git/commit.md`.

## Filosofía

1. **Una responsabilidad: el commit.** No hace staging, no hace push, no escribe handoff. Asume staging hecho y termina con el commit ejecutado (o el mensaje listo para que el caller lo intercepte).
2. **No-interactiva por diseño.** No usa `AskUserQuestion` ni pregunta nada. Si falta información para decidir tipo o footer, aplica la regla por defecto documentada y continúa.
3. **El caller decide si ejecuta el Paso 5.** Sub-agentes ejecutan los 5 pasos. El command `/git:commit` aplica solo los Pasos 1–4 y ejecuta el commit por su cuenta tras confirmación del usuario.

## Reglas duras

- No hacer `git add` — el staging debe estar hecho antes.
- No hacer `git push`, `git amend`, `git rebase`, `git reset`.
- No reintentar automáticamente si el commit falla — DETENER y reportar.
- No usar punto final, tiempo pasado ni mensajes genéricos en el asunto.

## Flujo de trabajo

### Paso 1 — Leer el diff staged

```bash
git diff --cached --stat
git diff --cached
```

Si `git diff --cached --stat` está vacío → DETENER: "No hay cambios staged. No commiteo."

### Paso 2 — Analizar los cambios y detectar ticket

Determinar qué cambió, por qué, impacto y scope. Ejecutar:

```bash
git branch --show-current
```

Buscar referencia a ticket en el nombre de la rama. Patrones reconocidos:
- `feat/123-descripcion` → ticket `#123`
- `fix/PROJ-456-...` → ticket `PROJ-456`
- `JIRA-789-...` → ticket `JIRA-789`
- `<tipo>/<TICKET>-...` con `TICKET` en mayúsculas + guión + número

Si se encuentra → guardar para el footer del Paso 4. Si no → footer sin referencia.

### Paso 3 — Seleccionar tipo de Conventional Commit

| Tipo | Cuándo usarlo |
|---|---|
| `feat` | Nueva funcionalidad (bump MINOR) |
| `fix` | Corrección de bug (bump PATCH) |
| `docs` | Solo documentación |
| `style` | Formateo sin cambio de lógica |
| `refactor` | Reestructuración sin feature ni bug |
| `test` | Agregar o actualizar tests únicamente |
| `chore` | Mantenimiento (deps, configs, tooling) |
| `perf` | Mejora de rendimiento |
| `ci` | Cambios en pipeline CI/CD |
| `build` | Sistema de build o dependencias externas |

Si el diff mezcla tipos → elegir el principal por volumen e intención. Si es genuinamente mixto → usar `chore` con scope explícito.

### Paso 4 — Redactar el mensaje

Formato: `<type>(<scope>): <description>`

**Reglas del asunto:**
- Tipo en minúsculas.
- Scope opcional, entre paréntesis (área afectada: módulo, package, capa).
- Descripción en imperativo presente, en minúscula, sin punto final.
- Asunto ≤ 50 caracteres.

**Cuerpo (si el cambio no es trivial):**
- Separar del asunto por una línea en blanco.
- Máx 72 caracteres por línea.
- Explicar QUÉ y POR QUÉ, no CÓMO.

**Footer:**
- `Refs <TICKET-ID>` si se detectó ticket en la rama y el tipo no es `fix`.
- `Fixes <TICKET-ID>` si el tipo es `fix` y hay ticket.
- `BREAKING CHANGE: <descripción>` si aplica.

**Anti-patrones — NUNCA:**
- "fix bug", "update code", "changes", "WIP", "misc", "stuff".
- Tiempo pasado: "added", "fixed", "updated" → usar imperativo: "add", "fix", "update".
- Punto final en el asunto.
- Mezclar idioma (todo en inglés o todo en español, consistente con el repo).

**Checklist pre-commit:**
- [ ] Tipo corresponde al propósito principal del diff
- [ ] Asunto ≤ 50 caracteres, imperativo, sin punto
- [ ] Cuerpo separado del asunto por línea en blanco (si aplica)
- [ ] Sin anti-patrones
- [ ] Footer con ticket si se detectó en la rama

### Paso 5 — Ejecutar el commit

> **Importante:** este paso lo ejecutan sub-agentes no-interactivos. El command `/git:commit` intercepta antes de este paso para pedir confirmación al usuario y ejecuta el commit por su cuenta. Si eres un sub-agente, continúa. Si el caller te indicó "no ejecutar el commit", devuelve el mensaje compuesto y termina aquí.

```bash
git commit -m "$(cat <<'EOF'
<mensaje completo aquí>
EOF
)"
```

- Éxito → capturar `commit_hash` con `git rev-parse HEAD` y `commit_subject` con `git log -1 --format=%s`. Reportar al caller.
- Fallo (pre-commit hook, lint, build) → DETENER. Capturar stdout + stderr completos. Reportar al caller con el error exacto. No reintentar.

## Formato de salida

Al terminar, reportar (máx 80 palabras):

- `commit_hash`: hash corto (`git rev-parse --short HEAD`)
- `commit_subject`: primera línea del mensaje
- Notas si las hay (ej. "ticket detectado en rama: PROJ-456")

Si el caller solicitó modo "solo redactar" (sin ejecutar Paso 5), reportar:

- `commit_message`: mensaje completo verbatim
- `detected_ticket`: ticket detectado o `null`
- `type`: tipo Conventional Commit elegido

## Modos de uso

| Caller | Ejecuta Paso 5 | Notas |
|---|---|---|
| Sub-agente (ej. `committer` vía `committer-flow`) | Sí | Flujo completo end-to-end, no-interactivo. |
| Command `/git:commit` | No | El command intercepta tras Paso 4, pide confirmación al usuario y ejecuta el commit por su cuenta en su propio Paso 4. |
