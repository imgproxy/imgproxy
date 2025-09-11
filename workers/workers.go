package workers

import (
	"context"

	"github.com/imgproxy/imgproxy/v3/monitoring"
	"golang.org/x/sync/semaphore"
)

// Workers controls how many concurrent image processings are allowed.
// Requests exceeding this limit will be queued.
//
// It can also optionally limit the number of requests in the queue.
type Workers struct {
	// queue semaphore: limits the queue size
	queue *semaphore.Weighted

	// workers semaphore: limits the number of concurrent image processings
	workers *semaphore.Weighted
}

// New creates new semaphores instance
func New(config *Config) (*Workers, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	var queue *semaphore.Weighted

	if config.RequestsQueueSize > 0 {
		queue = semaphore.NewWeighted(int64(config.RequestsQueueSize + config.WorkersNumber))
	}

	workers := semaphore.NewWeighted(int64(config.WorkersNumber))

	return &Workers{
		queue:   queue,
		workers: workers,
	}, nil
}

// Acquire acquires a worker.
// It returns a worker release function and an error if any.
func (s *Workers) Acquire(ctx context.Context) (context.CancelFunc, error) {
	defer monitoring.StartQueueSegment(ctx)()

	// First, try to acquire the queue semaphore if configured.
	// If the queue is full, return an error immediately.
	releaseQueue, err := s.acquireQueue()
	if err != nil {
		return nil, err
	}

	// Next, acquire the workers semaphore.
	err = s.workers.Acquire(ctx, 1)
	if err != nil {
		releaseQueue()
		return nil, err
	}

	release := func() {
		s.workers.Release(1)
		releaseQueue()
	}

	return release, nil
}

// acquireQueue acquires the queue semaphore and returns release function and error.
// If queue semaphore is not configured, it returns a noop anonymous function to make
// semaphore usage transparent.
func (s *Workers) acquireQueue() (context.CancelFunc, error) {
	if s.queue == nil {
		return func() {}, nil // return no-op cancel function if semaphore is disabled
	}

	acquired := s.queue.TryAcquire(1)
	if !acquired {
		return nil, newTooManyRequestsError()
	}

	return func() { s.queue.Release(1) }, nil
}
