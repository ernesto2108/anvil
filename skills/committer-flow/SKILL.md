---
name: committer-flow
description: Flujo completo de tres fases para commit, push y PR seguro — inputs, Conventional Commits, captura de rama destino, push, PR estructurado y manejo de errores. Úsalo cuando el usuario pida "commit y push", "haz el commit", "sube los cambios", "push a la rama", o cuando cierres una tarea que necesita quedar commiteada y empujada al remoto.
---

# Committer Flow

Flujo de tres fases: Fase 1 genera el commit y captura la rama destino; Fase 2 ejecuta el push; Fase 3 crea o reutiliza un PR estructurado.

## Reglas duras

- **`git push --force` / `-f` / `--force-with-lease`** — prohibido bajo cualquier circunstancia.
- **Rebase, reset, amend, filter-branch** — no reescribir historia.
- **Borrar ramas remotas** — fuera de scope.
- **Modificar código, tests, configs o specs** — operación solo-lectura sobre el repo.
- **Reintentar automáticamente un commit o push fallido** — reportar al humano y detener.
- **Inferir rama destino sin preguntar** — siempre es decisión del usuario (ver Paso 1.4).
- **Escribir contenido de git en un idioma distinto del español** — mensajes de commit y, si el flujo abre un PR, su título y cuerpo van completamente en español (ver "Idioma del contenido de git").

## Idioma del contenido de git

Regla transversal a ambas fases:

- Todo contenido que queda escrito en git o en el remoto se redacta **en español, siempre**: mensaje de commit (asunto, cuerpo, footer) y título + cuerpo del PR cuando se abre uno (`gh pr create`).
- Aplica sin excepción, independientemente del idioma del historial previo del repo y del idioma de la conversación.
- Los tipos de Conventional Commits (`feat`, `fix`, `chore`, `refactor`, etc.), identificadores de código, paths, nombres de archivos, flags, comandos y nombres propios se citan verbatim, sin traducir.
- Nombres de rama: slugs ASCII — palabras en español pero sin tildes ni ñ (ej. `feat/LIN-123-gestion-pagos`).
- La interacción con el humano (preguntas, reportes, notas de handoff) sigue en español — esta regla solo cubre contenido de git.

## Inputs requeridos — Fase 1

| Campo | Esperado | Fallback si falta |
|---|---|---|
| `Phase` | `1` | Asumir `Phase: 1`. |
| `TASK-ID` | Para resolver `.handoff/<TASK-ID>.md` | OMITIR Paso 1.0. Anotar: "Corrí sin TASK-ID — gate de handoff omitido". |
| `run_id` | Ubicar handoff propio en `.project-context/runs/<run_id>/` | Usar `ad-hoc`. Anotar: "run_id ausente — handoff propio en `ad-hoc/`". |
| `ANVIL_REPO` | Ruta al repo Anvil para `verify-handoff.sh` | OMITIR Paso 1.0. Anotar: "ANVIL_REPO ausente — gate de handoff omitido". |
| `PROJECT_ROOT` | Raíz del proyecto activo | OMITIR Paso 1.0. |
| Path al handoff del developer | `.handoff/<TASK-ID>.md` | OMITIR validación cruzada. Usar `git status --porcelain`. Anotar: "Sin handoff del developer — staging basado en `git status` puro". |
| Lista de archivos modificados | Lista del humano | Caer a `git status --porcelain`. |

Los fallbacks degradan funcionalidad opcional, nunca la operación core del commit.

## Inputs requeridos — Fase 2

| Campo | Requerido | Notas |
|---|---|---|
| `Phase` | siempre | `2` (literal) |
| `TASK-ID` | siempre | Mismo TASK-ID que Fase 1 |
| `run_id` | siempre | Necesario para leer el handoff propio |
| Path al handoff propio de Fase 1 | siempre | `.project-context/runs/<run_id>/committer-handoff.md` |
| Estado de los gates posteriores | siempre | "reviewer: PASS", "qa: PASS-WITH-NOTES sin bloqueadores", etc. |

Si falta algún campo en Fase 2, preguntar al humano antes de continuar.

## Flujo — Fase 1

> Cargar la skill `handoff` antes del Paso 1.0 si `TASK-ID` está presente.

### Paso 1.0 — Gate: verificar integridad del handoff

Solo corre si `ANVIL_REPO`, `PROJECT_ROOT` y `TASK-ID` están presentes. Si falta cualquiera → saltar al Paso 1.1.

```bash
bash <ANVIL_REPO>/scripts/verify-handoff.sh <PROJECT_ROOT> <TASK-ID>
```

- Exit 0 → continuar.
- Exit ≠ 0 → DETENER. Reportar al humano con stderr completo. No continuar con git ni escribir handoff.

### Paso 1.1 — Verificar estado del repo

Ejecutar `git status --porcelain`. Validar contra lista de archivos del humano.

- Archivos modificados no listados por el humano → preguntar si incluirlos.
- Archivos listados que no aparecen modificados → reportar y continuar con los que sí están.
- Working tree limpio → preguntar al humano qué archivos incluir antes de continuar.

### Paso 1.2 — Stage de archivos

`git add` SOLO sobre archivos explícitos. Nunca `git add .` ni `git add -A`.
Verificar con `git diff --cached --stat`.

### Paso 1.3 — Commit (no-interactivo)

Asume que el staging ya está hecho. Si `git diff --cached --stat` está vacío → DETENER: "No hay cambios staged. No commiteo."

**1. Leer el diff staged:**

```bash
git diff --cached --stat
git diff --cached
```

**2. Analizar los cambios:** determinar qué cambió, por qué, impacto y scope. Ejecutar `git branch --show-current` y buscar referencia a ticket en el nombre de rama (patrones: `feat/123-…`, `fix/PROJ-456`, `JIRA-789-…`). Si existe → guardarla para el footer.

**3. Seleccionar tipo de Conventional Commit:**

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

Si el diff mezcla tipos, elegir el principal. Si es genuinamente mixto, usar `chore` con scope explícito.

**4. Redactar el mensaje:**

Formato: `<type>(<scope>): <description>`

Idioma: el mensaje completo (asunto, cuerpo y footer) se escribe **en español, siempre**, sin importar el idioma del historial del repo ni el de la conversación. Los tipos de Conventional Commits, identificadores de código, paths y flags van verbatim.

Reglas: tipo en minúsculas · scope opcional (área afectada) · descripción en imperativo, minúscula, sin punto · asunto ≤ 50 caracteres.

Cuerpo (si el cambio no es trivial): separar con una línea en blanco · máx 72 chars por línea · explicar QUÉ y POR QUÉ, no CÓMO.

Footer: `Refs <TICKET-ID>` o `Fixes <TICKET-ID>` si se detectó ticket en rama. `BREAKING CHANGE: <desc>` si aplica.

Anti-patrones — NUNCA: "fix bug", "update code", "changes", "WIP", tiempo pasado ("added"), punto final en el asunto.

**5. Ejecutar el commit:**

```bash
git commit -m "$(cat <<'EOF'
<mensaje completo>
EOF
)"
```

- Éxito → capturar `commit_hash` (`git rev-parse HEAD`) y `commit_subject` (`git log -1 --format=%s`). Continuar al Paso 1.4.
- Fallo (pre-commit hook, lint, build) → DETENER. Capturar stdout + stderr completo. Reportar al humano con el error exacto. No escribir handoff propio.

**Checklist pre-commit:**
- [ ] Hay cambios staged
- [ ] Tipo corresponde al propósito principal
- [ ] Asunto ≤ 50 caracteres, imperativo, sin punto
- [ ] Mensaje completo en español (tipos de Conventional Commits e identificadores de código verbatim)
- [ ] Cuerpo separado del asunto por línea en blanco (si aplica)
- [ ] Sin anti-patrones
- [ ] Footer con ticket si se detectó en la rama

### Paso 1.4 — Preguntar rama destino

Usar `AskUserQuestion`:

- **Question:** "¿A qué rama hago push?"
- **Header:** "Rama destino"
- **Options:** (construir dinámicamente)
  1. `<rama actual>` — `git branch --show-current`
  2. Si `delivery-state.yaml` tiene `parent_branch`, incluirla etiquetada como "rama padre del milestone `<milestone>`" y ofrecerla como primera alternativa a la rama actual
  3. Hasta 3 ramas locales adicionales — `git branch --format='%(refname:short)' --sort=-committerdate`
  4. "Otra (la escribo)"

La pregunta nunca se omite: el humano siempre decide, incluso cuando hay `parent_branch`.

Si elige "Otra" → segunda llamada `AskUserQuestion` para capturar el nombre. Validar: sin espacios ni caracteres `~^:?*[\`.

### Paso 1.5 — Escribir handoff propio

Crear `.project-context/runs/<run_id>/committer-handoff.md`:

```markdown
# Committer handoff — Fase 1 → Fase 2

- TASK-ID: <TASK-ID>
- run_id: <run_id>
- Commit hash: <git rev-parse HEAD>
- Commit subject: <primera línea del mensaje>
- Rama destino: <rama elegida>
- Remoto: <git remote get-url origin>
- Fecha Fase 1: <ISO 8601>

## Mensaje del commit (verbatim)

<git log -1 --format=%B>

## Notas

<ej: "rama destino es nueva, push creará upstream" / vacío>
```

### Paso 1.6 — Reportar

Máx 100 palabras: commit hash corto, subject, rama destino, path al handoff, notas relevantes. DETENERSE — el humano continúa con `reviewer`.

## Flujo — Fase 2

### Paso 2.1 — Leer handoff propio

`Read` sobre `.project-context/runs/<run_id>/committer-handoff.md`. Si no existe → preguntar al humano dónde está antes de continuar.

### Paso 2.2 — Verificar HEAD

`git rev-parse HEAD` vs `Commit hash` del handoff:

- Igual → continuar.
- Diferente pero ancestor (`git merge-base --is-ancestor <hash> HEAD`) → esperado si `qa-fixer` añadió commits. Continuar.
- Diferente y no ancestor → preguntar al humano si proceder con HEAD actual.

### Paso 2.3 — Verificar rama destino

`git branch --show-current` vs `Rama destino` del handoff:

- Coinciden → continuar.
- No coinciden y la rama existe → preguntar al humano si hace checkout.
- No existe localmente → rama nueva; el push la creará en el remoto. Anotar en output.

### Paso 2.4 — Verificar working tree

`git status --porcelain` debe estar vacío.

- Limpio → continuar.
- Hay cambios → DETENER. Reportar al humano: fixes de `qa-fixer` sin commitear. Pedir mini-Fase-1 antes de continuar.

### Paso 2.5 — Push (no-interactivo)

```bash
git push origin <rama_destino>
```

Sin `--force`, `--force-with-lease` ni variantes bajo ninguna circunstancia.

- Exit 0 → capturar `git rev-parse HEAD` post-push. Continuar al Paso 2.6.
- Fallo `non-fast-forward` → DETENER: "Push rechazado — rama remota tiene commits que el local no tiene. [stderr]. ¿Cómo procedo?"
- Fallo `auth` → DETENER: "Push bloqueado por fallo de autenticación. [stderr]. ¿Cómo procedo?"
- Fallo hook remoto → DETENER: "Hook remoto rechazó el push. [stderr]. ¿Cómo procedo?"
- Cualquier otro fallo → DETENER: "Push falló. [stderr]. ¿Cómo procedo?"

### Paso 2.6 — Reportar

Máx 100 palabras: confirmación de push, rama remota, commit hash post-push, notas (ej. rama nueva creada).

## Flujo — Fase 3: PR

### Inputs requeridos

| Campo | Requerido | Fallback si falta |
|---|---|---|
| `Phase` | siempre | `3` (literal) |
| URL o path de `delivery-state.yaml` | sí para entrega trazada | Si falta, crear PR solo cuando el humano lo solicita explícitamente y anotar ausencia de tracking. |
| Rama base | no | Resolver por cascada (ver Paso 3.1). |
| Evidencia de validación | sí | DETENER: no crear un PR sin resultados. |

### Paso 3.1 — Verificar precondiciones

- Confirmar que el push de Fase 2 terminó y que `git status --porcelain` está vacío.
- Ejecutar `gh auth status`; si falla, DETENER y reportar la autenticación pendiente.
- Resolver la rama base por esta cascada, en orden, deteniéndose en la primera que aplique:
  1. `parent_branch` de `delivery-state.yaml` — rama padre del milestone.
  2. `pr_target_branch` de la configuración del proyecto en `.project-context/` (config estática del repo, no un campo del estado de entrega).
  3. Default remoto (`origin/HEAD`).
  4. Preguntar al humano.
- **Guardia:** si el estado tiene `parent_branch` y la base resuelta o solicitada es otra (por ejemplo `develop`), advertir explícitamente al humano —"el trabajo pertenece al milestone `<milestone>` y debería integrarse primero en `<parent_branch>`"— y continuar solo con su confirmación.
- Nunca abrir el PR contra una rama inferida de forma ambigua.

### Paso 3.2 — Reutilizar o crear

Buscar primero un PR abierto para la rama actual:

```bash
gh pr view <branch> --json url,title,body,state
```

- Si existe, reutilizar su URL. Actualizar título/cuerpo si no cumple el formato estructurado o contiene placeholders.
- Si no existe, crear con `gh pr create --base <base> --head <branch> --title <title> --body-file <file>`.
- Nunca crear un segundo PR para la misma rama.

### Paso 3.3 — Redactar el PR

Todo título y cuerpo del PR se escribe en español. El título sigue Conventional Commits e incluye el `TASK-ID` cuando exista. El cuerpo DEBE tener contenido específico derivado del diff y de las validaciones:

```markdown
## Contexto
<Por qué se necesita el cambio.>

## Cambios
- <Cambio concreto y observable>

## Validación
- `<comando>` — <resultado pass/fail>

## Riesgo y rollback
<Nivel de riesgo y acción concreta de revert/rollback.>

## Trazabilidad
- Linear: <URL del issue o motivo explícito de no-tracking>
- Documentación: <URL o reporter delta-only>
```

Prohibido: cuerpo de una línea, `TBD`, `N/A` sin motivo, o copiar el mensaje del commit como descripción completa.

### Paso 3.4 — Persistir y reportar

Guardar `pr_url` en el estado de entrega si existe. Reportar URL, rama base, task ID y si el PR se creó o reutilizó.

## Manejo de errores

Para cualquier fallo: capturar output textual completo → DETENER → reportar al humano con comando ejecutado, código de salida, output y paso donde ocurrió. Sin reintentos automáticos.

## Auto-QA

### Fase 1
- [ ] `verify-handoff.sh` devolvió exit 0 (o se omitió por fallback documentado)
- [ ] Commit hash capturado y válido
- [ ] Mensaje del commit redactado íntegramente en español
- [ ] Commit hash registrado en `committer-handoff.md`
- [ ] Rama destino registrada (no vacía)
- [ ] Handoff propio existe en `.project-context/runs/<run_id>/committer-handoff.md`

### Fase 2
- [ ] `git push` devolvió código 0
- [ ] Commit hash del handoff es ancestor (o igual) de HEAD post-push

### Fase 3
- [ ] `gh auth status` devolvió código 0
- [ ] Existe exactamente un PR abierto o creado para la rama
- [ ] Si el estado tiene `parent_branch`, la base del PR es esa rama (o el humano confirmó otra tras la advertencia)
- [ ] El PR tiene Context, Changes, Validation, Risk and rollback y Tracking
- [ ] URL del PR persistida en el estado de entrega cuando aplica

Si algún check falla → reportar al humano antes de cerrar.

## Presupuesto

- **Fase 1:** Objetivo 4K | Máximo 8K | Máx tool calls: 12
- **Fase 2:** Objetivo 3K | Máximo 6K | Máx tool calls: 8
- **Fase 3:** Objetivo 3K | Máximo 6K | Máx tool calls: 8
