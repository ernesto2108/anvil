---
name: db-engines
description: Conocimiento específico por motor para PostgreSQL, SQLite y MySQL. PRAGMAs, limitaciones, mejores prácticas, drivers, ajuste de conexiones y particularidades de migración por motor. Cargar antes de escribir migraciones u optimizar queries.
---

# Referencia de Motores de Base de Datos

> Reglas, limitaciones y mejores prácticas específicas por motor. El agente DBA carga este skill para adaptar las migraciones y decisiones de schema al motor objetivo.

## Cuándo Usar

- Antes de escribir cualquier migración — detectar el motor primero, luego cargar la referencia correspondiente
- Al optimizar queries — las técnicas difieren por motor
- Al configurar una nueva base de datos — ajuste de conexiones, PRAGMAs, pooling

## Detección del Motor

Detectar el motor a partir de señales del proyecto (verificar en orden):

1. Archivos de migración existentes — la sintaxis revela el motor (ej. `SERIAL` = Postgres, `AUTOINCREMENT` = SQLite)
2. Imports de driver en el código — `pq` / `pgx` = Postgres, `go-sqlite3` = SQLite, `go-sql-driver/mysql` = MySQL
3. Cadenas de conexión en la configuración — `postgres://`, `file:*.db`, `mysql://`
4. `docker-compose.yml` u otros archivos de infra — nombres de imagen de servicio
5. Documento de diseño del arquitecto — puede especificar la elección del motor

## Referencias de Motor

Cargar SOLO el o los motores relevantes para la tarea actual:

- **PostgreSQL** → leer `engines/postgresql.md`
- **SQLite** → leer `engines/sqlite.md`
- **MySQL** → leer `engines/mysql.md`

Cada referencia cubre: tipos, limitaciones, patrones de migración, ajuste de rendimiento, patrones multi-tenant y drivers de Go.

## Herramientas de Migración por Lenguaje

| Lenguaje | Herramienta | Notas |
|----------|------|-------|
| Go | `github.com/golang-migrate/migrate/v4` | Usar driver de fuente `file://` con `NewWithDatabaseInstance`. NO usar `embed.FS` / `iofs` — mantener los archivos SQL como archivos planos en `migrations/`. NO llamar `m.Close()` cuando se usa `WithInstance` (cierra el `*sql.DB` compartido). Ver `/go-conventions` para código de configuración |
| Node.js | `knex` o `prisma migrate` | |
| Python | `alembic` (SQLAlchemy) o `django.db.migrations` | |
| Rust | `sqlx migrate` o `diesel migrations` | |

## Referencia Rápida — Comparación de Motores

| Característica | PostgreSQL | SQLite | MySQL |
|---------|-----------|--------|-------|
| UUID nativo | Sí | No (TEXT) | No (CHAR(36)) |
| Tipo ENUM | Sí | No (CHECK) | Sí |
| ALTER TABLE DROP COLUMN | Sí | No (4 pasos) | Sí |
| Índice concurrente | Sí | No | No |
| RLS | Sí | No | No |
| Enforcement de FK | Siempre activo | Por conexión PRAGMA | Siempre activo (InnoDB) |
| Transacciones en migraciones | Explícitas ok | Solo implícitas | Explícitas ok |
| Máx escritores concurrentes | Muchos | Uno | Muchos |
| Driver Go recomendado | pgx/v5 | mattn/go-sqlite3 | go-sql-driver/mysql |
