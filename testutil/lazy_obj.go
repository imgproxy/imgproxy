package testutil

import (
	"github.com/stretchr/testify/require"
)

// LazyObj is a function that returns an object of type T.
type LazyObj[T any] func() T

// LazyObjInit is a function that initializes and returns an object of type T and an error if any.
type LazyObjInit[T any] func() (T, error)

// NewLazyObj creates a new LazyObj that initializes the object on the first call.
func NewLazyObj[T any](t require.TestingT, init LazyObjInit[T]) LazyObj[T] {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}

	var obj *T

	return func() T {
		if obj != nil {
			return *obj
		}

		o, err := init()
		require.NoError(t, err)

		obj = &o
		return o
	}
}
