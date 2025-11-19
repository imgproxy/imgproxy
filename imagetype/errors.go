package imagetype

import (
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errctx"
)

type (
	TypeDetectionError struct{ *errctx.WrappedError }
	UnknownFormatError struct{ *errctx.TextError }
)

func newTypeDetectionError(err error) error {
	return TypeDetectionError{errctx.NewWrappedError(
		err,
		1,
		errctx.WithPrefix("failed to detect image type"),
		errctx.WithStatusCode(http.StatusUnprocessableEntity),
		errctx.WithPublicMessage("Failed to detect source image type"),
		errctx.WithShouldReport(false),
	)}
}

func newUnknownFormatError() error {
	return UnknownFormatError{errctx.NewTextError(
		"Source image type not supported",
		1,
		errctx.WithStatusCode(http.StatusUnprocessableEntity),
		errctx.WithPublicMessage("Invalid source image"),
		errctx.WithShouldReport(false),
	)}
}
