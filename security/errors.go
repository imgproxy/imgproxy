package security

import (
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errctx"
)

const (
	processingDocsUrl = "https://imgproxy.net/docs/processing/"
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
		errctx.WithDocsURL(processingDocsUrl),
	)}
}

func newMalformedSignatureError() error {
	msg := "The signature appears to be a processing option. The signature section should always be present in the URL."

	return SignatureError{errctx.NewTextError(
		msg,
		1,
		errctx.WithStatusCode(http.StatusForbidden),
		errctx.WithPublicMessage(msg),
		errctx.WithShouldReport(false),
		errctx.WithDocsURL(processingDocsUrl),
	)}
}

func newImageResolutionError(msg string) error {
	return ImageResolutionError{errctx.NewTextError(
		msg,
		1,
		errctx.WithStatusCode(http.StatusUnprocessableEntity),
		errctx.WithPublicMessage("Invalid source image"),
		errctx.WithShouldReport(false),
		errctx.WithDocsURL("https://docs.imgproxy.net/configuration/options#security"),
	)}
}

func newSourceURLError(imageURL string) error {
	return SourceURLError{errctx.NewTextError(
		fmt.Sprintf("Source URL is not allowed: %s", imageURL),
		1,
		errctx.WithStatusCode(http.StatusNotFound),
		errctx.WithPublicMessage("Invalid source URL"),
		errctx.WithShouldReport(false),
		errctx.WithDocsURL("https://docs.imgproxy.net/configuration/options#IMGPROXY_ALLOWED_SOURCES"),
	)}
}
