package core

import "sync"

type Container struct {
	prv, next *Container
	Value     any
}

type Queue struct {
	front *Container
	back  *Container
	size  int
	mu    sync.Mutex
	ready chan struct{}
}

func GetContainer(value any) *Container {
	return &Container{prv: nil, next: nil, Value: value}
}

func NewQueue() *Queue {
	return &Queue{front: nil, back: nil, size: 0, mu: sync.Mutex{}}
}

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

func (q *Queue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.front == nil
}

func (q *Queue) PeekFront() (Container, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.front == nil {
		var empty Container
		return empty, false
	}
	return q.front.Value.(Container), true
}

func (q *Queue) PeekBack() (Container, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.back == nil {
		var empty Container
		return empty, false
	}
	return q.back.Value.(Container), true
}

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

func (q *Queue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.size
}

func (q *Queue) IsReady() {

	for q.Size() != 0 {
		q.ready <- struct{}{}
		return
	}
}
