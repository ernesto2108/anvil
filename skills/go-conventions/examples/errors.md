# Error Handling Examples

## Good: Error Handling with Context

```go
// Wrap errors with operation context — callers can still use errors.Is/As
func (r *pgUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
    var u User
    err := r.db.QueryRowContext(ctx,
        `SELECT id, email, name FROM users WHERE id = $1`, id,
    ).Scan(&u.ID, &u.Email, &u.Name)

    if errors.Is(err, sql.ErrNoRows) {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("get user %s: %w", id, err)
    }
    return &u, nil
}
```

**Why:** Every error includes what operation failed. `%w` preserves the chain for `errors.Is` upstream.

---

## Bad: Panic in Library Code

```go
// BAD — crashes the entire application
func MustParseConfig(path string) Config {
    data, err := os.ReadFile(path)
    if err != nil {
        panic(err) // caller has no chance to recover gracefully
    }
    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        panic(err)
    }
    return cfg
}

// GOOD — return error, let caller decide
func ParseConfig(path string) (Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return Config{}, fmt.Errorf("read config %s: %w", path, err)
    }
    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return Config{}, fmt.Errorf("parse config %s: %w", path, err)
    }
    return cfg, nil
}
```

---

## Bad: Bare Error Returns

```go
// BAD — no context about what failed
func (s *OrderService) PlaceOrder(ctx context.Context, req OrderReq) error {
    user, err := s.users.Get(ctx, req.UserID)
    if err != nil {
        return err // which operation failed? Get user? Get product? Save order?
    }
    product, err := s.products.Get(ctx, req.ProductID)
    if err != nil {
        return err
    }
    return s.orders.Save(ctx, NewOrder(user, product))
}

// GOOD — each error includes operation context
func (s *OrderService) PlaceOrder(ctx context.Context, req OrderReq) error {
    user, err := s.users.Get(ctx, req.UserID)
    if err != nil {
        return fmt.Errorf("get user %s: %w", req.UserID, err)
    }
    product, err := s.products.Get(ctx, req.ProductID)
    if err != nil {
        return fmt.Errorf("get product %s: %w", req.ProductID, err)
    }
    if err := s.orders.Save(ctx, NewOrder(user, product)); err != nil {
        return fmt.Errorf("save order for user %s: %w", req.UserID, err)
    }
    return nil
}
```

---

## Bad: Ignoring Errors

```go
// BAD — silently discarding errors
json.Unmarshal(data, &result)
f.Close()
resp.Body.Close()

// GOOD — handle or explicitly acknowledge
if err := json.Unmarshal(data, &result); err != nil {
    return fmt.Errorf("unmarshal response: %w", err)
}
defer func() {
    if err := f.Close(); err != nil {
        slog.Warn("close file", "err", err)
    }
}()
```
