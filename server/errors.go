package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type (
	RouteNotDefinedError  string
	RequestCancelledError string
	RequestTimeoutError   string
	InvalidSecretError    struct{}
)

func newRouteNotDefinedError(path string) *ierrors.Error {
	return ierrors.Wrap(
		RouteNotDefinedError(fmt.Sprintf("Route for %s is not defined", path)),
		1,
		ierrors.WithStatusCode(http.StatusNotFound),
		ierrors.WithPublicMessage("Not found"),
		ierrors.WithShouldReport(false),
	)
}

func (e RouteNotDefinedError) Error() string { return string(e) }

func newRequestCancelledError(after time.Duration) *ierrors.Error {
	return ierrors.Wrap(
		RequestCancelledError(fmt.Sprintf("Request was cancelled after %v", after)),
		1,
		ierrors.WithStatusCode(499),
		ierrors.WithPublicMessage("Cancelled"),
		ierrors.WithShouldReport(false),
	)
}

func (e RequestCancelledError) Error() string { return string(e) }

func newRequestTimeoutError(after time.Duration) *ierrors.Error {
	return ierrors.Wrap(
		RequestTimeoutError(fmt.Sprintf("Request was timed out after %v", after)),
		1,
		ierrors.WithStatusCode(http.StatusServiceUnavailable),
		ierrors.WithPublicMessage("Gateway Timeout"),
		ierrors.WithCategory("timeout"),
		ierrors.WithShouldReport(false),
	)
}

func (e RequestTimeoutError) Error() string { return string(e) }

func newInvalidSecretError() error {
	return ierrors.Wrap(
		InvalidSecretError{},
		1,
		ierrors.WithStatusCode(http.StatusForbidden),
		ierrors.WithPublicMessage("Forbidden"),
		ierrors.WithShouldReport(false),
	)
}

func (e InvalidSecretError) Error() string { return "Invalid secret" }
