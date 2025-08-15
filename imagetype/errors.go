package imagetype

import (
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type (
	TypeDetectionError struct{ error }
	UnknownFormatError struct{}
)

func newTypeDetectionError(err error) error {
	return ierrors.Wrap(
		TypeDetectionError{err},
		1,
		ierrors.WithStatusCode(http.StatusUnprocessableEntity),
		ierrors.WithPublicMessage("Failed to detect source image type"),
		ierrors.WithShouldReport(false),
	)
}

func (e TypeDetectionError) Error() string {
	return "Failed to detect image type: " + e.error.Error()
}

func (e TypeDetectionError) Unwrap() error { return e.error }

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
