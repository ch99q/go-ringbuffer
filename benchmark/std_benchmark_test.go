package ringbuffer

import (
	"container/ring"
	"testing"

	ringbuffer "github.com/ch99q/go-ringbuffer"
)

// BenchmarkComparison runs 1:1 comparison benchmarks between our RingBuffer and Go's container/ring package.
// These benchmarks are designed to be as fair as possible, performing equivalent operations on both implementations.

// Benchmark sizes to test with
var benchSizes = []int{16, 128, 1024}

// setupStdRing creates a container/ring.Ring filled with integers 0 to size-1
func setupStdRing(size int) *ring.Ring {
	r := ring.New(size)
	for i := 0; i < size; i++ {
		r.Value = i
		r = r.Next()
	}
	return r
}

// BenchmarkGet compares element access between implementations
func BenchmarkGet(b *testing.B) {
	for _, size := range benchSizes {
		b.Run("StdRing-"+itoa(size), func(b *testing.B) {
			r := setupStdRing(size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Move to a position based on i (to avoid caching effects)
				pos := i % size
				temp := r
				for j := 0; j < pos; j++ {
					temp = temp.Next()
				}
				_ = temp.Value
			}
		})

		b.Run("RingBuffer-"+itoa(size), func(b *testing.B) {
			// Create initial data
			data := make([]int, size)
			for i := 0; i < size; i++ {
				data[i] = i
			}
			rb := ringbuffer.New(data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = rb.Get(i % size)
			}
		})
	}
}

// BenchmarkAdd compares adding elements between implementations
func BenchmarkAdd(b *testing.B) {
	for _, size := range benchSizes {
		b.Run("StdRing-"+itoa(size), func(b *testing.B) {
			r := setupStdRing(size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// In container/ring, we need to replace a value in an existing slot and move forward
				r.Value = i
				r = r.Next()
			}
		})

		b.Run("RingBuffer-"+itoa(size), func(b *testing.B) {
			data := make([]int, size)
			rb := ringbuffer.New(data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rb.Add(i)
			}
		})
	}
}

// BenchmarkIterate compares iterating through all elements in both implementations
func BenchmarkIterate(b *testing.B) {
	for _, size := range benchSizes {
		b.Run("StdRing-"+itoa(size), func(b *testing.B) {
			r := setupStdRing(size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var sum int
				r.Do(func(v interface{}) {
					if val, ok := v.(int); ok {
						sum += val
					}
				})
			}
		})

		b.Run("RingBuffer-"+itoa(size), func(b *testing.B) {
			// Create initial data
			data := make([]int, size)
			for i := 0; i < size; i++ {
				data[i] = i
			}
			rb := ringbuffer.New(data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var sum int
				for j := 0; j < rb.Len(); j++ {
					sum += rb.Get(j)
				}
			}
		})
	}
}

// BenchmarkBatchOperations compares batch operations where possible
func BenchmarkBatchOperations(b *testing.B) {
	for _, size := range benchSizes {
		batchSize := size / 4

		b.Run("StdRing-"+itoa(size), func(b *testing.B) {
			r := setupStdRing(size)
			batch := make([]int, batchSize)
			for i := 0; i < batchSize; i++ {
				batch[i] = i + 1000
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Standard ring doesn't have batch operations, so do individual operations
				for _, val := range batch {
					r.Value = val
					r = r.Next()
				}
			}
		})

		b.Run("RingBuffer-"+itoa(size), func(b *testing.B) {
			data := make([]int, size)
			rb := ringbuffer.New(data)
			batch := make([]int, batchSize)
			for i := 0; i < batchSize; i++ {
				batch[i] = i + 1000
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rb.AddAll(batch)
			}
		})
	}
}

// BenchmarkGetRange compares retrieving a range of elements
func BenchmarkGetRange(b *testing.B) {
	for _, size := range benchSizes {
		rangeSize := size / 4

		b.Run("StdRing-"+itoa(size), func(b *testing.B) {
			r := setupStdRing(size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Create a slice from the ring (no direct equivalent in std ring)
				result := make([]interface{}, rangeSize)
				temp := r
				for j := 0; j < rangeSize; j++ {
					result[j] = temp.Value
					temp = temp.Next()
				}
			}
		})

		b.Run("RingBuffer-"+itoa(size), func(b *testing.B) {
			data := make([]int, size)
			for i := 0; i < size; i++ {
				data[i] = i
			}
			rb := ringbuffer.New(data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result := rb.GetRange(rangeSize)
				_ = result
			}
		})
	}
}

// BenchmarkConcurrentAccess tests concurrent reads and writes
// Note: container/ring is not thread-safe, so we only test our RingBuffer
func BenchmarkConcurrentAccess(b *testing.B) {
	for _, size := range benchSizes {
		b.Run("RingBuffer-"+itoa(size), func(b *testing.B) {
			data := make([]int, size)
			rb := ringbuffer.New(data)

			// Pre-fill with some data
			for i := 0; i < size*2; i++ {
				rb.Add(i)
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				counter := 0
				for pb.Next() {
					if counter%2 == 0 {
						// Half the goroutines read
						_ = rb.Get(counter % size)
					} else {
						// Half the goroutines write
						rb.Add(counter)
					}
					counter++
				}
			})
		})
	}
}

// Helper function to convert int to string
func itoa(i int) string {
	switch i {
	case 16:
		return "small"
	case 128:
		return "medium"
	case 1024:
		return "large"
	default:
		return "custom"
	}
}
