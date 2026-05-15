# Dominio: deploy

last_updated: 2026-05-08

## Responsabilidad

Gestionar el despliegue de archivos de configuración de Anvil (skills, pipelines, CLAUDE.md, registry) a los directorios de configuración de los distintos clientes AI (Claude Code, Codex, Gemini, OpenCode). También implementa runners de agentes por proveedor.

## Archivos clave

```
internal/deploy/
├── deploy.go          — ResolvePaths() para targets (Claude, OpenCode, Gemini, Codex); SnapshotItem()
├── claude.go          — runner/deployer para Claude Code
├── codex.go           — runner/deployer para Codex
├── gemini.go          — runner/deployer para Gemini
├── opencode.go        — runner/deployer para OpenCode
├── cursor.go          — runner/deployer para Cursor
├── ownership.go       — gestión de ownership de archivos desplegados
├── summary.go         — resumen de despliegue
└── integration_test.go — tests de integración (630 líneas)
```

## Flujo principal

```
anvil deploy / anvil targets →
→ ResolvePaths() resuelve paths por target según env vars y home dir
→ SnapshotItem() copia archivos y symlinks al directorio de configuración del AI client
→ registro de ownership para tracking
```

## Dependencias de este dominio

- `pkg/config` — configuración de targets
- `pkg/fileutil` — CopyDir, operaciones de archivo
- `pkg/output` — mensajes de output formateados

## Quién depende de este dominio

- `internal/cli/targets.go` — subcomandos de gestión de targets
- `internal/cli/run.go` — runners de agentes por proveedor

## Variables de entorno relevantes

- `CLAUDE_HOME` — override del home de Claude Code (default: `~/.claude`)

## Gotchas

- `ResolvePaths()` devuelve paths hardcoded relativos a `$HOME` para todos los providers excepto Claude (que respeta `CLAUDE_HOME`) — si un provider cambia su config dir, el código debe actualizarse manualmente
- Los tests de integración (`integration_test.go`, 630 líneas) requieren el sistema de archivos real — no usan mocks
