package asyncbuffer

import (
	"sync"
	"sync/atomic"
)

type condCh = chan struct{}

// Cond signals that an event has occurred to multiple waiters.
// Uses a cursor to detect if Tick() occurred between checking a condition and waiting.
type Cond struct {
	mu        sync.RWMutex
	ch        condCh
	cursor    atomic.Uint64 // Incremented on each Tick
	closeOnce sync.Once
}

// NewCond creates a new Cond instance with an initialized channel.
func NewCond() *Cond {
	return &Cond{
		ch: make(condCh),
	}
}

// Tick signals that an event has occurred by closing the channel and incrementing cursor.
func (t *Cond) Tick() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.ch != nil {
		t.cursor.Add(1)     // Increment FIRST so waiters see new cursor when they wake
		close(t.ch)         // Then close channel to wake waiters
		t.ch = make(condCh) // Create new channel for next Tick
	}
}

// Cursor returns the current cursor value.
// Capture this before checking your condition, then pass it to Wait().
func (t *Cond) Cursor() uint64 {
	return t.cursor.Load()
}

// Wait waits for a Tick() to occur after the given cursor value.
// Returns the new cursor value when awakened.
// If the cursor has already advanced beyond the given value, returns immediately with the current cursor.
func (t *Cond) Wait(cursor uint64) uint64 {
	t.mu.RLock()
	ch := t.ch
	current := t.cursor.Load()
	t.mu.RUnlock()

	// If cursor already advanced, return immediately
	if current != cursor || ch == nil {
		return current
	}

	// Wait for the channel to close (Tick happened)
	// If Tick happens after t.mu.RUnlock() and here, we will just wait on a
	// closed channel and return updated cursor value.
	<-ch

	// NOTE: Another Tick() could happen here before we read cursor.
	// This is acceptable - we return the current cursor value, which may be
	// newer than the Tick() that woke us. The caller will use this cursor
	// for the next wait cycle and re-check their condition.
	return t.cursor.Load()
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
