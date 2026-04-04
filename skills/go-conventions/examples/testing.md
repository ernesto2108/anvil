# Testing Examples

## Good: Table-Driven Test with Subtests

```go
func Test_ValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {name: "valid email", email: "user@example.com", wantErr: false},
        {name: "missing @", email: "userexample.com", wantErr: true},
        {name: "missing domain", email: "user@", wantErr: true},
        {name: "empty string", email: "", wantErr: true},
        {name: "unicode local", email: "usuario@ejemplo.com", wantErr: false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if tt.wantErr && err == nil {
                t.Error("expected error, got nil")
            }
            if !tt.wantErr && err != nil {
                t.Errorf("unexpected error: %v", err)
            }
        })
    }
}
```

**Why:** Each case is a data row. Adding new cases is one line. Subtests give you `go test -run Test_ValidateEmail/missing_@`.

---

## Bad: Tests Without Structure — Sequential Assertions

```go
// BAD — sequential assertions, can't isolate failures or run selectively
func Test_CalculatePrice(t *testing.T) {
    result := CalculatePrice(100, 0.1)
    if result != 90 {
        t.Errorf("got %d want 90", result)
    }
    result = CalculatePrice(200, 0.5)
    if result != 100 {
        t.Errorf("got %d want 100", result)
    }
    result = CalculatePrice(0, 0.1)
    if result != 0 {
        t.Errorf("got %d want 0", result)
    }
}

// GOOD — table-driven with subtests, isolate each case
func Test_CalculatePrice(t *testing.T) {
    tests := []struct {
        name     string
        price    int
        discount float64
        want     int
    }{
        {name: "10% off 100", price: 100, discount: 0.1, want: 90},
        {name: "50% off 200", price: 200, discount: 0.5, want: 100},
        {name: "zero price", price: 0, discount: 0.1, want: 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := CalculatePrice(tt.price, tt.discount)
            if got != tt.want {
                t.Errorf("CalculatePrice(%d, %.1f) = %d, want %d", tt.price, tt.discount, got, tt.want)
            }
        })
    }
}
```
