package semaphores

import (
	"context"

	"github.com/imgproxy/imgproxy/v3/monitoring"
	"golang.org/x/sync/semaphore"
)

// Semaphores is a container for the queue and processing semaphores
type Semaphores struct {
	// queueSize semaphore: limits the queueSize size
	queueSize *semaphore.Weighted

	// processing semaphore: limits the number of concurrent image processings
	processing *semaphore.Weighted
}

// New creates new semaphores instance
func New(config *Config) (*Semaphores, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	var queue *semaphore.Weighted

	if config.RequestsQueueSize > 0 {
		queue = semaphore.NewWeighted(int64(config.RequestsQueueSize + config.Workers))
	}

	processing := semaphore.NewWeighted(int64(config.Workers))

	return &Semaphores{
		queueSize:  queue,
		processing: processing,
	}, nil
}

// AcquireQueue acquires the queue semaphore and returns release function and error.
// if queue semaphore is not configured, it returns a noop anonymous function to make
// semaphore usage transparent.
func (s *Semaphores) AcquireQueue() (context.CancelFunc, error) {
	if s.queueSize == nil {
		return func() {}, nil // return no-op cancel function if semaphore is disabled
	}

	acquired := s.queueSize.TryAcquire(1)
	if !acquired {
		return nil, newTooManyRequestsError()
	}

	return func() { s.queueSize.Release(1) }, nil
}

// AcquireProcessing acquires the processing semaphore
func (s *Semaphores) AcquireProcessing(ctx context.Context) (context.CancelFunc, error) {
	defer monitoring.StartQueueSegment(ctx)()

	err := s.processing.Acquire(ctx, 1)
	if err != nil {
		return nil, err
	}

	return func() { s.processing.Release(1) }, nil
}
