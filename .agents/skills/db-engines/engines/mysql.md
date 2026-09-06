# MySQL

## Tipos y Convenciones
- Sin `UUID` nativo — usa `CHAR(36)` o `BINARY(16)` con conversión a nivel de aplicación
- Usa `DATETIME` no `TIMESTAMP` (TIMESTAMP tiene el problema del año 2038 y se auto-actualiza)
- Usa `VARCHAR(n)` con longitudes explícitas — MySQL las aplica
- `BOOLEAN` es alias de `TINYINT(1)` — 0/1, no true/false
- `TEXT` / `LONGTEXT` para strings largas — el máximo de `VARCHAR` es 65,535 bytes

## Patrones de Migración
- `ALTER TABLE ... ALGORITHM=INPLACE` cuando sea posible (evita copiar la tabla)
- Sin `CREATE INDEX CONCURRENTLY` — la creación de índices bloquea la tabla
- `IF NOT EXISTS` / `IF EXISTS` soportado
- `ALTER TABLE ... ADD COLUMN` soporta `AFTER <column>` para ordenar columnas
- DDL Online: para tablas grandes, considera `pt-online-schema-change` o `gh-ost`

## Ajuste de Rendimiento
- InnoDB es el motor por defecto y correcto — nunca usar MyISAM
- `innodb_buffer_pool_size` = 70-80% de la RAM disponible para servidores DB dedicados
- `EXPLAIN FORMAT=JSON` para análisis detallado de consultas
- Índices de cobertura con `INCLUDE` no soportados — pon todas las columnas necesarias en el índice mismo
- `FORCE INDEX(idx_name)` como último recurso para selección de índice incorrecta

## Multi-Tenant
- Sin RLS nativo — aplica aislamiento de tenant a nivel de aplicación/consulta
- Índices compuestos: `(tenant_id, ...)` para todas las consultas con scope de tenant
- Considera schema-por-tenant para aislamiento fuerte (agrega complejidad operativa)

## Drivers de Go
- `github.com/go-sql-driver/mysql` — driver estándar `database/sql`
- Connection string: `user:pass@tcp(host:port)/dbname?parseTime=true&multiStatements=true`
- `parseTime=true` es OBLIGATORIO — sin él, las columnas `DATETIME` devuelven `[]byte` en lugar de `time.Time`
- Traducción de errores: `*mysql.MySQLError` con números de error (1062 = duplicado, 1452 = violación FK)
