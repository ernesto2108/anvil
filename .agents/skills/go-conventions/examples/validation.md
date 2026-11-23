# Ejemplos de Validación

## Bien: Patrón de Validación en la Entidad

La validación pertenece a la entidad, no al servicio. Cada entidad de input tiene un método `Validate()` que limpia y valida sus propios campos.

```go
// domain/entities/workflow.go

type CreateWorkflow struct {
    TenantID    string
    Name        string
    Description string
    CreatedBy   string
}

func (c *CreateWorkflow) Validate() error {
    c.TenantID = strings.TrimSpace(c.TenantID)
    c.Name = strings.TrimSpace(c.Name)
    c.Description = strings.TrimSpace(c.Description)
    c.CreatedBy = strings.TrimSpace(c.CreatedBy)

    if c.TenantID == "" {
        return fmt.Errorf("tenant_id is required")
    }
    if c.Name == "" {
        return fmt.Errorf("name is required")
    }
    if c.CreatedBy == "" {
        return fmt.Errorf("created_by is required")
    }
    return nil
}
```

El servicio queda limpio — solo lógica de negocio:

```go
// application/create_workflow.go

func (s svc) CreateWorkflow(ctx context.Context, req entities.CreateWorkflow) error {
    if err := req.Validate(); err != nil {
        return err
    }
    // business logic only — no field validation here
    exists, err := s.db.ExistsWorkflowByName(ctx, req.TenantID, req.Name)
    if err != nil {
        return fmt.Errorf("check workflow exists: %w", err)
    }
    if exists {
        return fmt.Errorf("workflow %q already exists", req.Name)
    }
    return s.db.SaveWorkflow(ctx, req)
}
```

Para endpoints GET con parámetros de string crudos, crear una entidad filter:

```go
// domain/entities/filters.go

type GetByIDFilter struct {
    ID       string
    TenantID string
}

func (f *GetByIDFilter) Validate() error {
    f.ID = strings.TrimSpace(f.ID)
    f.TenantID = strings.TrimSpace(f.TenantID)

    if f.ID == "" {
        return fmt.Errorf("id is required")
    }
    if f.TenantID == "" {
        return fmt.Errorf("tenant_id is required")
    }
    return nil
}
```

**Flujo:** Handler (binding tags) → DTO.ToBusiness() → Entity.Validate() → Service (solo lógica de negocio)

**Por qué:** La validación está centralizada, es testeable de forma aislada y reutilizable entre servicios. El servicio tiene una sola responsabilidad: orquestar la lógica de negocio.

---

## Mal: Validación en la Capa de Servicio

Dispersar `strings.TrimSpace` + verificaciones `== ""` en los métodos del servicio es un error común. Duplica la validación, la hace inestable de forma aislada y filtra preocupaciones de input en la lógica de negocio.

```go
// BAD — validation scattered in every service method
func (s svc) CreateWorkflow(ctx context.Context, req entities.CreateWorkflow) error {
    req.TenantID = strings.TrimSpace(req.TenantID)
    req.Name = strings.TrimSpace(req.Name)
    req.CreatedBy = strings.TrimSpace(req.CreatedBy)
    if req.TenantID == "" { return fmt.Errorf("tenant_id is required") }
    if req.Name == "" { return fmt.Errorf("name is required") }
    if req.CreatedBy == "" { return fmt.Errorf("created_by is required") }
    return s.db.Save(ctx, req)
}

// BAD — same pattern repeated in get/update/delete/list services
func (s svc) GetByID(ctx context.Context, id, tenantID string) (*Entity, error) {
    id = strings.TrimSpace(id)
    tenantID = strings.TrimSpace(tenantID)
    if id == "" { return nil, fmt.Errorf("id is required") }
    if tenantID == "" { return nil, fmt.Errorf("tenant_id is required") }
    return s.db.GetByID(ctx, id, tenantID)
}
```

Por qué está mal:
- Duplicado en cada método — la misma verificación de TenantID en más de 10 lugares
- No se puede hacer unit test de la validación sin llamar al servicio
- El servicio tiene dos responsabilidades: validación + lógica de negocio
- Fácil de olvidar en nuevos métodos — cobertura inconsistente

Ver `examples/validation.md` → "Bien: Patrón de Validación en la Entidad" para el enfoque correcto.
