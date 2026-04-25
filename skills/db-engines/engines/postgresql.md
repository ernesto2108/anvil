# PostgreSQL

## Tipos y Convenciones
- Usa `UUID` con `uuid_generate_v7()` o `gen_random_uuid()` para primary keys
- Usa `TIMESTAMP WITH TIME ZONE` (no `DATETIME`, no `TIMESTAMP` sin TZ)
- Usa tipos `ENUM` o constraints `CHECK` para columnas de status
- Prefiere `TEXT` sobre `VARCHAR(n)` a menos que necesites el constraint de longitud por razones de negocio
- `BOOLEAN` es nativo — úsalo directamente

## Patrones de Migración
- `CREATE INDEX CONCURRENTLY` para índices en tablas grandes (evita bloqueo de tabla)
- `IF NOT EXISTS` / `IF EXISTS` para migraciones idempotentes
- `ALTER TABLE ... ADD COLUMN ... DEFAULT <value>` en una sola sentencia (Postgres 11+ no reescribe filas)
- Transacciones explícitas `BEGIN` / `COMMIT` están permitidas y son recomendadas para migraciones multi-sentencia

## Ajuste de Rendimiento
- Connection pooling: usa `pgxpool` o pooler externo (PgBouncer) — nunca abras conexiones ilimitadas
- `EXPLAIN ANALYZE` para verificar el uso de índices antes y después de cambios
- Índices parciales: `CREATE INDEX ... WHERE deleted_at IS NULL` para tablas con soft-delete
- Índices de cobertura: `INCLUDE (col)` para evitar búsqueda en heap
- `GROUPING SETS` para múltiples agregaciones en un solo paso

## Multi-Tenant
- Políticas de Row Level Security (RLS) para aislamiento de tenant
- `current_setting('app.tenant_id')` para contexto de RLS
- Los índices compuestos lideran con `tenant_id`: `idx_table_tenant_status`

## Drivers de Go
- `github.com/jackc/pgx/v5` — preferido (Go puro, connection pooling via `pgxpool`)
- `github.com/lib/pq` — legacy, compatible con `database/sql`
- Traducción de errores: `pq.Error` o códigos `pgconn.PgError` a errores de dominio (23505 = violación de unique, 23503 = violación FK)
