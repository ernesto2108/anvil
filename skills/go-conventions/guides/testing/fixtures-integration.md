# Test Fixtures & Integration Tests

## Test Fixtures and testdata/

```
mypackage/
├── handler.go
├── handler_test.go
└── testdata/
    ├── valid_request.json
    └── invalid_request.json
```

```go
// Load fixture
func loadFixture(t *testing.T, name string) []byte {
    t.Helper()
    data, err := os.ReadFile(filepath.Join("testdata", name))
    if err != nil {
        t.Fatalf("load fixture %s: %v", name, err)
    }
    return data
}
```

---

## Integration Tests

Use build tags to separate integration tests:

```go
//go:build integration

package repo_test

func Test_UserRepo_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    db := setupTestDB(t)
    repo := NewUserRepo(db)

    t.Run("create and retrieve", func(t *testing.T) {
        user := newTestUser(t, "test@example.com")
        err := repo.Save(context.Background(), user)
        if err != nil {
            t.Fatalf("save: %v", err)
        }

        got, err := repo.GetByID(context.Background(), user.ID)
        if err != nil {
            t.Fatalf("get: %v", err)
        }
        if got.Email != user.Email {
            t.Errorf("email = %q, want %q", got.Email, user.Email)
        }
    })

    t.Run("duplicate email", func(t *testing.T) {
        user := newTestUser(t, "dup@example.com")
        if err := repo.Save(context.Background(), user); err != nil {
            t.Fatalf("save first: %v", err)
        }

        user2 := newTestUser(t, "dup@example.com")
        err := repo.Save(context.Background(), user2)
        if err == nil {
            t.Error("expected error for duplicate email, got nil")
        }
    })
}
```

Run: `go test ./... -tags integration`
