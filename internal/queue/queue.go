package queue

import (
	"context"
	"sync"
)

// BuildQueue manages build jobs in a queue
type BuildQueue interface {
	Enqueue(deploymentID uint) error
	Dequeue(ctx context.Context) (uint, error)
	Size() int
}

// InMemoryQueue is a simple in-memory queue (for development)
// In production, use Redis or RabbitMQ
type InMemoryQueue struct {
	items []uint
	mu    sync.Mutex
	cond  *sync.Cond
}

func NewInMemoryQueue() *InMemoryQueue {
	q := &InMemoryQueue{
		items: make([]uint, 0),
	}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *InMemoryQueue) Enqueue(deploymentID uint) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, deploymentID)
	q.cond.Signal()
	return nil
}

func (q *InMemoryQueue) Dequeue(ctx context.Context) (uint, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.items) == 0 {
		// Wait for item or context cancellation
		done := make(chan struct{})
		go func() {
			q.cond.Wait()
			close(done)
		}()

		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-done:
			// Item available
		}
	}

	item := q.items[0]
	q.items = q.items[1:]
	return item, nil
}

func (q *InMemoryQueue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}
