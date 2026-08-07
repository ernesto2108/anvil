---
name: db-engines
description: Conocimiento específico por motor para bases de datos relacionales (PostgreSQL, SQLite, MySQL), Redis, vectoriales, documentales, time-series, messaging y search engines. Mejores prácticas, drivers, migraciones y particularidades por motor. Cargar antes de escribir migraciones u optimizar queries. Úsalo cuando se tome una decisión sobre qué motor de base de datos usar, o cuando el architect o dba evalúen opciones de persistencia.
---

# Referencia de Motores de Base de Datos

> Reglas, limitaciones y mejores prácticas específicas por motor. El agente DBA carga este skill para adaptar las migraciones y decisiones de schema al motor objetivo.

## Cuándo Usar

- Antes de escribir cualquier migración — detectar el motor primero, luego cargar la referencia correspondiente
- Al optimizar queries — las técnicas difieren por motor
- Al configurar una nueva base de datos — ajuste de conexiones, PRAGMAs, pooling

## Detección del Motor

Detectar el motor a partir de señales del proyecto (verificar en orden):

### Bases de datos relacionales
1. Archivos de migración existentes — la sintaxis revela el motor (ej. `SERIAL` = Postgres, `AUTOINCREMENT` = SQLite)
2. Imports de driver en el código — `pq` / `pgx` = Postgres, `go-sqlite3` = SQLite, `go-sql-driver/mysql` = MySQL
3. Cadenas de conexión en la configuración — `postgres://`, `file:*.db`, `mysql://`

### No relacionales
4. Imports de driver — `go-redis` = Redis, `mongo-driver` = MongoDB, `kafka-go` = Kafka, `elastic` = Elasticsearch, `qdrant` / `pgvector` = vectorial
5. `docker-compose.yml` u otros archivos de infra — nombres de imagen de servicio (`redis`, `mongo`, `elasticsearch`, `qdrant`, `kafka`, `influxdb`)
6. Configuración / env vars — `REDIS_URL`, `MONGO_URI`, `KAFKA_BROKERS`, `ELASTICSEARCH_URL`, `QDRANT_URL`
7. Documento de diseño del arquitecto — puede especificar la elección del motor

## Referencias de Motor

Cargar SOLO el o los motores relevantes para la tarea actual:

### Relacionales
- **PostgreSQL** → leer `engines/postgresql.md`
- **SQLite** → leer `engines/sqlite.md`
- **MySQL** → leer `engines/mysql.md`

### No relacionales
- **Redis** → leer `engines/redis.md` — caché, sesiones, colas, rate limiting, distributed locks
- **Vector DBs** → leer `engines/vector.md` — pgvector, Qdrant, Pinecone, Weaviate, embeddings, RAG
- **Document DBs** → leer `engines/document.md` — MongoDB, DynamoDB, Firestore
- **Time-Series** → leer `engines/timeseries.md` — TimescaleDB, InfluxDB, QuestDB
- **Messaging** → leer `engines/messaging.md` — Kafka, RabbitMQ, NATS, Schema Registry
- **Search Engines** → leer `engines/search.md` — Elasticsearch, Meilisearch, Typesense

Cada referencia cubre: cuándo usar, modelo de datos, convenciones de naming, migraciones/versionado, pitfalls de producción, optimización y drivers por stack.

## Herramientas de Migración por Lenguaje

| Lenguaje | Herramienta | Notas |
|----------|------|-------|
| Go | `github.com/golang-migrate/migrate/v4` | Usar driver de fuente `file://` con `NewWithDatabaseInstance`. NO usar `embed.FS` / `iofs` — mantener los archivos SQL como archivos planos en `migrations/`. NO llamar `m.Close()` cuando se usa `WithInstance` (cierra el `*sql.DB` compartido). Ver `/go-conventions` para código de configuración |
| Node.js | `knex` o `prisma migrate` | |
| Python | `alembic` (SQLAlchemy) o `django.db.migrations` | |
| Rust | `sqlx migrate` o `diesel migrations` | |

## Referencia Rápida — Comparación de Motores Relacionales

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

## Referencia Rápida — Motores No Relacionales

| Motor | Migraciones | Flujo del DBA | Driver Go recomendado |
|-------|-------------|---------------|----------------------|
| Redis | No (keyspace versionado) | TTL audit, memory analysis, keyspace naming | go-redis/v9 |
| pgvector | SQL (es Postgres) | Igual que Postgres + tracking de modelo de embedding | pgx/v5 + pgvector-go |
| Qdrant | Collection versioning | Re-embed al cambiar modelo, metadata schema | qdrant/go-client |
| MongoDB | Lazy/batch (schema_version) | Schema version en docs, bulk migration scripts | mongo-driver |
| DynamoDB | No (schema-on-read) | Table design review, GSI planning | aws-sdk-go-v2 |
| TimescaleDB | SQL (es Postgres) | Igual que Postgres + hypertables, retention, aggregates | pgx/v5 |
| InfluxDB | No (schema-on-write) | Measurement/tag design, cardinality review | influxdb-client-go |
| Kafka | Schema Registry | Schema evolution, compatibility validation | kafka-go o confluent-kafka-go |
| Elasticsearch | Alias + reindex | Mapping versioning, alias swap, sync strategy | go-elasticsearch |
| Meilisearch | Settings update | Searchable/filterable attributes, sync strategy | meilisearch-go |
