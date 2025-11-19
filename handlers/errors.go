package handlers

import (
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/imagetype"
)

// Monitoring error categories
const (
	ErrCategoryTimeout       = "timeout"
	ErrCategoryImageDataSize = "image_data_size"
	ErrCategoryPathParsing   = "path_parsing"
	ErrCategorySecurity      = "security"
	ErrCategoryQueue         = "queue"
	ErrCategoryDownload      = "download"
	ErrCategoryProcessing    = "processing"
	ErrCategoryIO            = "IO"
	ErrCategoryConfig        = "config(tmp)" // NOTE: THIS IS TEMPORARY
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

func newInvalidURLErrorf(status int, format string, args ...interface{}) errctx.Error {
	return InvalidURLError{errctx.NewTextError(
		fmt.Sprintf(format, args...),
		2,
		errctx.WithStatusCode(status),
		errctx.WithPublicMessage("Invalid URL"),
		errctx.WithShouldReport(false),
	)}
}

// NewInvalidPathError creates "invalid path" error
func NewInvalidPathError(path string) errctx.Error {
	return newInvalidURLErrorf(
		http.StatusNotFound,
		"Invalid path: %s", path,
	)
}

// NewCantSaveError creates "resulting image not supported" error
func NewCantSaveError(format imagetype.Type) errctx.Error {
	return newInvalidURLErrorf(
		http.StatusUnprocessableEntity,
		"Resulting image format is not supported: %s", format,
	)
}

// NewCantLoadError creates "source image not supported" error
func NewCantLoadError(format imagetype.Type) errctx.Error {
	return newInvalidURLErrorf(
		http.StatusUnprocessableEntity,
		"Source image format is not supported: %s", format,
	)
}
