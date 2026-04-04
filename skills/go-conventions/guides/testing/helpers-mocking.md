# Test Helpers & Mocking with Interfaces

## Test Helpers

```go
// Always mark helpers with t.Helper() — errors report the caller's line
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()

    db, err := sql.Open("postgres", testDSN)
    if err != nil {
        t.Fatalf("open test db: %v", err)
    }

    t.Cleanup(func() {
        db.Close()
    })

    return db
}

// Factory helpers for common test objects
func newTestOrder(t *testing.T, opts ...func(*Order)) *Order {
    t.Helper()
    o := &Order{
        ID:     uuid.New(),
        Status: StatusPending,
        Amount: 1000,
    }
    for _, opt := range opts {
        opt(o)
    }
    return o
}
```

- `t.Helper()` on every helper — makes error output point to the test, not the helper
- `t.Cleanup()` instead of `defer` — runs after the test regardless of subtest nesting
- Factory functions use functional options for flexibility

---

## Mocking with Interfaces

No mocking frameworks. Define interfaces, implement hand-written test doubles.

### Pattern A: Function-pointer fakes

Best for fine-grained control — configure behavior per test case:

```go
// Production interface (defined by consumer)
type UserRepository interface {
    Save(ctx context.Context, u *User) error
    GetByID(ctx context.Context, id string) (*User, error)
}

// Function-pointer fake — each field controls one method
type repoFake struct {
    saveFn    func(ctx context.Context, u *User) error
    getByIDFn func(ctx context.Context, id string) (*User, error)
}

func (f repoFake) Save(ctx context.Context, u *User) error {
    if f.saveFn == nil {
        return nil
    }
    return f.saveFn(ctx, u)
}

func (f repoFake) GetByID(ctx context.Context, id string) (*User, error) {
    if f.getByIDFn == nil {
        return nil, nil
    }
    return f.getByIDFn(ctx, id)
}

// Usage
func Test_CreateUser_success(t *testing.T) {
    repo := repoFake{
        saveFn: func(_ context.Context, u *User) error {
            if u.Email == "" {
                t.Error("expected email to be set")
            }
            return nil
        },
    }
    svc := NewUserService(repo)
    err := svc.Create(context.Background(), "test@example.com", "Test")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}

func Test_CreateUser_repoError(t *testing.T) {
    repo := repoFake{
        saveFn: func(_ context.Context, _ *User) error {
            return errors.New("mock error")
        },
    }
    svc := NewUserService(repo)
    err := svc.Create(context.Background(), "test@example.com", "Test")
    if err == nil {
        t.Fatal("expected error, got nil")
    }
}
```

### Pattern B: Embedding + panic stubs

Best for compile-time safety — ensures new interface methods are noticed:

```go
// Base stub that panics on unimplemented methods
type serviceMock struct{}

func (s serviceMock) Create(ctx context.Context, email, name string) (*User, error) {
    panic("implement me")
}
func (s serviceMock) GetByID(ctx context.Context, id string) (*User, error) {
    panic("implement me")
}

// Override only the methods you need — compiler catches missing ones
type createMock struct {
    serviceMock
    createResp *User
    createErr  error
}

func (m createMock) Create(_ context.Context, _, _ string) (*User, error) {
    return m.createResp, m.createErr
}

// Usage
func Test_Handler_Create_success(t *testing.T) {
    mock := createMock{
        createResp: &User{ID: "123", Email: "test@example.com"},
    }
    h := NewHandler(mock)
    // ... test handler
}
```

**When to use which:**
- Function-pointer fakes → control per test case, nil = no-op default
- Embedding + panic stubs → compile-time safety, interface changes break fast
