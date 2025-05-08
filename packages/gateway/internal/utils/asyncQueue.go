package utils

import (
	"context"
	"log"
	"sync"
	"time"
)

type Queue[T any] struct {
	pending    []T
	processing chan T
	nWorkers   int
	mu         sync.Mutex
}

func NewQueue[T any]() *Queue[T] {
	nWorkers := 8

	return &Queue[T]{
		pending:    make([]T, 0),
		processing: make(chan T, nWorkers),
		nWorkers:   nWorkers,
		mu:         sync.Mutex{},
	}
}

func (q *Queue[T]) Run(ctx context.Context, f func(T)) {
	for i := range q.nWorkers {
		go q.work(ctx, i, f)
	}

	go q.manageItems(ctx)

	go func() {
		<-ctx.Done()

		for len(q.processing) > 0 {
			<-q.processing
		}

		time.Sleep(200 * time.Millisecond) // wait for other workers to stop
		close(q.processing)
	}()
}

func (q *Queue[T]) Add(items ...T) {
	q.mu.Lock()
	q.pending = append(q.pending, items...)
	q.mu.Unlock()
}

func (q *Queue[T]) work(ctx context.Context, workerId int, f func(T)) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("stopping queue worker: %d\n", workerId)
			return
		case it, ok := <-q.processing:
			if !ok {
				return
			}

			f(it)
		}
	}
}

// move item to processing channel if processing channel is empty
func (q *Queue[T]) manageItems(ctx context.Context) {
	// ticker := time.NewTicker(100 * time.Millisecond)
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			log.Println("shutting manage channel down")
			return
		case <-ticker.C:
			q.mu.Lock()
			for len(q.processing) < cap(q.processing) && len(q.pending) > 0 {
				q.processing <- q.pending[0]
				q.pending = q.pending[1:]
			}
			q.mu.Unlock()
		}
	}
}
