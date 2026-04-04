# Testing Repositories

Mock `sql.Rows` for unit-testing repository scan logic:

```go
type mockRows struct {
    data    [][]interface{}
    index   int
    scanErr error // inject Scan errors
    err     error // inject iteration errors
}

func (m *mockRows) Next() bool {
    m.index++
    return m.index <= len(m.data)
}

func (m *mockRows) Scan(dest ...interface{}) error {
    if m.scanErr != nil {
        return m.scanErr
    }
    row := m.data[m.index-1]
    for i, d := range dest {
        switch v := d.(type) {
        case *sql.NullInt64:
            if row[i] != nil {
                *v = sql.NullInt64{Int64: row[i].(int64), Valid: true}
            }
        case *sql.NullString:
            if row[i] != nil {
                *v = sql.NullString{String: row[i].(string), Valid: true}
            }
        case *sql.NullFloat64:
            if row[i] != nil {
                *v = sql.NullFloat64{Float64: row[i].(float64), Valid: true}
            }
        case *sql.NullTime:
            if row[i] != nil {
                *v = sql.NullTime{Time: row[i].(time.Time), Valid: true}
            }
        case *string:
            *v = row[i].(string)
        case *int64:
            *v = row[i].(int64)
        }
    }
    return nil
}

func (m *mockRows) Close() error { return nil }
func (m *mockRows) Err() error  { return m.err }

// Usage — test that Scan correctly maps sql.Null* columns to persistence DTOs
func Test_Repo_ScanUsers(t *testing.T) {
    rows := &mockRows{
        data: [][]interface{}{
            {"user-1", "alice@test.com", "Alice", time.Now()},
            {"user-2", "bob@test.com", nil, time.Now()}, // nullable name
        },
    }

    users, err := scanUsers(rows)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(users) != 2 {
        t.Fatalf("got %d users, want 2", len(users))
    }
    // Verify mapper extracted values correctly
    if users[0].Name != "Alice" {
        t.Errorf("users[0].Name = %q, want %q", users[0].Name, "Alice")
    }
    if users[1].Name != "" {
        t.Errorf("users[1].Name = %q, want empty (NULL)", users[1].Name)
    }
}
```

For error injection, add configurable error fields:

```go
type mockRows struct {
    data    [][]interface{}
    index   int
    scanErr error // inject Scan errors
    err     error // inject iteration errors
}

func (m *mockRows) Scan(dest ...interface{}) error {
    if m.scanErr != nil {
        return m.scanErr
    }
    // ... normal scan logic
}
```
