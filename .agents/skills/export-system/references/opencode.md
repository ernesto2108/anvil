# Referencia — OpenCode

Fuente: `opencode.ai/docs` (validado agosto 2026).

## Rutas de destino

| Artefacto fuente | Destino OpenCode |
|---|---|
| `CLAUDE.md` | `AGENTS.md` (raíz del proyecto) |
| `agents/<name>.md` | `.opencode/agents/<name>.md` |
| `skills/<name>/SKILL.md` | sin copia — OpenCode ya lee `.claude/skills/` |
| `commands/<name>.md` | `.opencode/commands/<name>.md` |

## Instrucciones

- OpenCode lee `AGENTS.md` en la raíz.
- Lee `CLAUDE.md` **solo como fallback oficial** si no existe `AGENTS.md`.
- ⚠️ **Consecuencia:** en cuanto se genera `AGENTS.md`, ese archivo **anula** el `CLAUDE.md` existente. Por eso el export debe incluir el contenido **completo** de `CLAUDE.md`, no un resumen ni un puntero.

## Skills

OpenCode descubre skills en: `.claude/skills/`, `~/.claude/skills/`, `.opencode/skills/` y `.agents/skills/`.

→ **No copiar skills para este destino.** Reportar en la salida: "skills leídas in-situ desde `.claude/skills/`".

Validar de todos modos el `name` de cada skill contra `^[a-z0-9]+(-[a-z0-9]+)*$`: OpenCode rechaza las que no cumplen. Reportar las inválidas.

## Agentes — mapeo de frontmatter

El **filename es el ID** del agente. No emitir campo `name`.

| Campo fuente | Campo OpenCode | Regla |
|---|---|---|
| `name` | — | Se omite; pasa a ser el nombre del archivo |
| `description` | `description` | Copia literal (obligatorio en OpenCode) |
| — | `mode` | Siempre `subagent` |
| `model` | — (se omite) | **No emitir `model`.** El agente hereda el modelo de la sesión activa de OpenCode. Anotar en el reporte el tier original de cada agente (`sonnet`/`opus`) para fijado manual opcional |
| `permissionMode` | `permission` | Ver tabla de permisos abajo |
| `skills` | — | Eliminar; mover al body (Paso 5 del SKILL.md) |
| body | body | Copia literal = system prompt |

Campos opcionales que **no** se emiten salvo que el usuario los pida: `temperature`, `top_p`, `steps`.

> El campo `tools:` booleano existe pero está **deprecado**. Usar siempre `permission`.

> **Por qué se omite `model`:** OpenCode exige un `provider/model-id` resoluble contra los providers **autenticados en la máquina destino**. Emitir `anthropic/...` rompe el subagente (falla antes de ejecutar) si el usuario tiene autenticado otro provider. El exportador no puede conocer esa configuración, y el tier de origen es una preferencia de capacidad, no un binding a provider.

## Tabla de permisos

Los nombres de tool van en minúsculas. Valores: `allow` | `deny` | `ask`.

| `permissionMode` | `permission` emitido |
|---|---|
| `read` | `edit: deny`, `write: deny`, `bash: deny` |
| `write` | `edit: allow`, `write: allow`, `bash: deny` |
| `execute` | `edit: allow`, `write: allow`, `bash: allow` |

Ejemplo de agente exportado:

```yaml
---
description: Agente de exploración e investigación. Lee código y docs locales...
mode: subagent
permission:
  edit: deny
  write: deny
  bash: deny
---
```

## Commands

`commands/<name>.md` → `.opencode/commands/<name>.md`.

| Campo fuente | Campo OpenCode |
|---|---|
| `name` | — (el filename es el nombre) |
| `description` | `description` |
| `tools` | — (sin equivalente; se omite) |
| — | `agent` — solo si el command delega a un agente concreto |
| — | `model` — omitir salvo que el usuario lo pida |
| — | `subtask` — `true` si el command debe correr aislado |

El body es un template: los placeholders de argumentos se escriben como `$ARGUMENTS`.

Los commands anidados (`commands/git/commit.md`, nombre `git:commit`) se aplanan a `.opencode/commands/git-commit.md`, y se anota el renombrado en el reporte.

## Pérdidas conocidas

- El tier de modelo (`sonnet`/`opus`) no se exporta: los agentes heredan el modelo de la sesión de OpenCode. El reporte lista el tier original por agente para fijado manual opcional (solo si el destino tiene autenticado un provider con modelos equivalentes).
- El campo `tools` de los commands se pierde.
- La jerarquía de namespaces de commands (`git:commit`) se aplana.
