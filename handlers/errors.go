package handlers

import (
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errctx"
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
	ResponseWriteError struct{ *errctx.WrappedError }
	InvalidURLError    struct{ *errctx.TextError }
)

func NewResponseWriteError(cause error) errctx.Error {
	return ResponseWriteError{errctx.NewWrappedError(
		cause,
		1,
		errctx.WithPrefix("failed to write response"),
		errctx.WithPublicMessage("Failed to write response"),
	)}
}

func newInvalidURLErrorf(status int, format string, args ...interface{}) error {
	return InvalidURLError{errctx.NewTextError(
		fmt.Sprintf(format, args...),
		2,
		errctx.WithStatusCode(status),
		errctx.WithPublicMessage("Invalid URL"),
		errctx.WithShouldReport(false),
		errctx.WithCategory(CategoryPathParsing),
	)}
}

// NewInvalidPathError creates "invalid path" error
func NewInvalidPathError(path string) error {
	return newInvalidURLErrorf(
		http.StatusNotFound,
		"Invalid path: %s", path,
	)
}

// NewCantSaveError creates "resulting image not supported" error
func NewCantSaveError(format imagetype.Type) error {
	return newInvalidURLErrorf(
		http.StatusUnprocessableEntity,
		"Resulting image format is not supported: %s", format,
	)
}

// NewCantLoadError creates "source image not supported" error
func NewCantLoadError(format imagetype.Type) error {
	return newInvalidURLErrorf(
		http.StatusUnprocessableEntity,
		"Source image format is not supported: %s", format,
	)
}
