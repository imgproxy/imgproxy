package security

import (
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type (
	SignatureError       string
	FileSizeError        struct{}
	ImageResolutionError string
	SecurityOptionsError struct{}
	SourceURLError       string
	SourceAddressError   string
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

func newFileSizeError() error {
	return ierrors.Wrap(
		FileSizeError{},
		1,
		ierrors.WithStatusCode(http.StatusUnprocessableEntity),
		ierrors.WithPublicMessage("Invalid source image"),
		ierrors.WithShouldReport(false),
	)
}

func (e FileSizeError) Error() string { return "Source image file is too big" }

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

func newSecurityOptionsError() error {
	return ierrors.Wrap(
		SecurityOptionsError{},
		1,
		ierrors.WithStatusCode(http.StatusForbidden),
		ierrors.WithPublicMessage("Invalid URL"),
		ierrors.WithShouldReport(false),
	)
}

func (e SecurityOptionsError) Error() string { return "Security processing options are not allowed" }

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

func newSourceAddressError(msg string) error {
	return ierrors.Wrap(
		SourceAddressError(msg),
		1,
		ierrors.WithStatusCode(http.StatusNotFound),
		ierrors.WithPublicMessage("Invalid source URL"),
		ierrors.WithShouldReport(false),
	)
}

func (e SourceAddressError) Error() string { return string(e) }
