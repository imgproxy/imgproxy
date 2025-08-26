package semaphores

import (
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type TooManyRequestsError struct{}

func newTooManyRequestsError() error {
	return ierrors.Wrap(
		TooManyRequestsError{},
		1,
		ierrors.WithStatusCode(http.StatusTooManyRequests),
		ierrors.WithPublicMessage("Too many requests"),
		ierrors.WithShouldReport(false),
	)
}

func (e TooManyRequestsError) Error() string { return "Too many requests" }
