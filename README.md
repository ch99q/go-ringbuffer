# RingBuffer

A high-performance, thread-safe, fixed-size circular buffer for Go.

## Features

- **High Performance**: Optimized for speed with lock-free reads and minimal contention
- **Thread-Safe**: Safely access from multiple goroutines concurrently
- **Zero Allocations**: Core operations don't allocate memory after initialization
- **Type Safe**: Uses Go generics for type safety
- **Power-of-2 Optimization**: Uses bitwise operations instead of modulo for better performance

## Benchmarks

This ring buffer offers excellent performance compared to other implementations, with different trade-offs:

```
// Standard Go slice vs. RingBuffer (element access)
BenchmarkSliceComparison/SliceAccess-large-10         	1000000000	         0.8261 ns/op	       0 B/op	       0 allocs/op
BenchmarkSliceComparison/RingBufferAccess-large-10             648359515	         1.852 ns/op	       0 B/op	       0 allocs/op

// container/ring vs. RingBuffer (element access)
BenchmarkGet/StdRing-large-10                                  2043105	       582.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/RingBuffer-large-10                               647848177	         1.853 ns/op	       0 B/op	       0 allocs/op

// Standard Go slice with mutex vs. RingBuffer (concurrent access)
BenchmarkSliceComparison/SliceConcurrent-large-10              14611782	        80.59 ns/op	       0 B/op	       0 allocs/op
BenchmarkSliceComparison/RingBufferConcurrent-large-10         19577772	        61.49 ns/op	       0 B/op	       0 allocs/op

// Batch operations - memory allocations
BenchmarkSliceComparison/SliceBatch-large-10                   1988011	       603.4 ns/op	    6118 B/op	       0 allocs/op
BenchmarkSliceComparison/RingBufferBatch-large-10              4963058	       248.7 ns/op	       0 B/op	       0 allocs/op
```

Full benchmark results are available in the [benchmark.txt](./benchmark.txt) file.

### Element Access
- **Up to 314x faster** than container/ring for element access on large buffers (582.0 ns/op vs 1.853 ns/op)
- **Over 26x faster** than smallnest/ringbuffer for read operations (49.48 ns/op vs 1.860 ns/op)
- **Constant-time access** regardless of buffer size
- **Zero memory allocations** for all read operations
- Only **~2.2x slower** than raw slice access, while providing thread safety (0.8261 ns/op vs 1.852 ns/op)

### Concurrent Performance
- **~24% faster** than mutex-protected slices for concurrent access (80.59 ns/op vs 61.49 ns/op)
- **Consistent ~60ns/op** performance regardless of buffer size
- Built-in thread safety with no additional code required

### Memory Efficiency
- **Zero allocations** for write operations vs. variable allocations for slices
- For large buffer batch operations: **0 bytes allocated** vs. **6118 bytes** for slices
- **~2.4x faster** for batch operations on large buffers compared to slices (603.4 ns/op vs 248.7 ns/op)

### When to Use
- When you need a **thread-safe circular buffer** without external synchronization
- When you want **near-slice performance** with built-in thread safety
- When **memory efficiency** is important for batch operations
- When you need a **fixed-size ring buffer** with wrap-around semantics

### Comparison with Other Implementations
- **vs raw Go slices**: ~2.3x slower for basic access, but thread-safe and zero allocations for writes
- **vs container/ring**: Dramatically faster for all operations (~318x for reads)
- **vs smallnest/ringbuffer**: ~25x faster for reads, slightly slower for writes, with the advantage of type safety through generics

## Installation

```bash
go get github.com/ch99q/go-ringbuffer
```

## Usage

### Creating a Ring Buffer

```go
import "github.com/ch99q/go-ringbuffer"

// Create with initial data (size will be rounded up to next power of 2)
data := []int{1, 2, 3, 4, 5}
rb := ringbuffer.New(data)

// Create with specified capacity (will be rounded up to next power of 2)
rb = ringbuffer.NewWithCapacity[int](100)
```

### Adding Elements

```go
// Add a single element
rb.Add(42)

// Add multiple elements at once (more efficient)
rb.AddAll([]int{10, 20, 30, 40, 50})
```

### Accessing Elements

```go
// Get the newest element (most recently added)
newest := rb.Get(0)

// Get the second newest element
secondNewest := rb.Get(1)

// Get multiple elements in reverse chronological order (newest first)
// Returns elements from newest to oldest, matching Get's behavior
recentItems := rb.GetRange(10) // Get the 10 most recent items
```

### Advanced Usage

```go
// Modify an element at a specific position
rb.Set(0, 100) // Set the newest element to 100

// Get the capacity
capacity := rb.Len()
```

## Performance Considerations

- **Power-of-2 Size Optimization**: The buffer's size is always a power of 2, which allows the use of bitwise AND operations (`&`) instead of modulo (`%`) for index wrapping. This is significantly faster on most CPU architectures and is a key factor in the impressive performance of this implementation.
- **Lock-Free Reads**: Reads are lock-free and can happen concurrently with writes
- `AddAll` is more efficient than multiple `Add` calls for bulk operations
- **Constant-time O(1) access** regardless of buffer size

## Thread Safety

The ring buffer is fully thread-safe:
- Multiple readers can access the buffer concurrently
- Writers are synchronized, so only one write can happen at a time
- Reads and writes can happen concurrently

## Contributing
Contributions are welcome! Please open an issue or submit a pull request for any bugs, features, or improvements.

### Benchmarks
To run the benchmarks, use the following command:

````
go test ./benchmark -bench=Benchmark -run='^$' -benchmem > benchmark.txt
````

This will generate a `benchmark.txt` file with the results.

## License

[MIT License](LICENSE)