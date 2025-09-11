package testutil

import (
	"context"

	"github.com/stretchr/testify/suite"
)

// LazySuite is a test suite that automatically resets [LazyObj] instances.
// It uses [LazySuite.AfterTest] to perform the reset after each test,
// so if you also use this function in your test suite, don't forget to call
// [LazySuite.AfterTest] or [LazySuite.ResetLazyObjects] explicitly.
type LazySuite struct {
	suite.Suite

	resets []context.CancelFunc
}

// Lazy returns the LazySuite instance itself.
// Needed to implement [LazySuiteFrom].
func (s *LazySuite) Lazy() *LazySuite {
	return s
}

// AfterTest is called by testify after each test.
// If you also use this function in your test suite, don't forget to call
// [LazySuite.AfterTest] or [LazySuite.ResetLazyObjects] explicitly.
func (s *LazySuite) AfterTest(_, _ string) {
	// Reset lazy objects after each test
	s.ResetLazyObjects()
}

// ResetLazyObjects resets all lazy objects created with [NewLazySuiteObj]
func (s *LazySuite) ResetLazyObjects() {
	for _, reset := range s.resets {
		reset()
	}
}

type LazySuiteFrom interface {
	Lazy() *LazySuite
}

// NewLazySuiteObj creates a new [LazyObj] instance and registers its cleanup function
// to a provided [LazySuite].
func NewLazySuiteObj[T any](
	s LazySuiteFrom,
	newFn LazyObjNew[T],
	dropFn ...LazyObjDrop[T],
) (LazyObj[T], context.CancelFunc) {
	// Get the [LazySuite] instance
	lazy := s.Lazy()
	// Create the [LazyObj] instance
	obj, cancel := newLazyObj(lazy, newFn, dropFn...)
	// Add cleanup function to the resets list
	lazy.resets = append(lazy.resets, cancel)

	return obj, cancel
}
