# Pipeline (Etapa1 -> Etapa2 -> Etapa3)

**Cuándo:** Los datos fluyen a través de etapas de transformación secuenciales, cada una potencialmente concurrente.

**Escenario real:** Pipeline ETL: leer filas de CSV -> validar/transformar -> inserción en lotes en base de datos.

```go
package main

import (
    "context"
    "fmt"
    "strings"
)

type RawRow struct {
    Line int
    Data string
}

type ValidRow struct {
    Line int
    Name string
    Age  int
}

type InsertResult struct {
    Line int
    OK   bool
}

// Etapa 1: Leer datos crudos
func readRows(ctx context.Context, lines []string) <-chan RawRow {
    out := make(chan RawRow)
    go func() {
        defer close(out)
        for i, line := range lines {
            select {
            case out <- RawRow{Line: i + 1, Data: line}:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

// Etapa 2: Validar y transformar (puede ejecutar múltiples workers)
func validate(ctx context.Context, in <-chan RawRow) <-chan ValidRow {
    out := make(chan ValidRow)
    go func() {
        defer close(out)
        for row := range in {
            parts := strings.SplitN(row.Data, ",", 2)
            if len(parts) != 2 {
                continue // omitir filas inválidas
            }
            select {
            case out <- ValidRow{Line: row.Line, Name: parts[0]}: // simplificado
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

// Etapa 3: Inserción en lotes
func batchInsert(ctx context.Context, in <-chan ValidRow) <-chan InsertResult {
    out := make(chan InsertResult)
    go func() {
        defer close(out)
        for row := range in {
            // Simula inserción en DB
            select {
            case out <- InsertResult{Line: row.Line, OK: true}:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    lines := []string{"Alice,30", "Bob,25", "bad-line", "Carol,28"}

    // Conecta el pipeline: leer -> validar -> insertar
    raw := readRows(ctx, lines)
    valid := validate(ctx, raw)
    results := batchInsert(ctx, valid)

    for r := range results {
        fmt.Printf("line %d: ok=%v\n", r.Line, r.OK)
    }
}
```

**Regla clave del blog de Go:** "Las etapas cierran sus canales de salida cuando todas las operaciones de envío están completas. Las etapas continúan recibiendo de los canales de entrada hasta que esos canales se cierran o los remitentes se desbloquean." (Fuente: [Go Concurrency Patterns: Pipelines](https://go.dev/blog/pipelines))

**Error común:** No cerrar los channels. Si la etapa 2 nunca cierra su channel de salida, la etapa 3 se bloquea en `range` indefinidamente. Cada goroutine que es dueña de un channel debe hacer `defer close(out)`.
