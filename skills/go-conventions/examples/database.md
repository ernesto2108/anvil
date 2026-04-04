# Database Examples

## Good: HTTP Handler — Clean Separation (Gin)

```go
// DTO for transport layer
type CreateUserRequest struct {
    Email string `json:"email" binding:"required,email"`
    Name  string `json:"name"  binding:"required,min=2"`
}

type CreateUserResponse struct {
    ID    string `json:"id"`
    Email string `json:"email"`
}

// Handler only does: bind → call service → respond
func (h *Handler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
        return
    }

    user, err := h.service.Create(c.Request.Context(), req.Email, req.Name)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, CreateUserResponse{
        ID:    user.ID,
        Email: user.Email,
    })
}
```

**Why:** Handler has no business logic. Gin's `ShouldBindJSON` handles decode + validation. DTOs prevent domain types from leaking into the API contract.

---

## Good: Repository Pattern — Query Separation + Persistence DTOs

### Query Function (in `queries/` package)

```go
// queries/user.go — each query is a pure function returning (string, []any, error)
func GetUserByID(id string) (string, []any, error) {
    query := `SELECT id, email, name, created_at FROM users WHERE id = $1`
    return query, []any{id}, nil
}

// Dynamic queries use strings.Builder + parameterized $N
func SearchUsers(filters SearchFilters) (string, []any, error) {
    var b strings.Builder
    var args []any
    argN := 1

    b.WriteString(`SELECT id, email, name, created_at FROM users WHERE 1=1`)

    if filters.Email != "" {
        b.WriteString(fmt.Sprintf(` AND email = $%d`, argN))
        args = append(args, filters.Email)
        argN++
    }
    if filters.Active != nil {
        b.WriteString(fmt.Sprintf(` AND active = $%d`, argN))
        args = append(args, *filters.Active)
        argN++
    }

    return b.String(), args, nil
}
```

### Persistence DTO + Mapper

```go
// dto/user.go — all fields are sql.Null* to absorb NULL values
type User struct {
    ID        sql.NullString
    Email     sql.NullString
    Name      sql.NullString
    CreatedAt sql.NullTime
}

// dto/user_mapper.go — ToBusiness extracts values from sql.Null* fields
func (u User) ToBusiness() entities.User {
    return entities.User{
        ID:        u.ID.String,
        Email:     u.Email.String,
        Name:      u.Name.String,
        CreatedAt: u.CreatedAt.Time,
    }
}

// Batch mapper for slices
func NewToBusiness(dtos []User) []entities.User {
    users := make([]entities.User, len(dtos))
    for i, d := range dtos {
        users[i] = d.ToBusiness()
    }
    return users
}
```

### Repository Method (query → execute → scan → map)

```go
type repository struct {
    client  PostgresSql    // custom DB interface, not *sql.DB
    timeout time.Duration
}

func NewRepository(client PostgresSql, timeout time.Duration) *repository {
    return &repository{client: client, timeout: timeout}
}

func (r *repository) GetByID(ctx context.Context, id string) (entities.User, error) {
    ctx, cancel := context.WithTimeout(ctx, r.timeout)
    defer cancel()

    query, args, err := queries.GetUserByID(id)
    if err != nil {
        return entities.User{}, err
    }

    var user dto.User
    err = r.client.QueryRowContext(ctx, query, args...).
        Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)

    switch {
    case err == nil:
        return user.ToBusiness(), nil
    case errors.Is(err, sql.ErrNoRows):
        return entities.User{}, ErrNotFound
    default:
        return entities.User{}, psql.PostgresError(err)
    }
}

// Transaction pattern — BeginTx + deferred rollback + commit
func (r *repository) Transfer(ctx context.Context, from, to string, amount int64) error {
    ctx, cancel := context.WithTimeout(ctx, r.timeout)
    defer cancel()

    tx, err := r.client.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback() // no-op after commit

    // ... execute queries within tx ...

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit transfer: %w", err)
    }
    return nil
}
```

**Why:** `WithTimeout` on every call. Query separation keeps SQL out of repo logic. Persistence DTO absorbs NULL. Mapper returns clean domain type. `PostgresError` translates DB-specific errors to domain errors.

---

## Bad: SQL Inline in Repository

```go
// BAD — SQL strings mixed with business logic, no persistence DTO, no error translation
func (r *repo) GetUser(ctx context.Context, id string) (*User, error) {
    var u User
    err := r.db.QueryRowContext(ctx,
        `SELECT id, email FROM users WHERE id = $1`, id,
    ).Scan(&u.ID, &u.Email)
    return &u, err
}

// GOOD — queries separated, persistence DTO absorbs NULL, mapper keeps domain clean
func (r *repository) GetUser(ctx context.Context, id string) (entities.User, error) {
    ctx, cancel := context.WithTimeout(ctx, r.timeout)
    defer cancel()

    query, args, err := queries.GetUser(id)
    if err != nil {
        return entities.User{}, err
    }

    var user dto.User
    err = r.client.QueryRowContext(ctx, query, args...).
        Scan(&user.ID, &user.Email)
    if err != nil {
        return entities.User{}, psql.PostgresError(err)
    }

    return user.ToBusiness(), nil
}
```
