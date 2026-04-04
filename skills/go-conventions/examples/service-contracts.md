# Service Contract Patterns

## Full flow: Handler → Service → Repository

Each layer has a clear responsibility and type signature. Never leak concerns across layers.

### Layer 1: Handler (HTTP boundary)

Constructs domain entity from HTTP context/request, passes to service.

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

### Layer 2: Service Interface (port)

Receives typed entity, never raw primitives.

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

### Layer 3: Service Implementation (application)

Validates via entity, delegates to repo. If just a bridge — return directly.

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

### Layer 4: Repository (infrastructure)

Uses domain error codes, sql.Null* for scanning, context with timeout.

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

### Layer 5: Persistence DTO (scan struct)

ALL fields use sql.Null* types.

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

### Layer 6: Entity (domain)

Request entity with Validate() that normalizes and checks fields.

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

## Decision table: when to wrap vs return errors

| Layer | Error source | Pattern |
|-------|-------------|---------|
| Repository | DB driver error | `errors.New(errors.QueryErr, errors.WithError(err))` |
| Repository | Row scan error | `errors.New(errors.ScanErr, errors.WithError(err))` |
| Repository | No rows found | `errors.New(errors.NotFoundErr)` |
| Service | Validation failed | `return err` (entity.Validate already returns domain error) |
| Service | Repo returned error | `return err` (repo already mapped to domain error) |
| Service | Business logic error | `errors.New(errors.SomeBusinessErr)` |
| Handler | Binding/parse error | `errors.New(errors.BadRequestErr)` |
| Handler | Service returned error | `g.Error(err)` (middleware handles HTTP status) |
