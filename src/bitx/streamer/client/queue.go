package client

import "sync"

// Queue implements a FIFO queue of interface{}s.
type Queue struct {
	mu    sync.Mutex
	queue []interface{}
}

// NewQueue returns a new Queue struct.
func NewQueue() *Queue {
	return &Queue{queue: make([]interface{}, 0)}
}

// Enqueue adds an interface{} to the back of the queue.
func (q *Queue) Enqueue(obj interface{}) {
	q.mu.Lock()
	q.queue = append(q.queue, obj)
	q.mu.Unlock()
}

// Dequeue removes an interface{} from the front of the queue.
func (q *Queue) Dequeue() interface{} {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	obj := q.queue[0]
	q.queue = q.queue[1:]

	return obj
}

// Len returns the size of the queue.
func (q *Queue) Len() int {
	return len(q.queue)
}
