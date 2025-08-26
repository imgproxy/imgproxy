package processing

import (
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagetype"
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
	categoryConfig        = "config(tmp)" // NOTE: THIS IS TEMPORARY
)

type (
	ResponseWriteError struct{ error }
	InvalidURLError    string
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

// newCantSaveError creates "resulting image not supported" error
func newCantSaveError(format imagetype.Type) error {
	return ierrors.Wrap(newInvalidURLErrorf(
		http.StatusUnprocessableEntity,
		"Resulting image format is not supported: %s", format,
	), 1, ierrors.WithCategory(categoryPathParsing))
}

// newCantLoadError creates "source image not supported" error
func newCantLoadError(format imagetype.Type) error {
	return ierrors.Wrap(newInvalidURLErrorf(
		http.StatusUnprocessableEntity,
		"Source image format is not supported: %s", format,
	), 1, ierrors.WithCategory(categoryProcessing))
}
