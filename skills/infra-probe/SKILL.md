---
name: infra-probe
description: Sondeo read-only del estado real de servicios de infraestructura (PostgreSQL, MySQL, Redis, Kafka, Elasticsearch, MongoDB, SQLite) via MCP, sin operar a ciegas. Detecta qué infra declara el proyecto, qué MCPs están disponibles en la sesión, y consulta health/schema/persistencia. Úsalo cuando el usuario o un agente diga "is the database up", "check redis", "inspect the live schema", "verify persistence", "health check de infra", "qué servicios están corriendo", o antes de escribir tests de integración o lógica de datos. Solo lectura — nunca writes ni DDL.
---

# Infra Probe — Sondeo read-only de servicios de infraestructura

> Consulta el estado real de Postgres, Redis, Kafka, Elasticsearch, MongoDB y SQLite via MCP, en lugar de asumir cómo está configurada la infra. Read-only siempre.

## Filosofía

1. **Nunca operar a ciegas** — antes de escribir un test de integración o lógica de datos, verificar que el servicio existe, está arriba, y tiene el schema que el código espera. Una suposición incorrecta sobre la infra contamina todo lo que se construye encima.
2. **Read-only por contrato, no por buena voluntad** — la skill solo emite operaciones de lectura (SELECT, EXPLAIN, INFO, PING, SHOW, DESCRIBE). Cualquier intento de escritura se rechaza y se reporta, sin importar quién lo pida.
3. **Detección en runtime, no hardcoding** — la skill no asume qué MCPs están configurados. Los descubre cada vez via ToolSearch, y degrada con elegancia a inspección de archivos cuando no hay MCP.

## Parámetros de entrada

| Parámetro | Requerido | Valores | Default | Notas |
|---|---|---|---|---|
| `service` | No | `postgres` \| `mysql` \| `redis` \| `kafka` \| `elasticsearch` \| `mongodb` \| `auto` | `auto` | `auto` detecta y sondea todos los servicios declarados |
| `env` | No | `dev` \| `staging` \| `prod` | `dev` | Si se omite, usar `dev` y **advertir explícitamente** en el output que se asumió `dev` |
| `purpose` | No | `health-check` \| `schema-inspect` \| `verify-persistence` | `health-check` | Determina la profundidad del sondeo y se loggea como intent |

**Advertencia de ambiente:** si `env` no se pasó, incluir en el output: `⚠️ env no especificado — se asumió "dev". Pasa env explícitamente para staging/prod.`

## Flujo de trabajo

### Paso 1 — Detección en 3 capas

**Capa 1 — ¿Qué infra declara el proyecto?** Leer en este orden (el primero que exista gana como fuente de verdad, los siguientes enriquecen):

1. `.project-context/infra-services.md` — cache caliente generado por `context-init`. Si existe, úsalo como punto de partida (lista de servicios + disponibilidad de MCP ya resuelta).
2. `docker-compose.yml`, `docker-compose.dev.yml`, `docker-compose.override.yml` — servicios y puertos declarados.
3. `.env.example`, `.env.local`, `.env` — URLs de conexión (`DATABASE_URL`, `REDIS_URL`, `KAFKA_BROKERS`, etc.).
4. Archivos de dependencias:
   - `go.mod` → buscar `pgx`, `redis`, `rueidis`, `sarama`, `kafka`, `elasticsearch`, `mongo`
   - `package.json` → buscar `pg`, `ioredis`, `kafkajs`, `@elastic/elasticsearch`, `mongodb`
   - `requirements.txt` → buscar `psycopg`, `redis`, `kafka-python`, `elasticsearch`, `pymongo`
   - `Cargo.toml` → buscar `sqlx`, `tokio-postgres`, `redis`, `rdkafka`, `elasticsearch`, `mongodb`

> **Si `.project-context/infra-services.md` no existe:** continuar con las fuentes 2-4. No detenerse — la skill funciona sin el cache, solo más lento.

**Capa 2 — ¿Qué MCPs están disponibles en la sesión?** Para cada servicio detectado en la Capa 1, intentar `ToolSearch` con el keyword correspondiente:

| Detección | ToolSearch keyword | MCP esperado |
|---|---|---|
| PostgreSQL | `"postgres"` o `"postgresql"` | `mcp__postgres` |
| MySQL | `"mysql"` | `mcp__mysql` |
| Redis | `"redis"` | `mcp__redis` |
| Kafka | `"kafka"` | `mcp__kafka` |
| Elasticsearch | `"elasticsearch"` | `mcp__elasticsearch` |
| MongoDB | `"mongodb"` | `mcp__mongodb` |
| SQLite | — (no necesita MCP) | filesystem — lee el archivo `.db` directo |

Si el MCP no aparece en ToolSearch → marcar `mcp_available: false` y planear fallback (file-inspection).

**Capa 3 — Tabla de mapeo (qué herramientas usar por tecnología):**

| Tecnología | MCP esperado | Herramientas permitidas (read-only) |
|---|---|---|
| PostgreSQL | `mcp__postgres__*` | `query` (SELECT/EXPLAIN/SHOW only), `describe_table`, `list_tables` |
| MySQL | `mcp__mysql__*` | `query` (SELECT/EXPLAIN/SHOW only), `describe_table`, `list_tables` |
| Redis | `mcp__redis__*` | `info`, `ping`, `memory_usage`, `dbsize` |
| Kafka | `mcp__kafka__*` | `list_topics`, `describe_topic`, `consumer_group_offsets` |
| Elasticsearch | `mcp__elasticsearch__*` | `get_mapping`, `get_settings`, `cat_indices` |
| MongoDB | `mcp__mongodb__*` | `list_collections`, `get_schema`, `count` |
| SQLite | filesystem | Read directo sobre el archivo `.db` |

### Paso 2 — Gate de seguridad read-only (CRÍTICO)

Antes de emitir cualquier operación, validar que es read-only. **Si la operación a ejecutar contiene `INSERT`, `UPDATE`, `DELETE`, `DROP`, `CREATE`, `ALTER`, `TRUNCATE`, `GRANT`, `FLUSHDB`, `FLUSHALL`, `SET`, `DEL`, o cualquier mutación → DETENER, rechazar la operación, y reportar el intento.**

Esta puerta no es opcional ni configurable. La skill jamás escribe en ninguna infraestructura, sin importar el `env`, el `purpose`, ni quién la invoque. Solo se permiten: `SELECT`, `EXPLAIN`, `SHOW`, `DESCRIBE`, `INFO`, `PING`, y los tool calls read-only listados en la Capa 3.

Reporte de rechazo:
```
🚫 Operación rechazada por infra-probe — no es read-only.
   Intento: <operación exacta>
   infra-probe solo emite lecturas. Para mutaciones, usa el agente DBA correspondiente.
```

### Paso 3 — Sondear por servicio

Para cada servicio objetivo (`service=auto` → todos los detectados):

1. **Si hay MCP disponible:**
   - `health-check`: emitir el ping/info más barato (`ping`, `SELECT 1`, `cat_indices`, `list_topics`). Determina `status: up | down`.
   - `schema-inspect`: además listar tablas/colecciones/topics/mappings con las tools read-only de la Capa 3.
   - `verify-persistence`: además contar filas/keys/mensajes (`SELECT count(*)`, `dbsize`, `count`, `consumer_group_offsets`).
2. **Si NO hay MCP disponible (fallback):**
   - Leer la URL de conexión desde `.env*` o `docker-compose` para confirmar que el servicio está **declarado**, pero marcar `status: unavailable` (no se puede confirmar liveness sin MCP).
   - Agregar el servicio a `fallback_used`.
   - `reason`: `MCP no configurado — solo se confirmó la declaración en <fuente>`.
3. **SQLite:** abrir el archivo `.db` via filesystem read (no requiere MCP). Si el archivo existe → `status: up`.

### Paso 4 — Emitir output estructurado

Ver "Formato de salida". Detenerse después de reportar — la skill no actúa sobre los hallazgos, solo los reporta al agente host.

## Formato de salida

```
{
  env: "dev",
  env_assumed: true,            // true si env no se pasó explícitamente
  purpose: "health-check",
  services: {
    postgres: {
      status: "up" | "down" | "unavailable",
      mcp_available: true | false,
      tables: [...],            // solo si status=up y purpose incluye schema-inspect
      row_counts: {...},        // solo si purpose=verify-persistence
      reason: "..."             // presente solo si status != "up"
    },
    redis: { ... },
    kafka: { ... }
  },
  fallback_used: ["redis"],     // servicios sin MCP donde se usó file-inspection
  warnings: [
    "env no especificado — se asumió dev"
  ]
}
```

## Reglas

- **Read-only absoluto** — solo `SELECT`, `EXPLAIN`, `SHOW`, `DESCRIBE`, `INFO`, `PING` y las tools read-only de la Capa 3. El gate del Paso 2 es inviolable.
- **Credentials nunca en la skill** — las connection strings viven en el MCP server o en `.env*`. La skill nunca las inventa ni las imprime en claro.
- **Detección en runtime** — nunca asumir que un MCP está configurado; siempre confirmar via ToolSearch.
- **Degradar con elegancia** — sin MCP, confirmar declaración del servicio y marcar `unavailable`, no fallar.
- **Advertir el ambiente asumido** — si `env` no se pasó, decirlo en `warnings`.
- **No actuar sobre los hallazgos** — la skill reporta estado; el agente host decide qué hacer.

## Ejemplos de uso por consumidor

### `tester` — Paso 0.5: health check antes de tests de integración
```
Invocar infra-probe con { service: "auto", env: "dev", purpose: "health-check" }.
Si postgres.status != "up" → no escribir tests de integración contra DB; reportar
al humano que la infra no está disponible y ofrecer tests unitarios con mocks.
```

### `developer-backend` — verificar schema real antes de implementar lógica de datos
```
Invocar infra-probe con { service: "postgres", env: "dev", purpose: "schema-inspect" }.
Usar services.postgres.tables para confirmar nombres reales de tablas/columnas antes
de escribir el repositorio. Evita queries contra columnas que no existen.
```

### `dba-reader` — alternativa live a db-schema-scan
```
Cuando los archivos de migración no reflejan el estado real (drift sospechado),
invocar infra-probe con { service: "postgres", purpose: "schema-inspect" } para leer
el schema vivo y compararlo contra lo que declaran las migraciones.
```

### `context-init` — enriquecer infra-services.md durante el scan inicial
```
Durante el scan de infraestructura, invocar infra-probe con { service: "auto",
purpose: "health-check" } para resolver mcp_available por servicio y escribir el
resultado en .project-context/infra-services.md.
```

## Checklist antes de reportar

- [ ] Se leyó `.project-context/infra-services.md` si existía (cache caliente)
- [ ] Se detectaron servicios desde docker-compose / .env / archivos de dependencias
- [ ] Se intentó ToolSearch por cada servicio detectado
- [ ] El gate read-only (Paso 2) se aplicó a toda operación emitida
- [ ] Servicios sin MCP marcados `unavailable` y listados en `fallback_used`
- [ ] Si `env` se asumió `dev`, hay un warning explícito en el output
- [ ] El output sigue el formato estructurado, sin connection strings en claro
