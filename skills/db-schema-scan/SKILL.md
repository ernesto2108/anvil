---
name: db-schema-scan
description: Inspección de solo lectura del schema de base de datos mediante archivos de migración y SQL de schema. Usar cuando el usuario diga "check the schema", "show tables", "inspect migrations", "what columns does X have", o necesite entender la estructura de la base de datos antes de escribir queries.
---

# Escaneo de Schema de Base de Datos

> Inspección de solo lectura del schema de base de datos actual. Prerequisito para el trabajo DBA y la optimización de queries.

## Cuándo Usar

- Antes de cualquier trabajo de migración DBA (entender el estado actual)
- Antes de que el developer escriba queries de repositorio (verificar nombres de tabla/columna)
- Cuando el usuario pregunta "qué columnas tiene X", "muéstrame el schema", "verifica las tablas"
- Como prerequisito para `/db-optimize`

## Workflow

### Paso 1 — Encontrar Archivos de Migración

Buscar archivos de migración en ubicaciones comunes:

```
migrations/*.sql
db/migrations/*.sql
sql/migrations/*.sql
database/migrations/*.sql
**/migrate/*.sql
```

También verificar dumps de schema: `schema.sql`, `init.sql`, `db/schema.sql`

### Paso 2 — Parsear el Schema

Leer los archivos de migración en orden (por número/timestamp) y construir un modelo mental de:

1. **Tablas** — nombre, columnas, tipos, restricciones
2. **Relaciones** — foreign keys, tablas de unión
3. **Índices** — nombre, columnas, unique/partial
4. **Enums/Tipos** — tipos personalizados definidos
5. **Políticas RLS** — si es multi-tenant
6. **Triggers** — si existen

### Paso 3 — Producir Resumen del Schema

Generar un resumen estructurado:

```markdown
## Resumen del Schema — <project>

### Tablas (<count>)

#### <table_name>
| Columna | Tipo | Nullable | Default | Restricción |
|--------|------|----------|---------|------------|
| id | UUID | NO | uuid_generate_v7() | PK |
| email | VARCHAR(255) | NO | — | UNIQUE |
| tenant_id | UUID | YES | — | FK → tenants(id) |
| created_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | — |

**Índices:** idx_users_email, idx_users_tenant_id
**RLS:** tenant_id = current_setting('app.tenant_id')

### Relaciones
- users.tenant_id → tenants.id
- user_roles.user_id → users.id
- user_roles.role_id → roles.id

### Cantidad de Migraciones: <N> (última: <filename>)

### Problemas Potenciales
- [ ] La tabla X no tiene tenant_id (brecha multi-tenant)
- [ ] La tabla Y no tiene índice en la columna Z consultada frecuentemente
- [ ] La columna A es VARCHAR(255) pero solo almacena códigos cortos
```

### Paso 4 — Verificar Desajustes Schema/Código (opcional)

Si el orquestador lo solicita, comparar el schema contra los archivos de queries del repositorio:
- Columnas referenciadas en queries que no existen en el schema
- Tablas en el schema que no tienen repositorio correspondiente
- Desajustes de tipo (schema dice UUID, el código escanea como string)

## Reglas

- **SOLO LECTURA** — nunca modificar el schema ni los archivos de migración
- **Reportar, no corregir** — marcar problemas para que el agente DBA los maneje
- **El orden importa** — leer las migraciones en secuencia para entender la evolución
- **Verificar rollbacks** — notar migraciones que tienen `.up.sql` pero no `.down.sql`
