package svg

import (
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type (
	SanitizeError string
)

func newSanitizeError(msg string) error {
	return ierrors.Wrap(
		SanitizeError(msg),
		1,
		ierrors.WithStatusCode(http.StatusUnprocessableEntity),
		ierrors.WithPublicMessage("Invalid URL"),
		ierrors.WithShouldReport(false),
	)
}

func (e SanitizeError) Error() string { return string(e) }
