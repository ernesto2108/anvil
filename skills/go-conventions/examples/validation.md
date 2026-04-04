# Validation Examples

## Good: Entity Validation Pattern

Validation belongs in the entity, not in the service. Each input entity has a `Validate()` method that cleans and validates its own fields.

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

The service is clean — only business logic:

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

For GET endpoints with raw string parameters, create a filter entity:

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

**Flow:** Handler (binding tags) → DTO.ToBusiness() → Entity.Validate() → Service (business logic only)

**Why:** Validation is centralized, testable in isolation, reusable across services. The service has one responsibility: orchestrating business logic.

---

## Bad: Validation in Service Layer

Scattering `strings.TrimSpace` + `== ""` checks across service methods is a common mistake. It duplicates validation, makes it untestable in isolation, and leaks input concerns into business logic.

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

Why it's wrong:
- Duplicated in every method — same TenantID check in 10+ places
- Can't unit test validation without calling the service
- Service has two responsibilities: validation + business logic
- Easy to forget in new methods — inconsistent coverage

See `examples/validation.md` → "Good: Entity Validation Pattern" for the correct approach.
