package asyncbuffer

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TestCondSuite struct {
	suite.Suite

	cond *Cond
}

func (s *TestCondSuite) SetupTest() {
	s.cond = NewCond()
}

func (s *TestCondSuite) TearDownTest() {
	if s.cond != nil {
		s.cond.Close()
	}
}

// TestBasicWaitAndTick tests the basic functionality of the Cond
func (s *TestCondSuite) TestBasicWaitAndTick() {
	ch := s.cond.ch

	// Start a goroutine that will tick after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		s.cond.Tick()
	}()

	// Start a goroutine that will wait for the tick
	var done atomic.Bool
	go func() {
		s.cond.Wait()
		done.Store(true)
	}()

	s.Require().Eventually(done.Load, 200*time.Millisecond, 10*time.Millisecond)

	// Means that and old channel was closed and a new one has been created
	s.Require().NotEqual(ch, s.cond.ch)
}

// TestWaitMultipleWaiters tests that multiple waiters can be unblocked by a single tick
func (s *TestCondSuite) TestWaitMultipleWaiters() {
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
			s.cond.Wait()
			results[index] = true
		}(i)
	}

	// Wait for all goroutines to start waiting
	startWg.Wait()

	// Wait for all waiters to complete
	var done atomic.Bool
	go func() {
		s.cond.Tick() // Signal that execution can proceed
		wg.Wait()
		done.Store(true)
	}()

	s.Require().Eventually(done.Load, 200*time.Millisecond, 10*time.Millisecond)

	// Check that all waiters were unblocked
	for _, completed := range results {
		s.Require().True(completed)
	}
}

// TestClose tests the behavior of the Cond when closed
func (s *TestCondSuite) TestClose() {
	s.cond.Close()
	s.cond.Close() // Should not panic
	s.cond.Wait()  // Should eventually return
	s.cond.Tick()  // Should not panic

	s.Require().Nil(s.cond.ch)
}

func (s *TestCondSuite) TestRapidTicksAndWaits() {
	const iterations = 1000

	var wg sync.WaitGroup

	// Start a goroutine that will rapidly tick
	wg.Add(1)
	go func() {
		defer wg.Done()

		timeTicker := time.Tick(time.Microsecond)

		for range iterations {
			<-timeTicker
			s.cond.Tick()
		}

		s.cond.Close() // Close after all ticks
	}()

	// Start multiple waiters
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range iterations / 10 {
				s.cond.Wait()
			}
		}()
	}

	var done atomic.Bool
	go func() {
		wg.Wait()
		done.Store(true)
	}()

	s.Require().Eventually(done.Load, 200*time.Millisecond, 10*time.Millisecond)
}

func TestCond(t *testing.T) {
	suite.Run(t, new(TestCondSuite))
}
