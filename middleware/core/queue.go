// Package core provides a simple in-memory doubly-linked queue used by
// the middleware to hold work items (Solr jobs) while they are being
// processed. The implementation is intentionally small; consider using
// a tested concurrent queue implementation for production workloads.

package core

import "sync"

// Container represents a node inside the Queue and holds an arbitrary
// Value. The fields `prv` and `next` link to neighboring nodes.
type Container struct {
	prv, next *Container
	Value     any
}

// Queue is a thread-safe double-ended queue supporting push/pop
// operations at both ends. A simple mutex guards concurrent access.
type Queue struct {
	front *Container
	back  *Container
	size  int
	mu    sync.Mutex
	ready chan struct{}
}

// GetContainer creates a new Container wrapping the provided value.
func GetContainer(value any) *Container {
	return &Container{prv: nil, next: nil, Value: value}
}

// NewQueue constructs an empty Queue.
func NewQueue() *Queue {
	return &Queue{front: nil, back: nil, size: 0, mu: sync.Mutex{}}
}

// PushBack appends a Container to the back of the queue.
func (q *Queue) PushBack(value *Container) {
	if q.back == nil {
		q.front = value
		q.back = q.front
	} else {
		value.prv = q.back
		newContainer := value
		q.back.next = newContainer
		q.back = newContainer
	}
	q.size += 1
}

// PushFront inserts a Container at the front of the queue.
func (q *Queue) PushFront(value *Container) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.front == nil {
		q.back = value
		q.front = q.back
		return
	} else {
		value.next = q.front
		newContainer := value
		q.front.prv = newContainer
		q.front = newContainer
	}
	q.size += 1
}

// PopFront removes and returns the front element. The boolean indicates
// whether an element was returned.
func (q *Queue) PopFront() (Container, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.front == nil {
		var empty Container
		return empty, false
	}

	value := q.front.Value.(Container)

	q.front = q.front.next
	if q.front != nil {
		q.front.prv = nil
	} else {
		q.back = nil
	}
	q.size -= 1
	return value, true
}

// IsEmpty reports whether the queue has no elements.
func (q *Queue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.front == nil
}

// PeekFront returns the front element without removing it.
func (q *Queue) PeekFront() (Container, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.front == nil {
		var empty Container
		return empty, false
	}
	return q.front.Value.(Container), true
}

// PeekBack returns the back element without removing it.
func (q *Queue) PeekBack() (Container, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.back == nil {
		var empty Container
		return empty, false
	}
	return q.back.Value.(Container), true
}

// PopBack removes and returns the back element.
func (q *Queue) PopBack() (Container, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.back == nil {
		var empty Container
		return empty, false
	}

	value := q.back.Value.(Container)

	q.back = q.back.prv
	if q.back != nil {
		q.back.next = nil
	} else {
		q.front = nil
	}
	q.size -= 1
	return value, true
}

// Size returns the current number of elements in the queue.
func (q *Queue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.size
}

// IsReady signals readiness by writing to the ready channel until the
// queue is drained. This helper is simplistic and should be adapted for
// real signaling usage.
func (q *Queue) IsReady() {

	for q.Size() != 0 {
		q.ready <- struct{}{}
		return
	}
}
