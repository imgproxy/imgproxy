package ierrors

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

type Option func(*Error)

type Error struct {
	err error

	prefix        string
	statusCode    int
	publicMessage string
	shouldReport  bool

	stack []uintptr
}

func (e *Error) Error() string {
	if len(e.prefix) > 0 {
		return fmt.Sprintf("%s: %s", e.prefix, e.err.Error())
	}

	return e.err.Error()
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) Cause() error {
	return e.err
}

func (e *Error) StatusCode() int {
	if e.statusCode <= 0 {
		return http.StatusInternalServerError
	}

	return e.statusCode
}

func (e *Error) PublicMessage() string {
	if len(e.publicMessage) == 0 {
		return "Internal error"
	}

	return e.publicMessage
}

func (e *Error) ShouldReport() bool {
	return e.shouldReport
}

func (e *Error) StackTrace() []uintptr {
	return e.stack
}

func (e *Error) Callers() []uintptr {
	return e.stack
}

func (e *Error) FormatStackLines() []string {
	lines := make([]string, len(e.stack))

	for i, pc := range e.stack {
		f := runtime.FuncForPC(pc)
		file, line := f.FileLine(pc)
		lines[i] = fmt.Sprintf("%s:%d %s", file, line, f.Name())
	}

	return lines
}

func (e *Error) FormatStack() string {
	return strings.Join(e.FormatStackLines(), "\n")
}

func Wrap(err error, stackSkip int, opts ...Option) *Error {
	if err == nil {
		return nil
	}

	var e *Error

	if ierr, ok := err.(*Error); ok {
		// if we have some options, we need to copy the error to not modify the original one
		if len(opts) > 0 {
			ecopy := *ierr
			e = &ecopy
		} else {
			return ierr
		}
	} else {
		e = &Error{
			err:          err,
			shouldReport: true,
		}
	}

	for _, opt := range opts {
		opt(e)
	}

	if len(e.stack) == 0 {
		e.stack = callers(stackSkip + 1)
	}

	return e
}

func WithStatusCode(code int) Option {
	return func(e *Error) {
		e.statusCode = code
	}
}

func WithPublicMessage(msg string) Option {
	return func(e *Error) {
		e.publicMessage = msg
	}
}

func WithPrefix(prefix string) Option {
	return func(e *Error) {
		if len(e.prefix) > 0 {
			e.prefix = fmt.Sprintf("%s: %s", prefix, e.prefix)
		} else {
			e.prefix = prefix
		}
	}
}

func WithShouldReport(report bool) Option {
	return func(e *Error) {
		e.shouldReport = report
	}
}

func callers(skip int) []uintptr {
	stack := make([]uintptr, 10)
	n := runtime.Callers(skip+2, stack)
	return stack[:n]
}
