package imagemeta

import (
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type (
	UnknownFormatError struct{}
	FormatError        string
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

func newFormatError(format, msg string) error {
	return ierrors.Wrap(
		FormatError(fmt.Sprintf("Invalid %s file: %s", format, msg)),
		1,
		ierrors.WithStatusCode(http.StatusUnprocessableEntity),
		ierrors.WithPublicMessage("Invalid source image"),
		ierrors.WithShouldReport(false),
	)
}

func (e FormatError) Error() string { return string(e) }
