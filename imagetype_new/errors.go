package imagetype_new

import (
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type (
	UnknownFormatError struct{}
)

func newUnknownFormatError() error {
	return ierrors.Wrap(
		UnknownFormatError{},
		1,
		ierrors.WithStatusCode(http.StatusUnprocessableEntity),
		ierrors.WithPublicMessage("Invalid source image"),
		ierrors.WithShouldReport(false),
	)
}

func (e UnknownFormatError) Error() string { return "Source image type not supported" }
