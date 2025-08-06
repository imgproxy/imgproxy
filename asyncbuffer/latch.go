package asyncbuffer

import (
	"sync"
)

// Latch is once-releasing semaphore.
type Latch struct {
	once sync.Once
	done chan struct{}
}

// NewLatch creates a new Latch.
func NewLatch() *Latch {
	return &Latch{done: make(chan struct{})}
}

// Release releases the latch, allowing all waiting goroutines to proceed.
func (g *Latch) Release() {
	g.once.Do(func() { close(g.done) })
}

// Wait blocks until the latch is released.
func (g *Latch) Wait() {
	<-g.done
}
