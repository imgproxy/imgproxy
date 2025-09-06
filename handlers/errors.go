package handlers

import (
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagetype"
)

// Monitoring error categories
const (
	CategoryTimeout       = "timeout"
	CategoryImageDataSize = "image_data_size"
	CategoryPathParsing   = "path_parsing"
	CategorySecurity      = "security"
	CategoryQueue         = "queue"
	CategoryDownload      = "download"
	CategoryProcessing    = "processing"
	CategoryIO            = "IO"
	CategoryConfig        = "config(tmp)" // NOTE: THIS IS TEMPORARY
)

type (
	ResponseWriteError struct{ error }
	InvalidURLError    string
)

func NewResponseWriteError(cause error) *ierrors.Error {
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

func NewInvalidURLErrorf(status int, format string, args ...interface{}) error {
	return ierrors.Wrap(
		InvalidURLError(fmt.Sprintf(format, args...)),
		1,
		ierrors.WithStatusCode(status),
		ierrors.WithPublicMessage("Invalid URL"),
		ierrors.WithShouldReport(false),
	)
}

func (e InvalidURLError) Error() string { return string(e) }

// NewCantSaveError creates "resulting image not supported" error
func NewCantSaveError(format imagetype.Type) error {
	return ierrors.Wrap(NewInvalidURLErrorf(
		http.StatusUnprocessableEntity,
		"Resulting image format is not supported: %s", format,
	), 1, ierrors.WithCategory(CategoryPathParsing))
}

// NewCantLoadError creates "source image not supported" error
func NewCantLoadError(format imagetype.Type) error {
	return ierrors.Wrap(NewInvalidURLErrorf(
		http.StatusUnprocessableEntity,
		"Source image format is not supported: %s", format,
	), 1, ierrors.WithCategory(CategoryProcessing))
}
