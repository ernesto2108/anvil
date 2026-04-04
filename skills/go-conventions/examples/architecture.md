# Architecture Examples

## Good: Small Interfaces Defined by Consumer

```go
// In the package that USES the dependency, not the one that implements it
package notification

// Only the methods this package actually calls
type UserFinder interface {
    GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)
}

type Service struct {
    users UserFinder
}
```

**Why:** Small interfaces are easier to mock, test, and swap. The consumer knows what it needs.

---

## Good: Constructor with Functional Options

```go
type Server struct {
    port         int
    readTimeout  time.Duration
    writeTimeout time.Duration
    logger       *slog.Logger
}

type Option func(*Server)

func WithPort(port int) Option {
    return func(s *Server) { s.port = port }
}

func WithTimeouts(read, write time.Duration) Option {
    return func(s *Server) {
        s.readTimeout = read
        s.writeTimeout = write
    }
}

func NewServer(logger *slog.Logger, opts ...Option) *Server {
    s := &Server{
        port:         8080,
        readTimeout:  5 * time.Second,
        writeTimeout: 10 * time.Second,
        logger:       logger,
    }
    for _, o := range opts {
        o(s)
    }
    return s
}
```

**Why:** Required dependencies are explicit parameters. Optional config uses functional options with sensible defaults.

---

## Bad: God Interfaces

```go
// BAD — forces implementors to implement everything, hard to mock
type UserService interface {
    Create(ctx context.Context, u *User) error
    Update(ctx context.Context, u *User) error
    Delete(ctx context.Context, id string) error
    GetByID(ctx context.Context, id string) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    List(ctx context.Context, filter Filter) ([]*User, error)
    Activate(ctx context.Context, id string) error
    Deactivate(ctx context.Context, id string) error
    ResetPassword(ctx context.Context, id string) error
    ChangeRole(ctx context.Context, id, role string) error
}

// GOOD — small interfaces, defined by consumer
// In the notification package:
type UserFinder interface {
    GetByID(ctx context.Context, id string) (*User, error)
}

// In the admin package:
type UserActivator interface {
    Activate(ctx context.Context, id string) error
    Deactivate(ctx context.Context, id string) error
}
```

---

## Bad: Global Mutable State

```go
// BAD — global DB connection, untestable, race-prone
var db *sql.DB

func init() {
    var err error
    db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        panic(err)
    }
}

func GetUser(id string) (*User, error) {
    return queryUser(db, id) // depends on global state
}

// GOOD — injected dependency
type UserRepo struct {
    db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
    return &UserRepo{db: db}
}

func (r *UserRepo) GetUser(ctx context.Context, id string) (*User, error) {
    return queryUser(ctx, r.db, id)
}
```

---

## Bad: Cross-Boundary Imports

```go
// BAD — domain imports from infrastructure
package domain

import "myapp/internal/infrastructure/postgres" // WRONG direction

type UserService struct {
    repo *postgres.UserRepo // domain depends on infra
}

// GOOD — domain defines interface, infra implements it
package domain

type UserRepository interface {
    Save(ctx context.Context, u *User) error
}

type UserService struct {
    repo UserRepository // depends on abstraction
}
```

---

## Bad: Unnecessary Pointer Fields

```go
// BAD — pointer fields when zero value is fine
type Config struct {
    Port    *int
    Host    *string
    Debug   *bool
    Timeout *time.Duration
}

// GOOD — use pointer only when nil has meaning (optional/nullable)
type Config struct {
    Port    int           // zero = use default
    Host    string        // empty = use default
    Debug   bool          // false = default
    Timeout time.Duration // zero = no timeout
}

// When nil DOES have meaning (truly optional)
type UserProfile struct {
    Name     string
    Nickname *string // nil = not set, "" = explicitly blank
}
```
