# Safe Concurrent Map Access

**When:** Multiple goroutines reading/writing to a shared map.

**Real scenario:** In-memory cache accessed by HTTP handlers.

## Using sync.RWMutex (recommended for most cases)

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

## Using sync.Map (specific use cases only)

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

**Common mistake:** Using a bare `map` across goroutines. Maps are NOT safe for concurrent writes. The runtime will panic with "concurrent map writes" (if you are lucky) or silently corrupt data. Always protect with a mutex or use `sync.Map`.
