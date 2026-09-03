# Referencia — Codex CLI

Fuente: `learn.chatgpt.com/docs` (ex `developers.openai.com/codex`), validado agosto 2026.

## Rutas de destino

| Artefacto fuente | Destino Codex |
|---|---|
| `CLAUDE.md` | `AGENTS.md` (raíz del proyecto) |
| `agents/<name>.md` | `.codex/agents/<name>.toml` — **TOML, no markdown** |
| `skills/<name>/SKILL.md` | `.agents/skills/<name>/SKILL.md` |
| `commands/<name>.md` | `.agents/skills/<name>/SKILL.md` (se convierten a skills) |

Ámbito global equivalente: `~/.codex/AGENTS.md`, `~/.codex/agents/`, `~/.agents/skills/`. Esta receta escribe solo en el proyecto.

## Instrucciones

- Codex lee `AGENTS.md` global (`~/.codex/AGENTS.md`), de proyecto (raíz) y de subdirectorios.
- El merge es por **concatenación raíz → hoja**.
- ⚠️ **Límite default: 32 KiB combinados.** Antes de escribir, medir el tamaño del `AGENTS.md` generado. Si se acerca o supera el límite → DETENER y reportar: el contenido excedente se trunca silenciosamente y las reglas del final se pierden.

## Agentes — formato TOML

Codex NO usa markdown para agentes. El body markdown del agente fuente va dentro del campo `developer_instructions` como string triple-quoted.

**Campos requeridos:** `name`, `description`, `developer_instructions`.

| Campo fuente | Campo Codex | Regla |
|---|---|---|
| `name` | `name` | Copia literal |
| `description` | `description` | Copia literal; escapar comillas |
| body | `developer_instructions` | Body markdown completo dentro de `"""…"""` |
| `permissionMode` | `sandbox_mode` | `read` → `read-only`; `write` y `execute` → `workspace-write` |
| `skills` | — | Eliminar; mover al inicio de `developer_instructions` (Paso 5) |

> **Sin campo `model`.** Los agentes fuente no declaran modelo. **No emitir `model` ni `model_reasoning_effort`** en el TOML: Codex usa el modelo y el esfuerzo de razonamiento de su propia configuración. Si un agente fuente conserva una línea `model:` residual, omitirla sin traducir.

Opcional: `mcp_servers` — solo si el usuario lo pide explícitamente.

> **No existe allowlist granular de tools en Codex.** El único control es `sandbox_mode`. Declararlo como pérdida en el reporte.

### Ejemplo de agente exportado

```toml
# GENERADO por la skill export-system. NO EDITAR A MANO.
# Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.

name = "explorer"
description = "Agente de exploración e investigación. Lee código y docs locales..."
sandbox_mode = "read-only"

developer_instructions = """
## Skills requeridas

Antes de trabajar, carga estas skills y sigue sus instrucciones: `read-files`.

# Agente — Explorer

## Rol
...
"""
```

**Cuidado con el escapado:** si el body contiene la secuencia `"""`, escaparla o reemplazarla antes de emitir el TOML. Si el body contiene `\` en contextos que TOML interpretaría, usar el literal multilínea `'''…'''` en su lugar.

## Skills

- Descubrimiento: `.agents/skills/` (cwd, directorios padres y raíz del repo) o `~/.agents/skills/`.
- Formato idéntico al estándar Agent Skills: **copia literal** del `SKILL.md`.
- Invocación: `$nombre-skill`, o implícita cuando la `description` matchea el contexto.

## Commands → skills

Los custom prompts de Codex están **deprecados oficialmente**. Los commands se convierten a skills:

1. Crear `.agents/skills/<nombre-command>/SKILL.md`.
2. Frontmatter: `name` (slug del command, `:` → `-`), `description` (la del command).
3. Body: el body del command.
4. Invocable como `$nombre-command`.

## Pérdidas conocidas

- Sin allowlist granular de tools: solo `sandbox_mode` (`read-only` / `workspace-write`).
- `write` y `execute` colapsan al mismo `sandbox_mode`: un agente de solo escritura obtiene también capacidad de shell.
- Riesgo de truncado del `AGENTS.md` por el límite de 32 KiB.
- Los namespaces de commands se aplanan (`git:commit` → `git-commit`).
