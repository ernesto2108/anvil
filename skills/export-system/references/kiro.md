# Referencia — Kiro

Fuente: `kiro.dev/docs` (validado agosto 2026).

## Rutas de destino

| Artefacto fuente | Destino Kiro |
|---|---|
| `CLAUDE.md` | `AGENTS.md` (raíz, tratado como always-included) |
| `agents/<name>.md` | `.kiro/agents/<name>.md` |
| `skills/<name>/SKILL.md` | `.kiro/skills/<name>/SKILL.md` |
| `commands/<name>.md` | `.kiro/steering/<name>.md` con `inclusion: manual` |

## Instrucciones

- Kiro soporta `AGENTS.md` en la raíz y lo trata como **always-included**.
- **No lee `CLAUDE.md`.** Generar `AGENTS.md` es obligatorio para este destino.
- Alternativa/complemento: steering files en `.kiro/steering/*.md` con frontmatter `inclusion`:

| `inclusion` | Campos extra | Comportamiento |
|---|---|---|
| `always` | — | Siempre en contexto |
| `fileMatch` | `fileMatchPattern` | Se carga al tocar archivos que matchean |
| `manual` | — | Se invoca con `#nombre` |
| `auto` | `name`, `description` | El modelo decide cuándo cargarlo |

> ⚠️ Bug conocido: `fileMatch` **no funciona** en steering global. Usar solo steering de proyecto para `fileMatch`.

## Agentes — mapeo de frontmatter

Markdown con frontmatter YAML; el body es el system prompt.

| Campo fuente | Campo Kiro | Regla |
|---|---|---|
| `name` | `name` | Copia literal |
| `description` | `description` | Copia literal |
| `permissionMode` | `tools` / `excludedTools` | Vocabulario propio — ver tabla abajo |
| `skills` | — | Eliminar; mover al body (Paso 5) |
| body | body | Copia literal = system prompt |

Campos opcionales que **no** se emiten salvo petición explícita: `resources`, `permissions` (reglas match/effect), `includeMcpJson`, `includePowers`.

> **Sin campo `model`.** Los agentes fuente no lo declaran: **no emitir `model`**; Kiro usa su modelo por defecto. Una línea `model:` residual en la fuente se omite sin traducir.

## Vocabulario de tools (propio de Kiro)

Valores válidos: `read`, `write`, `shell`, `subagent`.

| `permissionMode` | `tools` | `excludedTools` |
|---|---|---|
| `read` | `[read]` | `[write, shell]` |
| `write` | `[read, write]` | `[shell]` |
| `execute` | `[read, write, shell]` | — |

Agregar `subagent` a `tools` **solo** si el agente orquesta a otros agentes (su body invoca subagentes). Los subagentes de Kiro corren con contexto aislado.

Ejemplo de agente exportado:

```yaml
---
name: explorer
description: Agente de exploración e investigación. Lee código y docs locales...
tools:
  - read
excludedTools:
  - write
  - shell
---
```

## Skills

Kiro soporta el estándar Agent Skills y documenta explícitamente la importación de skills de la comunidad y de otras herramientas compatibles.

- **Copia literal** del `SKILL.md` a `.kiro/skills/<nombre>/SKILL.md`.
- Límites a validar antes de copiar: `name` máx **64** caracteres, `description` máx **1024** caracteres.
- Si una skill excede alguno de esos límites → omitirla y reportarla; no truncar contenido de forma silenciosa.

## Commands → steering manual

Kiro no tiene directorio de commands. El equivalente más cercano es un steering file de inclusión manual:

1. Crear `.kiro/steering/<nombre-command>.md`.
2. Frontmatter: `inclusion: manual`.
3. Body: el body del command.
4. Se invoca escribiendo `#<nombre-command>` en el chat.

Alternativa válida si el command debe dispararse por evento en lugar de a petición: un hook Manual. No generarla por defecto.

## Pérdidas conocidas

- La granularidad de tools se reduce a 4 valores (`read`/`write`/`shell`/`subagent`).
- Los commands pierden su naturaleza de comando invocable con argumentos: pasan a ser contexto inyectable con `#nombre`.
- El campo `tools` de los commands se pierde.
