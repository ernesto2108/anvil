# Go Patterns

## Functional Options

```go
type Option func(*Server)

func WithPort(port int) Option {
    return func(s *Server) { s.port = port }
}

func NewServer(opts ...Option) *Server {
    s := &Server{port: 8080} // defaults
    for _, o := range opts {
        o(s)
    }
    return s
}
```

## Constructor Functions

```go
func NewUserService(repo UserRepository, logger *slog.Logger) *UserService {
    return &UserService{
        repo:   repo,
        logger: logger,
    }
}
```

## Guard Clauses

```go
// bad — nested
func Process(u *User) error {
    if u != nil {
        if u.Active {
            // ... deep nesting
        }
    }
    return nil
}

// good — guard clauses
func Process(u *User) error {
    if u == nil {
        return ErrNilUser
    }
    if !u.Active {
        return ErrInactiveUser
    }
    // ... flat logic
    return nil
}
```

## Graceful Shutdown

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

srv := &http.Server{Addr: ":8080", Handler: mux}
go func() { srv.ListenAndServe() }()

<-ctx.Done()
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
srv.Shutdown(shutdownCtx)
```
