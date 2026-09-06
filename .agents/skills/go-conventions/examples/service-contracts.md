# Patrones de Contratos de Servicio

## Flujo completo: Handler → Service → Repository

Cada capa tiene una responsabilidad clara y una firma de tipos definida. Nunca filtrar preocupaciones entre capas.

### Capa 1: Handler (límite HTTP)

Construye la entidad de dominio a partir del contexto/request HTTP, la pasa al servicio.

```go
func (h *Handler) GetDashboardStats(g *gin.Context) {
    ctx := g.Request.Context()

    tenantID, err := middleware.TenantIDFromContext(g)
    if err != nil {
        g.Errors = append(g.Errors, g.Error(err))
        return
    }

    // Handler constructs the entity — service never sees raw strings
    request := entities.GetDashboardStatsRequest{TenantID: tenantID}

    stats, err := h.svc.GetDashboardStats(ctx, request)
    if err != nil {
        g.Errors = append(g.Errors, g.Error(err))
        return
    }

    g.JSON(http.StatusOK, dto.NewDashboardStatsResponse(stats))
}
```

### Capa 2: Interfaz de Servicio (puerto)

Recibe entidad tipada, nunca primitivos crudos.

```go
// GOOD — service receives entity
type DashboardServiceInterface interface {
    GetDashboardStats(ctx context.Context, request entities.GetDashboardStatsRequest) (entities.DashboardStats, error)
}

// BAD — service receives raw strings
type DashboardServiceInterface interface {
    GetDashboardStats(ctx context.Context, tenantID string) (entities.DashboardStats, error)
}
```

### Capa 3: Implementación de Servicio (aplicación)

Valida a través de la entidad, delega al repo. Si es solo un puente — retornar directamente.

```go
// GOOD — bridge pattern, no unnecessary wrapping
func (s *svc) GetDashboardStats(ctx context.Context, request entities.GetDashboardStatsRequest) (entities.DashboardStats, error) {
    if err := request.Validate(); err != nil {
        return entities.DashboardStats{}, err
    }

    return s.db.GetDashboardStats(ctx, request.TenantID)
}

// BAD — wraps error that repo already mapped
func (s *svc) GetDashboardStats(ctx context.Context, request entities.GetDashboardStatsRequest) (entities.DashboardStats, error) {
    if err := request.Validate(); err != nil {
        return entities.DashboardStats{}, err
    }

    stats, err := s.db.GetDashboardStats(ctx, request.TenantID)
    if err != nil {
        return entities.DashboardStats{}, fmt.Errorf("get dashboard stats: %w", err)  // WRONG
    }

    return stats, nil
}
```

### Capa 4: Repositorio (infraestructura)

Usa códigos de error de dominio, sql.Null* para scanning, contexto con timeout.

```go
func (r repository) GetDashboardStats(ctx context.Context, tenantID string) (entities.DashboardStats, error) {
    ctx, cancel := context.WithTimeout(ctx, r.timeout)
    defer cancel()

    query, args := queries.GetDashboardStats(tenantID)

    rows, err := r.client.QueryContext(ctx, query, args...)
    if err != nil {
        return entities.DashboardStats{}, errors.New(errors.QueryErr, errors.WithError(err))
    }

    defer rows.Close() //nolint:errcheck

    // ... scan into sql.Null* DTO, map to entity
}
```

### Capa 5: DTO de Persistencia (struct de scan)

TODOS los campos usan tipos sql.Null*.

```go
// GOOD
type DashboardStatsRow struct {
    TotalWorkflows sql.NullInt64
    StatusGroup    sql.NullString
    Count          sql.NullInt64
}

// BAD — plain types cause silent zero-value bugs on NULL
type DashboardStatsRow struct {
    TotalWorkflows int
    StatusGroup    string
    Count          int
}
```

### Capa 6: Entidad (dominio)

Entidad de request con Validate() que normaliza y verifica los campos.

```go
type GetDashboardStatsRequest struct {
    TenantID string
}

func (r *GetDashboardStatsRequest) Validate() error {
    r.TenantID = strings.TrimSpace(r.TenantID)
    if r.TenantID == "" {
        return errors.New(errors.BadRequestErr, errors.WithMessage("tenant_id is required"))
    }
    return nil
}
```

## Tabla de decisión: cuándo envolver vs retornar errores

| Capa | Fuente del error | Patrón |
|-------|-------------|---------|
| Repository | Error del driver DB | `errors.New(errors.QueryErr, errors.WithError(err))` |
| Repository | Error de scan de fila | `errors.New(errors.ScanErr, errors.WithError(err))` |
| Repository | No rows found | `errors.New(errors.NotFoundErr)` |
| Service | Validación fallida | `return err` (entity.Validate ya retorna error de dominio) |
| Service | Error retornado por repo | `return err` (repo ya mapeó a error de dominio) |
| Service | Error de lógica de negocio | `errors.New(errors.SomeBusinessErr)` |
| Handler | Error de binding/parse | `errors.New(errors.BadRequestErr)` |
| Handler | Error retornado por servicio | `g.Error(err)` (middleware maneja el status HTTP) |
