package process

import (
	"sync"
)

// RingBuffer is a thread-safe fixed-size circular buffer for log lines.
// Once full, new writes overwrite the oldest entries.
type RingBuffer struct {
	lines    []string
	capacity int
	head     int
	size     int
	mu       sync.RWMutex
}

// NewRingBuffer creates a ring buffer with the specified capacity.
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = 1000
	}
	return &RingBuffer{
		lines:    make([]string, capacity),
		capacity: capacity,
	}
}

// Write appends a line to the buffer.
func (rb *RingBuffer) Write(line string) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.lines[rb.head] = line
	rb.head = (rb.head + 1) % rb.capacity

	if rb.size < rb.capacity {
		rb.size++
	}
}

// ReadAll returns all lines in chronological order (oldest to newest).
func (rb *RingBuffer) ReadAll() []string {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.size == 0 {
		return []string{}
	}

	result := make([]string, rb.size)
	if rb.size < rb.capacity {
		// Buffer not yet full - head is the next write position
		copy(result, rb.lines[:rb.size])
	} else {
		// Buffer is full - head points to oldest entry
		copy(result, rb.lines[rb.head:])
		copy(result[rb.capacity-rb.head:], rb.lines[:rb.head])
	}

	return result
}

// Size returns the current number of lines in the buffer.
func (rb *RingBuffer) Size() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.size
}

// Clear empties the buffer.
func (rb *RingBuffer) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.head = 0
	rb.size = 0
	rb.lines = make([]string, rb.capacity)
}
