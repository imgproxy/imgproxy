package svg

import (
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errctx"
)

type (
	SanitizeError struct{ *errctx.WrappedError }
)

func newSanitizeError(err error) error {
	return SanitizeError{errctx.NewWrappedError(
		err,
		1,
		errctx.WithStatusCode(http.StatusUnprocessableEntity),
		errctx.WithPublicMessage("Broken or unsupported SVG image"),
		errctx.WithShouldReport(true),
	)}
}
