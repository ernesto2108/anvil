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
| `model` | `model` | Formato `provider/model-id`. `low|medium|high` NO son válidos. Default sugerido: `anthropic/claude-sonnet-4-5`. Anotar en el reporte que requiere revisión manual |
| `permissionMode` | `permission` | Ver tabla de permisos abajo |
| `skills` | — | Eliminar; mover al body (Paso 5 del SKILL.md) |
| body | body | Copia literal = system prompt |

Campos opcionales que **no** se emiten salvo que el usuario los pida: `temperature`, `top_p`, `steps`.

> El campo `tools:` booleano existe pero está **deprecado**. Usar siempre `permission`.

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
model: anthropic/claude-sonnet-4-5
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

- El tier de modelo (`low|medium|high`) no se traduce: requiere revisión manual.
- El campo `tools` de los commands se pierde.
- La jerarquía de namespaces de commands (`git:commit`) se aplana.
