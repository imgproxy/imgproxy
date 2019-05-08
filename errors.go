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
}

func (e *imgproxyError) Error() string {
	return e.Message
}

func newError(status int, msg string, pub string) *imgproxyError {
	return &imgproxyError{status, msg, pub}
}

func newUnexpectedError(msg string, skip int) *imgproxyError {
	return &imgproxyError{
		500,
		fmt.Sprintf("Unexpected error: %s\n%s", msg, stacktrace(skip+3)),
		"Internal error",
	}
}

func stacktrace(skip int) string {
	callers := make([]uintptr, 10)
	n := runtime.Callers(skip, callers)

	lines := make([]string, n)
	for i, pc := range callers[:n] {
		f := runtime.FuncForPC(pc)
		file, line := f.FileLine(pc)
		lines[i] = fmt.Sprintf("%s:%d %s", file, line, f.Name())
	}

	return strings.Join(lines, "\n")
}
