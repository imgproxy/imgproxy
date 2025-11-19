package errctx

// TextError is an implementation of [Error] that holds a simple text message.
//
// When implementing a custom error type that does not wrap another error,
// embed [TextError] to provide standard behavior.
type TextError struct {
	msg string
	*ErrorContext
}

// NewTextError creates a new [TextError] with the given message and options.
func NewTextError(msg string, stackSkip int, opts ...Option) *TextError {
	return &TextError{
		msg:          msg,
		ErrorContext: newErrorContext(stackSkip+1, opts...),
	}
}

// Error returns the error message with prefix if set.
func (e *TextError) Error() string {
	if len(e.prefix) > 0 {
		return e.prefix + ": " + e.msg
	}
	return e.msg
}
