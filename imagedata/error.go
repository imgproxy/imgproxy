package imagedata

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/security"
)

type httpError interface {
	Timeout() bool
}

func wrapError(err error) error {
	isTimeout := false

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		isTimeout = true
	case errors.Is(err, context.Canceled):
		return ierrors.New(
			499,
			fmt.Sprintf("The image request is cancelled: %s", err),
			msgSourceImageIsUnreachable,
		)
	case errors.Is(err, security.ErrSourceAddressNotAllowed), errors.Is(err, security.ErrInvalidSourceAddress):
		return ierrors.New(
			404,
			err.Error(),
			msgSourceImageIsUnreachable,
		)
	default:
		if httpErr, ok := err.(httpError); ok {
			isTimeout = httpErr.Timeout()
		}
	}

	if !isTimeout {
		return err
	}

	ierr := ierrors.New(
		http.StatusGatewayTimeout,
		fmt.Sprintf("The image request timed out: %s", err),
		msgSourceImageIsUnreachable,
	)
	ierr.Unexpected = true

	return ierr
}
