# Patrones de Base de Datos

Estos patrones reflejan la capa de acceso a DB real usada en los proyectos:

- **Funciones de query en el paquete `queries/`** — cada query es una función que retorna `(string, []any, error)`. Las queries complejas usan `strings.Builder` con placeholders `$N` parametrizados
- **DTOs de persistencia** — structs con `sql.Null*` para TODOS los campos (`NullString`, `NullInt64`, `NullFloat64`, `NullTime`), separados de las entidades de dominio. Viven en el paquete `dto/` o `persistence/`
- **Mappers** — método `ToBusiness()` en un DTO individual, función por lotes `NewToBusiness()` para slices. Extraer `.String`, `.Int64`, `.Time`, etc. de los campos `sql.Null*`
- **Struct del repositorio** — tiene `client` (interfaz DB personalizada) + `timeout time.Duration`. Cada método llama `context.WithTimeout(ctx, r.timeout)` con cancel diferido
- **Interfaz DB** — interfaz personalizada `PostgresSql` que envuelve `*sql.DB` con su propia interfaz `Rows` para testabilidad. Nunca depender de `*sql.DB` directamente en los repositorios
- **Traducción de errores** — `PostgresError(err)` traduce códigos `pq.Error` a errores de dominio (duplicate key → conflict, foreign key → not found, etc.)
- **Transacciones** — `BeginTx` + rollback diferido + commit explícito al final. El rollback es no-op después del commit
- **Dos capas de DTO** — DTOs HTTP/input (json tags, binding tags) vs DTOs de persistencia/output (campos sql.Null*). Nunca mezclarlos

Flujo del método de repositorio: `WithTimeout → query() → execute → scan into DTO → map to domain`

Ver `examples/good-patterns.md` para ejemplos de código completos.

## Migraciones — `golang-migrate`

Usar `github.com/golang-migrate/migrate/v4` para todas las migraciones de base de datos. Nunca escribir `CREATE TABLE` ad-hoc o DDL de schema en el código de la aplicación.

- **Embeber migraciones** con `//go:embed migrations/*.sql` para que se incluyan dentro del binario
- **Naming de archivos:** `<number>_<action>_<target>.up.sql` / `.down.sql` (ej., `000001_create_users.up.sql`)
- **Un cambio por par** — un par up/down hace una sola cosa lógica
- **El agente DBA es dueño de los archivos de migración** — el desarrollador NO crea ni modifica archivos `.up.sql`/`.down.sql`. Si tu tarea necesita cambios de schema, decirle al orchestrator que invoque al DBA primero
- **Reglas específicas del motor** — ver `/db-engines` para PRAGMAs, configuración de drivers, y peculiaridades de migración por motor
