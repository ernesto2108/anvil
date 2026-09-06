# Acceso Seguro a Mapas Concurrentes

**Cuándo usar:** Múltiples goroutines leyendo/escribiendo en un mapa compartido.

**Escenario real:** Caché en memoria accedida por handlers HTTP.

## Usando sync.RWMutex (recomendado para la mayoría de los casos)

```go
type Cache struct {
    mu    sync.RWMutex
    items map[string]Item
}

func NewCache() *Cache {
    return &Cache{items: make(map[string]Item)}
}

func (c *Cache) Get(key string) (Item, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    item, ok := c.items[key]
    return item, ok
}

func (c *Cache) Set(key string, item Item) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = item
}

func (c *Cache) Delete(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.items, key)
}

// GetOrSet atomically checks and sets -- holds write lock entire time
func (c *Cache) GetOrSet(key string, create func() Item) Item {
    c.mu.Lock()
    defer c.mu.Unlock()
    if item, ok := c.items[key]; ok {
        return item
    }
    item := create()
    c.items[key] = item
    return item
}
```

## Usando sync.Map (solo casos de uso específicos)

```go
// sync.Map is optimized for two specific cases:
// 1. Key set is stable (read-heavy, rarely written)
// 2. Multiple goroutines read/write disjoint key sets
//
// For everything else, use map + sync.RWMutex.

var cache sync.Map

// Store
cache.Store("key", value)

// Load
if val, ok := cache.Load("key"); ok {
    item := val.(Item) // Type assertion required
}

// LoadOrStore (atomic check-and-set)
actual, loaded := cache.LoadOrStore("key", newItem)
```

**Error común:** Usar un `map` desnudo entre goroutines. Los mapas NO son seguros para escrituras concurrentes. El runtime entrará en pánico con "concurrent map writes" (si tienes suerte) o corromperá los datos silenciosamente. Siempre proteger con un mutex o usar `sync.Map`.
