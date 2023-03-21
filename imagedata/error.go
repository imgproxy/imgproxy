package imagedata

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type httpError interface {
	Timeout() bool
}

func wrapError(err error) error {
	isTimeout := false

	if errors.Is(err, context.Canceled) {
		return ierrors.New(
			499,
			fmt.Sprintf("The image request is cancelled: %s", err),
			msgSourceImageIsUnreachable,
		)
	} else if errors.Is(err, context.DeadlineExceeded) {
		isTimeout = true
	} else if httpErr, ok := err.(httpError); ok {
		isTimeout = httpErr.Timeout()
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
