# Fan-Out / Fan-In

**Cuándo:** Tienes N elementos independientes a procesar y quieres paralelizarlos entre workers, luego recopilar todos los resultados.

**Escenario real:** Obtener datos de enriquecimiento desde una API externa para 1000 registros.

```go
package main

import (
    "context"
    "fmt"
    "sync"
)

// Etapa 1: Generator -- produce elementos de trabajo
func generate(ctx context.Context, items []string) <-chan string {
    out := make(chan string)
    go func() {
        defer close(out)
        for _, item := range items {
            select {
            case out <- item:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

// Etapa 2: Worker -- procesa un elemento a la vez
func process(ctx context.Context, in <-chan string) <-chan string {
    out := make(chan string)
    go func() {
        defer close(out)
        for item := range in {
            // Simula llamada API o cómputo
            result := "processed:" + item
            select {
            case out <- result:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

// Fan-in: combina múltiples channels en uno
func merge(ctx context.Context, channels ...<-chan string) <-chan string {
    out := make(chan string)
    var wg sync.WaitGroup

    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan string) {
            defer wg.Done()
            for val := range c {
                select {
                case out <- val:
                case <-ctx.Done():
                    return
                }
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()
    return out
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    items := []string{"a", "b", "c", "d", "e", "f"}
    in := generate(ctx, items)

    // Fan-out: 3 workers leyendo desde el mismo channel
    w1 := process(ctx, in)
    w2 := process(ctx, in)
    w3 := process(ctx, in)

    // Fan-in: combina resultados
    for result := range merge(ctx, w1, w2, w3) {
        fmt.Println(result)
    }
}
```

**Error común:** Olvidar `select` con `ctx.Done()` en los envíos por channel. Sin esto, si el consumidor deja de leer, los productores se bloquean indefinidamente (goroutine leak). Cada envío de channel dentro de una goroutine debe estar envuelto en un select con un case de cancelación.
