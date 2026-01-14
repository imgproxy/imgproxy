package handlers

import (
	"context"
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

	defaultDocsUrl = "https://docs.imgproxy.net/usage/processing"
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

func newInvalidURLErrorf(docsUrl string, status int, format string, args ...any) errctx.Error {
	return InvalidURLError{errctx.NewTextError(
		fmt.Sprintf(format, args...),
		2,
		errctx.WithStatusCode(status),
		errctx.WithPublicMessage("Invalid URL"),
		errctx.WithShouldReport(false),
		errctx.WithDocsURL(docsUrl),
	)}
}

// NewInvalidPathError creates "invalid path" error
func NewInvalidPathError(ctx context.Context, path string) errctx.Error {
	return newInvalidURLErrorf(
		errctx.DocsBaseURL(ctx, defaultDocsUrl),
		http.StatusNotFound,
		"Invalid path: %s", path,
	)
}

// NewCantSaveError creates "resulting image not supported" error
func NewCantSaveError(format imagetype.Type) errctx.Error {
	return newInvalidURLErrorf(
		"https://docs.imgproxy.net/image_formats_support",
		http.StatusUnprocessableEntity,
		"Resulting image format is not supported: %s", format,
	)
}

// NewCantLoadError creates "source image not supported" error
func NewCantLoadError(ctx context.Context, format imagetype.Type) errctx.Error {
	return newInvalidURLErrorf(
		"https://docs.imgproxy.net/image_formats_support",
		http.StatusUnprocessableEntity,
		"Source image format is not supported: %s", format,
	)
}
