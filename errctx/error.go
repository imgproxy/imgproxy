package errctx

import "reflect"

// Error is an interface for errors that carry additional context information.
type Error interface {
	// Error returns the error message.
	Error() string

	// StatusCode returns the HTTP status code associated with the error.
	StatusCode() int
	// PublicMessage returns the public message associated with the error.
	PublicMessage() string
	// ShouldReport indicates whether the error should be reported.
	ShouldReport() bool
	// DocsURL returns the documentation URL associated with the error.
	DocsURL() string

	// StackTrace returns the stack trace associated with the error.
	// This method is traditionally used to retrieve the stack trace
	// by packages like error reporters.
	StackTrace() []uintptr
	// Callers returns the stack trace associated with the error.
	// This method is traditionally used to retrieve the stack trace
	// by packages like error reporters.
	Callers() []uintptr
	// FormatStack returns the stack trace as a formatted string.
	FormatStack() string

	// CloneErrorContext returns a copy of the error context
	// with applied options.
	CloneErrorContext(opts ...Option) *ErrorContext
}

// ErrorType returns the type name of the given error.
// If the error is [WrappedError], it returns the type of the inner error.
func ErrorType(err error) string {
	if ew, ok := err.(*WrappedError); ok {
		return ErrorType(ew.Unwrap())
	}
	return reflect.TypeOf(err).String()
}
