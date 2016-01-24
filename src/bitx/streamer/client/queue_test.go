package client

import "testing"

func expectSize(t *testing.T, q *Queue, size int) {
	if l := q.Len(); l != size {
		t.Errorf("Expected %d, got %d", size, l)
	}
}

func TestQueue(t *testing.T) {
	q := NewQueue()
	expectSize(t, q, 0)
	q.Enqueue(1)
	expectSize(t, q, 1)
	q.Enqueue(2)
	expectSize(t, q, 2)
	obj := q.Dequeue()
	expectSize(t, q, 1)
	if obj.(int) != 1 {
		t.Errorf("Expected 1, got %v", obj)
	}
	obj = q.Dequeue()
	expectSize(t, q, 0)
	if obj.(int) != 2 {
		t.Errorf("Expected 1, got %v", obj)
	}
}
