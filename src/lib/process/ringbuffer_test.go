package process

import (
	"testing"
)

func TestRingBuffer_Write_ReadAll(t *testing.T) {
	rb := NewRingBuffer(3)

	// Test initial state
	if rb.Size() != 0 {
		t.Errorf("expected size 0, got %d", rb.Size())
	}

	lines := rb.ReadAll()
	if len(lines) != 0 {
		t.Errorf("expected empty buffer, got %d lines", len(lines))
	}

	// Write less than capacity
	rb.Write("line1")
	rb.Write("line2")

	if rb.Size() != 2 {
		t.Errorf("expected size 2, got %d", rb.Size())
	}

	lines = rb.ReadAll()
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "line1" || lines[1] != "line2" {
		t.Errorf("unexpected lines: %v", lines)
	}

	// Write to fill capacity
	rb.Write("line3")
	lines = rb.ReadAll()
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	// Write beyond capacity - should overwrite oldest
	rb.Write("line4")
	lines = rb.ReadAll()
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "line2" || lines[1] != "line3" || lines[2] != "line4" {
		t.Errorf("expected [line2, line3, line4], got %v", lines)
	}
}

func TestRingBuffer_Clear(t *testing.T) {
	rb := NewRingBuffer(3)
	rb.Write("line1")
	rb.Write("line2")

	rb.Clear()

	if rb.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", rb.Size())
	}

	lines := rb.ReadAll()
	if len(lines) != 0 {
		t.Errorf("expected empty buffer after clear, got %d lines", len(lines))
	}
}

func TestRingBuffer_ZeroCapacity(t *testing.T) {
	rb := NewRingBuffer(0)
	if rb.capacity != 1000 {
		t.Errorf("expected default capacity 1000, got %d", rb.capacity)
	}
}
