package ringbuffer

import (
	"testing"

	ringbuffer "github.com/ch99q/go-ringbuffer"
	smallnest "github.com/smallnest/ringbuffer"
)

// Benchmark sizes for the tests
var smallnestBenchSizes = []int{16, 128, 1024}

// BenchmarkBytes compares byte operations between implementations
func BenchmarkByteOperations(b *testing.B) {
	for _, size := range smallnestBenchSizes {
		// Benchmark Write operation
		b.Run("SmallnestWrite-"+itoa(size), func(b *testing.B) {
			rb := smallnest.New(size)
			data := []byte("x")
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Write can fail if buffer is full, but we ignore that for benchmark
				rb.Write(data)
			}
		})

		b.Run("RingBufferAdd-"+itoa(size), func(b *testing.B) {
			rb := ringbuffer.NewWithCapacity[byte](size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rb.Add(byte('x'))
			}
		})

		// Benchmark Read operation
		b.Run("SmallnestRead-"+itoa(size), func(b *testing.B) {
			rb := smallnest.New(size)
			// Pre-fill the buffer
			data := make([]byte, size/2)
			for i := 0; i < size/2; i++ {
				data[i] = byte(i % 256)
			}
			rb.Write(data)

			readBuf := make([]byte, 1)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Rewind to avoid emptying the buffer
				if rb.Length() < 1 {
					rb.Write(data)
				}
				rb.Read(readBuf)
			}
		})

		b.Run("RingBufferGet-"+itoa(size), func(b *testing.B) {
			// Create and fill buffer
			initialData := make([]byte, size/2)
			for i := 0; i < size/2; i++ {
				initialData[i] = byte(i % 256)
			}
			rb := ringbuffer.New(initialData)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = rb.Get(i % (size / 2))
			}
		})

		// Benchmark batch operations
		batchSize := size / 4
		b.Run("SmallnestBatch-"+itoa(size), func(b *testing.B) {
			rb := smallnest.New(size)
			data := make([]byte, batchSize)
			for i := 0; i < batchSize; i++ {
				data[i] = byte(i % 256)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rb.Write(data)
			}
		})

		b.Run("RingBufferBatch-"+itoa(size), func(b *testing.B) {
			rb := ringbuffer.NewWithCapacity[byte](size)
			data := make([]byte, batchSize)
			for i := 0; i < batchSize; i++ {
				data[i] = byte(i % 256)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rb.AddAll(data)
			}
		})
	}
}

// BenchmarkConcurrency compares concurrent access patterns
func BenchmarkBytesConcurrency(b *testing.B) {
	for _, size := range smallnestBenchSizes {
		// Note: smallnest/ringbuffer uses mutexes for both reads and writes
		b.Run("SmallnestConcurrent-"+itoa(size), func(b *testing.B) {
			rb := smallnest.New(size)
			// Pre-fill with data
			data := make([]byte, size/2)
			for i := 0; i < size/2; i++ {
				data[i] = byte(i % 256)
			}
			rb.Write(data)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				counter := 0
				readBuf := make([]byte, 1)
				writeBuf := []byte{byte('x')}

				for pb.Next() {
					if counter%2 == 0 {
						// Read if there's data
						if rb.Length() > 0 {
							rb.Read(readBuf)
						}
					} else {
						// Write if there's space
						if rb.Free() > 0 {
							rb.Write(writeBuf)
						}
					}
					counter++
				}
			})
		})

		b.Run("RingBufferConcurrent-"+itoa(size), func(b *testing.B) {
			rb := ringbuffer.NewWithCapacity[byte](size)
			// Pre-fill with data
			for i := 0; i < size/2; i++ {
				rb.Add(byte(i % 256))
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				counter := 0
				for pb.Next() {
					if counter%2 == 0 {
						// Read operation
						_ = rb.Get(counter % (size / 2))
					} else {
						// Write operation
						rb.Add(byte('x'))
					}
					counter++
				}
			})
		})
	}
}
