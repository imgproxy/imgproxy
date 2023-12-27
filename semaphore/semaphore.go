package semaphore

import (
	"context"
	"sync"
)

type Semaphore struct {
	sem chan struct{}
}

func New(n int) *Semaphore {
	return &Semaphore{
		sem: make(chan struct{}, n),
	}
}

func (s *Semaphore) Acquire(ctx context.Context) (*Token, bool) {
	select {
	case s.sem <- struct{}{}:
		return &Token{release: s.release}, true
	case <-ctx.Done():
		return &Token{release: func() {}}, false
	}
}

func (s *Semaphore) TryAcquire() (*Token, bool) {
	select {
	case s.sem <- struct{}{}:
		return &Token{release: s.release}, true
	default:
		return &Token{release: func() {}}, false
	}
}

func (s *Semaphore) release() {
	<-s.sem
}

type Token struct {
	release     func()
	releaseOnce sync.Once
}

func (t *Token) Release() {
	t.releaseOnce.Do(t.release)
}
