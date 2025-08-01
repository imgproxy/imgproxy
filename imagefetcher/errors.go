package imagefetcher

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/imgproxy/imgproxy/v3/errwrap"
	"github.com/imgproxy/imgproxy/v3/security"
)

const msgSourceImageIsUnreachable = "Source image is unreachable"

type (
	RequestError          struct{ error }
	RequestSchemeError    struct{ error }
	PartialResponseError  struct{ error }
	ResponseStatusError   struct{ error }
	TooManyRedirectsError struct{ error }
	RequestCanceledError  struct{ error }
	RequestTimeoutError   struct{ error }
)

type NotModifiedError struct {
	headers http.Header
}

type httpError interface {
	Timeout() bool
}

func newRequestError(err error) error {
	return errwrap.From(
		RequestError{err}, 1,
	).
		WithStatusCode(http.StatusNotFound).
		WithPublicMessage(msgSourceImageIsUnreachable).
		WithShouldReport(false)
}

func newRequestSchemeError(scheme string) error {
	return errwrap.From(
		RequestSchemeError{fmt.Errorf("unknown scheme: %s", scheme)}, 1,
	).
		WithStatusCode(http.StatusNotFound).
		WithPublicMessage(msgSourceImageIsUnreachable).
		WithShouldReport(false)
}

func newPartialResponseError(msg string) error {
	return errwrap.From(
		PartialResponseError{errors.New(msg)}, 1,
	).
		WithStatusCode(http.StatusNotFound).
		WithPublicMessage(msgSourceImageIsUnreachable).
		WithShouldReport(false)
}

func newResponseStatusError(status int, body string) error {
	var err error

	if len(body) > 0 {
		err = fmt.Errorf("status: %d; %s", status, body)
	} else {
		err = fmt.Errorf("status: %d", status)
	}

	statusCode := http.StatusNotFound
	if status >= 500 {
		statusCode = http.StatusInternalServerError
	}

	return errwrap.From(
		ResponseStatusError{err}, 1,
	).
		WithStatusCode(statusCode).
		WithPublicMessage(msgSourceImageIsUnreachable).
		WithShouldReport(false)
}

func newTooManyRedirectsError(n int) error {
	return errwrap.From(
		TooManyRedirectsError{fmt.Errorf("stopped after %d redirects", n)}, 1,
	).
		WithStatusCode(http.StatusNotFound).
		WithPublicMessage(msgSourceImageIsUnreachable).
		WithShouldReport(false)
}

func newRequestCanceledError(err error) error {
	// stack 2
	return errwrap.Wrapf(
		RequestCanceledError{err},
		"the image request is cancelled",
	).WithStatusCode(499).
		WithPublicMessage(msgSourceImageIsUnreachable).
		WithShouldReport(false)
}

func newImageRequestTimeoutError(err error) error {
	return errwrap.From(
		RequestTimeoutError{err},
		2,
	).
		WithStatusCode(http.StatusGatewayTimeout).
		WithPublicMessage(msgSourceImageIsUnreachable).
		WithShouldReport(false)
}

func newNotModifiedError(headers http.Header) error {
	return errwrap.From(
		NotModifiedError{headers},
		1,
	).
		WithStatusCode(http.StatusNotModified).
		WithPublicMessage("Not modified").
		WithShouldReport(false)
}

func (e NotModifiedError) Error() string { return "not modified" }

func (e NotModifiedError) Headers() http.Header {
	return e.headers
}

// Is performs comparison of two notModifiedError instances.
// Any error should be Comparable, http.Header is not comparable,
// hence, we need to compare headers manually.
func (nm NotModifiedError) Is(target error) bool {
	m, ok := target.(NotModifiedError)
	return ok && reflect.DeepEqual(nm.headers, m.headers)
}

func wrapError(err error) error {
	isTimeout := false

	var secArrdErr security.SourceAddressError

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		isTimeout = true
	case errors.Is(err, context.Canceled):
		return newRequestCanceledError(err)
	case errors.As(err, &secArrdErr):
		return errwrap.From(err, 1).
			WithStatusCode(404).
			WithPublicMessage(msgSourceImageIsUnreachable).
			WithShouldReport(false)
	default:
		if httpErr, ok := err.(httpError); ok {
			isTimeout = httpErr.Timeout()
		}
	}

	if isTimeout {
		return errwrap.From(newImageRequestTimeoutError(err), 1)
	}

	// shift stack by 1 (NOTE: START WRAPPING FROM ORIGIN)
	return errwrap.Wrap(err)
}
