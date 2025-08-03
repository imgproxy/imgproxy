package ierrors

import (
	"github.com/imgproxy/imgproxy/v3/errwrap"
)

type Error = errwrap.ErrWrap
type Option func(*Error) *Error

// Now, it's fallback to the original Wrap function
func Wrap(err error, stackSkip int, opts ...Option) *Error {
	if err == nil {
		return nil
	}

	var e *Error

	x, ok := err.(*Error)
	if ok {
		e = errwrap.Wrap(x)
	} else {
		e = errwrap.From(err, stackSkip)
	}

	for _, opt := range opts {
		e = opt(e)
	}

	return e
}

func WithStatusCode(code int) Option {
	return func(e *Error) *Error {
		x := e.WithStatusCode(code)
		return x
	}
}

func WithPublicMessage(msg string) Option {
	return func(e *Error) *Error {
		return e.WithPublicMessage(msg)
	}
}

func WithPrefix(prefix string) Option {
	return func(e *Error) *Error {
		return errwrap.Wrapf(e, "%s", prefix)
	}
}

func WithShouldReport(report bool) Option {
	return func(e *Error) *Error {
		return e.WithShouldReport(report)
	}
}
