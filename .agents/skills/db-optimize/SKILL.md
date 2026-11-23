---
name: db-optimize
description: Identifica queries lentas y sugiere optimizaciones de schema, índices o queries. Usar cuando el usuario diga "slow query", "optimize SQL", "add index", "query performance", "EXPLAIN", o al investigar cuellos de botella en la base de datos.
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->


# Optimización de Base de Datos

> Analiza el rendimiento de queries y sugiere mejoras de schema, índices o queries.

## Prerequisito

Ejecutar `/db-schema-scan` primero (o recibir contexto del schema en línea del orquestador) para entender las tablas, índices y relaciones actuales.

## Cuándo Usar

- El usuario reporta queries lentas o timeouts
- Después de agregar una nueva query a un repositorio
- Durante la revisión QA de features con mucha carga de datos
- Auditoría proactiva de patrones de queries

## Workflow

### Paso 1 — Identificar Queries Objetivo

Encontrar queries para analizar:
- Leer archivos de repositorio (`queries/`, `*_psql.go`, `*_repository.go`)
- Buscar: full table scans, cláusulas WHERE faltantes, JOINs sin índices, patrones N+1
- Si el usuario señaló una query específica, comenzar por ahí

### Paso 2 — Analizar Cada Query

Para cada query, evaluar:

| Verificación | Qué buscar | Corrección |
|-------|-----------------|-----|
| **Índice faltante en WHERE/JOIN** | Columna en WHERE o JOIN ON sin índice | Crear índice |
| **Full table scan** | SELECT sin WHERE en tabla grande | Agregar filtrado o paginación |
| **Queries N+1** | Loop llamando query de fila única N veces | Query en batch con cláusula IN o JOIN |
| **SELECT *** | Obtener todas las columnas cuando solo se necesitan 2-3 | Listar columnas específicas |
| **LIMIT faltante** | Queries de lista sin paginación | Agregar LIMIT + OFFSET o cursor |
| **ORDER BY sin índice** | Ordenar en columna sin índice | Agregar índice o índice compuesto |
| **Cast de tipo implícito** | WHERE varchar_col = 123 (sin comillas) | Corregir tipo para evitar cast |
| **OR en WHERE** | Múltiples condiciones OR impiden uso de índice | UNION o reestructurar |
| **COUNT(*)** en tabla grande | Full scan para contar | Conteo aproximado o contador cacheado |
| **Aislamiento de tenant** | Query multi-tenant sin filtro tenant_id | Agregar tenant_id al WHERE + índice compuesto |

### Paso 3 — Recomendaciones de Índices

Al sugerir índices:

```markdown
## Recomendaciones de Índices

| Tabla | Índice Sugerido | Columnas | Tipo | Justificación |
|-------|----------------|---------|------|-----------|
| instances | idx_instances_tenant_status | (tenant_id, status) | btree | La query filtra por tenant + status, actualmente full scan |
| events | idx_events_instance_created | (instance_id, created_at DESC) | btree | La query de timeline ordena por fecha por instancia |

### Trade-offs de Índices
- Cada índice ralentiza INSERT/UPDATE ~5-10%
- Solo agregar índices para queries que se ejecutan frecuentemente
- Índices compuestos: poner columna de alta cardinalidad primero (excepto tenant_id para aislamiento)
- Índices parciales para subconjuntos filtrados: `WHERE deleted_at IS NULL`
```

### Paso 4 — Sugerencias de Reescritura de Queries

Si la query misma puede mejorarse:

```markdown
## Reescrituras de Queries

### Antes (patrón N+1)
for each workflow_id:
  SELECT * FROM instances WHERE workflow_id = $1

### Después (batch)
SELECT * FROM instances WHERE workflow_id = ANY($1::uuid[])

### Impacto: N queries → 1 query
```

## Optimizaciones Específicas de PostgreSQL

| Técnica | Cuándo | Cómo |
|-----------|------|-----|
| `EXPLAIN ANALYZE` | Verificar uso de índices | Prefixar la query con EXPLAIN ANALYZE |
| `CREATE INDEX CONCURRENTLY` | Tablas grandes en producción | Evita bloqueo de tabla |
| Índice parcial | Columna tiene muchos NULLs que se filtran | `CREATE INDEX ... WHERE deleted_at IS NULL` |
| Covering index | Evitar heap lookup | `INCLUDE (col)` en el índice |
| `GROUPING SETS` | Múltiples agregaciones en un paso | Reemplazar múltiples queries GROUP BY |
| Ajuste del pool de conexiones | Errores de agotamiento del pool | `max_open_conns`, `max_idle_conns`, `conn_max_lifetime` |

## Salida

```markdown
## Reporte de Rendimiento de Queries — <project>

### Queries Analizadas: <N>
### Problemas Encontrados: <N>

### Críticos (corregir ahora)
1. **[file:line]** — <descripción> → <corrección>

### Recomendados (mejorar rendimiento)
1. **[file:line]** — <descripción> → <corrección>

### Cambios de Índices
| Acción | SQL |
|--------|-----|
| Agregar | `CREATE INDEX CONCURRENTLY idx_x ON table (col1, col2)` |
| Eliminar (sin uso) | `DROP INDEX idx_y` |

### Impacto Estimado
- Query X: ~500ms → ~5ms (índice en columnas WHERE)
- Query Y: N+1 eliminado, ~N*10ms → ~15ms
```

## Reglas

- **Solo lectura** — sugerir cambios, no ejecutarlos. El agente DBA crea las migraciones
- **Medir antes de optimizar** — no agregar índices especulativamente. Identificar la query lenta real primero
- **Índices compuestos > múltiples índices individuales** — uno (tenant_id, status, created_at) supera a tres índices separados
- **No sobre-indexar** — cada índice tiene un costo en rendimiento de escritura. Solo indexar lo que se consulta frecuentemente
