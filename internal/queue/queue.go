package queue

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrQueueEmpty = errors.New("queue is empty")
)

type Queue struct {
	queues map[string]chan interface{}
	mu     sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		queues: make(map[string]chan interface{}),
	}
}

func (q *Queue) Push(key string, values ...interface{}) {
	q.mu.Lock()

	if _, ok := q.queues[key]; !ok {
		q.queues[key] = make(chan interface{}, 1000) // Arbitrary buffer size
	}

	for _, value := range values {
		q.queues[key] <- value
	}

	q.mu.Unlock()
}

func (q *Queue) Pop(key string) (interface{}, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if queue, ok := q.queues[key]; ok {
		select {
		case value := <-queue:
			return value, nil
		default:
			return nil, ErrQueueEmpty
		}
	}

	return nil, ErrQueueEmpty
}

func (q *Queue) BPop(key string, timeout time.Duration) (interface{}, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if queue, ok := q.queues[key]; ok {
		select {
		case value := <-queue:
			return value, nil
		case <-time.After(timeout):
			return nil, ErrQueueEmpty
		}
	}

	return nil, ErrQueueEmpty
}
