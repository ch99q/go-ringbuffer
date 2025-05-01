package ringbuffer

import (
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	// Test with normal data
	items := []int{1, 2, 3, 4, 5}
	rb := New(items)
	if rb.Len() != 8 { // Next power of 2 after 5
		t.Errorf("Expected length 8, got %d", rb.Len())
	}

	// Test with empty slice
	rb = New([]int{})
	if rb.Len() != 1 {
		t.Errorf("Expected length 1 for empty slice, got %d", rb.Len())
	}
}

func TestNewWithCapacity(t *testing.T) {
	// Test with positive capacity
	rb := NewWithCapacity[int](5)
	if rb.Len() != 8 { // Next power of 2 after 5
		t.Errorf("Expected length 8, got %d", rb.Len())
	}

	// Test with zero capacity
	rb = NewWithCapacity[int](0)
	if rb.Len() != 1 {
		t.Errorf("Expected length 1 for zero capacity, got %d", rb.Len())
	}

	// Test with negative capacity
	rb = NewWithCapacity[int](-5)
	if rb.Len() != 1 {
		t.Errorf("Expected length 1 for negative capacity, got %d", rb.Len())
	}
}

func TestGet(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	rb := New(items)

	// Add some new items to create a wrapping scenario
	for i := 6; i <= 10; i++ {
		rb.Add(i)
	}

	// Now the buffer should contain 6, 7, 8, 9, 10 in positions 0-4
	expected := []int{10, 9, 8, 7, 6}
	for i := 0; i < 5; i++ {
		if rb.Get(i) != expected[i] {
			t.Errorf("At position %d: expected %d, got %d", i, expected[i], rb.Get(i))
		}
	}
}

func TestSet(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	rb := New(items)

	// Set some values
	rb.Set(0, 100)
	rb.Set(2, 300)

	// Verify the values were set
	if rb.data[0] != 100 {
		t.Errorf("Expected data[0] to be 100, got %d", rb.data[0])
	}
	if rb.data[2] != 300 {
		t.Errorf("Expected data[2] to be 300, got %d", rb.data[2])
	}
}

func TestAdd(t *testing.T) {
	rb := New([]int{1, 2, 3, 4})

	// Add a new value
	rb.Add(5)

	// The newest value should be 5
	if rb.Get(0) != 5 {
		t.Errorf("Expected newest value to be 5, got %d", rb.Get(0))
	}

	// Add enough values to wrap around
	for i := 6; i <= 12; i++ {
		rb.Add(i)
	}

	// Check the newest values
	expected := []int{12, 11, 10, 9}
	for i := 0; i < 4; i++ {
		if rb.Get(i) != expected[i] {
			t.Errorf("At position %d: expected %d, got %d", i, expected[i], rb.Get(i))
		}
	}
}

func TestAddAll(t *testing.T) {
	rb := New([]int{1, 2, 3, 4})

	// Add multiple values
	rb.AddAll([]int{5, 6, 7})

	// Check the newest values
	expected := []int{7, 6, 5, 4}
	for i := 0; i < 4; i++ {
		if rb.Get(i) != expected[i] {
			t.Errorf("At position %d: expected %d, got %d", i, expected[i], rb.Get(i))
		}
	}

	// Test adding more items than buffer size
	rb = New([]int{1, 2, 3, 4})
	rb.AddAll([]int{5, 6, 7, 8, 9, 10})
	expected = []int{10, 9, 8, 7}
	for i := 0; i < 4; i++ {
		if rb.Get(i) != expected[i] {
			t.Errorf("At position %d: expected %d, got %d", i, expected[i], rb.Get(i))
		}
	}

	// Test adding empty slice
	rb.AddAll([]int{})
	for i := 0; i < 4; i++ {
		if rb.Get(i) != expected[i] {
			t.Errorf("After empty add, at position %d: expected %d, got %d", i, expected[i], rb.Get(i))
		}
	}
}

func TestGetRange(t *testing.T) {
	// Initialize buffer
	rb := New([]int{1, 2, 3, 4})
	
	// Add more items to wrap around
	for i := 5; i <= 8; i++ {
		rb.Add(i)
	}

	// Buffer should contain 5, 6, 7, 8

	// Get all elements
	all := rb.GetRange(4)
	expected := []int{8, 7, 6, 5} // Newest to oldest
	for i := 0; i < 4; i++ {
		if all[i] != expected[i] {
			t.Errorf("GetRange(4)[%d]: expected %d, got %d", i, expected[i], all[i])
		}
	}

	// Get subset of elements
	subset := rb.GetRange(2)
	expected = []int{8, 7} // Newest to oldest
	for i := 0; i < 2; i++ {
		if subset[i] != expected[i] {
			t.Errorf("GetRange(2)[%d]: expected %d, got %d", i, expected[i], subset[i])
		}
	}

	// Try to get more elements than available
	oversized := rb.GetRange(10)
	if len(oversized) != 4 {
		t.Errorf("Expected oversized GetRange to return %d elements, got %d", 4, len(oversized))
	}
}

func TestConcurrency(t *testing.T) {
	rb := New([]int{0, 0, 0, 0, 0, 0, 0, 0})
	const iterations = 1000
	var wg sync.WaitGroup

	// Start readers
	for r := 0; r < 5; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				_ = rb.Get(i % rb.Len())
				_ = rb.GetRange(rb.Len() / 2)
			}
		}()
	}

	// Start writers
	for w := 0; w < 5; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				rb.Add(id*iterations + i)
				if i%100 == 0 {
					rb.AddAll([]int{id, id, id})
				}
			}
		}(w)
	}

	wg.Wait()
	// The test passes if there are no race failures (panics)
}

// Benchmarks copied from the original benchmark suite
func BenchmarkRingBufferGet(b *testing.B) {
	const size = 100
	data := make([]int, size)
	for i := 0; i < size; i++ {
		data[i] = i
	}
	
	rb := New(data)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rb.Get(i % size)
	}
}

func BenchmarkRingBufferAdd(b *testing.B) {
	const size = 100
	data := make([]int, size)
	rb := New(data)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Add(i)
	}
}

func BenchmarkRingBufferConcurrentAccess(b *testing.B) {
	const size = 100
	data := make([]int, size)
	rb := New(data)
	
	// Pre-fill with some data
	for i := 0; i < size*2; i++ {
		rb.Add(i)
	}
	
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
}