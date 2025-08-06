package asyncbuffer

import (
	"sync"
)

type tickCh = chan struct{}

// Ticker signals that an event has occurred to a multiple waiters.
type Ticker struct {
	_         noCopy
	mu        sync.Mutex
	ch        tickCh
	closeOnce sync.Once
}

// NewTicker creates a new Ticker instance with an initialized channel.
func NewTicker() *Ticker {
	return &Ticker{
		ch: make(tickCh),
	}
}

// Tick signals that an event has occurred by closing the channel.
func (t *Ticker) Tick() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.ch != nil {
		close(t.ch)
		t.ch = make(tickCh)
	}
}

// Wait blocks until the channel is closed, indicating that an event has occurred.
func (t *Ticker) Wait() {
	t.mu.Lock()
	ch := t.ch
	t.mu.Unlock()

	if ch == nil {
		return
	}

	<-ch
}

// Close closes the ticker channel and prevents further ticks.
func (t *Ticker) Close() {
	t.closeOnce.Do(func() {
		t.mu.Lock()
		defer t.mu.Unlock()

		if t.ch != nil {
			close(t.ch)
			t.ch = nil
		}
	})
}
