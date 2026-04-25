# Guía de Cobertura y Tests de Benchmark

## Guía de Cobertura

**Qué cubrir:**
- Lógica de negocio y reglas de dominio
- Caminos de error y casos borde
- Validación de entrada
- Transiciones de estado

**Qué NO obsesionarse:**
- Getters/setters simples
- Código de wire-up / DI
- Código generado
- Wrappers de librerías de terceros (testear vía integración)

Meta: 80%+ en paquetes de lógica de negocio. No perseguir el 100% — los retornos son decrecientes.

Verificar cobertura: `go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out`

---

## Tests de Benchmark

```go
func BenchmarkParseAmount(b *testing.B) {
    for b.Loop() {
        ParseAmount("12345.67")
    }
}

// Con sub-benchmarks
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

Ejecutar: `go test -bench=. -benchmem ./...`
