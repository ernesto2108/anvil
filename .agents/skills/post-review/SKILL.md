---
name: post-review
description: "Skill de revisión post-desarrollo con checklists específicos por stack para Go, React, React Native, Terraform y PostgreSQL. Úsalo cuando el reviewer o qa terminen una revisión de código y se necesite el checklist post-revisión por stack (Go, React, React Native, Terraform, PostgreSQL)."
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->


# Post-Review Skill — Dispatcher

## Propósito

Proporcionar checklists de revisión específicos por stack para el agente Reviewer. Este archivo enruta al checklist correcto según el stack detectado.

## Tabla de Enrutamiento

| Stack detectado | Checklist a cargar |
|---|---|
| Go | `skills/post-review/checklists/go.md` |
| React (JS/TS) | `skills/post-review/checklists/react.md` |
| React Native | `skills/post-review/checklists/react-native.md` |
| Terraform | `skills/post-review/checklists/terraform.md` |
| PostgreSQL | `skills/post-review/checklists/postgres.md` |
| Servidor MCP (por propósito) | `skills/mcp-dev/anti-patterns.md` (checklist fuente; para HTTP/auth también `skills/mcp-dev/security-and-auth.md`) |
| Integración LLM (por propósito) | `skills/ai-engineering/anti-patterns.md` (checklist fuente) |

Las dos últimas ramas **referencian** las tablas de anti-patrones que ya viven en las skills `mcp-dev` y `ai-engineering` — son la fuente única del checklist, no se copian aquí. Cargar el archivo indicado y aplicar su tabla (`error`/`warning` siempre; `suggestion` solo en modo improve/refactor).

### Detección de la rama IA/MCP (por PROPÓSITO, no por extensión)

Los stacks IA/MCP se detectan por propósito, con los MISMOS marcadores que la inferencia de `task-writer`:
- **Servidor MCP** — path bajo `mcp-server/` o `servers/*/`, o keywords `@modelcontextprotocol/sdk`, `FastMCP`, `mcp.server`, SDK Go `modelcontextprotocol/go-sdk`, `registerTool`/`mcp.tool`/`AddTool`.
- **Integración LLM** (CUALQUIER proveedor: Claude, OpenAI-compatible, Ollama/local) — keywords `anthropic`, `claude-agent-sdk`, `openai`, `ollama`, `llama.cpp`, `vllm`, `openai-compatible`, `messages.create`, `/api/chat`, `output_config`, structured outputs, prompts como artefactos, evals/RAG/embeddings, o llamadas a cualquier endpoint LLM local o remoto.
- **Excepción:** `.mcp.json` / `.mcp.json.example` es consumo de MCPs de infra, NO construcción → no dispara esta rama.

Estas ramas son **aditivas**: un mismo archivo `.py`/`.ts`/`.go` recibe su checklist de stack base (Go/React) Y el checklist IA/MCP cuando el propósito coincide.

## Revisiones Multi-Stack

Cuando un diff contiene archivos de múltiples stacks, cargar TODOS los checklists relevantes. Aplicar cada checklist únicamente a sus archivos correspondientes.

## Tabla de Detección de Lint

| Stack | Archivos de configuración | Linter | Instalación |
|---|---|---|---|
| Go | `.golangci.yml`, `.golangci.yaml` | golangci-lint | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |
| React/TS | `.eslintrc.*`, `eslint.config.*`, `eslint` en package.json | ESLint | `npm install -D eslint` / `pnpm add -D eslint` |
| React Native | Igual que React | ESLint | Igual que React |
| Terraform | `.tflint.hcl` | TFLint | `brew install tflint` / `curl -s https://raw.githubusercontent.com/terraform-linters/tflint/master/install_linux.sh \| bash` |
| PostgreSQL | N/A | N/A | — |

## Archivos de Soporte

| Archivo | Propósito |
|---|---|
| `rubric.md` | Criterios de puntuación universales (escala 1-10) |
| `report-format.md` | Especificación del formato de salida en consola |

## Uso

1. El agente Reviewer detecta los stacks a partir de las extensiones de archivo
2. Verificar que existe configuración de lint por stack (ver Tabla de Detección de Lint)
3. Ejecutar el linter si existe configuración; señalar ausencia como CRÍTICO si no
4. Cargar los checklists correspondientes de la tabla de enrutamiento
5. Cargar `rubric.md` y `report-format.md`
6. Ejecutar la revisión según el checklist
7. Agregar puntuación (incluyendo hallazgos de lint) e imprimir el reporte
