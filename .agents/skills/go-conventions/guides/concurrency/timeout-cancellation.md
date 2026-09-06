# Timeout y Cancelación con Context

**Cuándo:** Toda llamada externa (HTTP, DB, gRPC, file I/O) en código de producción.

**Escenario real:** Handler HTTP que llama 3 microservicios, debe responder en máximo 5 segundos en total.

```go
func handleOrder(w http.ResponseWriter, r *http.Request) {
    // Hereda el contexto de la request (cancelado cuando el cliente se desconecta)
    ctx := r.Context()

    // Agrega timeout global para este handler
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel() // Siempre defer cancel

    g, ctx := errgroup.WithContext(ctx)

    var (
        user    User
        inv     Inventory
        pricing Price
    )

    g.Go(func() error {
        var err error
        user, err = fetchUser(ctx, r.FormValue("user_id"))
        return err
    })
    g.Go(func() error {
        var err error
        inv, err = checkInventory(ctx, r.FormValue("item_id"))
        return err
    })
    g.Go(func() error {
        var err error
        pricing, err = getPrice(ctx, r.FormValue("item_id"))
        return err
    })

    if err := g.Wait(); err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            http.Error(w, "request timed out", http.StatusGatewayTimeout)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Usa user, inv, pricing para construir la respuesta...
    json.NewEncoder(w).Encode(map[string]any{
        "user": user.Name, "available": inv.Qty, "price": pricing.Amount,
    })
}
```

## Reglas de Context (del [blog de Go: Context](https://go.dev/blog/context)):
- Pasa el context como primer parámetro a toda función en el call path
- Nunca almacenes un context en un struct
- Usa `context.Background()` solo en `main()`, `init()`, y configuración de tests de alto nivel
- Siempre `defer cancel()` después de `WithTimeout` / `WithCancel` / `WithDeadline`
- Usa tipos de clave no exportados para valores de context para evitar colisiones

```go
// Valores de context con type-safety
type ctxKey int
const requestIDKey ctxKey = 0

func WithRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey, id)
}

func RequestID(ctx context.Context) string {
    id, _ := ctx.Value(requestIDKey).(string)
    return id
}
```

**Error común:** Llamar cancel en un context recibido como parámetro. Solo el creador de un context derivado debe llamar su función cancel. Las sub-operaciones deben retornar errores, no cancelar contextos padre.
