package server

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
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
	ErrorType     string
	RequestID     string
	PublicMessage string
	Version       string
	ErrorChain    []errorChainItem
	StackTrace    []stackFrame
}

// errorChainItem represents a single item in the error chain.
type errorChainItem struct {
	Type       string
	Message    string
	StackFrame *stackFrame // First stack frame where this error occurred
}

// stackFrame represents a single frame in the stack trace.
type stackFrame struct {
	File       string
	Line       int
	Function   string
	ErrorIndex int  // Index in the error chain (0 = outermost, increases to root cause)
	IsLast     bool // True if this frame belongs to the last error in the chain
}

// unwrapErrorChain traverses the error chain using errors.Unwrap and returns
// all error types and messages from the outermost wrapper to the root cause.
func unwrapErrorChain(err error) []errorChainItem {
	var chain []errorChainItem

	for err != nil {
		// Get the error type and extract the first stack frame if available
		errorType := errctx.ErrorType(err)

		var firstFrame *stackFrame
		var errWithCtx errctx.Error

		if errors.As(err, &errWithCtx) {
			stackStr := errWithCtx.FormatStack()

			if stackStr != "" {
				frames := parseStackString(stackStr)

				if len(frames) > 0 {
					firstFrame = &frames[0]
				}
			}
		}

		chain = append(chain, errorChainItem{
			Type:       cleanErrorType(errorType),
			Message:    err.Error(),
			StackFrame: firstFrame,
		})

		err = errors.Unwrap(err)
	}

	return chain
}

// cleanErrorType removes Go syntax artifacts from error type string
func cleanErrorType(t string) string {
	return errorTypeReplacer.Replace(t)
}

// buildStackTrace gets the stack trace from the outermost error and marks
// which frames correspond to which errors in the chain.
func buildStackTrace(err error, chain []errorChainItem) []stackFrame {
	// Get stack trace from the outermost error (it has the full trace)
	var errWithCtx errctx.Error

	if !errors.As(err, &errWithCtx) {
		return nil
	}

	stackStr := errWithCtx.FormatStack()
	if stackStr == "" {
		return nil
	}

	frames := parseStackString(stackStr)

	// Match stack frames with errors in the chain
	for i := range frames {
		frames[i].ErrorIndex = findErrorIndexForFrame(&frames[i], chain)
		frames[i].IsLast = frames[i].ErrorIndex == len(chain)-1
	}

	return frames
}

// findErrorIndexForFrame determines which error in the chain this frame belongs to
func findErrorIndexForFrame(frame *stackFrame, chain []errorChainItem) int {
	// Match the frame with errors in the chain by comparing file and line
	for i := len(chain) - 1; i >= 0; i-- {
		if chain[i].StackFrame == nil {
			continue
		}

		if chain[i].StackFrame.File == frame.File && chain[i].StackFrame.Line == frame.Line {
			return i
		}
	}

	// If no match found, assign to outermost error
	return 0
}

// parseStackString parses a formatted stack trace string into stack frames
func parseStackString(stackStr string) []stackFrame {
	if stackStr == "" {
		return nil
	}

	lines := bytes.Split([]byte(stackStr), []byte("\n"))
	frames := make([]stackFrame, 0, len(lines))

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		// Parse line format: "/path/to/file.go:123 package.FunctionName"
		parts := bytes.SplitN(line, []byte(" "), 2)
		if len(parts) != 2 {
			continue
		}

		// Split file path and line number
		fileLine := bytes.Split(parts[0], []byte(":"))
		if len(fileLine) != 2 {
			continue
		}

		var lineNum int
		fmt.Sscanf(string(fileLine[1]), "%d", &lineNum)

		frames = append(frames, stackFrame{
			File:     string(fileLine[0]),
			Line:     lineNum,
			Function: string(parts[1]),
		})
	}

	return frames
}

// generateErrorHTML renders the error page template with the given error information.
func generateErrorHTML(err errctx.Error, reqID string) (string, error) {
	errorChain := unwrapErrorChain(err)

	data := errorPageData{
		StatusCode:    err.StatusCode(),
		ErrorType:     cleanErrorType(errctx.ErrorType(err)),
		RequestID:     reqID,
		PublicMessage: err.PublicMessage(),
		Version:       version.Version,
		ErrorChain:    errorChain,
		StackTrace:    buildStackTrace(err, errorChain),
	}

	var buf bytes.Buffer
	if err := errorPageTemplate.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
