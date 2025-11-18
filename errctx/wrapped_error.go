package errctx

// WrappedError is an implementation of [Error] that wraps another error.
//
// It also implements the Unwrap and Cause methods to retrieve the wrapped error.
//
// When implementing a custom error type that wraps another error,
// embed [WrappedError] to provide standard behavior.
type WrappedError struct {
	error
	*ErrorContext
}

// NewWrappedError creates a new [WrappedError] with the given error and options.
// It wraps the provided even if it already is an [WrappedError] and always
// creates a new context.
// To avoid double wrapping and reuse existing context, use [Wrap] instead.
func NewWrappedError(err error, stackSkip int, opts ...Option) *WrappedError {
	return &WrappedError{
		error:        err,
		ErrorContext: newErrorContext(stackSkip+1, opts...),
	}
}

// Wrap creates a new [WrappedError] with the given error and options.
// If the provided error already an [Error], it clones its context and applies
// the new options to it.
// If the provided error is already an [WrappedError], it rewraps its inner error
// to avoid double wrapping.
func Wrap(err error, stackSkip int, opts ...Option) Error {
	if err == nil {
		return nil
	}

	var ec *ErrorContext

	// Check if the error already has context
	if ewc, ok := err.(Error); ok {
		// If the error already has context and no new options are provided,
		// we don't need to wrap it again and can return it as is.
		if len(opts) == 0 {
			return ewc
		}

		// Clone the existing context to avoid modifying the original one
		// and apply new options
		ec = ewc.CloneErrorContext(opts...)

		// If the error is already an WrappedError, get its inner error
		// to avoid double wrapping
		if ew, ok := err.(*WrappedError); ok {
			err = ew.Unwrap()
		}
	} else {
		// If the error does not have context, create a new one
		ec = newErrorContext(stackSkip+1, opts...)
	}

	return &WrappedError{
		error:        err,
		ErrorContext: ec,
	}
}

// Error returns the error message with prefix if set.
func (e *WrappedError) Error() string {
	if len(e.prefix) > 0 {
		return e.prefix + ": " + e.error.Error()
	}
	return e.error.Error()
}

// Unwrap returns the wrapped error.
func (e *WrappedError) Unwrap() error {
	return e.error
}

// Cause returns the wrapped error.
// This method is provided for compatibility with github.com/pkg/errors.
func (e *WrappedError) Cause() error {
	return e.error
}
