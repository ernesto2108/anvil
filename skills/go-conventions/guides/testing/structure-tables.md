# Test File Structure & Table-Driven Tests

## Test File Structure

Two common styles depending on project conventions:

### stdlib-only (no external test dependencies)

```go
package user_test  // use external test package for black-box testing

import (
    "context"
    "testing"

    "myapp/internal/user"
)

// Test helpers at the top
func newTestUser(t *testing.T, email string) *user.User {
    t.Helper()
    u, err := user.New(email, "Test User")
    if err != nil {
        t.Fatalf("newTestUser: %v", err)
    }
    return u
}

// Tests grouped by function/method
func Test_New(t *testing.T) { ... }
func Test_User_Activate(t *testing.T) { ... }
func Test_User_ChangeEmail(t *testing.T) { ... }
```

### testify (projects that use stretchr/testify)

```go
package user_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "myapp/internal/user"
)

func newTestUser(t *testing.T, email string) *user.User {
    t.Helper()
    u, err := user.New(email, "Test User")
    require.NoError(t, err)
    return u
}

func Test_New(t *testing.T) { ... }
func Test_User_Activate(t *testing.T) { ... }
```

- Use `_test` package suffix for black-box tests (tests the public API)
- Use same package for white-box tests only when testing unexported logic
- Name tests: `Test_FunctionName` or `Test_Type_Method` (underscore after Test)
- Choose one assertion style per project and be consistent

---

## Table-Driven Tests

The default pattern for any function with multiple input scenarios:

```go
func Test_ParseAmount(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    int64
        wantErr string // empty = no error expected
    }{
        {name: "valid dollars and cents", input: "12.50", want: 1250},
        {name: "whole number", input: "100", want: 10000},
        {name: "zero", input: "0", want: 0},
        {name: "negative", input: "-5.00", want: -500},
        {name: "too many decimals", input: "1.234", wantErr: "invalid amount"},
        {name: "not a number", input: "abc", wantErr: "invalid amount"},
        {name: "empty string", input: "", wantErr: "empty input"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseAmount(tt.input)

            if tt.wantErr != "" {
                // stdlib: use t.Fatalf/t.Errorf
                if err == nil {
                    t.Fatalf("expected error containing %q, got nil", tt.wantErr)
                }
                if !strings.Contains(err.Error(), tt.wantErr) {
                    t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
                }
                return

                // testify alternative:
                // require.Error(t, err)
                // assert.Contains(t, err.Error(), tt.wantErr)
            }

            // stdlib
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if got != tt.want {
                t.Errorf("ParseAmount(%q) = %d, want %d", tt.input, got, tt.want)
            }

            // testify alternative:
            // require.NoError(t, err)
            // assert.Equal(t, tt.want, got)
        })
    }
}
```

Rules:
- Every test case has a descriptive `name`
- stdlib: use `t.Fatalf` for fatal checks (stop on failure), `t.Errorf` for non-fatal
- testify: use `require` for fatal checks, `assert` for non-fatal
- Error cases check the error message/type, not just `err != nil`
- Keep test data inline when simple, use `testdata/` for complex fixtures
