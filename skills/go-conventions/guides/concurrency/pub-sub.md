# Pub/Sub (Event Broadcasting)

**When:** Multiple consumers need to receive the same event (config reload, price updates, notification fanout).

**Real scenario:** Broadcasting config changes to all active HTTP handlers.

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

// Subscribe returns a channel and an unsubscribe function
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

// Publish sends event to all subscribers (non-blocking)
func (b *Broker[T]) Publish(event T) {
    b.mu.RLock()
    defer b.mu.RUnlock()

    for _, ch := range b.subscribers {
        select {
        case ch <- event:
        default:
            // Subscriber is slow -- drop event (or log warning)
        }
    }
}

func main() {
    broker := NewBroker[string]()

    // Subscriber 1
    ch1, unsub1 := broker.Subscribe(10)
    defer unsub1()

    // Subscriber 2
    ch2, unsub2 := broker.Subscribe(10)
    defer unsub2()

    // Consumer goroutines
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

    // Publish events
    broker.Publish("config updated")
    broker.Publish("price changed")

    time.Sleep(100 * time.Millisecond)
    cancel()
    wg.Wait()
}
```

**Common mistake:** Blocking on publish when a subscriber's channel is full. This can freeze the publisher and all other subscribers. Always use `select` with `default` for non-blocking send, or use buffered channels with appropriate capacity.
