package testutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// LazyObj is a function that returns an object of type T.
type LazyObj[T any] func() T

// LazyObjT is an interface that provides access to the [testing.T] object.
type LazyObjT interface {
	T() *testing.T
}

// LazyObjNew is a function that creates and returns an object of type T and an error if any.
type LazyObjNew[T any] func() (T, error)

// LazyObjDrop is a callback that is called when [LazyObj] is reset.
// It receives a pointer to the object being dropped.
// If the object was not yet initialized, the callback is not called.
type LazyObjDrop[T any] func(T) error

// NewLazyObj creates a new [LazyObj] that initializes the object on the first call.
// It returns a function that can be called to get the object and a cancel function
// that can be called to reset the object.
func NewLazyObj[T any](
	s LazyObjT,
	newFn LazyObjNew[T],
	dropFn ...LazyObjDrop[T],
) (LazyObj[T], context.CancelFunc) {
	s.T().Helper()

	var obj *T

	init := func() T {
		if obj != nil {
			return *obj
		}

		o, err := newFn()
		require.NoError(s.T(), err, "Failed to initialize lazy object")

		obj = &o
		return o
	}

	cancel := func() {
		if obj == nil {
			return
		}

		for _, fn := range dropFn {
			if fn == nil {
				continue
			}

			require.NoError(s.T(), fn(*obj), "Failed to reset lazy object")
		}

		obj = nil
	}

	return init, cancel
}
