// handlererr package exposes helper functions for error handling in request handlers
// (like streaming or processing).
package handlererr

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/metrics"
)

// Error types for categorizing errors in the error collector
const (
	ErrTypeTimeout         = "timeout"
	ErrTypeStreaming       = "streaming"
	ErrTypeDownload        = "download"
	ErrTypeSvgProcessing   = "svg_processing"
	ErrTypeProcessing      = "processing"
	ErrTypePathParsing     = "path_parsing"
	ErrTypeSecurity        = "security"
	ErrTypeQueue           = "queue"
	ErrTypeIO              = "IO"
	ErrTypeDownloadTimeout = "download"
)

// Send sends an error to the error collector if the error is not a 499 (client closed request)
func Send(ctx context.Context, errType string, err error) {
	if ierr, ok := err.(*ierrors.Error); ok {
		switch ierr.StatusCode() {
		case http.StatusServiceUnavailable:
			errType = ErrTypeTimeout
		case 499:
			return // no need to report request closed by the client
		}
	}

	metrics.SendError(ctx, errType, err)
}

// SendAndPanic sends an error to the error collector and panics with that error.
func SendAndPanic(ctx context.Context, errType string, err error) {
	Send(ctx, errType, err)
	panic(err)
}

// Check checks if the error is not nil and sends it to the error collector, panicking if it is not nil.
func Check(ctx context.Context, errType string, err error) {
	if err == nil {
		return
	}
	SendAndPanic(ctx, errType, err)
}
