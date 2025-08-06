package asyncbuffer

import (
	"sync"
)

type condCh = chan struct{}

// Cond signals that an event has occurred to a multiple waiters.
type Cond struct {
	mu        sync.RWMutex
	ch        condCh
	closeOnce sync.Once
}

// NewCond creates a new Ticker instance with an initialized channel.
func NewCond() *Cond {
	return &Cond{
		ch: make(condCh),
	}
}

// Tick signals that an event has occurred by closing the channel.
func (t *Cond) Tick() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.ch != nil {
		close(t.ch)
		t.ch = make(condCh)
	}
}

// Wait blocks until the channel is closed, indicating that an event has occurred.
func (t *Cond) Wait() {
	t.mu.RLock()
	ch := t.ch
	t.mu.RUnlock()

	if ch == nil {
		return
	}

	<-ch
}

// Close closes the ticker channel and prevents further ticks.
func (t *Cond) Close() {
	t.closeOnce.Do(func() {
		t.mu.Lock()
		defer t.mu.Unlock()

		if t.ch != nil {
			close(t.ch)
			t.ch = nil
		}
	})
}
