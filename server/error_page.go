package server

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/version"
)

var (
	// errorPageTemplateStr is the HTML template for the error page
	//go:embed error_page.html
	errorPageTemplateStr string

	// errorPageTemplate is the parsed HTML template for the error page
	errorPageTemplate = template.Must(
		template.New("error_page").Funcs(template.FuncMap{
			"add":     func(a, b int) int { return a + b },
			"replace": strings.ReplaceAll,
		}).Parse(errorPageTemplateStr),
	)

	// errorTypeReplacer is used to clean up error type strings
	errorTypeReplacer = strings.NewReplacer("*", "", "{", "", "}", "")
)

// errorPageData holds the data passed to the error page template.
type errorPageData struct {
	StatusCode    int
	RequestID     string
	PublicMessage string
	Version       string
	Commit        string
	GoVersion     string
	ErrorChain    []errorChainItem
}

// errorChainItem represents a single item in the error chain.
type errorChainItem struct {
	Type       string
	Message    string
	StackTrace []stackFrame
}

// stackFrame represents a single frame in the stack trace.
type stackFrame struct {
	File     string
	Line     int
	Function string
}

// unwrapErrorChain traverses the error chain using errors.Unwrap and returns
// all error types and messages from the outermost wrapper to the root cause.
func unwrapErrorChain(err error) []errorChainItem {
	var chain []errorChainItem

	for err != nil {
		chain = append(chain, errorChainItem{
			Type:       errorTypeReplacer.Replace(errctx.ErrorType(err)),
			Message:    err.Error(),
			StackTrace: buildStackTrace(err),
		})

		err = errors.Unwrap(err)
	}

	return chain
}

// buildStackTrace gets the stack trace from the outermost error and marks
// which frames correspond to which errors in the chain.
func buildStackTrace(err error) []stackFrame {
	type stackTracer interface {
		StackTrace() []uintptr
	}

	type callers interface {
		Callers() []uintptr
	}

	var stack []uintptr

	//nolint:errorlint
	switch t := err.(type) {
	case stackTracer:
		stack = t.StackTrace()
	case callers:
		stack = t.Callers()
	default:
		return nil
	}

	frames := make([]stackFrame, 0, len(stack))
	for _, pc := range stack {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		file, line := fn.FileLine(pc)
		frames = append(frames, stackFrame{
			File:     file,
			Line:     line,
			Function: fn.Name(),
		})
	}

	return frames
}

// getBuildInfo returns the Go version and Git commit hash.
func getBuildInfo() (string, string) {
	gover := runtime.Version()

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return gover, setting.Value
			}
		}
	}

	return gover, ""
}

// generateErrorHTML renders the error page template with the given error information.
// It returns the generated HTML as a byte slice and a Content-Type string.
func generateErrorHTML(err errctx.Error, reqID string) ([]byte, string) {
	errorChain := unwrapErrorChain(err)
	gover, commit := getBuildInfo()

	data := errorPageData{
		StatusCode:    err.StatusCode(),
		RequestID:     reqID,
		PublicMessage: err.PublicMessage(),
		Version:       version.Version,
		Commit:        commit,
		GoVersion:     gover,
		ErrorChain:    errorChain,
	}

	var buf bytes.Buffer
	if errorPageTemplate.Execute(&buf, data) != nil {
		// In case of template execution error, return the error message and stack trace as plain text.
		body := fmt.Appendf([]byte{}, "%s\n\n%s", err.Error(), err.FormatStack())
		return body, "text/plain; charset=utf-8"
	}

	return buf.Bytes(), "text/html; charset=utf-8"
}
