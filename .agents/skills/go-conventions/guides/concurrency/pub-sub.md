# Pub/Sub (Difusión de Eventos)

**Cuándo:** Múltiples consumidores necesitan recibir el mismo evento (recarga de config, actualizaciones de precio, fanout de notificaciones).

**Escenario real:** Difundir cambios de configuración a todos los handlers HTTP activos.

```go
package main

import (
    "context"
    "fmt"
    "sync"
    "time"
)

type Broker[T any] struct {
    mu          sync.RWMutex
    subscribers map[int]chan T
    nextID      int
}

func NewBroker[T any]() *Broker[T] {
    return &Broker[T]{
        subscribers: make(map[int]chan T),
    }
}

// Subscribe retorna un channel y una función para cancelar la suscripción
func (b *Broker[T]) Subscribe(bufSize int) (<-chan T, func()) {
    b.mu.Lock()
    defer b.mu.Unlock()

    ch := make(chan T, bufSize)
    id := b.nextID
    b.nextID++
    b.subscribers[id] = ch

    unsubscribe := func() {
        b.mu.Lock()
        defer b.mu.Unlock()
        delete(b.subscribers, id)
        close(ch)
    }
    return ch, unsubscribe
}

// Publish envía el evento a todos los suscriptores (no bloqueante)
func (b *Broker[T]) Publish(event T) {
    b.mu.RLock()
    defer b.mu.RUnlock()

    for _, ch := range b.subscribers {
        select {
        case ch <- event:
        default:
            // El suscriptor es lento -- se descarta el evento (o se registra un warning)
        }
    }
}

func main() {
    broker := NewBroker[string]()

    // Suscriptor 1
    ch1, unsub1 := broker.Subscribe(10)
    defer unsub1()

    // Suscriptor 2
    ch2, unsub2 := broker.Subscribe(10)
    defer unsub2()

    // Goroutines consumidoras
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    var wg sync.WaitGroup
    consume := func(name string, ch <-chan string) {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for {
                select {
                case msg, ok := <-ch:
                    if !ok {
                        return
                    }
                    fmt.Printf("%s received: %s\n", name, msg)
                case <-ctx.Done():
                    return
                }
            }
        }()
    }

    consume("sub1", ch1)
    consume("sub2", ch2)

    // Publica eventos
    broker.Publish("config updated")
    broker.Publish("price changed")

    time.Sleep(100 * time.Millisecond)
    cancel()
    wg.Wait()
}
```

**Error común:** Bloquearse en publish cuando el channel de un suscriptor está lleno. Esto puede congelar al publisher y a todos los demás suscriptores. Usa siempre `select` con `default` para envíos no bloqueantes, o usa channels con buffer de capacidad apropiada.
