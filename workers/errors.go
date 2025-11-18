package workers

import (
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errctx"
)

type TooManyRequestsError struct{ *errctx.TextError }

func newTooManyRequestsError() error {
	return TooManyRequestsError{errctx.NewTextError(
		"Too many requests",
		1,
		errctx.WithStatusCode(http.StatusTooManyRequests),
		errctx.WithPublicMessage("Too many requests"),
		errctx.WithShouldReport(false),
	)}
}
