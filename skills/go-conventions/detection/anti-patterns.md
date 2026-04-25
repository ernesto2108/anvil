# Anti-Patrones Go — Referencia de Detección

## Detección Pasiva

Al revisar código Go, escanear estos patrones y reportar usando el formato:
`[file:line] [severity] [category] anti-pattern-name`

Reportar solo `error` y `warning` por defecto. Reportar `suggestion` solo cuando el usuario pide mejorar/refactorizar/optimizar.

## Tabla de Anti-Patrones

| Patrón de código | Anti-patrón | Severidad | Categoría | Corrección → Patrón |
|---|---|---|---|---|
| `panic()` fuera de `main()` | panic-in-library | error | reliability | Retornar `error` — ver Error Handling |
| `init()` haciendo trabajo real | hidden-init | error | reliability | Constructor injection — ver Patterns > Constructor Functions |
| `_ = f.Close()` o error ignorado | ignored-error | error | reliability | Manejar o loguear — ver Error Handling |
| `var db *sql.DB` a nivel de paquete | global-mutable-state | error | concurrency | Inyectar via constructor — ver Architecture Rules #7 |
| `setInterval`/ticker sin stop | resource-leak | error | memory | `defer ticker.Stop()` en el setup |
| `sync.Mutex` protegiendo operaciones de canal | double-sync | error | concurrency | Usar un solo mecanismo — ver Concurrency |
| `defer` en el cuerpo de un loop | deferred-in-loop | error | memory | Cerrar explícitamente en el cuerpo del loop |
| `return err` desnudo de errores crudos/sin tipo | unwrapped-error | warning | errors | Mapear a `errors.New(errors.SomeCode)` o `fmt.Errorf("op: %w", err)` en pkgs standalone — ver Error Handling |
| `fmt.Errorf` envolviendo un `*Errors` ya mapeado | double-wrapped-error | warning | errors | Solo `return err` — la infra ya lo mapeó — ver Error Handling |
| `context.Background()` en handlers | missing-context | warning | reliability | Usar `r.Context()` — ver Context |
| `context.Context` almacenado en struct | stored-context | warning | reliability | Pasar como primer parámetro — ver Context |
| `interface{}` / `any` en el dominio | untyped-domain | warning | types | Tipos concretos o generics — ver domain-entity-guardrails |
| Interfaz dios (>5 métodos) | god-interface | warning | design | Interfaces pequeñas por consumidor — ver Architecture Rules #2 |
| Importaciones circulares | circular-import | warning | design | Extraer interfaz compartida — ver Architecture Rules #5 |
| >3 niveles de anidado if/else | deep-nesting | warning | readability | Guard clauses — ver Patterns > Guard Clauses |
| Función larga (>5 parámetros) | param-bloat | warning | readability | Options struct — ver Patterns > Functional Options |
| Tests sin subtests `t.Run()` | flat-tests | warning | testing | Table-driven con subtests — ver Testing Patterns |
| Tests dependientes del orden de ejecución | coupled-tests | warning | testing | Setup independiente por test — ver Testing Patterns |
| Estado de test compartido entre `t.Run` | shared-test-state | warning | testing | Cada subtest crea sus propios fixtures |
| `time.Sleep` en tests | sleep-in-tests | warning | testing | Canales, sync, o polling con timeout |
| Falta `t.Helper()` en helpers | missing-t-helper | suggestion | testing | Agregar `t.Helper()` en la primera línea |
| Exportar símbolos no usados | over-export | suggestion | design | No exportar lo que no se usa externamente |
| Paquete `log` en vez de `slog` | unstructured-logging | suggestion | observability | Usar `slog` (stdlib) |
| String typing para enums | string-enum | suggestion | types | `type Status int` con `iota` |
| `reflect` para tareas simples | unnecessary-reflect | suggestion | performance | Type switches, generics, o código concreto |
| Strings de error con mayúscula/punto | error-format | suggestion | style | Minúsculas, sin puntuación al final |
| Nombre de paquete `users` o `user_service` | bad-package-name | suggestion | style | Corto, singular, sin underscores: `user` |
| `math/rand` para tokens/keys/sesiones | insecure-random | error | security | Usar `crypto/rand` — ver security-guide.md |
| `fmt.Sprintf` con input de usuario en SQL | sql-injection | error | security | Queries parametrizadas con `$N` — ver security-guide.md |
| Sin límite de tamaño en request body | missing-body-limit | warning | security | `http.MaxBytesReader(w, r.Body, limit)` — ver security-guide.md |
| Sin endpoint `/healthz` o `/readyz` | missing-health-check | warning | observability | Agregar liveness + readiness — ver observability-guide.md |
| `os.Getenv` profundo en el call stack | scattered-config | warning | design | Cargar config una vez en `main()`, inyectar via constructor |
| Logueando PII/secretos (password, token) | logged-secrets | error | security | Implementar `LogValuer`, redactar campos — ver slog-guide.md |
| `strings.TrimSpace` + `== ""` en capa de servicio | validation-in-service | warning | architecture | Mover a método `entity.Validate()` — ver Architecture Rules #8 |
| Método de servicio con >2 validaciones de campo antes de lógica de negocio | scattered-validation | warning | architecture | Crear entidad input con `Validate()` — ver Architecture Rules #8 |
| `g.Param("id")` con literal de string inline | magic-param-string | warning | architecture | Usar constante `dto.ParamXxx` de `dto/constants.go` — ver Architecture Rules #9 |
| `g.Query("status")` con literal de string inline | magic-query-string | warning | architecture | Usar constante `dto.QueryXxx` de `dto/constants.go` — ver Architecture Rules #9 |
| TrimSpace + verificación vacía para parámetros de ruta URL en capa de aplicación | param-validation-wrong-layer | warning | architecture | Validar params de ruta en el handler, no en el servicio — ver Architecture Rules #9 |
| `errors.WithMessage("foo: " + variable)` concatenación de strings | concat-in-error-message | warning | style | Usar `fmt.Sprintf("foo: %s", variable)` — nunca concatenar con `+` en mensajes de error |
| `http.Get()`/`http.Post()` (sin timeout, sin context) | http-no-timeout | error | reliability | `http.NewRequestWithContext(ctx, ...)` — ver context-cleanup-guide.md |
| `http.DefaultClient` sin Timeout | default-client | warning | reliability | `&http.Client{Timeout: 15*time.Second}` — ver context-cleanup-guide.md |
| `db.Query()` sin Context | query-no-context | error | reliability | `db.QueryContext(ctx, ...)` — ver context-cleanup-guide.md |
| `Query()` para INSERT/UPDATE/DELETE | query-for-exec | error | memory | Usar `ExecContext()` — Query retorna rows que deben cerrarse |
| Falta `rows.Close()` después de QueryContext | unclosed-rows | error | memory | `defer rows.Close()` inmediatamente después de la verificación de error |
| Falta `rows.Err()` después del loop de iteración | unchecked-rows-err | warning | reliability | Siempre verificar `rows.Err()` después de `for rows.Next()` |
| `defer` en el cuerpo de un loop | defer-in-loop-cleanup | error | memory | Extraer a función helper — ver context-cleanup-guide.md |
| Falta config de pool en `sql.DB` (MaxOpenConns) | no-pool-config | warning | reliability | Configurar MaxOpenConns, MaxIdleConns, ConnMaxLifetime |
| `resp.Body.Close()` antes de verificar error | close-before-check | error | crashes | Verificar error primero, luego `defer resp.Body.Close()` |
| `func AsError(err error, target interface{}) bool { return errors.As(err, &target) }` | errors-as-double-pointer | error | crashes | `target` ya es un puntero dentro de `interface{}` — `&target` crea `*interface{}` que `errors.As` no puede desenvolver. Usar `errors.As(err, &customErr)` directamente en el call site, nunca envolverlo |
| Error de servicio externo descartado: `if err != nil { return errors.New(domainErr) }` sin loguear `err` | swallowed-external-error | warning | observability | Loguear el error original antes de retornar el error de dominio: `log.Error("service X failed", log.WithError(err))` — de lo contrario el debugging es imposible |
| Método de interfaz de servicio recibe >1 parámetro `string` crudo (ej. `GetStats(ctx, tenantID string, status string)`) | service-accepts-raw-strings | warning | architecture | Crear un struct de entidad request/filter con `Validate()`. El servicio recibe la entidad, no strings crudos. Ver `examples/service-contracts.md` — coincide con el flujo `DTO → Entity.ToBusiness() → Service(entity) → entity.Validate()` |
| Tipo llamado `*DTO` dentro del paquete `domain/entities/` | dto-in-domain | warning | naming | `DTO` pertenece a los límites de transporte (`http/dto/`, `psql/dto/`). El dominio usa nombres descriptivos: `Detail`, `Summary`, `Filter` — ver Architecture Rules #11 |
| Struct de agregado de dominio duplica campos de entidades hijas en vez de componerlos | flattened-aggregate | warning | architecture | Componer con tipos de entidad existentes: `OrderDetail{Order, []Item}` no copia campo por campo — ver Architecture Rules #10 |
| Struct DTO de persistencia (en `infrastructure/output/persistencia/*/dto/`) usa `string`, `int` o `time.Time` planos en vez de tipos `sql.Null*` | dto-without-sql-null | warning | architecture | TODOS los campos en DTOs de persistencia deben usar `sql.NullString`, `sql.NullInt64`, `sql.NullTime`, etc. El método `ToBusiness()` del mapper extrae los valores reales. Los tipos planos causan bugs silenciosos de valor-cero en NULL provenientes de JOINs, COALESCE o GROUPING SETS |
