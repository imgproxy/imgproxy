package security

import (
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type (
	SignatureError       string
	ImageResolutionError string
	SourceURLError       string
)

func newSignatureError(msg string) error {
	return ierrors.Wrap(
		SignatureError(msg),
		1,
		ierrors.WithStatusCode(http.StatusForbidden),
		ierrors.WithPublicMessage("Forbidden"),
		ierrors.WithShouldReport(false),
	)
}

func (e SignatureError) Error() string { return string(e) }

func newImageResolutionError(msg string) error {
	return ierrors.Wrap(
		ImageResolutionError(msg),
		1,
		ierrors.WithStatusCode(http.StatusUnprocessableEntity),
		ierrors.WithPublicMessage("Invalid source image"),
		ierrors.WithShouldReport(false),
	)
}

func (e ImageResolutionError) Error() string { return string(e) }

func newSourceURLError(imageURL string) error {
	return ierrors.Wrap(
		SourceURLError(fmt.Sprintf("Source URL is not allowed: %s", imageURL)),
		1,
		ierrors.WithStatusCode(http.StatusNotFound),
		ierrors.WithPublicMessage("Invalid source URL"),
		ierrors.WithShouldReport(false),
	)
}

func (e SourceURLError) Error() string { return string(e) }
