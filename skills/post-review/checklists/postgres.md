# PostgreSQL Review Checklist

## Correctness

- [ ] Migraciones son idempotentes (IF NOT EXISTS, IF EXISTS)
- [ ] No hay ALTER TABLE que bloquee la tabla en produccion (agregar columna con DEFAULT en PG < 11)
- [ ] Foreign keys tienen ON DELETE definido (CASCADE, SET NULL, o RESTRICT segun el caso)
- [ ] No hay columnas NOT NULL nuevas sin DEFAULT en tablas con datos existentes
- [ ] Tipos de datos correctos (no usar VARCHAR para UUIDs, no TEXT para enums)
- [ ] Migraciones tienen rollback (down migration)
- [ ] Transacciones usadas correctamente en migraciones multi-statement

## Security

- [ ] No hay queries con string concatenation (SQL injection)
- [ ] Roles de base de datos siguen least privilege
- [ ] No hay passwords en archivos de migracion
- [ ] Connection strings usan SSL (`sslmode=require` o `verify-full`)
- [ ] No hay GRANT ALL a usuarios de aplicacion
- [ ] RLS (Row Level Security) donde hay multi-tenancy

## Performance

- [ ] Indices creados para columnas usadas en WHERE, JOIN, ORDER BY frecuentes
- [ ] Indices creados con CONCURRENTLY para no bloquear en produccion
- [ ] No hay indices duplicados o redundantes
- [ ] No hay SELECT * en queries de produccion (seleccionar columnas especificas)
- [ ] JOINs tienen indices en ambos lados de la relacion
- [ ] No hay funciones en WHERE que impidan uso de indices (`WHERE LOWER(email)` sin indice funcional)
- [ ] EXPLAIN ANALYZE verificado en queries complejas
- [ ] Paginacion con cursor (keyset), no OFFSET en tablas grandes

## Conventions

- [ ] Tablas en plural, snake_case (`user_accounts`, no `UserAccount`)
- [ ] Columnas en snake_case
- [ ] Primary keys named `id` o `{table}_id`
- [ ] Timestamps: `created_at`, `updated_at` presentes
- [ ] Indices con nombre descriptivo, no generado (`idx_users_email`, no `idx_12345`)
- [ ] Enums como tipos de PostgreSQL o lookup tables, no magic strings
- [ ] Comentarios en columnas/tablas con logica de negocio no obvia

## Data Integrity

- [ ] Constraints CHECK donde aplique (rangos, formatos)
- [ ] UNIQUE constraints en campos que deben ser unicos
- [ ] No hay columnas nullable sin razon justificada
- [ ] Soft delete implementado consistentemente si es patron del proyecto
- [ ] Audit trail (quien, cuando, que cambio) en tablas criticas
