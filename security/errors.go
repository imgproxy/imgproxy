package security

import (
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errctx"
)

type (
	SignatureError       struct{ *errctx.TextError }
	ImageResolutionError struct{ *errctx.TextError }
	SourceURLError       struct{ *errctx.TextError }
)

func newSignatureError(msg string) error {
	return SignatureError{errctx.NewTextError(
		msg,
		1,
		errctx.WithStatusCode(http.StatusForbidden),
		errctx.WithPublicMessage("Forbidden"),
		errctx.WithShouldReport(false),
	)}
}

func newImageResolutionError(msg string) error {
	return ImageResolutionError{errctx.NewTextError(
		msg,
		1,
		errctx.WithStatusCode(http.StatusUnprocessableEntity),
		errctx.WithPublicMessage("Invalid source image"),
		errctx.WithShouldReport(false),
	)}
}

func newSourceURLError(imageURL string) error {
	return SourceURLError{errctx.NewTextError(
		fmt.Sprintf("Source URL is not allowed: %s", imageURL),
		1,
		errctx.WithStatusCode(http.StatusNotFound),
		errctx.WithPublicMessage("Invalid source URL"),
		errctx.WithShouldReport(false),
	)}
}
