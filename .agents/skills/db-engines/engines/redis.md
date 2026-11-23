# Redis

## Cuándo Usar

Caché de resultados costosos, sesiones de usuario, rate limiting, pub/sub lightweight, colas simples (LPUSH/BRPOP), leaderboards (sorted sets), distributed locks. Redis Stack agrega: JSON nativo (RedisJSON), búsqueda full-text (RediSearch), series de tiempo (RedisTimeSeries).

## Modelo de Datos

Redis NO tiene schema. El **keyspace** ES el schema — la convención de nombres define la estructura.

### Estructuras de datos
| Estructura | Caso de uso | Ejemplo |
|---|---|---|
| String | Caché, contadores, flags | `cache:user:123:profile` |
| Hash | Objetos con campos | `user:123` → `{name, email, role}` |
| List | Colas, feeds recientes | `queue:emails` |
| Set | Membresías, tags únicos | `role:admin:members` |
| Sorted Set | Rankings, feeds ordenados por score | `leaderboard:weekly` |
| Stream | Event log, pub/sub persistente | `events:orders` |
| JSON (Stack) | Documentos anidados con JSONPath | `doc:order:456` |

### Convenciones de Naming (Keyspace)

Separador `:` estándar. Formato: `{app}:{env}:{entity}:{id}:{field}`

```
myapp:prod:session:{token}            # sesiones
myapp:prod:cache:user:{id}:profile    # caché de perfil
myapp:prod:rate:{user_id}:{window}    # rate limiting
myapp:prod:lock:{resource}            # distributed locks
myapp:prod:queue:{queue_name}         # colas
myapp:prod:counter:{metric}           # contadores
```

- Siempre prefijo de app para evitar colisiones en Redis compartido
- Env en el prefijo permite inspeccionar por entorno
- Nunca usar espacios ni caracteres especiales en keys

## TTL — Parte del Diseño, No Afterthought

**Regla:** cada key DEBE tener TTL o justificación explícita de por qué es permanente.

| Tipo de dato | TTL recomendado |
|---|---|
| Caché de API | 5min – 1hr según frecuencia de cambio |
| Sesiones | Duración de sesión (ej. 24hr) |
| Rate limiting | Tamaño de ventana (ej. 60s) |
| Locks | Timeout del lock + margen (ej. 30s) |
| Colas | Sin TTL (se consume), pero TTL en DLQ |
| Datos permanentes | Sin TTL — documentar por qué |

Keys sin TTL son memory leaks silenciosos. En auditorías, verificar con `OBJECT IDLETIME` y `MEMORY USAGE`.

## Migraciones / Versionado

Redis NO tiene migraciones formales. Estrategias de evolución:

### Versionado de keyspace
```
cache:v1:user:{id}  → formato viejo
cache:v2:user:{id}  → formato nuevo (campos adicionales)
```
TTL como herramienta de migración: dejar expirar v1, escribir solo v2.

### Cambio de estructura
Si un Hash necesita campos nuevos: agregarlos con `HSET` — los campos existentes no se afectan. Si se necesita cambiar el tipo de estructura (String → Hash), usar versionado de key.

### Auditoría de keyspace
En lugar de migraciones, el DBA audita:
1. Keys sin TTL → `redis-cli --scan --pattern '*' | xargs -L 1 redis-cli TTL`
2. Keys de gran tamaño → `redis-cli --bigkeys`
3. Distribución de memoria → `redis-cli --memkeys`
4. Patrones obsoletos → `SCAN` con pattern de versión vieja

## Pitfalls de Producción

| # | Pitfall | Consecuencia | Prevención |
|---|---|---|---|
| 1 | Keys sin TTL | OOM kill | TTL obligatorio por política |
| 2 | `KEYS *` en producción | Bloquea servidor (single-threaded) | Usar `SCAN` con cursor |
| 3 | Hotkeys | Latencia en toda la instancia | Sharding de key o read replicas |
| 4 | Large values (Sets/Lists >10K elementos) | Bloquea event loop | Particionar en chunks |
| 5 | Sin pipelining en batch | Latencia acumulada por roundtrip | `Pipeline()` para operaciones batch |
| 6 | Conexión por request | Agota file descriptors | Connection pooling obligatorio |
| 7 | Pub/Sub sin consumers | Mensajes descartados silenciosamente | Usar Streams si necesitas persistencia |

## Persistencia

| Modo | Descripción | Cuándo usar |
|---|---|---|
| RDB (snapshots) | Dump periódico a disco | Caché que tolera pérdida de datos recientes |
| AOF (append-only) | Log de cada escritura | Datos críticos (sesiones, colas) |
| RDB + AOF | Ambos | Producción con datos importantes |
| Sin persistencia | Solo RAM | Caché puro, datos reconstruibles |

## Optimización de Rendimiento

- **Pipelining**: agrupar comandos en batch — reduce roundtrips de red
- **Lua scripts**: operaciones atómicas complejas sin roundtrips múltiples
- **Connection pooling**: tamaño de pool = número de goroutines/workers concurrentes
- **Eviction policy**: `allkeys-lru` para caché puro, `volatile-lru` para mix caché + datos persistentes
- **`OBJECT ENCODING`**: verificar representación interna — ziplist es más eficiente que hashtable para hashes pequeños
- **Cluster**: sharding automático para datasets que no caben en una instancia

## Patrones Comunes

### Distributed Lock (Redlock simplificado)
```
SET lock:{resource} {token} NX EX 30
# ... trabajo ...
# Liberar solo si el token coincide (Lua script)
```

### Rate Limiting (Sliding Window)
```
key = rate:{user_id}:{window}
MULTI
  INCR key
  EXPIRE key {window_seconds}
EXEC
# Si count > limit → rechazar
```

### Cache-Aside
```
1. GET cache:entity:{id}
2. Si miss → consultar DB → SET cache:entity:{id} EX 300
3. En write → DEL cache:entity:{id} (invalidar)
```

## Drivers por Stack

| Stack | Librería | Notas |
|---|---|---|
| Go | `github.com/redis/go-redis/v9` | Context en todas las ops, `PoolSize` explícito |
| TypeScript | `ioredis` | Más robusto que `redis` oficial para clusters |
| Python | `redis-py` con `ConnectionPool` | Async: `redis.asyncio` |
| Rust | `fred` o `redis-rs` | `fred` para clusters, `redis-rs` para standalone |
