// errwrap package provides a way to wrap errors with additional context
package errwrap

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

const (
	// defaultPublicMessage is the default public message for errors
	defaultPublicMessage = "Internal error"

	// defaultStatusCode is the default HTTP status code for errors
	// (500 internal server error)
	defaultStatusCode = http.StatusInternalServerError
)

type ErrWrap struct {
	// err is the underlying error which is wrapped originally
	error

	// statusCode represents the HTTP status code for the error
	statusCode int

	// publicMessage is a message that is shown in the error reporting system
	publicMessage string

	// shouldReport indicates whether the error should be reported to the error reporting system
	shouldReport bool

	// stack is the original stack trace of the error
	stack []uintptr

	// messages is a list of messages associated with the error
	messages []string
}

// New creates a New ErrWrap instance with the provided message
func New(message string, skipStackFrames int) *ErrWrap {
	return From(fmt.Errorf("%s", message), skipStackFrames)
}

// Newf creates a new ErrWrap instance with a formatted message
func Errorf(skipStackFrames int, format string, args ...any) *ErrWrap {
	return From(fmt.Errorf(format, args...), skipStackFrames)
}

// From creates a new ErrWrap instance from an existing error.
// In case underlying error is already an ErrWrap, it will start from scratch,
// except that it will keep the original error.
func From(err error, skipStackFrames int) *ErrWrap {
	// Do not re-wrap an already wrapped error
	e := err

	o, ok := err.(*ErrWrap)
	if ok {
		e = o.Unwrap()
	}

	return &ErrWrap{
		error:         e,
		statusCode:    defaultStatusCode,
		publicMessage: defaultPublicMessage,
		shouldReport:  true,
		stack:         callers(skipStackFrames),
		messages:      make([]string, 0),
	}
}

// Clone creates a copy of the ErrWrap instance.
func (e *ErrWrap) Clone() *ErrWrap {
	if e == nil {
		return nil
	}

	clone := &ErrWrap{
		error:         e.error,
		statusCode:    e.statusCode,
		publicMessage: e.publicMessage,
		shouldReport:  e.shouldReport,
		stack:         e.stack,
		messages:      make([]string, len(e.messages)),
	}
	copy(clone.messages, e.messages)

	return clone
}

// Wrap wraps an existing error into an ErrWrap instance
func Wrap(err error) *ErrWrap {
	if err == nil {
		return nil
	}

	var wrapped *ErrWrap
	if existing, ok := err.(*ErrWrap); ok {
		wrapped = existing.Clone()
	} else {
		wrapped = &ErrWrap{
			error:         err,
			statusCode:    defaultStatusCode,
			publicMessage: defaultPublicMessage,
			shouldReport:  true,
			stack:         callers(0),
			messages:      make([]string, 0),
		}
	}

	return wrapped
}

// Wrapf wraps an existing error into an ErrWrap instance with a formatted message
func Wrapf(err error, msg string, args ...any) *ErrWrap {
	if err == nil {
		return nil
	}

	wrapped := Wrap(err)
	formatted := fmt.Sprintf(msg, args...)
	wrapped.messages = append(wrapped.messages, formatted)
	return wrapped
}

// Error returns the error message
func (e *ErrWrap) Error() string {
	if len(e.messages) > 0 {
		return fmt.Sprintf("%s: %s", e.error.Error(), strings.Join(e.messages, ": "))
	}
	return e.error.Error()
}

// Unwrap returns the underlying error
func (e *ErrWrap) Unwrap() error {
	return e.error
}

// StatusCode returns the HTTP status code
func (e *ErrWrap) StatusCode() int {
	return e.statusCode
}

// PublicMessage returns the public message
func (e *ErrWrap) PublicMessage() string {
	return e.publicMessage
}

// ShouldReport returns whether the error should be reported
func (e *ErrWrap) ShouldReport() bool {
	return e.shouldReport
}

// WithStatusCode sets the HTTP status code and returns a new instance
func (e *ErrWrap) WithStatusCode(code int) *ErrWrap {
	newErr := e.Clone()
	newErr.statusCode = code
	return newErr
}

// WithPublicMessage sets the public message and returns a new instance
func (e *ErrWrap) WithPublicMessage(msg string) *ErrWrap {
	newErr := e.Clone()
	newErr.publicMessage = msg
	return newErr
}

// WithShouldReport sets whether the error should be reported and returns a new instance
func (e *ErrWrap) WithShouldReport(report bool) *ErrWrap {
	// Create a copy to maintain immutability
	newErr := e.Clone()
	newErr.shouldReport = report
	return newErr
}

// FormatStack formats the stack trace into a human-readable string
func (e *ErrWrap) FormatStack() string {
	lines := make([]string, len(e.stack))

	for i, pc := range e.stack {
		f := runtime.FuncForPC(pc)
		file, line := f.FileLine(pc)
		lines[i] = fmt.Sprintf("%s:%d %s", file, line, f.Name())
	}

	return strings.Join(lines, "\n")
}

// callers captures the stack trace
func callers(skip int) []uintptr {
	stack := make([]uintptr, 10)
	n := runtime.Callers(skip+2, stack)
	return stack[:n]
}
