package ensure

type EnsureFunc[T any] func() T

// Ensure ensures that the returned value is not nil.
// If the provided pointer is nil, the function calls the provided
// EnsureFunc to obtain a new value.
// Otherwise, it returns the original value.
func Ensure[T any](val *T, f EnsureFunc[T]) *T {
	if val == nil {
		v := f()
		return &v
	}
	return val
}
