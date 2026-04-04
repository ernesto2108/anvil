# Coverage Guidelines & Benchmark Tests

## Coverage Guidelines

**What to cover:**
- Business logic and domain rules
- Error paths and edge cases
- Input validation
- State transitions

**What NOT to obsess over:**
- Simple getters/setters
- Wire-up / DI code
- Generated code
- Third-party library wrappers (test via integration)

Target: 80%+ on business logic packages. Don't chase 100% — diminishing returns.

Check coverage: `go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out`

---

## Benchmark Tests

```go
func BenchmarkParseAmount(b *testing.B) {
    for b.Loop() {
        ParseAmount("12345.67")
    }
}

// With sub-benchmarks
func BenchmarkHash(b *testing.B) {
    sizes := []int{64, 256, 1024, 4096}
    for _, size := range sizes {
        data := make([]byte, size)
        b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
            for b.Loop() {
                Hash(data)
            }
        })
    }
}
```

Run: `go test -bench=. -benchmem ./...`
