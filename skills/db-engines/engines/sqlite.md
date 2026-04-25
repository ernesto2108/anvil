# SQLite

## Limitaciones
- `ALTER TABLE` solo puede `ADD COLUMN` — sin DROP, sin RENAME (antes de 3.35), sin ALTER TYPE
- Sin tipo `ENUM` — usa constraints `CHECK`: `status TEXT CHECK(status IN ('active','inactive'))`
- Sin creación de índices concurrente — sin opción `CONCURRENTLY`
- Sin tipo `UUID` nativo — almacena como `TEXT`
- Sin tipo `TIMESTAMP` — almacena como `TEXT` en formato RFC3339
- `AUTOINCREMENT` es opcional y más lento que `INTEGER PRIMARY KEY` (que auto-incrementa por defecto)
- Para cambios de esquema más allá de ADD COLUMN: **patrón de 4 pasos** — crea nueva tabla, copia datos, elimina la antigua, renombra

## PRAGMAs de Producción (OBLIGATORIOS al abrir conexión)

Aplica estos en orden en cada nueva conexión:

```sql
PRAGMA journal_mode = WAL;          -- concurrent reads, non-blocking writes
PRAGMA synchronous = NORMAL;        -- safe in WAL mode, avoids FSYNC on every write
PRAGMA busy_timeout = 5000;         -- wait up to 5s on lock contention instead of failing
PRAGMA foreign_keys = ON;           -- enforce FK constraints (OFF by default!)
PRAGMA cache_size = -65536;         -- 64MB page cache (default ~2MB is too small)
PRAGMA temp_store = MEMORY;         -- temp tables/indices in memory
PRAGMA mmap_size = 268435456;       -- 256MB memory-mapped I/O for read performance
```

## Detalles del Modo WAL
- WAL habilita lecturas concurrentes mientras una escritura está en progreso — crítico para cualquier acceso multi-goroutine o multi-proceso
- Auto-checkpoint en 1000 páginas (por defecto) — suficiente para herramientas CLI/desktop
- Para cargas de trabajo con alta escritura: `PRAGMA wal_autocheckpoint = <pages>` o `PRAGMA wal_checkpoint(TRUNCATE)` manual en shutdown graceful
- Siempre hace checkpoint antes del backup (`sqlite3 .backup`)
- El modo WAL persiste entre conexiones — configúralo una vez, se mantiene

## Patrones de Rotación de Datos
- Usa `ON DELETE CASCADE` en todos los FKs hijos para que un solo `DELETE FROM parent` limpie los hijos
- Rotación basada en cantidad: `DELETE FROM parent WHERE id NOT IN (SELECT id FROM parent ORDER BY created_at DESC LIMIT ?)` — tamaño de DB predecible
- Ejecuta `PRAGMA optimize` periódicamente (al cerrar la app) para mantener estadísticas del query planner
- `VACUUM` recupera espacio después de eliminaciones grandes — pero bloquea la DB, así que ejecútalo durante ventanas de mantenimiento o en shutdown

## Quirks de Migración
- Las migraciones NO deben contener `BEGIN` / `COMMIT` explícitos — `golang-migrate` y la mayoría de herramientas envuelven cada archivo en una transacción implícita
- `CREATE TABLE IF NOT EXISTS` para primera migración idempotente
- `ON DELETE CASCADE` debe declararse en la creación de la tabla — no se puede agregar después via ALTER
- Incluye sentencias `CREATE INDEX` en la misma migración que la tabla que indexan
- Las foreign keys NO se aplican a menos que `PRAGMA foreign_keys = ON` — esto es por conexión, no persistente
- Al usar `golang-migrate` con `WithInstance(db, ...)`: NO llames `m.Close()` — cierra el `*sql.DB` compartido. Deja que el llamador maneje el ciclo de vida de la conexión
- Cuando las tablas tienen relaciones FK, ordena las migraciones para que las tablas padre se creen primero (ej., `000001_create_runs`, `000002_create_agents` donde agents referencia runs)

## Fuentes de migración: `iofs` vs `file://` (CRÍTICO para binarios distribuidos)

`golang-migrate` soporta múltiples fuentes. La elección importa mucho para la distribución.

### Fuente `file://`
- Lee archivos `.sql` del disco en tiempo de ejecución
- Funciona solo cuando las migraciones viven en el filesystem del host — típicamente durante tests o desarrollo local
- **Falla en cualquier binario distribuido sin el repositorio** — el usuario no tiene `./migrations/` en su máquina

```go
m, err := migrate.NewWithDatabaseInstance(
    "file://"+migrationsPath, "sqlite3", driver,
)
```

### Fuente `iofs` (requerida para binarios distribuidos)
- Lee migraciones de cualquier `fs.FS` — incluyendo `embed.FS`
- Las migraciones se convierten en parte del binario via `//go:embed`
- Esta es la ÚNICA fuente que funciona para un binario que el usuario final descarga

```go
import (
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/sqlite3"
    "github.com/golang-migrate/migrate/v4/source/iofs"
)

// migrations is an embed.FS rooted at the migrations directory
src, err := iofs.New(migrations, ".")
if err != nil {
    return fmt.Errorf("create iofs source: %w", err)
}

driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
if err != nil {
    return fmt.Errorf("create sqlite3 driver: %w", err)
}

m, err := migrate.NewWithInstance("iofs", src, "sqlite3", driver)
if err != nil {
    return fmt.Errorf("create migrate instance: %w", err)
}
```

### Regla — diseña para iofs desde el primer día
Si el store alguna vez se distribuirá en un binario (CLI tool, desktop app, servidor), **usa `iofs` desde la primera migración**. Ofrecer AMBAS fuentes está bien (los tests pueden seguir usando `file://` por ergonomía), pero la ruta de producción por defecto debe ser `iofs`.

**Patrón recomendado — dos constructores compartiendo un helper:**
```go
// Shared DB setup — no duplication
func openDB(dbPath string) (*sql.DB, error) { /* ... */ }

// Production: embedded migrations
func NewFS(dbPath string, migrations fs.FS, subPath string, maxRuns int) (*Store, error) {
    db, err := openDB(dbPath)
    if err != nil {
        return nil, err
    }
    sub, err := fs.Sub(migrations, subPath)
    if err != nil {
        _ = db.Close()
        return nil, fmt.Errorf("resolve sub-fs: %w", err)
    }
    if err := RunMigrationsFS(db, sub); err != nil {
        _ = db.Close()
        return nil, err
    }
    return &Store{db: db}, nil
}

// Tests / dev: disk migrations
func New(dbPath, migrationsPath string, maxRuns int) (*Store, error) { /* uses file:// */ }
```

### Anti-patrón — `file://` en código de producción
Si el comando CLI abre el store con `store.New(dbPath, filepath.Join(home, ".app", "migrations"), ...)`, el binario fallará en el primer run para cualquier usuario que no tenga un directorio de migraciones instalado por separado. Esto es una **brecha de diseño**, no un bug — detéctala en la revisión de arquitectura, no en runtime.

## Driver de Go
- `github.com/mattn/go-sqlite3` — el único driver `database/sql` de grado producción (requiere CGO)
- Connection string: `file:<path>?_journal_mode=WAL&_foreign_keys=on&_busy_timeout=5000`
- Patrón de escritor único: abre un `*sql.DB` con `SetMaxOpenConns(1)` para escrituras
- Para cargas de trabajo con muchas lecturas: `*sql.DB` separado para lecturas con `SetMaxOpenConns(max(4, runtime.NumCPU()))`
