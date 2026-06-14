# heapedcache

Fixed-size, thread-safe in-memory cache for Go. When the cache is full, the least recently refreshed item is evicted automatically. Originally from [emersonary/heapedcache](https://github.com/emersonary/heapedcache).

```
github.com/emersonary/appkit/heapedcache
```

## Features

- Fixed capacity with automatic eviction of oldest entries
- O(1) map lookups, O(log n) heap maintenance
- Thread-safe (`sync.RWMutex`)
- Generic IDs and object types
- Optional loader via `GetOrAdd`

## Usage

```go
import "github.com/emersonary/appkit/heapedcache"

type Person struct {
    ID   int
    Name string
}

cache := heapedcache.New[int, Person](1000)

cache.Push(42, &Person{ID: 42, Name: "Ada"})
person := cache.Get(42)

loaded := cache.GetOrAdd(99, func(id int) *Person {
    return &Person{ID: id, Name: "Grace"}
})

removed := cache.Remove(42)
oldest := cache.Pop()
size := cache.Len()
```

## API

| Method | Description |
|--------|-------------|
| `New[TId, TObj](maxRows)` | Create a cache with fixed capacity |
| `Push(id, item)` | Insert or update an item |
| `Get(id)` | Retrieve by ID, or `nil` |
| `GetOrAdd(id, fn)` | Retrieve, or load and store via `fn` |
| `Remove(id)` | Invalidate one entry |
| `Pop()` | Remove and return the oldest item |
| `PopWithRefreshed()` | Like `Pop`, also returns last refresh time |
| `Len()` | Current number of cached items |

## Tests

```bash
go test ./heapedcache/...
go test -short ./heapedcache/...   # skips 1M-item stress tests
```
