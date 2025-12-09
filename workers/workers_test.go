package workers

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
)

type WorkersTestSuite struct {
	testutil.LazySuite

	config  testutil.LazyObj[*Config]
	workers testutil.LazyObj[*Workers]
}

func (s *WorkersTestSuite) SetupSuite() {
	s.config, _ = testutil.NewLazySuiteObj(s, func() (*Config, error) {
		return &Config{RequestsQueueSize: 0, WorkersNumber: 1}, nil
	})

	s.workers, _ = testutil.NewLazySuiteObj(s, func() (*Workers, error) {
		return New(s.config())
	})
}

func (s *WorkersTestSuite) acquire(ctx context.Context, n int, delay time.Duration) (int, error) {
	errg := new(errgroup.Group)
	acquired := int64(0)

	// Get the workers instance before running goroutines
	workers := s.workers()

	for range n {
		errg.Go(func() error {
			release, err := workers.Acquire(ctx)

			if err == nil {
				time.Sleep(delay)

				release()
				atomic.AddInt64(&acquired, 1)
			}

			return err
		})
	}

	err := errg.Wait()
	return int(acquired), err
}

func (s *WorkersTestSuite) TestQueueDisabled() {
	s.config().RequestsQueueSize = 0
	s.config().WorkersNumber = 2

	// Try to acquire workers that exceed allowed workers number
	acquired, err := s.acquire(s.T().Context(), 4, 10*time.Millisecond)
	s.Require().Equal(4, acquired, "All workers should be eventually acquired")
	s.Require().NoError(err, "All workers should be acquired without error")
}

func (s *WorkersTestSuite) TestQueueEnabled() {
	s.config().RequestsQueueSize = 1
	s.config().WorkersNumber = 2

	// Try to acquire workers that fit allowed workers number + queue size
	acquired, err := s.acquire(s.T().Context(), 3, 10*time.Millisecond)
	s.Require().Equal(3, acquired, "All workers should be eventually acquired")
	s.Require().NoError(err, "All workers should be acquired without error")

	// Try to acquire workers that exceed allowed workers number + queue size
	acquired, err = s.acquire(s.T().Context(), 6, 10*time.Millisecond)
	s.Require().Equal(3, acquired, "Only 4 workers should be acquired")
	s.Require().ErrorAs(err, new(TooManyRequestsError))
}

func (s *WorkersTestSuite) TestContextTimeout() {
	ctx, cancel := context.WithTimeout(s.T().Context(), 5*time.Millisecond)
	defer cancel()

	acquired, err := s.acquire(ctx, 2, 100*time.Millisecond)
	s.Require().Equal(1, acquired, "Only 1 worker should be acquired")
	s.Require().ErrorIs(err, context.DeadlineExceeded, "Context deadline exceeded error expected")
}

func (s *WorkersTestSuite) TestContextCanceled() {
	ctx, cancel := context.WithCancel(s.T().Context())
	cancel()

	acquired, err := s.acquire(ctx, 2, 100*time.Millisecond)
	s.Require().Equal(0, acquired, "No worker should be acquired")
	s.Require().ErrorIs(err, context.Canceled, "Context canceled error expected")
}

func (s *WorkersTestSuite) TestSemaphoresInvalidConfig() {
	_, err := New(&Config{RequestsQueueSize: 0, WorkersNumber: 0})
	s.Require().Error(err)

	_, err = New(&Config{RequestsQueueSize: -1, WorkersNumber: 1})
	s.Require().Error(err)
}

func TestWorkers(t *testing.T) {
	suite.Run(t, new(WorkersTestSuite))
}
