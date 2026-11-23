# Infra Services

Generado por: context-init
Última actualización: 2026-08-13

## Servicios detectados

### sqlite
- **Detectado via:** `go.mod` (`mattn/go-sqlite3`, `asg017/sqlite-vec-go-bindings`), `pkg/storage/sqlite.go`
- **MCP disponible:** false (no hay MCP de SQLite configurado en esta sesión — ToolSearch sin resultado)
- **Fallback:** filesystem — la base de datos vive en un archivo `.db` local, no requiere MCP para confirmar su existencia
- **Ambiente:** dev
- **Notas:** requiere build tag `fts5` para funcionalidad completa (FTS5 + sqlite-vec). No se detectó un archivo `.db` fijo commiteado — se genera en runtime.

### docker (solo CI de migraciones)
- **Detectado via:** `docker-compose.test.yml`, `docker/test-migrations.dockerfile`
- **MCP disponible:** N/A — no es una infra de runtime del producto, es un contenedor efímero para validar migraciones en CI
- **Notas:** no aplica sondeo de health-check — no corre en dev ni producción, solo en pipeline de CI.

## Sin servicios adicionales

No se detectaron Postgres, MySQL, Redis, Kafka, Elasticsearch ni MongoDB declarados en `go.mod`, `docker-compose.*` ni `.env*` — no se encontró ningún archivo `.env*` en la raíz del repo. anvil es un monolito CLI que solo depende de SQLite local y de la API externa de Anthropic (ver `Technical domain/dependencies.md`), la cual no es infraestructura sondeable vía `infra-probe` (no es DB/broker).

## Sugerencia

No se sugiere `mcp-setup` para SQLite en este momento — el patrón de acceso es archivo local, no requiere connection string ni MCP server dedicado para las operaciones actuales de este repo.
