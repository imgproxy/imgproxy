package ierrors

import (
	"fmt"
	"runtime"
	"strings"
)

type Error struct {
	StatusCode    int
	Message       string
	PublicMessage string
	Unexpected    bool

	stack []uintptr
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) FormatStack() string {
	if e.stack == nil {
		return ""
	}

	return formatStack(e.stack)
}

func (e *Error) StackTrace() []uintptr {
	return e.stack
}

func (e *Error) SetUnexpected(u bool) *Error {
	e.Unexpected = u
	return e
}

func New(status int, msg string, pub string) *Error {
	return &Error{
		StatusCode:    status,
		Message:       msg,
		PublicMessage: pub,
	}
}

func NewUnexpected(msg string, skip int) *Error {
	return &Error{
		StatusCode:    500,
		Message:       msg,
		PublicMessage: "Internal error",
		Unexpected:    true,

		stack: callers(skip + 3),
	}
}

func Wrap(err error, skip int) *Error {
	if ierr, ok := err.(*Error); ok {
		return ierr
	}
	return NewUnexpected(err.Error(), skip+1)
}

func WrapWithMessage(err error, skip int, msg string) *Error {
	if ierr, ok := err.(*Error); ok {
		newErr := *ierr
		ierr.Message = msg
		return &newErr
	}
	return NewUnexpected(err.Error(), skip+1)
}

func callers(skip int) []uintptr {
	stack := make([]uintptr, 10)
	n := runtime.Callers(skip, stack)
	return stack[:n]
}

func formatStack(stack []uintptr) string {
	lines := make([]string, len(stack))
	for i, pc := range stack {
		f := runtime.FuncForPC(pc)
		file, line := f.FileLine(pc)
		lines[i] = fmt.Sprintf("%s:%d %s", file, line, f.Name())
	}

	return strings.Join(lines, "\n")
}

func StatusCode(err error) int {
	if ierr, ok := err.(*Error); ok {
		return ierr.StatusCode
	}
	return 0
}
