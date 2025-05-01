// Package ringbuffer provides a high-performance, thread-safe, fixed-size circular buffer.
//
// This implementation is optimized for very fast reads (lock-free) and efficient writes,
// making it ideal for scenarios like time-series data, rolling windows, and circular logs.
// The implementation uses power-of-2 sized buffers with bitwise operations for maximum performance.
package ringbuffer

import (
	"sync"
	"sync/atomic"
)

// RingBuffer implements a fixed-size circular buffer with power-of-2 sized capacity.
// It provides thread-safe access with lock-free reads and synchronized writes.
// This is optimized for high-performance time series data and rolling windows.
type RingBuffer[T any] struct {
	data []T           // power-of-2 sized buffer
	mask int           // bitmask for efficient wrapping (len(data)-1)
	head int           // next write position
	mu   sync.Mutex    // serializes writers
	seq  atomic.Uint64 // low bit == "write in progress"
}

// New creates a fixed-size ring buffer with a size that's a power of 2.
// The size will be rounded up to the next power of 2 if necessary.
// If items is empty, it will create a buffer with capacity 1.
func New[T any](items []T) *RingBuffer[T] {
	if len(items) == 0 {
		// Create a buffer with capacity 1 instead of panicking
		return &RingBuffer[T]{
			data: make([]T, 1),
			mask: 0, // mask for size 1 is 0
		}
	}
	
	// Round up to next power of 2
	size := 1
	for size < len(items) {
		size *= 2
	}
	
	// Create buffer with power-of-2 size
	buf := make([]T, size)
	
	// Copy input items
	copy(buf, items)
	
	return &RingBuffer[T]{
		data: buf,
		mask: size - 1, // bitmask for wrapping (e.g., 7 for size 8)
	}
}

// NewWithCapacity creates a fixed-size ring buffer with the specified capacity.
// The capacity will be rounded up to the next power of 2.
func NewWithCapacity[T any](capacity int) *RingBuffer[T] {
	if capacity <= 0 {
		capacity = 1
	}
	
	// Round up to next power of 2
	size := 1
	for size < capacity {
		size *= 2
	}
	
	return &RingBuffer[T]{
		data: make([]T, size),
		mask: size - 1,
	}
}

// readBarrier runs your closure between two matching seq loads, retrying
// if a writer was active in between.
func (s *RingBuffer[T]) readBarrier(read func()) {
	for {
		before := s.seq.Load()
		if before&1 == 1 {
			continue
		}
		read()
		if before == s.seq.Load() {
			return
		}
	}
}

// Len returns the buffer capacity.
// This matches the standard Go naming convention (like len(slice))
func (s *RingBuffer[T]) Len() int {
	return len(s.data)
}

// Get returns elements in reverse chronological order.
// Get(0) is always the newest/latest, Get(1) the second newest, and so on.
func (s *RingBuffer[T]) Get(i int) T {
	var v T
	s.readBarrier(func() {
		// Use bitwise AND with mask instead of modulo for wrapping
		idx := (s.head - 1 - i) & s.mask
		v = s.data[idx]
	})
	return v
}

// Set overwrites the i'th element.
func (s *RingBuffer[T]) Set(i int, val T) {
	s.mu.Lock()
	s.seq.Add(1) // odd -> readers spin
	
	idx := (s.head + i) & s.mask
	s.data[idx] = val
	
	s.seq.Add(1) // even -> readers proceed
	s.mu.Unlock()
}

// Add inserts a new value, advancing the ring buffer.
// This is more consistent with standard library function naming.
func (s *RingBuffer[T]) Add(val T) {
	s.mu.Lock()
	s.seq.Add(1)
	
	s.data[s.head] = val
	s.head = (s.head + 1) & s.mask
	
	s.seq.Add(1)
	s.mu.Unlock()
}

// AddAll efficiently adds multiple elements at once.
// This follows the same naming pattern as Add but indicates multiple elements.
func (s *RingBuffer[T]) AddAll(values []T) {
	if len(values) == 0 {
		return
	}
	
	s.mu.Lock()
	s.seq.Add(1)
	
	size := len(s.data)
	valLen := len(values)
	
	// If we're adding more items than the buffer size,
	// just take the most recent ones
	if valLen >= size {
		copy(s.data, values[valLen-size:])
		s.head = 0
	} else {
		// Otherwise, add them one by one without releasing the lock
		for _, val := range values {
			s.data[s.head] = val
			s.head = (s.head + 1) & s.mask
		}
	}
	
	s.seq.Add(1)
	s.mu.Unlock()
}

// GetRange efficiently retrieves a range of elements in reverse chronological order.
// Returns count elements, starting with the newest and ending with the oldest.
// If count > Len(), only the available elements will be returned.
func (s *RingBuffer[T]) GetRange(count int) []T {
	var result []T
	
	s.readBarrier(func() {
		size := len(s.data)
		if count > size {
			count = size
		}
		
		result = make([]T, count)
		for i := 0; i < count; i++ {
			// Get items from newest to oldest, matching Get's order
			idx := (s.head - 1 - i) & s.mask
			result[i] = s.data[idx]
		}
	})
	
	return result
}