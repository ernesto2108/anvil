# Bases de Datos de Series de Tiempo

## Cuándo Usar

Métricas de infraestructura (CPU, memoria, latencia), IoT (sensores, dispositivos), datos financieros (OHLCV, ticks), logs con timestamp como dimensión primaria. Característica: volumen alto de writes append-only, queries siempre incluyen rango de tiempo.

## Opciones y Cuándo Elegir

| Motor | Cuándo elegir | Notas |
|---|---|---|
| **TimescaleDB** | Ya tienes Postgres, queries SQL complejas, JOINs con datos relacionales | Extensión de Postgres — hypertables con particionado automático por tiempo |
| **InfluxDB** | Métricas de infra, IoT, cardinality controlada | Measurement + tags + fields + timestamp. Tags son indexados, fields no |
| **QuestDB** | Ingestion masiva, SQL con extensiones de tiempo | Muy rápido para writes. Columnar storage |

## Conceptos Clave

### TimescaleDB
- **Hypertable**: tabla particionada automáticamente por tiempo. Semántica SQL normal
- **Chunks**: particiones internas (por defecto 7 días). Compresión automática de chunks viejos
- **Continuous aggregates**: vistas materializadas que se actualizan incrementalmente
- **Retention policies**: eliminación automática de datos viejos

### InfluxDB
- **Measurement**: equivalente a tabla
- **Tags**: campos indexados (string only). Definen la "serie". Cardinalidad = combinaciones únicas de tags
- **Fields**: campos NO indexados (cualquier tipo). Datos que cambian por punto
- **Timestamp**: siempre presente, nanosecond precision por defecto

### Modelo de cardinalidad (InfluxDB — CRÍTICO)
```
# CORRECTO — tags de baja cardinalidad
measurement: cpu_usage
tags: {host: "web-01", region: "us-east"}
fields: {value: 72.5, cores_used: 3}

# INCORRECTO — tag de alta cardinalidad
tags: {request_id: "uuid-unique-per-request"}  # ← MILLONES de series → OOM
```

**Regla**: si un tag tiene más de 100K valores únicos, es un field, no un tag.

## Migraciones / Versionado

### TimescaleDB — Migraciones SQL normales
Compatible con golang-migrate, Flyway, Alembic. Mismas herramientas que Postgres.

```sql
-- Crear hypertable
CREATE TABLE metrics (
  time        TIMESTAMPTZ NOT NULL,
  sensor_id   TEXT NOT NULL,
  temperature DOUBLE PRECISION,
  humidity    DOUBLE PRECISION
);
SELECT create_hypertable('metrics', 'time');

-- Agregar columna (migración normal)
ALTER TABLE metrics ADD COLUMN pressure DOUBLE PRECISION;

-- Retention policy
SELECT add_retention_policy('metrics', INTERVAL '90 days');

-- Continuous aggregate
CREATE MATERIALIZED VIEW metrics_hourly
WITH (timescaledb.continuous) AS
SELECT
  time_bucket('1 hour', time) AS bucket,
  sensor_id,
  AVG(temperature) AS avg_temp,
  MAX(temperature) AS max_temp
FROM metrics
GROUP BY bucket, sensor_id;
```

### InfluxDB — Sin migraciones formales
Los campos emergen de los writes. No hay DDL. El "schema" se define por convención:
- Documentar measurements, tags y fields en el repo
- Versionar la convención en un archivo `influx_schema.md`
- Cambio de tag a field (o viceversa) requiere reescribir toda la serie

### QuestDB — DDL básico
```sql
CREATE TABLE metrics (
  timestamp TIMESTAMP,
  sensor_id SYMBOL,  -- indexado, como tag de InfluxDB
  temperature DOUBLE,
  humidity DOUBLE
) TIMESTAMP(timestamp) PARTITION BY DAY;
```

## Convenciones de Naming

### TimescaleDB (SQL estándar)
```
# Tablas: plural, snake_case
sensor_readings, api_latencies, financial_ticks

# Continuous aggregates: {tabla}_{intervalo}
sensor_readings_hourly, api_latencies_5min

# Índices: idx_{tabla}_{columnas}
idx_sensor_readings_sensor_time
```

### InfluxDB
```
# Measurements: snake_case, singular descriptivo
cpu_usage, http_request, temperature_reading

# Tags: snake_case, sin prefijos
host, region, sensor_id, environment

# Fields: snake_case, descriptivos
value, response_time_ms, error_count
```

## Pitfalls de Producción

| # | Pitfall | Consecuencia | Prevención |
|---|---|---|---|
| 1 | InfluxDB: alta cardinalidad en tags | Memory explosion, queries lentas | Tags = dimensiones fijas (<100K valores), fields = datos variables |
| 2 | Sin retention policy | Disco lleno en semanas/meses | Configurar retention desde día 1 |
| 3 | TimescaleDB: olvidar `create_hypertable` | Funciona como Postgres normal sin particionado | Verificar con `SELECT * FROM timescaledb_information.hypertables` |
| 4 | Queries sin filtro de tiempo | Full scan de millones de rows | Siempre incluir `WHERE time > NOW() - INTERVAL '...'` |
| 5 | Insertar punto por punto | Throughput 10-100x menor que batch | Buffer y batch writes |
| 6 | InfluxDB: cambiar tag ↔ field | Requiere reescribir toda la serie | Planificar schema antes de escribir datos |
| 7 | No configurar downsampling | Queries históricas cada vez más lentas | Continuous aggregates (TimescaleDB) o tasks (InfluxDB) |

## Optimización de Rendimiento

### Writes
- **Batch**: nunca insertar punto por punto — buffer de 1000-5000 puntos y bulk insert
- **TimescaleDB**: `INSERT INTO ... VALUES (...), (...), (...)` — multi-row insert
- **InfluxDB**: Line Protocol con batch de líneas
- **QuestDB**: ILP (InfluxDB Line Protocol) sobre TCP para máximo throughput

### Queries
- **TimescaleDB**: `time_bucket()` para agregaciones eficientes por período
- **InfluxDB**: `GROUP BY time()` en InfluxQL o `window()` en Flux
- **Downsampling**: agregar datos históricos (minutos → horas → días) para queries históricas
- **Compression** (TimescaleDB): `ALTER TABLE metrics SET (timescaledb.compress)` — 90%+ de compresión

### Retention
```sql
-- TimescaleDB: eliminar datos >90 días automáticamente
SELECT add_retention_policy('metrics', INTERVAL '90 days');

-- InfluxDB: retention policy por database/bucket
CREATE RETENTION POLICY "three_months" ON "mydb" DURATION 90d REPLICATION 1;
```

## Drivers por Stack

| Stack | TimescaleDB | InfluxDB | QuestDB |
|---|---|---|---|
| Go | `pgx/v5` (es Postgres) | `influxdata/influxdb-client-go` | `questdb/go-questdb-client` |
| TypeScript | `pg` o cualquier driver Postgres | `@influxdata/influxdb-client` | `@questdb/nodejs-client` |
| Python | `psycopg2` / `asyncpg` | `influxdb-client` | `questdb` sender |
| Rust | `sqlx` con feature Postgres | `influxdb2` crate | Line Protocol sobre TCP |
