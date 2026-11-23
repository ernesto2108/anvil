# Worker Pool (Concurrencia Acotada)

**Cuándo:** Tienes muchas tareas pero necesitas limitar la ejecución concurrente (límites de rate de API, pools de conexiones DB, restricciones de memoria).

**Escenario real:** Procesar 10,000 registros de base de datos con máximo 20 llamadas HTTP concurrentes.

## Usando errgroup.SetLimit (preferido para código nuevo)

```go
package main

import (
    "context"
    "fmt"
    "time"

    "golang.org/x/sync/errgroup"
)

type Record struct {
    ID   int
    Name string
}

func processRecord(ctx context.Context, r Record) error {
    // Simula llamada a API externa
    select {
    case <-time.After(100 * time.Millisecond):
        fmt.Printf("processed record %d\n", r.ID)
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func processAll(ctx context.Context, records []Record) error {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(20) // Máximo 20 goroutines concurrentes

    for _, r := range records {
        r := r // captura la variable del loop (no requerido en Go 1.22+)
        g.Go(func() error {
            return processRecord(ctx, r)
        })
    }

    return g.Wait() // Retorna el primer error; cancela ctx en caso de error
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    records := make([]Record, 100)
    for i := range records {
        records[i] = Record{ID: i, Name: fmt.Sprintf("item-%d", i)}
    }

    if err := processAll(ctx, records); err != nil {
        fmt.Printf("failed: %v\n", err)
    }
}
```

## Usando channels (patrón clásico)

```go
func workerPool(ctx context.Context, jobs []Record, numWorkers int) error {
    jobsCh := make(chan Record)
    errCh := make(chan error, 1) // con buffer: gana el primer error
    var wg sync.WaitGroup

    // Inicia un número fijo de workers
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobsCh {
                if err := processRecord(ctx, job); err != nil {
                    select {
                    case errCh <- err: // envía el primer error
                    default: // ya hay un error
                    }
                    return
                }
            }
        }()
    }

    // Envía los jobs
    go func() {
        defer close(jobsCh)
        for _, job := range jobs {
            select {
            case jobsCh <- job:
            case <-ctx.Done():
                return
            }
        }
    }()

    wg.Wait()
    close(errCh)
    return <-errCh
}
```

**Error común:** Crear una goroutine por elemento sin límite. 10,000 goroutines haciendo llamadas HTTP agotarán los descriptores de archivo y saturarán los servicios downstream. Siempre acota la concurrencia.
