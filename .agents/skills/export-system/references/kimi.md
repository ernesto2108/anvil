# Referencia — Kimi Code

Fuente: `github.com/MoonshotAI/kimi-code`, `docs/en/` (validado agosto 2026).

## Rutas de destino

| Artefacto fuente | Destino Kimi Code |
|---|---|
| `CLAUDE.md` | `AGENTS.md` (raíz) — alternativa: `.kimi-code/AGENTS.md` |
| `agents/<name>.md` | `.agents/agents/<name>.md` (preferido) o `.kimi-code/agents/<name>.md` |
| `skills/<name>/SKILL.md` | `.agents/skills/<name>/SKILL.md` |
| `commands/<name>.md` | `.agents/skills/<name>/SKILL.md` (se convierten a skills) |

Preferir `.agents/` sobre `.kimi-code/`: es la convención cross-tool que también leen Codex y OpenCode, y evita duplicar archivos.

## Instrucciones

- Kimi Code lee `AGENTS.md` en la raíz (o `.kimi-code/AGENTS.md`).
- Global: `~/.kimi-code/AGENTS.md` o `~/.agents/AGENTS.md`.
- **No lee `CLAUDE.md`.** La generación de `AGENTS.md` es obligatoria para este destino.

## Agentes — mapeo de frontmatter

Kimi Code declara **compatibilidad explícita** con archivos de agente estilo Claude Code. La conversión es casi literal.

| Campo fuente | Campo Kimi | Regla |
|---|---|---|
| `name` | `name` | Copia literal |
| `description` | `description` | Copia literal — es el **único campo obligatorio** |
| `permissionMode` | `tools` / `disallowedTools` | Ver tabla de permisos abajo |
| `model` | — | `model:` de Claude se **ignora**. Opcionalmente emitir `model_preference: primary` (default) o `secondary` para agentes de baja complejidad |
| `skills` | — | Se **ignora silenciosamente**. Eliminar y mover al body (Paso 5) |
| body | body | Copia literal = system prompt |

Campos opcionales soportados que se pueden emitir si aportan: `whenToUse` (cuándo invocar el agente), `subagents` (lista de agentes que puede invocar).

`tools` acepta lista YAML o string separado por comas. Usar lista YAML.

> ⚠️ **Nunca emitir `override: true`.** Ese campo reemplaza el system prompt principal de Kimi Code; un agente exportado jamás debe hacerlo.

## Tabla de permisos

| `permissionMode` | `tools` | `disallowedTools` |
|---|---|---|
| `read` | `[Read, Glob, Grep]` | `[Write, Edit, Bash]` |
| `write` | `[Read, Glob, Grep, Write, Edit]` | `[Bash]` |
| `execute` | omitir (acceso completo) | omitir |

Ejemplo de agente exportado:

```yaml
---
name: explorer
description: Agente de exploración e investigación. Lee código y docs locales...
tools:
  - Read
  - Glob
  - Grep
disallowedTools:
  - Write
  - Edit
  - Bash
model_preference: primary
---
```

## Skills

- Formato idéntico al estándar Agent Skills: copia literal del `SKILL.md`.
- `name` + `description` obligatorios; `name` debe coincidir con el directorio contenedor.
- Cada skill se registra **automáticamente como slash command**: `/skill:<nombre>`.
- Soporta `arguments` en el frontmatter y `$ARGUMENTS` en el body.

## Commands → skills

Kimi Code no tiene directorio de commands. Como cada skill ya es un slash command, la conversión natural es:

1. Crear `.agents/skills/<nombre-command>/SKILL.md`.
2. Frontmatter: `name` (slug del command, con `:` reemplazado por `-`), `description` (la del command).
3. Body: el body del command, con los argumentos como `$ARGUMENTS`.
4. Si el command delegaba a una skill ("Cargar la skill X"), conservar esa instrucción en el body.

El command `git:commit` → skill `git-commit` → invocable como `/skill:git-commit`.

## Pérdidas conocidas

- El tier de modelo (`low|medium|high`) no existe; solo `model_preference: primary|secondary`.
- Los namespaces de commands se aplanan (`git:commit` → `git-commit`).
- El campo `tools` de los commands se pierde.
