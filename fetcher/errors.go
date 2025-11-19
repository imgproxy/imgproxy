package fetcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/fetcher/transport/generichttp"
)

const msgSourceIsUnreachable = "Source is unreachable"

type (
	RequestError          struct{ *errctx.WrappedError }
	RequstSchemeError     struct{ *errctx.TextError }
	PartialResponseError  struct{ *errctx.TextError }
	ResponseStatusError   struct{ *errctx.TextError }
	TooManyRedirectsError struct{ *errctx.TextError }
	RequestCanceledError  struct{ *errctx.WrappedError }
	RequestTimeoutError   struct{ *errctx.WrappedError }

	NotModifiedError struct {
		*errctx.TextError
		headers http.Header
	}
)

func newRequestError(err error) error {
	return RequestError{errctx.NewWrappedError(
		err,
		1,
		errctx.WithStatusCode(http.StatusNotFound),
		errctx.WithPublicMessage(msgSourceIsUnreachable),
		errctx.WithShouldReport(false),
	)}
}

func newRequestSchemeError(scheme string) error {
	return RequstSchemeError{errctx.NewTextError(
		fmt.Sprintf("unknown scheme: %s", scheme),
		1,
		errctx.WithStatusCode(http.StatusNotFound),
		errctx.WithPublicMessage(msgSourceIsUnreachable),
		errctx.WithShouldReport(false),
	)}
}

func newPartialResponseError(msg string) error {
	return PartialResponseError{errctx.NewTextError(
		msg,
		1,
		errctx.WithStatusCode(http.StatusNotFound),
		errctx.WithPublicMessage(msgSourceIsUnreachable),
		errctx.WithShouldReport(false),
	)}
}

func newResponseStatusError(status int, body string) error {
	var msg string

	if len(body) > 0 {
		msg = fmt.Sprintf("status: %d; %s", status, body)
	} else {
		msg = fmt.Sprintf("status: %d", status)
	}

	statusCode := 404
	if status >= 400 && status < 500 {
		statusCode = status
	} else if status >= 500 {
		statusCode = http.StatusBadGateway
	}

	return ResponseStatusError{errctx.NewTextError(
		msg,
		1,
		errctx.WithStatusCode(statusCode),
		errctx.WithPublicMessage(msgSourceIsUnreachable),
		errctx.WithShouldReport(false),
	)}
}

func newTooManyRedirectsError(n int) error {
	return TooManyRedirectsError{errctx.NewTextError(
		fmt.Sprintf("stopped after %d redirects", n),
		1,
		errctx.WithStatusCode(http.StatusNotFound),
		errctx.WithPublicMessage(msgSourceIsUnreachable),
		errctx.WithShouldReport(false),
	)}
}

func newRequestCanceledError(err error) error {
	return RequestCanceledError{errctx.NewWrappedError(
		err,
		2,
		errctx.WithPrefix("source request is cancelled"),
		errctx.WithStatusCode(499),
		errctx.WithPublicMessage(msgSourceIsUnreachable),
		errctx.WithShouldReport(false),
	)}
}

func newRequestTimeoutError(err error) error {
	return RequestTimeoutError{errctx.NewWrappedError(
		err,
		2,
		errctx.WithPrefix("source request timed out"),
		errctx.WithStatusCode(http.StatusGatewayTimeout),
		errctx.WithPublicMessage(msgSourceIsUnreachable),
		errctx.WithShouldReport(false),
	)}
}

func newNotModifiedError(headers http.Header) error {
	return NotModifiedError{
		errctx.NewTextError(
			"not modified",
			1,
			errctx.WithStatusCode(http.StatusNotModified),
			errctx.WithPublicMessage("Not modified"),
			errctx.WithShouldReport(false),
		),
		headers,
	}
}

func (e NotModifiedError) Headers() http.Header {
	return e.headers
}

// NOTE: make private when we remove download functions from imagedata package
func WrapError(err error, skipStack int) error {
	type httpError interface {
		Timeout() bool
	}

	var srcErr generichttp.SourceAddressError

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return newRequestTimeoutError(err)
	case errors.Is(err, context.Canceled):
		return newRequestCanceledError(err)
	case err == io.ErrUnexpectedEOF:
		return PartialResponseError{errctx.NewTextError(
			"response is incomplete",
			1,
			errctx.WithStatusCode(http.StatusUnprocessableEntity),
			errctx.WithPublicMessage("Source response is incomplete"),
			errctx.WithShouldReport(false),
		)}
	case errors.As(err, &srcErr):
		return srcErr
	default:
		if httpErr, ok := err.(httpError); ok && httpErr.Timeout() {
			return newRequestTimeoutError(err)
		}
	}

	return errctx.WrapWithStackSkip(err, skipStack+1)
}
