# Helpers de Test y Mocking con Interfaces

## Helpers de Test

```go
// Siempre marcar los helpers con t.Helper() — los errores reportan la línea del caller
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

// Factory helpers para objetos de test comunes
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

- `t.Helper()` en cada helper — hace que el output de errores apunte al test, no al helper
- `t.Cleanup()` en lugar de `defer` — ejecuta después del test sin importar el anidamiento de subtests
- Las factory functions usan functional options para mayor flexibilidad

---

## Mocking con Interfaces

Sin frameworks de mocking. Definir interfaces, implementar test doubles escritos a mano.

### Patrón A: Function-pointer fakes

Ideal para control fino — configurar comportamiento por caso de test:

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

### Patrón B: Embedding + panic stubs

Ideal para seguridad en tiempo de compilación — garantiza que se noten los nuevos métodos de interfaz:

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

**Cuándo usar cuál:**
- Function-pointer fakes → control por caso de test, nil = no-op por defecto
- Embedding + panic stubs → seguridad en tiempo de compilación, los cambios de interfaz rompen rápido
