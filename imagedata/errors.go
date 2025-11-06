package imagedata

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/security"
)

type (
	ImageRequestError          struct{ error }
	ImageRequstSchemeError     string
	ImagePartialResponseError  string
	ImageResponseStatusError   string
	ImageTooManyRedirectsError string
	ImageRequestCanceledError  struct{ error }
	ImageRequestTimeoutError   struct{ error }

	NotModifiedError struct {
		headers map[string]string
	}

	httpError interface {
		Timeout() bool
	}
)

func newImageRequestError(err error) error {
	return ierrors.Wrap(
		ImageRequestError{err},
		1,
		ierrors.WithStatusCode(http.StatusNotFound),
		ierrors.WithPublicMessage(msgSourceImageIsUnreachable),
		ierrors.WithShouldReport(false),
	)
}

func (e ImageRequestError) Unwrap() error {
	return e.error
}

func newImageRequstSchemeError(scheme string) error {
	return ierrors.Wrap(
		ImageRequstSchemeError(fmt.Sprintf("Unknown scheme: %s", scheme)),
		1,
		ierrors.WithStatusCode(http.StatusNotFound),
		ierrors.WithPublicMessage(msgSourceImageIsUnreachable),
		ierrors.WithShouldReport(false),
	)
}

func (e ImageRequstSchemeError) Error() string { return string(e) }

func newImagePartialResponseError(msg string) error {
	return ierrors.Wrap(
		ImagePartialResponseError(msg),
		1,
		ierrors.WithStatusCode(http.StatusNotFound),
		ierrors.WithPublicMessage(msgSourceImageIsUnreachable),
		ierrors.WithShouldReport(false),
	)
}

func (e ImagePartialResponseError) Error() string { return string(e) }

func newImageResponseStatusError(status int, body string) error {
	var msg string

	if len(body) > 0 {
		msg = fmt.Sprintf("Status: %d; %s", status, body)
	} else {
		msg = fmt.Sprintf("Status: %d", status)
	}

	statusCode := 404
	if status >= 400 && status < 500 {
		statusCode = status
	} else if status >= 500 {
		statusCode = http.StatusBadGateway
	}

	return ierrors.Wrap(
		ImageResponseStatusError(msg),
		1,
		ierrors.WithStatusCode(statusCode),
		ierrors.WithPublicMessage(msgSourceImageIsUnreachable),
		ierrors.WithShouldReport(false),
	)
}

func (e ImageResponseStatusError) Error() string { return string(e) }

func newImageTooManyRedirectsError(n int) error {
	return ierrors.Wrap(
		ImageTooManyRedirectsError(fmt.Sprintf("Stopped after %d redirects", n)),
		1,
		ierrors.WithStatusCode(http.StatusNotFound),
		ierrors.WithPublicMessage(msgSourceImageIsUnreachable),
		ierrors.WithShouldReport(false),
	)
}

func (e ImageTooManyRedirectsError) Error() string { return string(e) }

func newImageRequestCanceledError(err error) error {
	return ierrors.Wrap(
		ImageRequestCanceledError{err},
		2,
		ierrors.WithStatusCode(499),
		ierrors.WithPublicMessage(msgSourceImageIsUnreachable),
		ierrors.WithShouldReport(false),
	)
}

func (e ImageRequestCanceledError) Error() string {
	return fmt.Sprintf("The image request is cancelled: %s", e.error)
}

func (e ImageRequestCanceledError) Unwrap() error { return e.error }

func newImageRequestTimeoutError(err error) error {
	return ierrors.Wrap(
		ImageRequestTimeoutError{err},
		2,
		ierrors.WithStatusCode(http.StatusGatewayTimeout),
		ierrors.WithPublicMessage(msgSourceImageIsUnreachable),
		ierrors.WithShouldReport(false),
	)
}

func (e ImageRequestTimeoutError) Error() string {
	return fmt.Sprintf("The image request timed out: %s", e.error)
}

func (e ImageRequestTimeoutError) Unwrap() error { return e.error }

func newNotModifiedError(headers map[string]string) error {
	return ierrors.Wrap(
		NotModifiedError{headers},
		1,
		ierrors.WithStatusCode(http.StatusNotModified),
		ierrors.WithPublicMessage("Not modified"),
		ierrors.WithShouldReport(false),
	)
}

func (e NotModifiedError) Error() string { return "Not modified" }

func (e NotModifiedError) Headers() map[string]string {
	return e.headers
}

func wrapError(err error) error {
	isTimeout := false

	var secArrdErr security.SourceAddressError

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		isTimeout = true
	case errors.Is(err, context.Canceled):
		return newImageRequestCanceledError(err)
	case errors.As(err, &secArrdErr):
		return ierrors.Wrap(
			err,
			1,
			ierrors.WithStatusCode(404),
			ierrors.WithPublicMessage(msgSourceImageIsUnreachable),
			ierrors.WithShouldReport(false),
		)
	default:
		if httpErr, ok := err.(httpError); ok {
			isTimeout = httpErr.Timeout()
		}
	}

	if isTimeout {
		return newImageRequestTimeoutError(err)
	}

	return ierrors.Wrap(err, 1)
}
