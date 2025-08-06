package asyncbuffer

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TickerTestSuite struct {
	suite.Suite
	ticker *Ticker
}

func (s *TickerTestSuite) SetupTest() {
	s.ticker = NewTicker()
}

func (s *TickerTestSuite) TeardownTest() {
	if s.ticker != nil {
		s.ticker.Close()
	}
}

// TestBasicWaitAndTick tests the basic functionality of the Ticker
func (s *TickerTestSuite) TestBasicWaitAndTick() {
	done := make(chan struct{})

	ch := s.ticker.ch

	// Start a goroutine that will tick after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		s.ticker.Tick()
	}()

	// Start a goroutine that will wait for the tick
	go func() {
		s.ticker.Wait()
		close(done)
	}()

	s.Require().Eventually(func() bool {
		select {
		case <-done:
			return true
		default:
			return false
		}
	}, 100*time.Millisecond, 10*time.Millisecond)

	// Means that and old channel was closed and a new one has been created
	s.Require().NotEqual(ch, s.ticker.ch)
}

// TestWaitMultipleWaiters tests that multiple waiters can be unblocked by a single tick
func (s *TickerTestSuite) TestWaitMultipleWaiters() {
	const numWaiters = 10

	var wg sync.WaitGroup
	var startWg sync.WaitGroup
	results := make([]bool, numWaiters)

	// Start multiple waiters
	for i := range numWaiters {
		wg.Add(1)
		startWg.Add(1)
		go func(index int) {
			defer wg.Done()
			startWg.Done() // Signal that this goroutine is ready
			s.ticker.Wait()
			results[index] = true
		}(i)
	}

	// Wait for all goroutines to start waiting
	startWg.Wait()

	// Wait for all waiters to complete
	done := make(chan struct{})
	go func() {
		s.ticker.Tick() // Signal that execution can proceed
		wg.Wait()
		close(done)
	}()

	s.Require().Eventually(func() bool {
		select {
		case <-done:
			return true
		default:
			return false
		}
	}, 100*time.Millisecond, 10*time.Millisecond)

	// Check that all waiters were unblocked
	for _, completed := range results {
		s.Require().True(completed)
	}
}

// TestClose tests the behavior of the Ticker when closed
func (s *TickerTestSuite) TestClose() {
	s.ticker.Close()
	s.ticker.Close() // Should not panic
	s.ticker.Wait()  // Should eventually return
	s.ticker.Tick()  // Should not panic

	s.Require().Nil(s.ticker.ch)
}

func (s *TickerTestSuite) TestRapidTicksAndWaits() {
	const iterations = 1000

	var wg sync.WaitGroup

	// Start a goroutine that will rapidly tick
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range iterations {
			s.ticker.Tick()
			time.Sleep(time.Microsecond)
		}
		s.ticker.Close() // Close after all ticks
	}()

	// Start multiple waiters
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range iterations / 10 {
				s.ticker.Wait()
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	s.Require().Eventually(func() bool {
		select {
		case <-done:
			return true
		default:
			return false
		}
	}, 100*time.Millisecond, 10*time.Millisecond)
}

func TestTicker(t *testing.T) {
	suite.Run(t, new(TickerTestSuite))
}
