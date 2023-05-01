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
	queues    map[string][]interface{}
	mu        sync.Mutex
	pushChan  chan *PushRequest
	pushBatch int
}

type PushRequest struct {
	Key    string
	Values []interface{}
}

const (
	batchThreshold = 200
)

func NewQueue() *Queue {
	q := &Queue{
		queues:    make(map[string][]interface{}),
		pushChan:  make(chan *PushRequest, batchThreshold),
		pushBatch: 0,
	}

	go func() {
		for req := range q.pushChan {
			q.doPush(req.Key, req.Values)
			q.pushBatch++

			if q.pushBatch >= batchThreshold {
				q.mu.Lock()
				q.pushBatch = 0
				q.mu.Unlock()
			}
		}
	}()

	return q
}

func (q *Queue) QPush(key string, values ...interface{}) error {
	q.pushChan <- &PushRequest{Key: key, Values: values}
	return nil
}

func (q *Queue) doPush(key string, values []interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queues[key] = append(q.queues[key], values...)
}

func (q *Queue) Pop(key string) (interface{}, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if queue, ok := q.queues[key]; ok && len(queue) > 0 {
		value := queue[0]
		q.queues[key] = queue[1:]
		return value, nil
	}

	return nil, ErrQueueEmpty
}

func (q *Queue) BPop(key string, timeout time.Duration) (interface{}, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if queue, ok := q.queues[key]; ok {
		if len(queue) > 0 {
			value := queue[0]
			q.queues[key] = queue[1:]
			return value, nil
		}
	}

	// Wait for an item to be available or for the timeout to expire
	c := sync.NewCond(&q.mu)
	go func() {
		time.Sleep(timeout)
		c.Broadcast()
	}()
	c.Wait()

	// Check again after waiting
	if queue, ok := q.queues[key]; ok {
		if len(queue) > 0 {
			value := queue[0]
			q.queues[key] = queue[1:]
			return value, nil
		}
	}

	return nil, ErrQueueEmpty
}
