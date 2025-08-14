package svg

import (
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type (
	SanitizeError struct{ error }
)

func newSanitizeError(err error) error {
	return ierrors.Wrap(
		SanitizeError{err},
		1,
		ierrors.WithStatusCode(http.StatusUnprocessableEntity),
		ierrors.WithPublicMessage("Broken or unsupported SVG image"),
		ierrors.WithShouldReport(false),
	)
}

func (e SanitizeError) Unwrap() error { return e.error }
