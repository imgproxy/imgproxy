package errctx

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

type ErrorContext struct {
	prefix       string
	statusCode   int
	publicMsg    string
	shouldReport bool
	docsUrl      string

	stack []uintptr
}

func newErrorContext(stackSkip int, opts ...Option) *ErrorContext {
	ec := &ErrorContext{
		statusCode:   http.StatusInternalServerError,
		publicMsg:    "Internal error",
		shouldReport: true,

		stack: callers(stackSkip + 1),
	}

	return ec.applyOptions(opts...)
}

func (ec *ErrorContext) CloneErrorContext(opts ...Option) *ErrorContext {
	newEc := *ec
	return newEc.applyOptions(opts...)
}

func (ec *ErrorContext) StatusCode() int {
	if ec.statusCode <= 0 {
		return 500
	}

	return ec.statusCode
}

func (ec *ErrorContext) PublicMessage() string {
	if len(ec.publicMsg) == 0 {
		return "Internal error"
	}

	return ec.publicMsg
}

func (ec *ErrorContext) ShouldReport() bool {
	return ec.shouldReport
}

func (ec *ErrorContext) DocsURL() string {
	return ec.docsUrl
}

// StackTrace returns the stack trace associated with the error.
func (ec *ErrorContext) StackTrace() []uintptr {
	return ec.stack
}

// Callers returns the stack trace associated with the error.
func (ec *ErrorContext) Callers() []uintptr {
	return ec.stack
}

// FormatStack formats the stack trace into a single string.
func (ec *ErrorContext) FormatStack() string {
	lines := make([]string, len(ec.stack))

	for i, pc := range ec.stack {
		f := runtime.FuncForPC(pc)
		file, line := f.FileLine(pc)
		lines[i] = fmt.Sprintf("%s:%d %s", file, line, f.Name())
	}

	return strings.Join(lines, "\n")
}

func (ec *ErrorContext) applyOptions(opts ...Option) *ErrorContext {
	for _, opt := range opts {
		opt(ec)
	}

	return ec
}

// callers returns the stack trace, skipping the specified number of frames.
func callers(skip int) []uintptr {
	stack := make([]uintptr, 10)
	n := runtime.Callers(skip+2, stack)
	return stack[:n]
}
