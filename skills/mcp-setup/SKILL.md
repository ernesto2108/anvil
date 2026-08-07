---
name: mcp-setup
description: Receta determinista para activar MCP servers de infraestructura en un repo. Detecta servicios (PostgreSQL, MySQL, Redis, Kafka, Elasticsearch, MongoDB, SQLite), los mapea a sus packages MCP, y genera .mcp.json.example (commitable) y .mcp.json (local gitignored) siguiendo el patrón .env/.env.example — connection strings siempre vía ${ENV_VAR}, nunca hardcodeadas. Úsalo cuando context-init o infra-probe reporten "mcp_available: false", cuando hagas setup inicial de un proyecto nuevo, o cuando el usuario diga "activar MCP", "configurar mcp.json", "habilitar MCPs de infra", "setup de mcp servers". Pausa con confirmación antes de escribir.
---

# MCP Setup — Activación de MCP servers de infraestructura

> Receta determinista que detecta los servicios de infra del repo, los mapea a sus MCP server packages, y genera `.mcp.json.example` (commitable) y `.mcp.json` (local, gitignored). Un único paso de confirmación antes de escribir. Credenciales siempre vía `${ENV_VAR}`.

## Filosofía

1. **Patrón `.env` / `.env.example`** — el `.mcp.json.example` es la referencia versionada que cualquier dev copia; el `.mcp.json` real es de cada desarrollador y nunca se commitea. Tratar la config de MCP como se trata cualquier secreto de proyecto.
2. **Credenciales fuera del archivo, siempre** — el `.mcp.json` referencia variables de entorno (`${POSTGRES_URL}`), nunca valores reales. Una credencial hardcodeada es un leak esperando a ocurrir; por eso jamás se lee `.env` con valores reales.
3. **Confirmar antes de escribir, mergear antes de pisar** — generar archivos de config es una mutación visible. Mostrar el plan y esperar OK; al actualizar archivos existentes, preservar las entradas manuales del humano en lugar de sobrescribirlas.

## Parámetros de entrada

| Parámetro | Requerido | Valores | Default | Notas |
|---|---|---|---|---|
| `services` | No | lista explícita, ej. `["postgres", "redis"]` | auto-detección | Si se omite, detectar todos los servicios declarados en el repo |
| `env` | No | `dev` \| `staging` | `dev` | Si se omite, asumir `dev` y advertirlo en el plan |

## Flujo de trabajo

### Paso 1 — Leer contexto del repo

Leer en este orden de prioridad (el primero que exista gana como fuente primaria; los siguientes enriquecen):

1. `.project-context/infra-services.md` — cache generado por `context-init`. Si existe, úsalo como punto de partida (servicios + `mcp_available` ya resuelto).
2. `docker-compose.yml` / `docker-compose.dev.yml` / `docker-compose.override.yml` — servicios y puertos declarados.
3. `.env.example` / `.env.local` — nombres de variables de conexión.
4. Archivos de dependencias — buscar clientes de BD:
   - `go.mod` → `pgx`, `redis`, `rueidis`, `sarama`, `kafka`, `elasticsearch`, `mongo`
   - `package.json` → `pg`, `ioredis`, `kafkajs`, `@elastic/elasticsearch`, `mongodb`
   - `requirements.txt` → `psycopg`, `redis`, `kafka-python`, `elasticsearch`, `pymongo`
   - `Cargo.toml` → `sqlx`, `tokio-postgres`, `redis`, `rdkafka`, `elasticsearch`, `mongodb`

> Si ya existe `.mcp.json.example`, leerlo para hacer **merge inteligente**: agregar entradas nuevas sin pisar las existentes.

### Paso 2 — Mapear servicios a packages MCP

| Servicio | Package MCP | Env vars esperadas |
|---|---|---|
| PostgreSQL | `@modelcontextprotocol/server-postgres` | `POSTGRES_URL` o `DATABASE_URL` |
| MySQL | `@modelcontextprotocol/server-mysql` | `MYSQL_URL` o `DATABASE_URL` |
| Redis | `@modelcontextprotocol/server-redis` | `REDIS_URL` |
| Kafka | `@modelcontextprotocol/server-kafka` | `KAFKA_BROKERS` |
| Elasticsearch | `@modelcontextprotocol/server-elasticsearch` | `ELASTICSEARCH_URL` |
| MongoDB | `@modelcontextprotocol/server-mongodb` | `MONGODB_URI` |
| SQLite | `@modelcontextprotocol/server-sqlite` | path al archivo `.db` detectado en el repo |

> **Si un servicio detectado no está en esta tabla → reportarlo como no soportado.** No inventar un package.

### Paso 3 — Verificar variables de entorno

1. Leer `.env.example` (y `.env.local` si existe) para saber qué variables están declaradas.
2. **Gate de seguridad: NUNCA leer `.env` con valores reales.** Solo se leen nombres de variables desde `.env.example` / `.env.local`.
3. Variable esperada no encontrada en ningún `.env*` → marcarla `⚠️ PENDIENTE` en el plan. **No bloquear** la generación.

### Paso 4 — PAUSA: mostrar plan y pedir confirmación

Antes de escribir nada, mostrar el plan y **esperar confirmación explícita del usuario**:

```
Servicios detectados: postgres, redis

Archivos a generar:
  .mcp.json.example  → commitable (template para otros devs)
  .mcp.json          → local gitignored (Claude Code lo usa)

Entradas:
  ✓ postgres → @modelcontextprotocol/server-postgres (POSTGRES_URL en .env.example)
  ⚠️ redis   → @modelcontextprotocol/server-redis (REDIS_URL no encontrada)

También verificaré que .mcp.json esté en .gitignore.

¿Procedo?
```

> **Si el usuario no confirma → DETENER. No escribir nada.**

Si `env` se asumió `dev`, incluir en el plan: `⚠️ env no especificado — se asumió "dev".`

### Paso 5 — Escribir archivos

**Formato exacto por entrada** (idéntico en ambos archivos):

```json
{
  "_readme": "Copia este archivo a .mcp.json y completa las variables de entorno. El archivo .mcp.json está gitignoreado.",
  "mcpServers": {
    "postgres": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-postgres"],
      "env": {
        "POSTGRES_URL": "${POSTGRES_URL}"
      }
    }
  }
}
```

1. **`.mcp.json.example`** — siempre generar/mergear. Es el artefacto commitable.
   - Env vars como `${ENV_VAR}` (placeholders, nunca valores reales).
   - Si ya existía → merge inteligente: agregar entradas nuevas, **NO pisar las manuales existentes**.

2. **`.mcp.json`** — mismo contenido que el `.example`.
   - Solo generar si **NO existe** localmente.
   - **Si ya existe → preguntar al usuario** antes de pisarlo (¿regenerar, o solo actualizar el `.example`?). No pisar sin confirmación.

3. **`.gitignore`** — verificar:
   - Si `.mcp.json` no está listado → agregarlo.
   - Si `.mcp.json.example` está gitignoreado por error → **advertir al usuario, NO modificar** sin confirmación.

### Paso 6 — Indicar próximos pasos

Tras escribir, indicar al usuario:

- **Reiniciar Claude Code** para que reconozca el nuevo `.mcp.json`.
- Tras el reinicio: correr `context-init` para que `infra-probe` verifique que los MCPs quedaron disponibles.
- Para otros devs que clonen el repo: `cp .mcp.json.example .mcp.json` y completar las variables de entorno.

## Reglas

- **NUNCA** hardcodear credenciales (connection strings, passwords, URLs reales) — siempre `${ENV_VAR}`.
- **NUNCA** leer `.env` con valores reales — solo `.env.example` y `.env.local` para nombres de variables.
- **NUNCA** commitear `.mcp.json` — siempre debe estar en `.gitignore`.
- **NUNCA** pisar `.mcp.json` existente sin confirmación del usuario.
- **NUNCA** pisar entradas manuales en `.mcp.json.example` existente — mergear, no sobrescribir.
- **SIEMPRE** generar `.mcp.json.example` como el artefacto commitable.
- **No instalar packages** — `npx` los resuelve en runtime; la instalación, si hace falta, es manual.
- Solo se escriben `.mcp.json`, `.mcp.json.example` y (si falta) la línea `.mcp.json` en `.gitignore`.

## Consumidores

- `context-init` — al final del scan de infraestructura, cuando detecta `mcp_available: false` para algún servicio.
- `devops` — en el setup inicial de un proyecto nuevo.
- Invocación directa por el humano para activar MCPs de infra en un repo existente.

## Checklist antes de cerrar

- [ ] Se leyó `.project-context/infra-services.md` si existía (cache caliente)
- [ ] Servicios detectados desde docker-compose / .env.example / archivos de dependencias
- [ ] Cada servicio mapeado al package exacto de la tabla del Paso 2
- [ ] Solo se leyó `.env.example` / `.env.local` — nunca `.env` con valores reales
- [ ] Variables ausentes marcadas `⚠️ PENDIENTE`, sin bloquear
- [ ] Hubo confirmación explícita del usuario antes de escribir (Paso 4)
- [ ] Ningún valor en `env` es una credencial literal — todos `${VAR}`
- [ ] `.mcp.json.example` generado/mergeado sin pisar entradas manuales
- [ ] `.mcp.json` no se pisó sin confirmación si ya existía
- [ ] `.mcp.json` listado en `.gitignore`
