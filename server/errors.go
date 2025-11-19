package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v3/errctx"
)

const (
	errCategoryUnexpected = "unexpected"
	errCategoryTimeout    = "timeout"
	errCategorySecurity   = "security"
)

type (
	// Error represents an error returned by RouteHandler with additional category information.
	// It intentionally does not implement the error interface to avoid using it as a regular error.
	Error struct {
		Err      errctx.Error
		Category string
	}

	RouteNotDefinedError  struct{ *errctx.TextError }
	RequestCancelledError struct{ *errctx.TextError }
	RequestTimeoutError   struct{ *errctx.TextError }
	InvalidSecretError    struct{ *errctx.TextError }
)

// NewError creates a new [Error] instance wrapping the given errctx.Error and category.
// If err is nil, it returns nil.
//
// If the error or any of its causes is [context.DeadlineExceeded] or [context.Canceled],
// the category is set to "timeout" regardless of the provided category.
func NewError(err errctx.Error, category string) *Error {
	if err == nil {
		return nil
	}

	// If the error or any of its causes is context.DeadlineExceeded or context.Canceled,
	// enforce the timeout category.
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		category = errCategoryTimeout
	}

	return &Error{
		Err:      err,
		Category: category,
	}
}

func newRouteNotDefinedError(path string) errctx.Error {
	return RouteNotDefinedError{errctx.NewTextError(
		fmt.Sprintf("Route for %s is not defined", path),
		1,
		errctx.WithStatusCode(http.StatusNotFound),
		errctx.WithPublicMessage("Not found"),
		errctx.WithShouldReport(false),
	)}
}

func newRequestCancelledError(after time.Duration) errctx.Error {
	return RequestCancelledError{errctx.NewTextError(
		fmt.Sprintf("Request was cancelled after %v", after),
		1,
		errctx.WithStatusCode(499),
		errctx.WithPublicMessage("Cancelled"),
		errctx.WithShouldReport(false),
	)}
}

func (e RequestCancelledError) Unwrap() error {
	return context.Canceled
}

func newRequestTimeoutError(after time.Duration) errctx.Error {
	return RequestTimeoutError{errctx.NewTextError(
		fmt.Sprintf("Request was timed out after %v", after),
		1,
		errctx.WithStatusCode(http.StatusServiceUnavailable),
		errctx.WithPublicMessage("Gateway Timeout"),
		errctx.WithShouldReport(false),
	)}
}

func (e RequestTimeoutError) Unwrap() error {
	return context.DeadlineExceeded
}

func newInvalidSecretError() errctx.Error {
	return InvalidSecretError{errctx.NewTextError(
		"Invalid secret",
		1,
		errctx.WithStatusCode(http.StatusForbidden),
		errctx.WithPublicMessage("Forbidden"),
		errctx.WithShouldReport(false),
	)}
}
