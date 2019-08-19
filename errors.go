package main

import (
	"fmt"
	"runtime"
	"strings"
)

type imgproxyError struct {
	StatusCode    int
	Message       string
	PublicMessage string
	Unexpected    bool

	stack []uintptr
}

func (e *imgproxyError) Error() string {
	return e.Message
}

func (e *imgproxyError) ErrorWithStack() string {
	if e.stack == nil {
		return e.Message
	}

	return fmt.Sprintf("%s\n%s", e.Message, formatStack(e.stack))
}

func (e *imgproxyError) StackTrace() []uintptr {
	return e.stack
}

func newError(status int, msg string, pub string) *imgproxyError {
	return &imgproxyError{
		StatusCode:    status,
		Message:       msg,
		PublicMessage: pub,
	}
}

func newUnexpectedError(msg string, skip int) *imgproxyError {
	return &imgproxyError{
		StatusCode:    500,
		Message:       msg,
		PublicMessage: "Internal error",
		Unexpected:    true,

		stack: callers(skip + 3),
	}
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
