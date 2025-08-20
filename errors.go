package main

import (
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

// Monitoring error categories
const (
	categoryTimeout       = "timeout"
	categoryImageDataSize = "image_data_size"
	categoryPathParsing   = "path_parsing"
	categorySecurity      = "security"
	categoryQueue         = "queue"
	categoryDownload      = "download"
	categoryProcessing    = "processing"
	categoryIO            = "IO"
	categoryStreaming     = "streaming"
)

type (
	ResponseWriteError   struct{ error }
	InvalidURLError      string
	TooManyRequestsError struct{}
)

func newResponseWriteError(cause error) *ierrors.Error {
	return ierrors.Wrap(
		ResponseWriteError{cause},
		1,
		ierrors.WithPublicMessage("Failed to write response"),
	)
}

func (e ResponseWriteError) Error() string {
	return fmt.Sprintf("Failed to write response: %s", e.error)
}

func (e ResponseWriteError) Unwrap() error {
	return e.error
}

func newInvalidURLErrorf(status int, format string, args ...interface{}) error {
	return ierrors.Wrap(
		InvalidURLError(fmt.Sprintf(format, args...)),
		1,
		ierrors.WithStatusCode(status),
		ierrors.WithPublicMessage("Invalid URL"),
		ierrors.WithShouldReport(false),
	)
}

func (e InvalidURLError) Error() string { return string(e) }

func newTooManyRequestsError() error {
	return ierrors.Wrap(
		TooManyRequestsError{},
		1,
		ierrors.WithStatusCode(http.StatusTooManyRequests),
		ierrors.WithPublicMessage("Too many requests"),
		ierrors.WithShouldReport(false),
	)
}

func (e TooManyRequestsError) Error() string { return "Too many requests" }
