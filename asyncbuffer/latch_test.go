package asyncbuffer_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v4/asyncbuffer"
)

func TestNewLatch(t *testing.T) {
	latch := asyncbuffer.NewLatch()

	require.NotNil(t, latch)
	require.NotNil(t, asyncbuffer.LatchDone(latch))

	// Channel should be open (not closed) initially
	select {
	case <-asyncbuffer.LatchDone(latch):
		t.Fatal("Latch should not be released initially")
	default:
		// Expected - channel is not ready
	}
}

func TestLatchRelease(t *testing.T) {
	latch := asyncbuffer.NewLatch()

	// Release the latch
	latch.Release()

	// Channel should now be closed/ready
	select {
	case <-asyncbuffer.LatchDone(latch):
		// Expected - channel is ready after release
	default:
		t.Fatal("Latch should be released after Release() call")
	}
}

func TestLatchWait(t *testing.T) {
	latch := asyncbuffer.NewLatch()

	// Start a goroutine that will wait
	waitCompleted := make(chan bool, 1)
	go func() {
		latch.Wait()
		waitCompleted <- true
	}()

	// Give the goroutine a moment to start waiting
	time.Sleep(10 * time.Millisecond)

	// Wait should not complete yet
	select {
	case <-waitCompleted:
		t.Fatal("Wait should not complete before Release")
	default:
		// Expected
	}

	// Release the latch
	latch.Release()

	// Wait should complete now
	select {
	case <-waitCompleted:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Wait should complete after Release")
	}
}

func TestLatchMultipleWaiters(t *testing.T) {
	latch := asyncbuffer.NewLatch()
	const numWaiters = 10

	var wg sync.WaitGroup
	waitersCompleted := make(chan int, numWaiters)

	// Start multiple waiters
	for i := range numWaiters {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			latch.Wait()
			waitersCompleted <- id
		}(i)
	}

	// Give goroutines time to start waiting
	time.Sleep(10 * time.Millisecond)

	// No waiters should complete yet
	assert.Empty(t, waitersCompleted)

	// Release the latch
	latch.Release()

	// All waiters should complete
	wg.Wait()
	close(waitersCompleted)

	// Verify all waiters completed
	completed := make([]int, 0, numWaiters)
	for id := range waitersCompleted {
		completed = append(completed, id)
	}
	assert.Len(t, completed, numWaiters)
}

func TestLatchMultipleReleases(t *testing.T) {
	latch := asyncbuffer.NewLatch()

	// Release multiple times should be safe
	latch.Release()
	latch.Release()
	latch.Release()

	// Should still be able to wait
	select {
	case <-asyncbuffer.LatchDone(latch):
		// Expected - channel should be ready
	default:
		t.Fatal("Latch should be released")
	}
}
