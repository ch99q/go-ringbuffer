package ringbuffer

import (
	"sync"
	"testing"

	ringbuffer "github.com/ch99q/go-ringbuffer"
)

// BenchmarkSliceComparison compares ring buffer operations with standard Go slice operations
func BenchmarkSliceComparison(b *testing.B) {
	for _, size := range benchSizes {
		// Benchmark element access (index lookup)
		b.Run("SliceAccess-"+itoa(size), func(b *testing.B) {
			// Create a slice with the same data
			data := make([]int, size)
			for i := 0; i < size; i++ {
				data[i] = i
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				idx := i % size
				_ = data[idx]
			}
		})

		b.Run("RingBufferAccess-"+itoa(size), func(b *testing.B) {
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

		// Benchmark element update
		b.Run("SliceUpdate-"+itoa(size), func(b *testing.B) {
			data := make([]int, size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				idx := i % size
				data[idx] = i
			}
		})

		b.Run("RingBufferUpdate-"+itoa(size), func(b *testing.B) {
			rb := ringbuffer.NewWithCapacity[int](size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rb.Set(i%size, i)
			}
		})

		// Benchmark append operation (similar to Add)
		b.Run("SliceAppend-"+itoa(size), func(b *testing.B) {
			data := make([]int, 0, size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if len(data) >= size {
					// Simulate circular behavior by removing the oldest element
					data = append(data[1:], i)
				} else {
					data = append(data, i)
				}
			}
		})

		b.Run("RingBufferAdd-"+itoa(size), func(b *testing.B) {
			rb := ringbuffer.NewWithCapacity[int](size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rb.Add(i)
			}
		})

		// Benchmark batch append (similar to AddAll)
		batchSize := size / 4
		b.Run("SliceBatch-"+itoa(size), func(b *testing.B) {
			data := make([]int, 0, size)
			batch := make([]int, batchSize)
			for i := 0; i < batchSize; i++ {
				batch[i] = i + 1000
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if len(data)+len(batch) > size {
					if len(batch) >= size {
						// If batch is bigger than or equal to slice capacity, just keep the latest items
						data = append([]int{}, batch[len(batch)-size:]...)
					} else {
						// Keep as many old elements as will fit with the new batch
						data = append(data[len(data)+len(batch)-size:], batch...)
					}
				} else {
					data = append(data, batch...)
				}
			}
		})

		b.Run("RingBufferBatch-"+itoa(size), func(b *testing.B) {
			rb := ringbuffer.NewWithCapacity[int](size)
			batch := make([]int, batchSize)
			for i := 0; i < batchSize; i++ {
				batch[i] = i + 1000
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rb.AddAll(batch)
			}
		})

		// Benchmark getting a subrange (similar to GetRange)
		rangeSize := size / 4
		b.Run("SliceRange-"+itoa(size), func(b *testing.B) {
			data := make([]int, size)
			for i := 0; i < size; i++ {
				data[i] = i
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				start := size - rangeSize
				result := make([]int, rangeSize)
				copy(result, data[start:])
			}
		})

		b.Run("RingBufferRange-"+itoa(size), func(b *testing.B) {
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

		// Benchmark concurrent access (slice with mutex vs ringbuffer)
		b.Run("SliceConcurrent-"+itoa(size), func(b *testing.B) {
			data := make([]int, size)
			for i := 0; i < size; i++ {
				data[i] = i
			}
			mu := sync.RWMutex{}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				counter := 0
				for pb.Next() {
					if counter%2 == 0 {
						// Read operation
						mu.RLock()
						_ = data[counter%size]
						mu.RUnlock()
					} else {
						// Write operation
						mu.Lock()
						data[counter%size] = counter
						mu.Unlock()
					}
					counter++
				}
			})
		})

		b.Run("RingBufferConcurrent-"+itoa(size), func(b *testing.B) {
			data := make([]int, size)
			for i := 0; i < size; i++ {
				data[i] = i
			}
			rb := ringbuffer.New(data)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				counter := 0
				for pb.Next() {
					if counter%2 == 0 {
						// Read operation
						_ = rb.Get(counter % size)
					} else {
						// Write operation
						rb.Add(counter)
					}
					counter++
				}
			})
		})
	}
}
