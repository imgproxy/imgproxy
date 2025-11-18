package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v3/errctx"
)

type (
	RouteNotDefinedError  struct{ *errctx.TextError }
	RequestCancelledError struct{ *errctx.TextError }
	RequestTimeoutError   struct{ *errctx.TextError }
	InvalidSecretError    struct{ *errctx.TextError }
)

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
		errctx.WithCategory(categoryTimeout),
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
		errctx.WithCategory(categoryTimeout),
	)}
}

func (e RequestTimeoutError) Unwrap() error {
	return context.DeadlineExceeded
}

func newInvalidSecretError() error {
	return InvalidSecretError{errctx.NewTextError(
		"Invalid secret",
		1,
		errctx.WithStatusCode(http.StatusForbidden),
		errctx.WithPublicMessage("Forbidden"),
		errctx.WithShouldReport(false),
	)}
}
